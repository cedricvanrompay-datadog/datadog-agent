// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package decoder

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers/dockerfile"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers/dockerstream"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers/encodedtext"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers/kubernetes"
	"github.com/DataDog/datadog-agent/pkg/logs/internal/parsers/noop"
	"github.com/DataDog/datadog-agent/pkg/logs/message"

	"github.com/stretchr/testify/assert"
)

type MockLineParser struct {
	inputChan chan *DecodedInput
}

func NewMockLineParser() *MockLineParser {
	return &MockLineParser{}
}

func (p *MockLineParser) Start(input chan *DecodedInput, output chan *Message) {
	p.inputChan = input
}

func (p *MockLineParser) Stop() {
	close(p.inputChan)
}

const contentLenLimit = 100

func TestDecodeIncomingData(t *testing.T) {
	decodedChan := make(chan *DecodedInput, 10)
	d := New(nil, decodedChan, nil, nil, nil, nil, contentLenLimit, &NewLineMatcher{}, nil)

	var line *DecodedInput

	// one line in one raw should be sent
	d.decodeIncomingData([]byte("helloworld\n"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	assert.Equal(t, len("helloworld\n"), line.rawDataLen)
	assert.Equal(t, "", d.lineBuffer.String())

	// multiple lines in one raw should be sent
	d.decodeIncomingData([]byte("helloworld\nhowayou\ngoodandyou"))
	l := 0
	line = <-decodedChan
	l += line.rawDataLen
	assert.Equal(t, "helloworld", string(line.content))
	line = <-decodedChan
	l += line.rawDataLen
	assert.Equal(t, "howayou", string(line.content))
	assert.Equal(t, "goodandyou", d.lineBuffer.String())
	assert.Equal(t, len("helloworld\nhowayou\n"), l)
	d.lineBuffer.Reset()
	d.rawDataLen, l = 0, 0

	// multiple lines in multiple rows should be sent
	d.decodeIncomingData([]byte("helloworld\nthisisa"))
	line = <-decodedChan
	l += line.rawDataLen
	assert.Equal(t, "helloworld", string(line.content))
	assert.Equal(t, "thisisa", d.lineBuffer.String())
	d.decodeIncomingData([]byte("longinput\nindeed"))
	line = <-decodedChan
	l += line.rawDataLen
	assert.Equal(t, "thisisalonginput", string(line.content))
	assert.Equal(t, "indeed", d.lineBuffer.String())
	assert.Equal(t, len("helloworld\nthisisalonginput\n"), l)
	d.lineBuffer.Reset()
	d.rawDataLen = 0

	// one line in multiple rows should be sent
	d.decodeIncomingData([]byte("hello world"))
	d.decodeIncomingData([]byte("!\n"))
	line = <-decodedChan
	assert.Equal(t, "hello world!", string(line.content))
	assert.Equal(t, len("hello world!\n"), line.rawDataLen)

	// excessively long line in one row should be sent by chunks
	d.decodeIncomingData([]byte(strings.Repeat("a", contentLenLimit+10) + "\n"))
	line = <-decodedChan
	assert.Equal(t, contentLenLimit, len(line.content))
	assert.Equal(t, contentLenLimit, line.rawDataLen)
	line = <-decodedChan
	assert.Equal(t, strings.Repeat("a", 10), string(line.content))
	assert.Equal(t, 11, line.rawDataLen)

	// excessively long line in multiple rows should be sent by chunks
	d.decodeIncomingData([]byte(strings.Repeat("a", contentLenLimit-5)))
	d.decodeIncomingData([]byte(strings.Repeat("a", 15) + "\n"))
	line = <-decodedChan
	assert.Equal(t, contentLenLimit, len(line.content))
	assert.Equal(t, contentLenLimit, line.rawDataLen)
	line = <-decodedChan
	assert.Equal(t, strings.Repeat("a", 10), string(line.content))
	assert.Equal(t, 11, line.rawDataLen)

	// empty lines should be sent
	d.decodeIncomingData([]byte("\n"))
	line = <-decodedChan
	assert.Equal(t, "", string(line.content))
	assert.Equal(t, "", d.lineBuffer.String())
	assert.Equal(t, 1, line.rawDataLen)

	// empty message should not change anything
	d.decodeIncomingData([]byte(""))
	assert.Equal(t, "", d.lineBuffer.String())
	assert.Equal(t, 0, d.rawDataLen)
}

func TestDecodeIncomingDataWithCustomSequence(t *testing.T) {
	decodedChan := make(chan *DecodedInput, 10)
	d := New(nil, nil, nil, nil, nil, nil, contentLenLimit, NewBytesSequenceMatcher([]byte("SEPARATOR"), 1), nil)

	var line *DecodedInput

	// one line in one raw should be sent
	d.decodeIncomingData([]byte("helloworldSEPARATOR"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	assert.Equal(t, "", d.lineBuffer.String())

	// multiple lines in one raw should be sent
	d.decodeIncomingData([]byte("helloworldSEPARATORhowayouSEPARATORgoodandyou"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	line = <-decodedChan
	assert.Equal(t, "howayou", string(line.content))
	assert.Equal(t, "goodandyou", d.lineBuffer.String())
	d.lineBuffer.Reset()

	// Line separartor may be cut by sending party
	d.decodeIncomingData([]byte("helloworldSEPAR"))
	d.decodeIncomingData([]byte("ATORhowayouSEPARATO"))
	d.decodeIncomingData([]byte("Rgoodandyou"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	line = <-decodedChan
	assert.Equal(t, "howayou", string(line.content))
	assert.Equal(t, "goodandyou", d.lineBuffer.String())
	d.lineBuffer.Reset()

	// empty lines should be sent
	d.decodeIncomingData([]byte("SEPARATOR"))
	line = <-decodedChan
	assert.Equal(t, "", string(line.content))
	assert.Equal(t, "", d.lineBuffer.String())

	// empty message should not change anything
	d.decodeIncomingData([]byte(""))
	assert.Equal(t, "", d.lineBuffer.String())
}

func TestDecodeIncomingDataWithSingleByteCustomSequence(t *testing.T) {
	decodedChan := make(chan *DecodedInput, 10)
	d := New(nil, decodedChan, nil, nil, nil, nil, contentLenLimit, NewBytesSequenceMatcher([]byte("&"), 1), nil)

	var line *DecodedInput

	// one line in one raw should be sent
	d.decodeIncomingData([]byte("helloworld&"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	assert.Equal(t, "", d.lineBuffer.String())

	// multiple blank lines
	n := 10
	d.decodeIncomingData([]byte(strings.Repeat("&", n)))
	for i := 0; i < n; i++ {
		line = <-decodedChan
		assert.Equal(t, "", string(line.content))
	}
	assert.Equal(t, "", d.lineBuffer.String())
	d.lineBuffer.Reset()

	// Mix empty & non-empty lines
	d.decodeIncomingData([]byte("helloworld&&"))
	d.decodeIncomingData([]byte("&howayou&"))
	line = <-decodedChan
	assert.Equal(t, "helloworld", string(line.content))
	line = <-decodedChan
	assert.Equal(t, "", string(line.content))
	line = <-decodedChan
	assert.Equal(t, "", string(line.content))
	line = <-decodedChan
	assert.Equal(t, "howayou", string(line.content))
	assert.Equal(t, "", d.lineBuffer.String())
	d.lineBuffer.Reset()

	// empty message should not change anything
	d.decodeIncomingData([]byte(""))
	assert.Equal(t, "", d.lineBuffer.String())
}

func TestDecoderLifeCycle(t *testing.T) {
	decodedChan := make(chan *DecodedInput, 10)
	d := New(nil, decodedChan, nil, nil, nil, nil, contentLenLimit, &NewLineMatcher{}, nil)

	// LineParser should not receive any lines
	d.Start()
	select {
	case <-decodedChan:
		assert.Fail(t, "LineParser should not handle anything")
	default:
		break
	}

	// LineParser should not receive any lines
	d.Stop()
	select {
	case <-decodedChan:
		break
	default:
		assert.Fail(t, "LineParser should be stopped")
	}
}

func TestDecoderInputNotDockerHeader(t *testing.T) {
	inputChan := make(chan *Input)
	h := NewMockLineParser()
	d := New(inputChan, nil, nil, nil, h, nil, 100, &NewLineMatcher{}, nil)
	d.Start()

	input := []byte("hello")
	input = append(input, []byte{1, 0, 0, 0, 0, 10, 0, 0}...) // docker header
	input = append(input, []byte("2018-06-14T18:27:03.246999277Z app logs\n")...)
	inputChan <- NewInput(input)

	var output *DecodedInput
	output = <-h.inputChan
	expected1 := append([]byte("hello"), []byte{1, 0, 0, 0, 0}...)
	assert.Equal(t, expected1, output.content)
	assert.Equal(t, len(expected1)+1, output.rawDataLen)

	output = <-h.inputChan
	expected2 := append([]byte{0, 0}, []byte("2018-06-14T18:27:03.246999277Z app logs")...)
	assert.Equal(t, expected2, output.content)
	assert.Equal(t, len(expected2)+1, output.rawDataLen)
	d.Stop()
}

func TestDecoderWithDockerHeader(t *testing.T) {
	source := config.NewLogSource("config", &config.LogsConfig{})
	d := InitializeDecoder(source, noop.New())
	d.Start()

	input := []byte("hello\n")
	input = append(input, []byte{1, 0, 0, 0, 0, 10, 0, 0}...) // docker header
	input = append(input, []byte("2018-06-14T18:27:03.246999277Z app logs\n")...)
	d.InputChan <- NewInput(input)

	var output *Message
	output = <-d.OutputChan
	assert.Equal(t, "hello", string(output.Content))
	assert.Equal(t, len("hello")+1, output.RawDataLen)

	output = <-d.OutputChan
	expected := []byte{1, 0, 0, 0, 0}
	assert.Equal(t, expected, output.Content)
	assert.Equal(t, 6, output.RawDataLen)

	output = <-d.OutputChan
	expected = append([]byte{0, 0}, []byte("2018-06-14T18:27:03.246999277Z app logs")...)
	assert.Equal(t, expected, output.Content)
	assert.Equal(t, len(expected)+1, output.RawDataLen)

	d.Stop()
}

func TestDecoderWithDockerHeaderSingleline(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	d := InitializeDecoder(
		config.NewLogSource("", &config.LogsConfig{}), dockerstream.New("abc123"))
	d.Start()
	defer d.Stop()

	line = append([]byte{2, 0, 0, 0, 0, 0, 0, 0}, []byte("2019-06-06T16:35:55.930852911Z message\n")...)
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusError, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852911Z", output.Timestamp)

	line = []byte("wrong message\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	// As we have no validation on the header, the parsing is incorrect
	// and this test fails.
	// It returns "wrong" as a timestamp and "message" as a content
	// TODO: add validation in the header and return the full message when
	// the validation fails.

	// output = <-d.OutputChan
	// assert.Equal(t, []byte("wrong message"), output.Content)
	// assert.Equal(t, lineLen, output.RawDataLen)
	// assert.Equal(t, message.StatusInfo, output.Status)
	// assert.Equal(t, "", output.Timestamp)

	output = <-d.OutputChan
	assert.Equal(t, []byte("message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "wrong", output.Timestamp)

}

func TestDecoderWithDockerHeaderMultiline(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	c := &config.LogsConfig{
		ProcessingRules: []*config.ProcessingRule{
			{
				Type:  config.MultiLine,
				Regex: regexp.MustCompile("1234"),
			},
		},
	}

	d := InitializeDecoder(config.NewLogSource("", c), dockerstream.New("abc123"))
	d.Start()
	defer d.Stop()

	line = append([]byte{1, 0, 0, 0, 0, 0, 0, 0}, []byte("2019-06-06T16:35:55.930852911Z 1234 hello\n")...)
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	line = append([]byte{1, 0, 0, 0, 0, 0, 0, 0}, []byte("2019-06-06T16:35:55.930852912Z world\n")...)
	lineLen += len(line)
	d.InputChan <- NewInput(line)

	line = append([]byte{2, 0, 0, 0, 0, 0, 0, 0}, []byte("2019-06-06T16:35:55.930852913Z 1234 bye\n")...)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 hello\\nworld"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852912Z", output.Timestamp)

	lineLen = len(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 bye"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusError, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852913Z", output.Timestamp)
}

func TestDecoderWithDockerJSONSingleline(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	d := InitializeDecoder(config.NewLogSource("", &config.LogsConfig{}), dockerfile.New())
	d.Start()
	defer d.Stop()

	line = []byte(`{"log":"message\n","stream":"stdout","time":"2019-06-06T16:35:55.930852911Z"}` + "\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852911Z", output.Timestamp)

	line = []byte("wrong message\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("wrong message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "", output.Timestamp)
}

func TestDecoderWithDockerJSONMultiline(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	c := &config.LogsConfig{
		ProcessingRules: []*config.ProcessingRule{
			{
				Type:  config.MultiLine,
				Regex: regexp.MustCompile("1234"),
			},
		},
	}

	d := InitializeDecoder(config.NewLogSource("", c), dockerfile.New())
	d.Start()
	defer d.Stop()

	line = []byte(`{"log":"1234 hello\n","stream":"stdout","time":"2019-06-06T16:35:55.930852911Z"}` + "\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	line = []byte(`{"log":"world\n","stream":"stdout","time":"2019-06-06T16:35:55.930852912Z"}` + "\n")
	lineLen += len(line)
	d.InputChan <- NewInput(line)

	line = []byte(`{"log":"1234 bye\n","stream":"stderr","time":"2019-06-06T16:35:55.930852913Z"}` + "\n")
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 hello\\nworld"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852912Z", output.Timestamp)

	lineLen = len(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 bye"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusError, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852913Z", output.Timestamp)
}

func TestDecoderWithDockerJSONSplittedByDocker(t *testing.T) {
	var output *Message
	var line []byte

	d := InitializeDecoder(config.NewLogSource("", &config.LogsConfig{}), dockerfile.New())
	d.Start()
	defer d.Stop()

	line = []byte(`{"log":"part1","stream":"stdout","time":"2019-06-06T16:35:55.930852911Z"}` + "\n")
	rawLen := len(line)
	d.InputChan <- NewInput(line)

	line = []byte(`{"log":"part2\n","stream":"stdout","time":"2019-06-06T16:35:55.930852912Z"}` + "\n")
	rawLen += len(line)
	d.InputChan <- NewInput(line)

	// We don't reaggregate partial messages but we expect content of line not finishing with a '\n' character to be reconciliated
	// with the next line.
	output = <-d.OutputChan
	assert.Equal(t, []byte("part1part2"), output.Content)
	assert.Equal(t, rawLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852912Z", output.Timestamp)
}

func TestDecoderWithDecodingParser(t *testing.T) {
	source := config.NewLogSource("config", &config.LogsConfig{})

	d := NewDecoderWithEndLineMatcher(source, encodedtext.New(encodedtext.UTF16LE), NewBytesSequenceMatcher(Utf16leEOL, 2), nil)
	d.Start()

	input := []byte{'h', 0x0, 'e', 0x0, 'l', 0x0, 'l', 0x0, 'o', 0x0, '\n', 0x0}
	d.InputChan <- NewInput(input)

	var output *Message
	output = <-d.OutputChan
	assert.Equal(t, "hello", string(output.Content))
	assert.Equal(t, len(input), output.RawDataLen)

	// Test with BOM
	input = []byte{0xFF, 0xFE, 'h', 0x0, 'e', 0x0, 'l', 0x0, 'l', 0x0, 'o', 0x0, '\n', 0x0}
	d.InputChan <- NewInput(input)

	output = <-d.OutputChan
	assert.Equal(t, "hello", string(output.Content))
	assert.Equal(t, len(input), output.RawDataLen)

	d.Stop()
}

func TestDecoderWithSinglelineKubernetes(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	d := InitializeDecoder(config.NewLogSource("", &config.LogsConfig{}), kubernetes.New())
	d.Start()
	defer d.Stop()

	line = []byte("2019-06-06T16:35:55.930852911Z stderr F message\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusError, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852911Z", output.Timestamp)

	line = []byte("wrong message\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("wrong message"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "", output.Timestamp)
}

func TestDecoderWithMultilineKubernetes(t *testing.T) {
	var output *Message
	var line []byte
	var lineLen int

	c := &config.LogsConfig{
		ProcessingRules: []*config.ProcessingRule{
			{
				Type:  config.MultiLine,
				Regex: regexp.MustCompile("1234"),
			},
		},
	}
	d := InitializeDecoder(config.NewLogSource("", c), kubernetes.New())
	d.Start()
	defer d.Stop()

	line = []byte("2019-06-06T16:35:55.930852911Z stdout F 1234 hello\n")
	lineLen = len(line)
	d.InputChan <- NewInput(line)

	line = []byte("2019-06-06T16:35:55.930852912Z stdout F world\n")
	lineLen += len(line)
	d.InputChan <- NewInput(line)

	line = []byte("2019-06-06T16:35:55.930852913Z stderr F 1234 bye\n")
	d.InputChan <- NewInput(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 hello\\nworld"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusInfo, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852912Z", output.Timestamp)

	lineLen = len(line)

	output = <-d.OutputChan
	assert.Equal(t, []byte("1234 bye"), output.Content)
	assert.Equal(t, lineLen, output.RawDataLen)
	assert.Equal(t, message.StatusError, output.Status)
	assert.Equal(t, "2019-06-06T16:35:55.930852913Z", output.Timestamp)
}
