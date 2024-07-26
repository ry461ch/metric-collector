package agent

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ry461ch/metric-collector/internal/app/agent/collector"
	"github.com/ry461ch/metric-collector/internal/app/agent/sender"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
)

type (
	TimeState struct {
		LastCollectMetricTime time.Time
		LastSendMetricTime    time.Time
	}

	Agent struct {
		metricSender    *sender.Sender
		metricCollector *collector.Collector
	}
)

func New(cfg *config.Config) *Agent {
	encrypter := encrypt.New(cfg.SecretKey)

	return &Agent{
		metricSender:    sender.New(encrypter, cfg),
		metricCollector: collector.New(cfg.PollIntervalSec),
	}
}

func (a *Agent) Run() {
	collectorCtx, collectorCtxCancel := context.WithCancel(context.Background())
	senderCtx, senderCtxCancel := context.WithCancel(context.Background())

	metricChannel := make(chan metrics.Metric, 10000)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		a.metricCollector.Run(collectorCtx, metricChannel)
		wg.Done()
	}()

	go func() {
		a.metricSender.Run(senderCtx, metricChannel)
		wg.Done()
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		collectorCtxCancel()
		senderCtxCancel()
		close(metricChannel)
		wg.Done()
	}()

	wg.Wait()
}
