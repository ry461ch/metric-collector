// Main module for running server
package server

import (
	"context"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/ry461ch/metric-collector/internal/app/server/crontasks/snapshotmaker"
	metricsgrpc "github.com/ry461ch/metric-collector/internal/app/server/grpc"
	"github.com/ry461ch/metric-collector/internal/app/server/handlers"
	"github.com/ry461ch/metric-collector/internal/app/server/router"
	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	pb "github.com/ry461ch/metric-collector/internal/proto"
	memstorage "github.com/ry461ch/metric-collector/internal/storage/memory"
	pgstorage "github.com/ry461ch/metric-collector/internal/storage/postgres"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/ipchecker"
	ipcheckermiddleware "github.com/ry461ch/metric-collector/pkg/ipchecker/middleware"
	"github.com/ry461ch/metric-collector/pkg/logging"
	"github.com/ry461ch/metric-collector/pkg/logging/middleware"
	"github.com/ry461ch/metric-collector/pkg/rsa"
	rsamiddleware "github.com/ry461ch/metric-collector/pkg/rsa/middleware"
)

// Сервер для сбора и сохранения метрик
type Server struct {
	cfg           *config.Config
	metricStorage Storage
	fileWorker    *fileworker.FileWorker
	snapshotMaker *snapshotmaker.SnapshotMaker
	server        *http.Server
	rsaDecrypter  *rsa.RsaDecrypter
	grpcServer    *metricsgrpc.MetricsGRPCServer
	ipChecker     *ipchecker.IPChecker
}

func getStorage(cfg *config.Config) Storage {
	if cfg.DBDsn != "" {
		return pgstorage.New(cfg.DBDsn)
	} else {
		return memstorage.New()
	}
}

// Init server instance
func New(cfg *config.Config) *Server {
	logging.Initialize(cfg.LogLevel)

	var rsaDecrypter *rsa.RsaDecrypter
	if cfg.CryptoKey != "" {
		rsaDecrypter = rsa.NewDecrypter(cfg.CryptoKey)
	}

	var ipChecker *ipchecker.IPChecker
	if cfg.TrustedSubnet != "" {
		ipChecker = ipchecker.New(cfg.TrustedSubnet)
	}

	// initialize storage
	metricStorage := getStorage(cfg)
	fileWorker := fileworker.New(cfg.FileStoragePath, metricStorage)
	handleService := handlers.New(cfg, metricStorage, fileWorker)
	handler := router.New(handleService, encrypt.New(cfg.SecretKey), rsaDecrypter, ipChecker)
	snapshotMaker := snapshotmaker.New(cfg.StoreInterval, fileWorker)
	server := &http.Server{Addr: cfg.Addr.Host + ":" + strconv.FormatInt(cfg.Addr.Port, 10), Handler: handler}
	grpcServer := metricsgrpc.New(cfg, metricStorage, fileWorker)

	return &Server{
		cfg:           cfg,
		metricStorage: metricStorage,
		fileWorker:    fileWorker,
		snapshotMaker: snapshotMaker,
		server:        server,
		rsaDecrypter:  rsaDecrypter,
		grpcServer:    grpcServer,
		ipChecker:     ipChecker,
	}
}

// Run server
func (s *Server) Run(ctx context.Context) {
	stopCtx, stopCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stopCancel()

	err := s.metricStorage.Initialize(stopCtx)
	if s.rsaDecrypter != nil {
		err = s.rsaDecrypter.Initialize(stopCtx)
		if err != nil {
			logging.Logger.Errorln("Can't parse private key file")
			return
		}
	}

	if err != nil {
		logging.Logger.Warnln("Db wasn't initialized")
	} else if externalStorage, ok := s.metricStorage.(ExternalStorage); ok {
		defer externalStorage.Close()
	}

	if s.cfg.Restore && s.cfg.DBDsn == "" {
		s.fileWorker.ExportFromFile(stopCtx)
	}

	// run server
	go func() {
		logging.Logger.Info("Server is running: ", s.cfg.Addr.String())
		s.server.ListenAndServe()
	}()

	// prepare grpc server
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		logging.Logger.Fatal(err)
	}

	var interceptors []grpc.StreamServerInterceptor
	interceptors = append(interceptors, requestlogger.LoggingStreamServerInterceptor)
	if s.ipChecker != nil {
		interceptors = append(interceptors, ipcheckermiddleware.CheckGRPCRequesterIP(s.ipChecker))
	}
	if s.rsaDecrypter != nil {
		interceptors = append(interceptors, rsamiddleware.DecryptStreamServerInterceptor(s.rsaDecrypter))
	}
	grpcServer := grpc.NewServer(grpc.ChainStreamInterceptor(interceptors...))
	pb.RegisterMetricsServer(grpcServer, s.grpcServer)

	// run grpc server
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logging.Logger.Fatal(err)
		}
	}()

	// run crontasks
	snapshotMakerCtx, snapshotMakerCtxCancel := context.WithCancel(stopCtx)
	defer snapshotMakerCtxCancel()
	go func() {
		if s.cfg.StoreInterval != int64(0) {
			s.snapshotMaker.Run(snapshotMakerCtx)
		}
	}()

	<-stopCtx.Done()
	grpcServer.GracefulStop()
	fileCtx, fileCtxCancel := context.WithTimeout(ctx, 1*time.Second)
	s.fileWorker.ImportToFile(fileCtx)
	fileCtxCancel()
	s.server.Shutdown(ctx)
	logging.Logger.Infoln("Gracefull shutdown")
}
