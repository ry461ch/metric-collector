package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/app/server/crontasks/snapshotmaker"
	"github.com/ry461ch/metric-collector/internal/app/server/handlers"
	"github.com/ry461ch/metric-collector/internal/app/server/router"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/internal/storage/postgres"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func getStorage(cfg *config.Config) (storage.Storage, error) {
	if cfg.DBDsn != "" {
		return pgstorage.NewPGStorage(context.Background(), cfg.DBDsn)
	} else {
		return memstorage.NewMemStorage(context.Background()), nil
	}
}

func Run() {
	// parse args and env
	cfg := config.NewConfig()
	config.ParseArgs(cfg)
	config.ParseEnv(cfg)

	// set logger
	logging.Initialize(cfg.LogLevel)

	// initialize storage
	mStorage, err := getStorage(cfg)
	if err == nil {
		defer mStorage.Close()
	}

	fileWorker := fileworker.New(cfg.FileStoragePath, mStorage)
	if cfg.Restore && cfg.DBDsn == "" {
		fileWorker.ExportFromFile(context.Background())
	}

	handleService := handlers.New(cfg, mStorage, fileWorker)
	router := router.New(handleService)

	logging.Logger.Info("Server is running: ", cfg.Addr.String())
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
			snapshotMaker.Run(context.Background())
			wg.Done()
		}()

		// wait for signal
		go func() {
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt)
			<-stop
			fileCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			fileWorker.ImportToFile(fileCtx)
			server.Shutdown(context.Background())
			snapshotMaker.Break()
			wg.Done()
		}()

		wg.Wait()
	} else {
		http.ListenAndServe(cfg.Addr.Host+":"+strconv.FormatInt(cfg.Addr.Port, 10), router)
	}
}
