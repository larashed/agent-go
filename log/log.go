package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelTrace = "trace"
)

const timeFormat = "2006-01-02T15:04:05.999999999"

func Bootstrap(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = timeFormat

	console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}
	log.Logger = zerolog.New(console).With().Caller().Timestamp().Logger()
}

func ParseLoggingLevel(level string) zerolog.Level {
	switch level {
	case levelDebug:
		return zerolog.DebugLevel
	case levelInfo:
		return zerolog.InfoLevel
	case levelTrace:
		return zerolog.TraceLevel
	}

	return zerolog.TraceLevel
}
