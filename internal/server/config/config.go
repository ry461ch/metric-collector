package config

import (
	"flag"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"

	"github.com/ry461ch/metric-collector/internal/config/netaddr"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevelStr     string `env:"LOG_LEVEL"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type Options struct {
	Addr            netaddr.NetAddress
	LogLevel        zapcore.Level
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
}

func ParseLogLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.InvalidLevel
	}
}

func ParseArgs(opt *Options) {
	flag.Var(&opt.Addr, "a", "Net address host:port")
	logLevelStr := flag.String("l", "INFO", "Log Level")
	flag.Int64Var(&opt.StoreInterval, "i", 300, "Store interval seconds")
	flag.StringVar(&opt.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&opt.Restore, "r", true, "Load data from fileStoragePath when server is starting")
	flag.Parse()
	opt.LogLevel = ParseLogLevel(*logLevelStr)
}

func ParseEnv(opt *Options) {
	var cfg Config
	env.Parse(&cfg)
	if cfg.Address != "" {
		addrParts := strings.Split(cfg.Address, ":")
		port, _ := strconv.ParseInt(addrParts[1], 10, 0)
		opt.Addr.Host = addrParts[0]
		opt.Addr.Port = port
	}
	if cfg.LogLevelStr != "" {
		opt.LogLevel = ParseLogLevel(cfg.LogLevelStr)
	}
}
