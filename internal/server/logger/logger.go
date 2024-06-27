package slogger

import (
	"go.uber.org/zap"
)

var Sugar zap.SugaredLogger

func TestInitialize() {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level.SetLevel(zap.DebugLevel)
	logger, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	Sugar = *logger.Sugar()
}
