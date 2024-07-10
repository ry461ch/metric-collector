package config

import (
	"flag"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/ry461ch/metric-collector/internal/config/netaddr"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevel	    string `env:"LOG_LEVEL"`
	StoreInterval   string `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         string `env:"RESTORE"`
}

type Options struct {
	Addr            netaddr.NetAddress
	LogLevel        string
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
}


func ParseArgs(opt *Options) {
	flag.Var(&opt.Addr, "a", "Net address host:port")
	flag.StringVar(&opt.LogLevel, "l", "INFO", "Log level")
	flag.Int64Var(&opt.StoreInterval, "i", 300, "Store interval seconds")
	flag.StringVar(&opt.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&opt.Restore, "r", true, "Load data from fileStoragePath when server is starting")
	flag.Parse()
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
	if cfg.LogLevel != "" {
		opt.LogLevel = cfg.LogLevel
	}
	if cfg.StoreInterval != "" {
		opt.StoreInterval, _ = strconv.ParseInt(cfg.StoreInterval, 10, 0)
	}
	if cfg.Restore != "" {
		opt.Restore, _ = strconv.ParseBool(cfg.Restore)
	}
	if cfg.FileStoragePath != "" {
		opt.FileStoragePath = cfg.FileStoragePath
	}
}
