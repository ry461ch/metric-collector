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
	agent := New(config.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		agent.Run(ctx)
	}()
	time.Sleep(2 * time.Second)
	signal.NotifyContext(ctx, os.Interrupt)
}
