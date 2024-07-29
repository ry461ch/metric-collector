package memstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type MemStorage struct {
	counterMutex sync.RWMutex
	counter      map[string]int64
	gaugeMutex   sync.RWMutex
	gauge        map[string]float64
}

func New() *MemStorage {
	return &MemStorage{counter: nil, gauge: nil}
}

func (ms *MemStorage) Initialize(ctx context.Context) error {
	ms.counter = map[string]int64{}
	ms.gauge = map[string]float64{}
	return nil
}

func (ms *MemStorage) SaveMetrics(ctx context.Context, metricList []metrics.Metric) error {
	// prepare arrays
	for _, metric := range metricList {
		if metric.ID == "" {
			return errors.New("INVALID_METRIC")
		}
		if metric.MType == "" {
			return errors.New("INVALID_METRIC")
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return errors.New("INVALID_METRIC")
			}
			ms.gaugeMutex.Lock()
			ms.gauge[metric.ID] = *metric.Value
			ms.gaugeMutex.Unlock()
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC")
			}
			ms.counterMutex.Lock()
			ms.counter[metric.ID] += *metric.Delta
			ms.counterMutex.Unlock()
		default:
			return errors.New("INVALID_METRIC")
		}
	}

	return nil
}

func (ms *MemStorage) ExtractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	metricList := []metrics.Metric{}

	ms.gaugeMutex.RLock()
	for key, val := range ms.gauge {
		metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
	}
	ms.gaugeMutex.RUnlock()

	ms.counterMutex.RLock()
	for key, val := range ms.counter {
		metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "counter",
			Delta: &val,
		})
	}
	ms.counterMutex.RUnlock()

	return metricList, nil
}

func (ms *MemStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	switch metric.MType {
	case "gauge":
		ms.gaugeMutex.RLock()
		val, ok := ms.gauge[metric.ID]
		ms.gaugeMutex.RUnlock()
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Value = &val
	case "counter":
		ms.counterMutex.RLock()
		val, ok := ms.counter[metric.ID]
		ms.counterMutex.RUnlock()
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Delta = &val
	default:
		return errors.New("INVALID_METRIC_TYPE")
	}
	return nil
}
