// Main module for running agent
package agent

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/ry461ch/metric-collector/internal/app/agent/collector"
	"github.com/ry461ch/metric-collector/internal/app/agent/sender"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/rsa"
)

// Agent запускает агента по сбору и отправки метрик на сервер
type Agent struct {
	metricSender    *sender.Sender
	metricCollector *collector.Collector
	rsaEncypter     *rsa.RsaEncrypter
}

// Init Agent instance
func New(cfg *config.Config) *Agent {
	var encrypter *encrypt.Encrypter
	if cfg.SecretKey != "" {
		encrypter = encrypt.New(cfg.SecretKey)
	}
	var rsaEncrypter *rsa.RsaEncrypter
	if cfg.CryptoKey != "" {
		rsaEncrypter = rsa.NewEncrypter(cfg.CryptoKey)
	}

	return &Agent{
		metricSender:    sender.New(encrypter, rsaEncrypter, cfg),
		metricCollector: collector.New(cfg.PollIntervalSec),
		rsaEncypter:     rsaEncrypter,
	}
}

// Run agent work
func (a *Agent) Run(ctx context.Context) {
	stopCtx, stopCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stopCancel()

	if a.rsaEncypter != nil {
		err := a.rsaEncypter.Initialize(stopCtx)
		if err != nil {
			log.Fatal("Can't parse public key file")
			return
		}
	}

	collectorCtx, collectorCtxCancel := context.WithCancel(stopCtx)
	defer collectorCtxCancel()
	senderCtx, senderCtxCancel := context.WithCancel(stopCtx)
	defer senderCtxCancel()

	metricChannel := a.metricCollector.CollectMetricsGenerator(collectorCtx)

	go func() {
		a.metricSender.Run(senderCtx, metricChannel)
	}()

	<-stopCtx.Done()
	log.Println("Gracefull shutdown")
}
