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

	"github.com/ry461ch/metric-collector/internal/agent/config"
	"github.com/ry461ch/metric-collector/internal/config/netaddr"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

type (
	TimeState struct {
		LastCollectMetricTime time.Time
		LastSendMetricTime    time.Time
	}

	Agent struct {
		timeState *TimeState
		options   config.Options
		// we don't need to use storage interface here, 
		// because we will always use memory storage in agent
		metricStorage  *memstorage.MemStorage
		metricService  *metricservice.MetricService
	}
)

func New(timeState *TimeState, options config.Options, metricStorage *memstorage.MemStorage) *Agent {
	return &Agent{
		timeState: timeState,
		options: options,
		metricStorage: metricStorage,
		metricService: metricservice.New(metricStorage),
	}
}

func (a *Agent) collectMetric() {
	log.Println("Trying to collect metrics")
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	a.metricStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	a.metricStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	a.metricStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	a.metricStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	a.metricStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	a.metricStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	a.metricStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	a.metricStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	a.metricStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	a.metricStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	a.metricStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	a.metricStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	a.metricStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	a.metricStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	a.metricStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	a.metricStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	a.metricStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	a.metricStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	a.metricStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	a.metricStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	a.metricStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	a.metricStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	a.metricStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	a.metricStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	a.metricStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	a.metricStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	a.metricStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	a.metricStorage.UpdateCounterValue("PollCount", 1)
	a.metricStorage.UpdateGaugeValue("RandomValue", rand.Float64())
	log.Println("Successfully got all metrics")
}

func (a *Agent) sendMetrics() error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + a.options.Addr.Host + ":" + strconv.FormatInt(a.options.Addr.Port, 10)

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
		time.Duration(time.Duration(a.options.PollIntervalSec)*time.Second) <= time.Since(a.timeState.LastCollectMetricTime) {
		a.collectMetric()
		a.timeState.LastCollectMetricTime = time.Now()
	}

	if a.timeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.options.ReportIntervalSec)*time.Second) <= time.Since(a.timeState.LastSendMetricTime) {
		a.sendMetrics()
		a.timeState.LastSendMetricTime = time.Now()
	}
}

func Run() {
	options := config.Options{Addr: netaddr.NetAddress{Host: "localhost", Port: 8080}}
	config.ParseArgs(&options)
	config.ParseEnv(&options)

	mAgent := New(&TimeState{}, options, &memstorage.MemStorage{})
	for {
		mAgent.runIteration()
		time.Sleep(time.Second)
	}
}
