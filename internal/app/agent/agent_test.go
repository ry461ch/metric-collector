package agent

import (
	"context"
	"os"
	"os/signal"
	"testing"
	"time"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
)

func TestBase(t *testing.T) {
	cfg := config.New()
	cfg.PollIntervalSec = 2
	cfg.ReportIntervalSec = 3
	agent := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		agent.Run(ctx)
	}()
	time.Sleep(3 * time.Second)
	signal.NotifyContext(ctx, os.Interrupt)
}
