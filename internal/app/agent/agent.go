package agent

import (
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
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

type (
	TimeState struct {
		LastCollectMetricTime time.Time
		LastSendMetricTime    time.Time
	}

	Agent struct {
		timeState *TimeState
		config   *config.Config
		metricService  *metricservice.MetricService
	}
)

func New(timeState *TimeState, config *config.Config, metricService *metricservice.MetricService) *Agent {
	return &Agent{
		timeState: timeState,
		config: config,
		metricService: metricService,
	}
}

func (a *Agent) collectMetric() {
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

	metricList := []metrics.Metrics{}
	for key, val := range metricGaugeMap {
		metricList = append(metricList, metrics.Metrics{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
	}
	for key, val := range metricCounterMap {
		metricList = append(metricList, metrics.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &val,
		})
	}
	a.metricService.SaveMetrics(metricList)
	log.Println("Successfully got all metrics")
}

func (a *Agent) sendMetrics() error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + a.config.Addr.Host + ":" + strconv.FormatInt(a.config.Addr.Port, 10)

	client := resty.New()

	metricList := a.metricService.ExtractMetrics()

	for _, metric := range metricList {
		req, _ := json.Marshal(metric)
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req).
			Post(serverURL + "/update/")
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("invalid request, server returned: %d", resp.StatusCode())
		}
	}
	log.Println("Successfully send all metrics")
	return nil
}

func (a *Agent) runIteration() {
	defaultTime := time.Time{}
	if a.timeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.PollIntervalSec)*time.Second) <= time.Since(a.timeState.LastCollectMetricTime) {
		a.collectMetric()
		a.timeState.LastCollectMetricTime = time.Now()
	}

	if a.timeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.config.ReportIntervalSec)*time.Second) <= time.Since(a.timeState.LastSendMetricTime) {
		a.sendMetrics()
		a.timeState.LastSendMetricTime = time.Now()
	}
}

func Run() {
	cfg := config.NewConfig()
	config.ParseArgs(cfg)
	config.ParseEnv(cfg)
	log.Println(cfg.Addr.String())

	metricService := metricservice.New(&memstorage.MemStorage{})
	mAgent := New(&TimeState{}, cfg, metricService)
	for {
		mAgent.runIteration()
		time.Sleep(time.Second)
	}
}
