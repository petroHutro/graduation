package logger

import (
	"fmt"
	"graduation/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newFileLogger(logFile string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{logFile}
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("cannot create file logger Build: %w", err)
	}
	return logger, nil
}

func newConsoleLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.EncodeCaller = nil
	config.DisableStacktrace = true
	logger, _ := config.Build()
	return logger, nil
}

func newMultiLogger(filePath string) (*zap.Logger, error) {
	consoleLogger, err := newConsoleLogger()
	if err != nil {
		return nil, fmt.Errorf("cannot create consol logger in multi logger: %w", err)
	}
	fileLogger, err := newFileLogger(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot create file logger in multi logger: %w", err)
	}
	multiCore := zapcore.NewTee(fileLogger.Core(), consoleLogger.Core())
	logger := zap.New(multiCore)
	return logger, nil
}

func InitLogger(conf config.Logger) error {
	if conf.MultiFlag {
		logger, err := newMultiLogger(conf.FilePath)
		if err != nil {
			return fmt.Errorf("cannot create multi logger: %w", err)
		}
		l.logger = logger
		return nil
	} else if conf.FileFlag {
		logger, err := newFileLogger(conf.FilePath)
		if err != nil {
			return fmt.Errorf("cannot create file logger: %w", err)
		}
		l.logger = logger
		return nil
	}
	logger, err := newConsoleLogger()
	if err != nil {
		return fmt.Errorf("cannot create consol logger: %w", err)
	}
	l.logger = logger
	return nil
}
