package server

import (
	"context"
	"testing"
	"time"

	config "github.com/ry461ch/metric-collector/internal/config/server"
)

func TestBase(t *testing.T) {
	server := New(config.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		server.Run(ctx)
	}()
	time.Sleep(2 * time.Second)
}
