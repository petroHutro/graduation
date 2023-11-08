package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

var l Logger

func Error(msg string, fields ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, fields...))
}

func Info(msg string, fields ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, fields...))
}

func Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(msg, fields...)
}

func Shutdown() {
	l.logger.Sync()
}
