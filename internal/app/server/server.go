package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/app/server/crontasks/snapshotmaker"
	"github.com/ry461ch/metric-collector/internal/app/server/handlers"
	"github.com/ry461ch/metric-collector/internal/app/server/router"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func Run() {
	// parse args and env
	cfg := config.NewConfig()
	config.ParseArgs(cfg)
	config.ParseEnv(cfg)

	// set logger
	logging.Initialize(cfg.LogLevel)

	// Init db
	db, _ := sql.Open("pgx", cfg.DBDsn)
	defer db.Close()

	// initialize storage
	metricStorage := memstorage.MemStorage{}
	metricService := metricservice.New(&metricStorage)
	fileWorker := fileworker.New(cfg.FileStoragePath, metricService)
	if cfg.Restore {
		err := fileWorker.ExportFromFile()
		if err != nil {
			panic(err)
		}
	}

	// prepare for serving
	handleService := handlers.New(cfg, metricService, fileWorker, db)
	router := router.New(handleService)

	logging.Logger.Info(cfg.Addr.String())
	if cfg.StoreInterval != int64(0) {
		var wg sync.WaitGroup
		wg.Add(3)
		
		// run server
		server := &http.Server{Addr: cfg.Addr.Host+":"+strconv.FormatInt(cfg.Addr.Port, 10), Handler: router}
		go func() {
			server.ListenAndServe()
			wg.Done()
		}()

		// run crontasks
		snapshotMaker := snapshotmaker.New(&snapshotmaker.TimeState{}, cfg, fileWorker)
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
		http.ListenAndServe(cfg.Addr.Host+":"+strconv.FormatInt(cfg.Addr.Port, 10), router)
	}
}
