package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

type Config struct {
	Addr	        netaddr.NetAddress `env:"ADDRESS"`
	LogLevel	    string 				`env:"LOG_LEVEL"`
	StoreInterval   int64 				`env:"STORE_INTERVAL"`
	FileStoragePath string 				`env:"FILE_STORAGE_PATH"`
	Restore         bool 				`env:"RESTORE"`
}

func NewConfig() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	return &Config{
		LogLevel: "INFO",
		StoreInterval: 300,
		FileStoragePath: "/tmp/metrics-db.json",
		Restore: true,
		Addr: addr,
	}
}


func ParseArgs(cfg *Config) {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.StringVar(&cfg.LogLevel, "l", "INFO", "Log level")
	flag.Int64Var(&cfg.StoreInterval, "i", 300, "Store interval seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "Load data from fileStoragePath when server is starting")
	flag.Parse()
}

func ParseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
