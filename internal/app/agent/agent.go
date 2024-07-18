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

	config "github.com/ry461ch/metric-collector/internal/config/agent"
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

func NewAgent(config *config.Config) *Agent {
	return &Agent{
		timeState:  &TimeState{},
		config:     config,
		memStorage: memstorage.NewMemStorage(),
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
	req, err := json.Marshal(metricList)
	if err != nil {
		return fmt.Errorf("can't convert model Metric to json")
	}

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
		log.Println("Server is not available")
		return fmt.Errorf("invalid request, server returned: %d", resp.StatusCode())
	}

	log.Println("Successfully send all metrics")
	return nil
}

func (a *Agent) collectAndSendMetrics(ctx context.Context) {
	defaultTime := time.Time{}
	if a.timeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.PollIntervalSec)*time.Second) <= time.Since(a.timeState.LastCollectMetricTime) {
		a.collectMetric(ctx)
		a.timeState.LastCollectMetricTime = time.Now()
	}

	if a.timeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.ReportIntervalSec)*time.Second) <= time.Since(a.timeState.LastSendMetricTime) {
		err := a.sendMetrics(ctx)
		if err != nil {
			return
		}
		a.timeState.LastSendMetricTime = time.Now()
	}
}

func (a *Agent) Run() {
	ctx := context.Background()
	for {
		iterCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		a.collectAndSendMetrics(iterCtx)
		cancel()
		time.Sleep(time.Second)
	}
}
