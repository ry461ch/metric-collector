package server

import (
	"net/http"
	"strconv"
	"os/signal"
	"os"
	"sync"
	"context"

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
		var wg sync.WaitGroup
		wg.Add(3)
		
		// run server
		server := &http.Server{Addr: options.Addr.Host+":"+strconv.FormatInt(options.Addr.Port, 10), Handler: router}
		go func() {
			server.ListenAndServe()
			wg.Done()
		}()

		// run crontasks
		snapshotMaker := snapshotmaker.New(&snapshotmaker.TimeState{}, options, &metricStorage)
		go func() {
			snapshotMaker.Run()
			wg.Done()
		}()

		// wait for signal
		go func() {
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt)
			<-stop
			fileWorker.ImportToFile()
			server.Shutdown(context.Background())
			snapshotMaker.Break()
			wg.Done()
		}()

		wg.Wait()
	} else {
		http.ListenAndServe(options.Addr.Host+":"+strconv.FormatInt(options.Addr.Port, 10), router)
	}
}
