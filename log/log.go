package log

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelTrace = "trace"
)

const timeFormat = "2006-01-02T15:04:05.999999999"

// Bootstrap logging
func Bootstrap(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = timeFormat
	var trimPrefixes = []string{
		"/github.com/larashed/agent-go",
		"go/src/larashed/",
		"/vendor",
		"/go/pkg/mod",
	}
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		var ok bool
		for _, prefix := range trimPrefixes {
			file, ok = trimLeftInclusive(file, prefix)
			if ok {
				break
			}
		}
		return fmt.Sprintf("%-41v", file+":"+strconv.Itoa(line))
	}
	console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}
	log.Logger = zerolog.New(console).With().Caller().Timestamp().Logger()
}

// ParseLoggingLevel maps internal log level to zerolog's log level
func ParseLoggingLevel(level string) zerolog.Level {
	switch level {
	case levelDebug:
		return zerolog.DebugLevel
	case levelInfo:
		return zerolog.InfoLevel
	case levelTrace:
		return zerolog.TraceLevel
	}

	return zerolog.DebugLevel
}

// trimLeftInclusive trims left part of the string up to and including the prefix.
func trimLeftInclusive(s string, prefix string) (string, bool) {
	start := strings.Index(s, prefix)
	if start != -1 {
		return s[start+len(prefix):], true
	}
	return s, false
}
