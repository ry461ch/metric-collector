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

	"github.com/ry461ch/metric-collector/internal/app/server/crontasks/snapshotmaker"
	"github.com/ry461ch/metric-collector/internal/app/server/handlers"
	"github.com/ry461ch/metric-collector/internal/app/server/router"
	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	memstorage "github.com/ry461ch/metric-collector/internal/storage/memory"
	pgstorage "github.com/ry461ch/metric-collector/internal/storage/postgres"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

type Server struct {
	cfg           *config.Config
	metricStorage Storage
	fileWorker    *fileworker.FileWorker
	snapshotMaker *snapshotmaker.SnapshotMaker
	server        *http.Server
}

func getStorage(cfg *config.Config) Storage {
	if cfg.DBDsn != "" {
		return pgstorage.New(cfg.DBDsn)
	} else {
		return memstorage.New()
	}
}

func New(cfg *config.Config) *Server {
	logging.Initialize(cfg.LogLevel)

	// initialize storage
	metricStorage := getStorage(cfg)
	fileWorker := fileworker.New(cfg.FileStoragePath, metricStorage)
	handleService := handlers.New(cfg, metricStorage, fileWorker)
	handler := router.New(handleService, encrypt.New(cfg.SecretKey))
	snapshotMaker := snapshotmaker.New(cfg.StoreInterval, fileWorker)
	server := &http.Server{Addr: cfg.Addr.Host + ":" + strconv.FormatInt(cfg.Addr.Port, 10), Handler: handler}

	return &Server{
		cfg:           cfg,
		metricStorage: metricStorage,
		fileWorker:    fileWorker,
		snapshotMaker: snapshotMaker,
		server:        server,
	}
}

func (s *Server) Run() {
	err := s.metricStorage.Initialize(context.Background())
	if err != nil {
		logging.Logger.Warnln("Db wasn't initialized")
	} else if externalStorage, ok := s.metricStorage.(ExternalStorage); ok {
		defer externalStorage.Close()
	}

	if s.cfg.Restore && s.cfg.DBDsn == "" {
		s.fileWorker.ExportFromFile(context.Background())
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// run server
	go func() {
		logging.Logger.Info("Server is running: ", s.cfg.Addr.String())
		s.server.ListenAndServe()
		wg.Done()
	}()

	// run crontasks
	snapshotMakerCtx, snapshotMakerCtxCancel := context.WithCancel(context.Background())
	go func() {
		if s.cfg.StoreInterval != int64(0) {
			s.snapshotMaker.Run(snapshotMakerCtx)
		}
		wg.Done()
	}()

	// wait for interrupting signal
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		fileCtx, fileCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
		s.fileWorker.ImportToFile(fileCtx)
		fileCtxCancel()
		s.server.Shutdown(context.Background())
		snapshotMakerCtxCancel()
		wg.Done()
	}()

	wg.Wait()
}
