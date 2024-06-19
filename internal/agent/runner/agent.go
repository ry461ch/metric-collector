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
)

type Agent struct {
	timeState *config.TimeState
	options   config.Options
	mStorage  storage
}

func NewAgent(timeState *config.TimeState, options config.Options, mStorage storage) Agent {
	return Agent{timeState: timeState, options: options, mStorage: mStorage}
}

func (a *Agent) collectMetric() {
	log.Println("Trying to collect metrics")
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	a.mStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	a.mStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	a.mStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	a.mStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	a.mStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	a.mStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	a.mStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	a.mStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	a.mStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	a.mStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	a.mStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	a.mStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	a.mStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	a.mStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	a.mStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	a.mStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	a.mStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	a.mStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	a.mStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	a.mStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	a.mStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	a.mStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	a.mStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	a.mStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	a.mStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	a.mStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	a.mStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	a.mStorage.UpdateCounterValue("PollCount", 1)
	a.mStorage.UpdateGaugeValue("RandomValue", rand.Float64())
	log.Println("Successfully got all metrics")
}

func (a *Agent) sendMetric() error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + a.options.Addr.Host + ":" + strconv.FormatInt(a.options.Addr.Port, 10)

	client := resty.New()
	for metricName, val := range a.mStorage.GetGaugeValues() {
		path := "/update/gauge/" + metricName + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	for metricName, val := range a.mStorage.GetCounterValues() {
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

func (a* Agent) runIteration() {
	defaultTime := time.Time{}
	if a.timeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(a.options.PollIntervalSec)*time.Second) <= time.Since(a.timeState.LastCollectMetricTime) {
		a.collectMetric()
		a.timeState.LastCollectMetricTime = time.Now()
	}

	if a.timeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(a.options.ReportIntervalSec)*time.Second) <= time.Since(a.timeState.LastSendMetricTime) {
		a.sendMetric()
		a.timeState.LastSendMetricTime = time.Now()
	}
}

func (a *Agent) Run() {
	for {
		a.runIteration()
		time.Sleep(time.Second)
	}
}
