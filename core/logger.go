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

// CreateBaseLogger creates a base logger for which other components and modules
// can extend upon, inheriting the base configuration and settings of a core
// logger. An error is returned if a non-empty log file path is given that
// cannot be created.
func CreateBaseLogger(logOut string, debug bool) (Logger, error) {
	logFile := os.Stdout

	if logOut != "" {
		file, err := os.OpenFile(logOut, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return Logger{}, err
		}

		logFile = file
	}

	return NewLogger(logFile, debug), nil
}
