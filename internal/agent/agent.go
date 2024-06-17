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

	"github.com/ry461ch/metric-collector/internal/models/agent_models"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricAgent struct {
	TimeState *agent_models.TimeState
	Options   agent_models.Options
	MStorage  storage.Storage
}

func (MA *MetricAgent) CollectMetric() {
	log.Println("Trying to collect metrics")
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	MA.MStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	MA.MStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	MA.MStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	MA.MStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	MA.MStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	MA.MStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	MA.MStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	MA.MStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	MA.MStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	MA.MStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	MA.MStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	MA.MStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	MA.MStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	MA.MStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	MA.MStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	MA.MStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	MA.MStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	MA.MStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	MA.MStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	MA.MStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	MA.MStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	MA.MStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	MA.MStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	MA.MStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	MA.MStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	MA.MStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	MA.MStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	MA.MStorage.UpdateCounterValue("PollCount", 1)
	MA.MStorage.UpdateGaugeValue("RandomValue", rand.Float64())
	log.Println("Successfully got all metrics")
}

func (MA *MetricAgent) SendMetric() error {
	log.Println("Trying to send metrics")
	serverURL := "http://" + MA.Options.Addr.Host + ":" + strconv.FormatInt(MA.Options.Addr.Port, 10)

	client := resty.New()
	for metricName, val := range MA.MStorage.GetGaugeValues() {
		path := "/update/gauge/" + metricName + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	for metricName, val := range MA.MStorage.GetCounterValues() {
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

func (MA *MetricAgent) Run() {

	defaultTime := time.Time{}
	if MA.TimeState.LastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(MA.Options.PollIntervalSec)*time.Second) <= time.Since(MA.TimeState.LastCollectMetricTime) {
		MA.CollectMetric()
		MA.TimeState.LastCollectMetricTime = time.Now()
	}

	if MA.TimeState.LastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(MA.Options.ReportIntervalSec)*time.Second) <= time.Since(MA.TimeState.LastSendMetricTime) {
		MA.SendMetric()
		MA.TimeState.LastSendMetricTime = time.Now()
	}
}
