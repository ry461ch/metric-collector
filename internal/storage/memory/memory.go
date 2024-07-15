package memstorage

import (
	"context"
	"errors"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}


func NewMemStorage(ctx context.Context) *MemStorage {
	return &MemStorage{counter: map[string]int64{}, gauge: map[string]float64{}}
}

func (ms *MemStorage) Ping(ctx context.Context) bool {
	return true
}

func (ms *MemStorage) Close() {}


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
			ms.gauge[metric.ID] = *metric.Value
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC")
			}
			ms.counter[metric.ID] += *metric.Delta
		default:
			return errors.New("INVALID_METRIC")
		}
	}

	return nil
}

func (ms *MemStorage) ExtractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	metricList := []metrics.Metric{}
	for key, val := range ms.gauge {
        metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
    }
	for key, val := range ms.counter {
        metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "counter",
			Delta: &val,
		})
    }
	return metricList, nil
}

func (ms *MemStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	switch (metric.MType) {
	case "gauge":
		val, ok := ms.gauge[metric.ID]
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Value = &val
	case "counter":
		val, ok := ms.counter[metric.ID]
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Delta = &val
	default:
		return errors.New("INVALID_METRIC_TYPE")
	}
	return nil
}
