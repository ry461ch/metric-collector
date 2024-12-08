package agent

import (
	"context"
	"testing"
	"time"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
)

func TestBase(t *testing.T) {
	publicKeyPath := "/tmp/public.test"

	cfg := config.New()
	cfg.PollIntervalSec = 2
	cfg.ReportIntervalSec = 3
	cfg.CryptoKey = publicKeyPath
	cfg.SecretKey = "secret"
	agent := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		agent.Run(ctx)
	}()
	time.Sleep(3 * time.Second)
}
