// Main module for running agent
package agent

import (
	"context"
	"os"
	"os/signal"

	"github.com/ry461ch/metric-collector/internal/app/agent/collector"
	"github.com/ry461ch/metric-collector/internal/app/agent/sender"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/rsaencrypt"
)

// Agent запускает агента по сбору и отправки метрик на сервер
type Agent struct {
	metricSender    *sender.Sender
	metricCollector *collector.Collector
	rsaencypter     *rsaencrypt.RsaEncrypter
}

// Init Agent instance
func New(cfg *config.Config) *Agent {
	var encrypter *encrypt.Encrypter
	if cfg.SecretKey != "" {
		encrypter = encrypt.New(cfg.SecretKey)
	}
	var rsaEncrypter *rsaencrypt.RsaEncrypter
	if cfg.CryptoKey != "" {
		rsaEncrypter = rsaencrypt.New(cfg.CryptoKey)
	}

	return &Agent{
		metricSender:    sender.New(encrypter, rsaEncrypter, cfg),
		metricCollector: collector.New(cfg.PollIntervalSec),
		rsaencypter:     rsaEncrypter,
	}
}

// Run agent work
func (a *Agent) Run(ctx context.Context) {
	if a.rsaencypter != nil {
		a.rsaencypter.Initialize(ctx)
	}

	collectorCtx, collectorCtxCancel := context.WithCancel(ctx)
	senderCtx, senderCtxCancel := context.WithCancel(ctx)

	metricChannel := a.metricCollector.CollectMetricsGenerator(collectorCtx)

	go func() {
		a.metricSender.Run(senderCtx, metricChannel)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
	case <-ctx.Done():
	}
	collectorCtxCancel()
	senderCtxCancel()
}
