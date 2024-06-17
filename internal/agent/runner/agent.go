package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/agent/config"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type Agent struct {
	TimeState *config.TimeState
	Options   config.Options
	MStorage  storage.Storage
}

func (a *Agent) CollectMetric() {
	log.Println("Trying to collect metrics")
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	a.MStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	a.MStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	a.MStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	a.MStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	a.MStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	a.MStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	a.MStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	a.MStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	a.MStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	a.MStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	a.MStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	a.MStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	a.MStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	a.MStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	a.MStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	a.MStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	a.MStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	a.MStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	a.MStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	a.MStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	a.MStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	a.MStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	a.MStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	a.MStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	a.MStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	a.MStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	a.MStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	a.MStorage.UpdateCounterValue("PollCount", 1)
	a.MStorage.UpdateGaugeValue("RandomValue", rand.Float64())
	log.Println("Successfully got all metrics")
}

func (ma *Agent) SendMetric() error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + ma.Options.Addr.Host + ":" + strconv.FormatInt(ma.Options.Addr.Port, 10)

	client := resty.New()
	for metricName, val := range ma.MStorage.GetGaugeValues() {
		path := "/update/gauge/" + metricName + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	for metricName, val := range ma.MStorage.GetCounterValues() {
		path := "/update/counter/" + metricName + "/" + strconv.FormatInt(val, 10)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	log.Println("Successfully send all metrics")
	return nil
}

func (a *Agent) Run() {
	defaultTime := time.Time{}
	if a.TimeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(a.Options.PollIntervalSec)*time.Second) <= time.Since(a.TimeState.LastCollectMetricTime) {
		a.CollectMetric()
		a.TimeState.LastCollectMetricTime = time.Now()
	}

	if a.TimeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.Options.ReportIntervalSec)*time.Second) <= time.Since(a.TimeState.LastSendMetricTime) {
		a.SendMetric()
		a.TimeState.LastSendMetricTime = time.Now()
	}
}
