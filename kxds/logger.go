package kxds

import (
	"fmt"

	"github.com/go-logr/logr"
)

type Logger struct {
	logger logr.Logger
}

func NewLogger(logger logr.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.V(5).Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}
