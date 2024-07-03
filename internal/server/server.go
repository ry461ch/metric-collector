package server

import (
	"net/http"
	"strconv"

	"github.com/ry461ch/metric-collector/internal/config/netaddr"
	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/internal/server/crontasks/snapshotmaker"
	"github.com/ry461ch/metric-collector/internal/server/handlers"
	"github.com/ry461ch/metric-collector/internal/server/router"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func Run() {
	// parse args and env
	options := config.Options{
		Addr:     netaddr.NetAddress{Host: "localhost", Port: 8080},
		LogLevel: "INFO",
	}
	config.ParseArgs(&options)
	config.ParseEnv(&options)

	// set logger
	logging.Initialize(options.LogLevel)

	// initialize storage
	metricStorage := memstorage.MemStorage{}
	fileWorker := fileworker.New(options.FileStoragePath, &metricStorage)
	if options.Restore {
		err := fileWorker.ExportFromFile()
		if err != nil {
			panic(err)
		}
	}

	// prepare for serving
	handleService := handlers.New(&metricStorage, options)
	router := router.New(handleService)

	if options.StoreInterval != int64(0) {
		go http.ListenAndServe(options.Addr.Host+":"+strconv.FormatInt(options.Addr.Port, 10), router)

		// run crontasks
		snapshotMaker := snapshotmaker.New(&snapshotmaker.TimeState{}, options, &metricStorage)
		snapshotMaker.Run()
	} else {
		http.ListenAndServe(options.Addr.Host+":"+strconv.FormatInt(options.Addr.Port, 10), router)
	}
}
