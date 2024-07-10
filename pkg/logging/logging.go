package logging

import (
	"errors"

	"go.uber.org/zap"
)

var Logger zap.SugaredLogger

func Initialize(level string) error {
	logCfg := zap.NewDevelopmentConfig()
	switch level {
	case "DEBUG":
		logCfg.Level.SetLevel(zap.DebugLevel)
	case "INFO":
		logCfg.Level.SetLevel(zap.InfoLevel)
	case "WARN":
		logCfg.Level.SetLevel(zap.WarnLevel)
	case "ERROR":
		logCfg.Level.SetLevel(zap.ErrorLevel)
	default:
		return errors.New("invalid logging level")
	}
	logger, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	Logger = *logger.Sugar()
	return nil
}
