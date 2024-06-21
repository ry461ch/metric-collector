package server

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/ry461ch/metric-collector/internal/server/logger"
	"github.com/ry461ch/metric-collector/internal/config/netaddr"
	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/internal/server/handlers"
	"github.com/ry461ch/metric-collector/internal/server/router"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func Run() {
	// parse args and env
	options := config.Options{
		Addr: netaddr.NetAddress{Host: "localhost", Port: 8080},
		LogLevel: config.ParseLogLevel("INFO"),
	}
	config.ParseArgs(&options)
	config.ParseEnv(&options)

	// set logger
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level.SetLevel(options.LogLevel)
	logger, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	slogger.Sugar = *logger.Sugar()

	// start serving
	handleService := handlers.New(&memstorage.MemStorage{})
	router := router.New(&handleService)
	err = http.ListenAndServe(options.Addr.Host+":"+strconv.FormatInt(options.Addr.Port, 10), router)
	if err != nil {
		panic(err)
	}
}