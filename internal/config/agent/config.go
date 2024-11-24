// Module for parsing agent environment and flags
package agentconfig

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

// Конфиг агента
type Config struct {
	ReportIntervalSec int64              `env:"REPORT_INTERVAL" json:"report_interval"`
	PollIntervalSec   int64              `env:"POLL_INTERVAL" json:"poll_interval"`
	Addr              netaddr.NetAddress `env:"ADDRESS" json:"address"`
	SecretKey         string             `env:"KEY"`
	RateLimit         int64              `env:"RATE_LIMIT"`
	CryptoKey         string             `env:"CRYPTO_KEY" json:"crypto_key"`
	Config            string             `env:"CONFIG"`
}

// Парсинг аргументов и переменных окружения для создания конфига агента
func New() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	cfg := &Config{ReportIntervalSec: 10, PollIntervalSec: 2, Addr: addr}

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
	flag.Int64Var(&cfg.ReportIntervalSec, "r", 10, "Interval of sending metrics to the server")
	flag.Int64Var(&cfg.PollIntervalSec, "p", 2, "Interval of polling metrics from runtime")
	flag.StringVar(&cfg.SecretKey, "k", "", "Secret key")
	flag.Int64Var(&cfg.RateLimit, "l", 1, "number of workers, which send metrics to the server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Crypto key file")
	flag.Parse()
}

func parseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
