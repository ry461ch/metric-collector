package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/client"
)

type HttpClient struct {
	url string
}

func (http_client HttpClient) Post(path string) (int64, error) {
	resp, err := http.Post(http_client.url, "text/plain", nil)
	if err != nil {
		return int64(0), fmt.Errorf("server broken or timeouted")
	}
	return int64(resp.StatusCode), nil

}

func CollectMetric(m_storage storage.Storage) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	m_storage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	m_storage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	m_storage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	m_storage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	m_storage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	m_storage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	m_storage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	m_storage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	m_storage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	m_storage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	m_storage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	m_storage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	m_storage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	m_storage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	m_storage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	m_storage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	m_storage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	m_storage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	m_storage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	m_storage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	m_storage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	m_storage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	m_storage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	m_storage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	m_storage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	m_storage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	m_storage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	m_storage.UpdateCounterValue("PollCount", 1)
	m_storage.UpdateGaugeValue("RandomValue", rand.Float64())
}

func SendMetric(m_storage storage.Storage, client client.ServerClient) {
	for metric_name, val := range m_storage.GetGaugeValues() {
		path := "/update/gauge/" + metric_name + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		status_code, err := client.Post(path)
		if err != nil {
			fmt.Printf("server broken or timeouted")
		}
		if status_code != http.StatusOK {
			fmt.Printf("an error occurred in the agent when sending metric %s", metric_name)
		}
	}
	for metric_name, val := range m_storage.GetCounterValues() {
		path := "/update/counter/" + metric_name + "/" + strconv.FormatInt(val, 10)
		status_code, err := client.Post(path)
		if err != nil {
			fmt.Printf("server broken or timeouted")
		}
		if status_code != http.StatusOK {
			fmt.Printf("an error occurred in the agent when sending metric %s", metric_name)
		}
	}
}

func Run(m_storage storage.Storage, client client.ServerClient) {
	CollectMetric(m_storage)

	if m_storage.GetCounterValue("PollCount")%5 == 0 {
		SendMetric(m_storage, client)
	}

}

func main() {
	server_url := "http://localhost:8080"
	internal_storage := storage.MetricStorage{}
	for {
		Run(&internal_storage, HttpClient{url: server_url})
		time.Sleep(2 * time.Second)
	}
}
