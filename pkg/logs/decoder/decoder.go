// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package decoder

import (
	"bytes"
	"regexp"
	"sync/atomic"
	"time"

	dd_conf "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// defaultContentLenLimit represents the max size for a line,
// if a line is bigger than this limit, it will be truncated.
const defaultContentLenLimit = 256 * 1000

// Input represents a chunk of line.
type Input struct {
	content []byte
}

// NewInput returns a new input.
func NewInput(content []byte) *Input {
	return &Input{
		content: content,
	}
}

// DecodedInput represents a decoded line and the raw length
type DecodedInput struct {
	content    []byte
	rawDataLen int
}

// NewDecodedInput returns a new decoded input.
func NewDecodedInput(content []byte, rawDataLen int) *DecodedInput {
	return &DecodedInput{
		content:    content,
		rawDataLen: rawDataLen,
	}
}

// Message represents a structured line.
type Message struct {
	Content            []byte
	Status             string
	RawDataLen         int
	Timestamp          string
	IngestionTimestamp int64
}

// NewMessage returns a new output.
func NewMessage(content []byte, status string, rawDataLen int, timestamp string) *Message {
	return &Message{
		Content:            content,
		Status:             status,
		RawDataLen:         rawDataLen,
		Timestamp:          timestamp,
		IngestionTimestamp: time.Now().UnixNano(),
	}
}

// Decoder wraps a collection of internal actors, joined by channels, representing the
// whole as a single actor with InputChan of type *decoder.Input and OutputChan of type
// *decoder.Message.
//
// Internally, it has three running actors:
//
// Decoder.run() takes data from InputChan, uses an EndlineMatcher to break it into lines,
// and passes those to the next actor via lineParser.Handle, which internally uses a channel.
//
// LineParser.run() takes data from its input channel, invokes the parser to convert it to
// parsers.Message, converts that to decoder.Message, and passes that to the next actor via
// lineHandler.Handle, which internally uses a channel.
//
// LineHandler.run() takes data from its input channel, processes it as necessary (as single
// lines, multiple lines, or auto-detecting the two), and sends the result to its output
// channel, which is the same channel as decoder.OutputChan.
type Decoder struct {
	// The number of raw lines decoded from the input before they are processed.
	// Needs to be first to ensure 64 bit alignment
	linesDecoded int64

	// InputChan represents the input to the decoder: a chunked stream of bytes
	InputChan chan *Input

	// decodedChan represents the data after it has been broken on newlines
	decodedChan chan *DecodedInput

	// lineChan represents the data after each line has been parsed into a message
	lineChan chan *Message

	// OutputChan is the output of the decoder, with any multi-line recognition applied
	// to the lines in lineChan.
	OutputChan chan *Message

	matcher         EndLineMatcher
	lineBuffer      *bytes.Buffer
	lineParser      LineParser
	lineHandler     LineHandler
	contentLenLimit int
	rawDataLen      int

	// The decoder holds on to an instace of DetectedPattern which is a thread safe container used to
	// pass a multiline pattern up from the line handler in order to surface it to the tailer.
	// The tailer uses this to determine if a pattern should be reused when a file rotates.
	detectedPattern *DetectedPattern
}

// InitializeDecoder returns a properly initialized Decoder
func InitializeDecoder(source *config.LogSource, parser parsers.Parser) *Decoder {
	return NewDecoderWithEndLineMatcher(source, parser, &NewLineMatcher{}, nil)
}

// NewDecoderWithEndLineMatcher initialize a decoder with given endline strategy.
func NewDecoderWithEndLineMatcher(source *config.LogSource, parser parsers.Parser, matcher EndLineMatcher, multiLinePattern *regexp.Regexp) *Decoder {
	inputChan := make(chan *Input)
	decodedChan := make(chan *DecodedInput)
	lineChan := make(chan *Message)
	outputChan := make(chan *Message)
	lineLimit := defaultContentLenLimit
	var lineHandler LineHandler
	var lineParser LineParser
	detectedPattern := &DetectedPattern{}

	for _, rule := range source.Config.ProcessingRules {
		if rule.Type == config.MultiLine {
			lh := NewMultiLineHandler(rule.Regex, config.AggregationTimeout(), lineLimit)

			// Since a single source can have multiple file tailers - each with their own decoder instance,
			// Make sure we keep track of the multiline match count info from all of the decoders so the
			// status page displays it correctly.
			if existingInfo, ok := source.GetInfo(lh.countInfo.InfoKey()).(*config.CountInfo); ok {
				// override the new decoders info to the instance we are already using
				lh.countInfo = existingInfo
			} else {
				// this is the first decoder we have seen for this source - use it's count info
				source.RegisterInfo(lh.countInfo)
			}
			lineHandler = lh
		}
	}
	if lineHandler == nil {
		if source.Config.AutoMultiLineEnabled() {
			log.Infof("Auto multi line log detection enabled")

			if multiLinePattern != nil {
				log.Info("Found a previously detected pattern - using multiline handler")

				// Save the pattern again for the next rotation
				detectedPattern.Set(multiLinePattern)

				lineHandler = NewMultiLineHandler(multiLinePattern, config.AggregationTimeout(), lineLimit)
			} else {
				lineHandler = buildAutoMultilineHandlerFromConfig(lineLimit, source, detectedPattern)
			}
		} else {
			lineHandler = NewSingleLineHandler(lineLimit)
		}
	}

	if parser.SupportsPartialLine() {
		lineParser = NewMultiLineParser(config.AggregationTimeout(), parser, lineLimit)
	} else {
		lineParser = NewSingleLineParser(parser)
	}

	return New(inputChan, decodedChan, lineChan, outputChan, lineParser, lineHandler, lineLimit, matcher, detectedPattern)
}

func buildAutoMultilineHandlerFromConfig(lineLimit int, source *config.LogSource, detectedPattern *DetectedPattern) *AutoMultilineHandler {
	linesToSample := source.Config.AutoMultiLineSampleSize
	if linesToSample <= 0 {
		linesToSample = dd_conf.Datadog.GetInt("logs_config.auto_multi_line_default_sample_size")
	}
	matchThreshold := source.Config.AutoMultiLineMatchThreshold
	if matchThreshold == 0 {
		matchThreshold = dd_conf.Datadog.GetFloat64("logs_config.auto_multi_line_default_match_threshold")
	}
	additionalPatterns := dd_conf.Datadog.GetStringSlice("logs_config.auto_multi_line_extra_patterns")
	additionalPatternsCompiled := []*regexp.Regexp{}

	for _, p := range additionalPatterns {
		compiled, err := regexp.Compile("^" + p)
		if err != nil {
			log.Warn("logs_config.auto_multi_line_extra_patterns containing value: ", p, " is not a valid regular expression")
			continue
		}
		additionalPatternsCompiled = append(additionalPatternsCompiled, compiled)
	}

	matchTimeout := time.Second * dd_conf.Datadog.GetDuration("logs_config.auto_multi_line_default_match_timeout")
	return NewAutoMultilineHandler(
		lineLimit,
		linesToSample,
		matchThreshold,
		matchTimeout,
		config.AggregationTimeout(),
		source,
		additionalPatternsCompiled,
		detectedPattern)
}

// New returns an initialized Decoder
func New(InputChan chan *Input, decodedChan chan *DecodedInput, lineChan chan *Message, OutputChan chan *Message, lineParser LineParser, lineHandler LineHandler, contentLenLimit int, matcher EndLineMatcher, detectedPattern *DetectedPattern) *Decoder {
	var lineBuffer bytes.Buffer
	return &Decoder{
		InputChan:       InputChan,
		decodedChan:     decodedChan,
		lineChan:        lineChan,
		OutputChan:      OutputChan,
		lineBuffer:      &lineBuffer,
		lineParser:      lineParser,
		lineHandler:     lineHandler,
		contentLenLimit: contentLenLimit,
		matcher:         matcher,
		detectedPattern: detectedPattern,
	}
}

// Start starts the Decoder
func (d *Decoder) Start() {
	d.lineParser.Start(d.decodedChan, d.lineChan)
	d.lineHandler.Start(d.lineChan, d.OutputChan)
	go d.run()
}

// Stop stops the Decoder
func (d *Decoder) Stop() {
	close(d.InputChan)
}

// GetLineCount returns the number of decoded lines
func (d *Decoder) GetLineCount() int64 {
	return atomic.LoadInt64(&d.linesDecoded)
}

// GetDetectedPattern returns a detected pattern (if any)
func (d *Decoder) GetDetectedPattern() *regexp.Regexp {
	if d.detectedPattern == nil {
		return nil
	}
	return d.detectedPattern.Get()
}

// run lets the Decoder handle data coming from InputChan
func (d *Decoder) run() {
	defer func() {
		// signal downstream to finish when this goroutine finishes
		close(d.decodedChan)
	}()
	for data := range d.InputChan {
		d.decodeIncomingData(data.content)
	}
}

// decodeIncomingData splits raw data based on '\n', creates and processes new lines
func (d *Decoder) decodeIncomingData(inBuf []byte) {
	i, j := 0, 0
	n := len(inBuf)
	maxj := d.contentLenLimit - d.lineBuffer.Len()

	for ; j < n; j++ {
		if j == maxj {
			// send line because it is too long
			d.lineBuffer.Write(inBuf[i:j])
			d.rawDataLen += (j - i)
			d.sendLine()
			i = j
			maxj = i + d.contentLenLimit
		} else if d.matcher.Match(d.lineBuffer.Bytes(), inBuf, i, j) {
			d.lineBuffer.Write(inBuf[i:j])
			d.rawDataLen += (j - i)
			d.rawDataLen++ // account for the matching byte
			d.sendLine()
			i = j + 1 // skip the last bytes of the matched sequence
			maxj = i + d.contentLenLimit
		}
	}
	d.lineBuffer.Write(inBuf[i:j])
	d.rawDataLen += (j - i)
}

// sendLine copies content from lineBuffer which is passed to lineHandler
func (d *Decoder) sendLine() {
	// Account for longer-than-1-byte line separator
	content := make([]byte, d.lineBuffer.Len()-(d.matcher.SeparatorLen()-1))
	copy(content, d.lineBuffer.Bytes())
	d.lineBuffer.Reset()
	d.decodedChan <- NewDecodedInput(content, d.rawDataLen)
	d.rawDataLen = 0
	atomic.AddInt64(&d.linesDecoded, 1)
}
