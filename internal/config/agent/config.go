// Module for parsing agent environment and flags
package agentconfig

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/jessevdk/go-flags"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

// Конфиг агента
type Config struct {
	ReportIntervalSec int64              `short:"r" env:"REPORT_INTERVAL" json:"report_interval"`
	PollIntervalSec   int64              `short:"p" env:"POLL_INTERVAL" json:"poll_interval"`
	Addr              netaddr.NetAddress `short:"a" env:"ADDRESS" json:"address"`
	SecretKey         string             `short:"k" env:"KEY"`
	RateLimit         int64              `short:"l" env:"RATE_LIMIT"`
	CryptoKey         string             `long:"crypto-key" env:"CRYPTO_KEY" json:"crypto_key"`
	Config            string             `long:"config" short:"c" env:"CONFIG"`
}

// Парсинг аргументов и переменных окружения для создания конфига агента
func New() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	cfg := &Config{ReportIntervalSec: 10, PollIntervalSec: 2, Addr: addr}

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
