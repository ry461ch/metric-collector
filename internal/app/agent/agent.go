package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/app/agent/config"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

type (
	TimeState struct {
		LastCollectMetricTime time.Time
		LastSendMetricTime    time.Time
	}

	Agent struct {
		timeState  *TimeState
		config     *config.Config
		memStorage *memstorage.MemStorage
	}
)

func New(timeState *TimeState, config *config.Config, memStorage *memstorage.MemStorage) *Agent {
	return &Agent{
		timeState:  timeState,
		config:     config,
		memStorage: memStorage,
	}
}

func (a *Agent) collectMetric(ctx context.Context) {
	log.Println("Trying to collect metrics")
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	metricGaugeMap := map[string]float64{}
	metricGaugeMap["Alloc"] = float64(rtm.Alloc)
	metricGaugeMap["BuckHashSys"] = float64(rtm.BuckHashSys)
	metricGaugeMap["Frees"] = float64(rtm.Frees)
	metricGaugeMap["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	metricGaugeMap["GCSys"] = float64(rtm.GCSys)
	metricGaugeMap["HeapAlloc"] = float64(rtm.HeapAlloc)
	metricGaugeMap["HeapIdle"] = float64(rtm.HeapIdle)
	metricGaugeMap["HeapInuse"] = float64(rtm.HeapInuse)
	metricGaugeMap["HeapObjects"] = float64(rtm.HeapObjects)
	metricGaugeMap["HeapReleased"] = float64(rtm.HeapReleased)
	metricGaugeMap["HeapSys"] = float64(rtm.HeapSys)
	metricGaugeMap["LastGC"] = float64(rtm.LastGC)
	metricGaugeMap["Lookups"] = float64(rtm.Lookups)
	metricGaugeMap["MCacheInuse"] = float64(rtm.MCacheInuse)
	metricGaugeMap["MCacheSys"] = float64(rtm.MCacheSys)
	metricGaugeMap["MSpanInuse"] = float64(rtm.MSpanInuse)
	metricGaugeMap["MSpanSys"] = float64(rtm.MSpanSys)
	metricGaugeMap["Mallocs"] = float64(rtm.Mallocs)
	metricGaugeMap["NextGC"] = float64(rtm.NextGC)
	metricGaugeMap["NumForcedGC"] = float64(rtm.NumForcedGC)
	metricGaugeMap["NumGC"] = float64(rtm.NumGC)
	metricGaugeMap["OtherSys"] = float64(rtm.OtherSys)
	metricGaugeMap["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	metricGaugeMap["StackInuse"] = float64(rtm.StackInuse)
	metricGaugeMap["StackSys"] = float64(rtm.StackSys)
	metricGaugeMap["Sys"] = float64(rtm.Sys)
	metricGaugeMap["TotalAlloc"] = float64(rtm.TotalAlloc)
	metricGaugeMap["RandomValue"] = rand.Float64()

	metricCounterMap := map[string]int64{}
	metricCounterMap["PollCount"] = 1

	metricList := []metrics.Metric{}
	for key, val := range metricGaugeMap {
		metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
	}
	for key, val := range metricCounterMap {
		metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "counter",
			Delta: &val,
		})
	}
	a.memStorage.SaveMetrics(ctx, metricList)
	log.Println("Successfully got all metrics")
}

func (a *Agent) sendMetrics(ctx context.Context) error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + a.config.Addr.Host + ":" + strconv.FormatInt(a.config.Addr.Port, 10)

	client := resty.New()

	metricList, _ := a.memStorage.ExtractMetrics(ctx)
	if len(metricList) == 0 {
		return nil
	}
	req, _ := json.Marshal(metricList)

	for i := 1; i < 7; i += 2 {
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req).
			Post(serverURL + "/updates/")
		if err != nil || resp.StatusCode() == http.StatusInternalServerError {
			continue
		}
		if resp.StatusCode() == http.StatusOK {
			log.Println("Successfully send all metrics")
			return nil
		}
		return fmt.Errorf("invalid request, server returned: %d", resp.StatusCode())
	}

	log.Println("Successfully send all metrics")
	return nil
}

func (a *Agent) runIteration(ctx context.Context) {
	iterCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	defaultTime := time.Time{}
	if a.timeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.PollIntervalSec)*time.Second) <= time.Since(a.timeState.LastCollectMetricTime) {
		a.collectMetric(iterCtx)
		a.timeState.LastCollectMetricTime = time.Now()
	}

	if a.timeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.ReportIntervalSec)*time.Second) <= time.Since(a.timeState.LastSendMetricTime) {
		a.sendMetrics(iterCtx)
		a.timeState.LastSendMetricTime = time.Now()
	}
}

func Run() {
	cfg := config.NewConfig()
	config.ParseArgs(cfg)
	config.ParseEnv(cfg)
	log.Println(cfg.Addr.String())
	ctx := context.Background()

	memStorage := memstorage.NewMemStorage(ctx)
	mAgent := New(&TimeState{}, cfg, memStorage)
	for {
		mAgent.runIteration(ctx)
		time.Sleep(time.Second)
	}
}
