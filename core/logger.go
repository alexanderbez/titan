package core

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// Logger is a simple wrapper around a Logrus logger.
type Logger struct {
	log.FieldLogger
}

// NewLogger returns a new initialized logger with the level, formatter, and
// output set.
func NewLogger(out io.Writer, debug bool) Logger {
	logger := log.New()
	logger.Out = out

	if out != os.Stdout {
		logger.Formatter = &log.JSONFormatter{}
	} else {
		logger.Formatter = &log.TextFormatter{FullTimestamp: true}
	}

	if debug {
		logger.Level = log.DebugLevel
	}

	return Logger{logger}
}

// With returns a new logger with the given key/value field mapping.
func (logger Logger) With(key, value string) Logger {
	ctxLogger := logger.WithField(key, value)
	return Logger{ctxLogger}
}
