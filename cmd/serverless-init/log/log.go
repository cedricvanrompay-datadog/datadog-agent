package log

import (
	"fmt"
	"os"
	"time"

	"github.com/DataDog/datadog-agent/cmd/serverless-init/tag"
	"github.com/DataDog/datadog-agent/pkg/config"
	logConfig "github.com/DataDog/datadog-agent/pkg/logs/config"
	serverlessLogs "github.com/DataDog/datadog-agent/pkg/serverless/logs"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	defaultFlushTimeout = 5 * time.Second
	loggerName          = "DD_CLOUDRUN_LOG_AGENT"
	logLevelEnvVar      = "DD_LOG_LEVEL"
	source              = "cloudrun"
	sourceName          = "Google Cloud Run"
)

type LogConfig struct {
	FlushTimeout time.Duration
	channel      chan *logConfig.ChannelMessage
	source       string
	loggerName   config.LoggerName
	containerId  string
}

type CustomWriter struct {
	LogConfig *LogConfig
}

func CreateConfig(containerId string) *LogConfig {
	return &LogConfig{
		FlushTimeout: defaultFlushTimeout,
		channel:      make(chan *logConfig.ChannelMessage),
		source:       source,
		loggerName:   loggerName,
		containerId:  containerId,
	}
}

func Write(conf *LogConfig, msgToSend []byte) {
	logMessage := &logConfig.ChannelMessage{
		Content: msgToSend,
	}
	conf.channel <- logMessage
}

func SetupLog(conf *LogConfig) {
	if err := config.SetupLogger(
		conf.loggerName,
		"error", // will be re-set later with the value from the env var
		"",      // logFile -> by setting this to an empty string, we don't write the logs to any file
		"",      // syslog URI
		false,   // syslog_rfc
		true,    // log_to_console
		false,   // log_format_json
	); err != nil {
		log.Errorf("Unable to setup logger: %s", err)
	}

	if logLevel := os.Getenv(logLevelEnvVar); len(logLevel) > 0 {
		if err := config.ChangeLogLevel(logLevel); err != nil {
			log.Errorf("While changing the loglevel: %s", err)
		}
	}
	serverlessLogs.SetupLogAgent(conf.channel, sourceName, source)
	serverlessLogs.SetLogsTags(getTagsWithRevision(tag.GetBaseTags(), conf.containerId))
}

func getTagsWithRevision(tags []string, containerId string) []string {
	var result []string
	result = append(result, tags...)
	result = append(result, fmt.Sprintf("containerid:%s", containerId))
	return result
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	Write(cw.LogConfig, p)
	return len(p), nil
}