package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/net_addr"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type Options struct {
	reportIntervalSec int64
	pollIntervalSec   int64
	addr              netaddr.NetAddress
}

type TimeState struct {
	lastCollectMetricTime time.Time
	lastSendMetricTime    time.Time
}

type MetricAgent struct {
	timeState *TimeState
	options   Options
	mStorage  storage.Storage
}

func (agent *MetricAgent) CollectMetric() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	agent.mStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	agent.mStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	agent.mStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	agent.mStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	agent.mStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	agent.mStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	agent.mStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	agent.mStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	agent.mStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	agent.mStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	agent.mStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	agent.mStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	agent.mStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	agent.mStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	agent.mStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	agent.mStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	agent.mStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	agent.mStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	agent.mStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	agent.mStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	agent.mStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	agent.mStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	agent.mStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	agent.mStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	agent.mStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	agent.mStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	agent.mStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	agent.mStorage.UpdateCounterValue("PollCount", 1)
	agent.mStorage.UpdateGaugeValue("RandomValue", rand.Float64())
}

func (agent *MetricAgent) SendMetric() error {
	serverURL := "http://" + agent.options.addr.Host + ":" + strconv.FormatInt(agent.options.addr.Port, 10)

	client := resty.New()
	for metricName, val := range agent.mStorage.GetGaugeValues() {
		path := "/update/gauge/" + metricName + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	for metricName, val := range agent.mStorage.GetCounterValues() {
		path := "/update/counter/" + metricName + "/" + strconv.FormatInt(val, 10)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			return fmt.Errorf("server broken or timeouted: %s", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("an error occurred in the agent when sending metric %s, server returned %d", metricName, resp.StatusCode())
		}
	}
	return nil
}

func (agent *MetricAgent) Run() {

	defaultTime := time.Time{}
	if agent.timeState.lastCollectMetricTime == defaultTime ||
		time.Duration(time.Duration(agent.options.pollIntervalSec)*time.Second) <= time.Since(agent.timeState.lastCollectMetricTime) {
		agent.CollectMetric()
		agent.timeState.lastCollectMetricTime = time.Now()
	}

	if agent.timeState.lastSendMetricTime == defaultTime ||
		time.Duration(time.Duration(agent.options.reportIntervalSec)*time.Second) <= time.Since(agent.timeState.lastSendMetricTime) {
		agent.SendMetric()
		agent.timeState.lastSendMetricTime = time.Now()
	}
}

func main() {
	options := Options{addr: netaddr.NetAddress{Host: "localhost", Port: 8080}}
	_ = flag.Value(&options.addr)
	flag.Var(&options.addr, "a", "Net address host:port")
	flag.Int64Var(&options.reportIntervalSec, "r", 10, "Interval of sending metrics to the server")
	flag.Int64Var(&options.pollIntervalSec, "p", 2, "Interval of polling metrics from runtime")
	flag.Parse()

	mAgent := MetricAgent{
		mStorage:  &storage.MetricStorage{},
		options:   options,
		timeState: &TimeState{},
	}
	for {
		mAgent.Run()
		time.Sleep(time.Second)
	}
}
