package collector

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type Collector struct {
	pollIntervalSec       int64
}

func New(pollIntervalSec int64) *Collector {
	return &Collector{
		pollIntervalSec:       pollIntervalSec,
	}
}

func collectRuntimeMetrics(ctx context.Context, metricChannel chan<- metrics.Metric) {
	log.Println("Trying to collect runtime metrics")
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

	for key, val := range metricGaugeMap {
		metric := metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		}
		select {
		case <-ctx.Done():
			return
		case metricChannel <- metric:
			continue
		}
	}
	for key, val := range metricCounterMap {
		metric := metrics.Metric{
			ID:    key,
			MType: "counter",
			Delta: &val,
		}
		select {
		case <-ctx.Done():
			return
		case metricChannel <- metric:
			continue
		}
	}

	log.Println("Successfully got all runtime metrics")
}

func collectGopsutilMetrics(ctx context.Context, metricChannel chan<- metrics.Metric) {
	log.Println("Trying to collect gopsutil metrics")
	virtualMemory, _ := mem.VirtualMemory()

	metricGaugeMap := map[string]float64{}

	metricGaugeMap["TotalMemory"] = float64(virtualMemory.Total)
	metricGaugeMap["FreeMemory"] = float64(virtualMemory.Free)

	cpuPercent, _ := cpu.Percent(0, true)
	for idx, val := range cpuPercent {
		metricGaugeMap[fmt.Sprintf("CPUutilization%d", idx+1)] = val
	}

	for key, val := range metricGaugeMap {
		metric := metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		}
		select {
		case <-ctx.Done():
			return
		case metricChannel <- metric:
			continue
		}
	}

	log.Println("Successfully got all gopsutil metrics")
}

func collectMetrics(ctx context.Context, metricChannel chan<- metrics.Metric) {
	log.Println("Trying to collect metrics")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		collectRuntimeMetrics(ctx, metricChannel)
		wg.Done()
	}()
	go func() {
		collectGopsutilMetrics(ctx, metricChannel)
		wg.Done()
	}()

	wg.Wait()
	log.Println("Successfully collect all metrics")
}

func (c *Collector) run(ctx context.Context, metricChannel chan<- metrics.Metric) {
	for {
		select {
		case <-ctx.Done():
			log.Println("collector done")
			return
		default:
		}
		collectCtx, collectCtxCancel := context.WithTimeout(ctx, 3*time.Second)
		collectMetrics(collectCtx, metricChannel)
		collectCtxCancel()
		time.Sleep(time.Duration(c.pollIntervalSec)*time.Second)
	}
}

func (c *Collector) CollectMetricsGenerator(ctx context.Context) chan metrics.Metric {
	metricChannel := make(chan metrics.Metric, 10000)

	go func() {
		defer close(metricChannel)
		c.run(ctx, metricChannel)
	}()

	return metricChannel
}
