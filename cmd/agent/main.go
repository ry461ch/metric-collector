package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/storage"
)


func CollectMetric(mStorage storage.Storage) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	mStorage.UpdateGaugeValue("Alloc", float64(rtm.Alloc))
	mStorage.UpdateGaugeValue("BuckHashSys", float64(rtm.BuckHashSys))
	mStorage.UpdateGaugeValue("Frees", float64(rtm.Frees))
	mStorage.UpdateGaugeValue("GCCPUFraction", float64(rtm.GCCPUFraction))
	mStorage.UpdateGaugeValue("GCSys", float64(rtm.GCSys))
	mStorage.UpdateGaugeValue("HeapAlloc", float64(rtm.HeapAlloc))
	mStorage.UpdateGaugeValue("HeapIdle", float64(rtm.HeapIdle))
	mStorage.UpdateGaugeValue("HeapInuse", float64(rtm.HeapInuse))
	mStorage.UpdateGaugeValue("HeapObjects", float64(rtm.HeapObjects))
	mStorage.UpdateGaugeValue("HeapReleased", float64(rtm.HeapReleased))
	mStorage.UpdateGaugeValue("HeapSys", float64(rtm.HeapSys))
	mStorage.UpdateGaugeValue("LastGC", float64(rtm.LastGC))
	mStorage.UpdateGaugeValue("Lookups", float64(rtm.Lookups))
	mStorage.UpdateGaugeValue("MCacheInuse", float64(rtm.MCacheInuse))
	mStorage.UpdateGaugeValue("MCacheSys", float64(rtm.MCacheSys))
	mStorage.UpdateGaugeValue("MSpanInuse", float64(rtm.MSpanInuse))
	mStorage.UpdateGaugeValue("MSpanSys", float64(rtm.MSpanSys))
	mStorage.UpdateGaugeValue("Mallocs", float64(rtm.Mallocs))
	mStorage.UpdateGaugeValue("NextGC", float64(rtm.NextGC))
	mStorage.UpdateGaugeValue("NumForcedGC", float64(rtm.NumForcedGC))
	mStorage.UpdateGaugeValue("NumGC", float64(rtm.NumGC))
	mStorage.UpdateGaugeValue("OtherSys", float64(rtm.OtherSys))
	mStorage.UpdateGaugeValue("PauseTotalNs", float64(rtm.PauseTotalNs))
	mStorage.UpdateGaugeValue("StackInuse", float64(rtm.StackInuse))
	mStorage.UpdateGaugeValue("StackSys", float64(rtm.StackSys))
	mStorage.UpdateGaugeValue("Sys", float64(rtm.Sys))
	mStorage.UpdateGaugeValue("TotalAlloc", float64(rtm.TotalAlloc))

	mStorage.UpdateCounterValue("PollCount", 1)
	mStorage.UpdateGaugeValue("RandomValue", rand.Float64())
}

func SendMetric(mStorage storage.Storage, serverURL string) {
	client := resty.New()
	for metricName, val := range mStorage.GetGaugeValues() {
		path := "/update/gauge/" + metricName + "/" + strconv.FormatFloat(val, 'f', -1, 64)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			fmt.Printf("server broken or timeouted: %s\n", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			fmt.Printf("an error occurred in the agent when sending metric %s, server returned %d\n", metricName, resp.StatusCode())
		}
	}
	for metricName, val := range mStorage.GetCounterValues() {
		path := "/update/counter/" + metricName + "/" + strconv.FormatInt(val, 10)
		resp, err := client.R().Post(serverURL + path)
		if err != nil {
			fmt.Printf("server broken or timeouted: %s\n", err.Error())
		}
		if resp.StatusCode() != http.StatusOK {
			fmt.Printf("an error occurred in the agent when sending metric %s, server returned %d\n", metricName, resp.StatusCode())
		}
	}
}

func Run(mStorage storage.Storage, serverURL string) {
	CollectMetric(mStorage)

	if mStorage.GetCounterValue("PollCount")%5 == 0 {
		SendMetric(mStorage, serverURL)
	}

}

func main() {
	serverURL := "http://localhost:8080"
	internalStorage := storage.MetricStorage{}
	for {
		Run(&internalStorage, serverURL)
		time.Sleep(2 * time.Second)
	}
}
