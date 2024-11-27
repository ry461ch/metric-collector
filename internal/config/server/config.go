// Module for parsing server environment and flags
package serverconfig

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/jessevdk/go-flags"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

// Конфиг сервера
type Config struct {
	DBDsn           string             `short:"d" env:"DATABASE_DSN" json:"database_dsn"`
	Addr            netaddr.NetAddress `short:"a" env:"ADDRESS" json:"address"`
	LogLevel        string             `short:"l" env:"LOG_LEVEL"`
	StoreInterval   int64              `short:"i" env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string             `short:"f" env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool               `short:"r" env:"RESTORE" json:"restore"`
	SecretKey       string             `short:"k" env:"KEY"`
	CryptoKey       string             `long:"crypto-key" env:"CRYPTO_KEY" json:"crypto_key"`
	Config          string             `long:"config" short:"c" env:"CONFIG"`
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

	args := os.Args[1:]
	if len(args) == 0 || strings.Contains(args[1], "test") {
		return cfg
	}

	cfgFile := os.Getenv("CONFIG")
	parseArgs(cfg)
	if cfgFile == "" && cfg.Config != "" {
		cfgFile = cfg.Config
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

	parseEnv(cfg)
	parseArgs(cfg)
	return cfg
}

func parseArgs(cfg *Config) {
	_, err := flags.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}

func parseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
