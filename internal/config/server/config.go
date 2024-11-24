// Module for parsing server environment and flags
package serverconfig

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

// Конфиг сервера
type Config struct {
	DBDsn           string             `env:"DATABASE_DSN" json:"database_dsn"`
	Addr            netaddr.NetAddress `env:"ADDRESS" json:"address"`
	LogLevel        string             `env:"LOG_LEVEL"`
	StoreInterval   int64              `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string             `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool               `env:"RESTORE" json:"restore"`
	SecretKey       string             `env:"KEY"`
	CryptoKey       string             `env:"CRYPTO_KEY" json:"crypto_key"`
	Config          string             `env:"CONFIG"`
}

// Парсинг аргументов и переменных окружения для создания конфига сервера
func New() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	cfg := &Config{
		LogLevel:        "INFO",
		StoreInterval:   10,
		FileStoragePath: "/tmp/metrics-db.json",
		Restore:         true,
		Addr:            addr,
	}

	cfgFile := os.Getenv("CONFIG")
	if cfgFile == "" {
		flag.StringVar(&cfgFile, "config", "", "Config file")
		flag.Parse()
	}
	if cfgFile != "" {
		cfgData, err := os.ReadFile(cfgFile)
		if err != nil {
			log.Fatalf("Can't parse env variables: %s", err)
			return nil
		}

		err = json.Unmarshal(cfgData, cfg)
		if err != nil {
			log.Fatalf("Can't parse env variables: %s", err)
			return nil
		}
	}

	parseArgs(cfg)
	parseEnv(cfg)
	return cfg
}

func parseArgs(cfg *Config) {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.StringVar(&cfg.LogLevel, "l", "INFO", "Log level")
	flag.StringVar(&cfg.DBDsn, "d", "", "database connection string")
	flag.Int64Var(&cfg.StoreInterval, "i", 10, "Store interval seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "Load data from fileStoragePath when server is starting")
	flag.StringVar(&cfg.SecretKey, "k", "", "Secret key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Crypto key file")
	flag.Parse()
}

func parseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
