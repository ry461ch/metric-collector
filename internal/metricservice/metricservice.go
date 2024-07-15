package metricservice

import (
	"errors"
	"context"
	"log"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricService struct {
	metricStorage 	storage.Storage
}

func New(metricStorage storage.Storage) *MetricService {
	return &MetricService{metricStorage: metricStorage}
}

func (ms *MetricService) SaveMetrics(ctx context.Context, metricList []metrics.Metrics) error {
	if ms.metricStorage == nil {
		log.Println("im here 1")
		return errors.New("INTERNAL_SERVER_ERROR")
	}
	for _, metric := range metricList {
		if metric.ID == "" {
			return errors.New("INVALID_METRIC_ID")
		}
		if metric.MType == "" {
			return errors.New("INVALID_METRIC_TYPE")
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return errors.New("INVALID_METRIC_VALUE")
			}
			err := ms.metricStorage.UpdateGaugeValue(ctx, metric.ID, *metric.Value)
			if err != nil {
				log.Println("im here 2", err.Error())
				return errors.New("INTERNAL_SERVER_ERROR")
			}
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC_VALUE")
			}
			err := ms.metricStorage.UpdateCounterValue(ctx, metric.ID, *metric.Delta)
			if err != nil {
				log.Println("im here 3", err.Error())
				return errors.New("INTERNAL_SERVER_ERROR")
			}
		default:
			return errors.New("INVALID_METRIC_TYPE")
		}
	}
	return nil
}

func (ms *MetricService) ExtractMetrics(ctx context.Context) ([]metrics.Metrics, error) {
	if ms.metricStorage == nil {
		return nil, errors.New("INTERNAL_SERVER_ERROR")
	}
	metricList := []metrics.Metrics{}
	gaugeValues, err := ms.metricStorage.GetGaugeValues(ctx)
	if err != nil {
		return nil, errors.New("INTERNAL_SERVER_ERROR")
	}
	for metricName, val := range gaugeValues {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &val,
		})
	}
	counterValues, err := ms.metricStorage.GetCounterValues(ctx)
	if err != nil {
		return nil, errors.New("INTERNAL_SERVER_ERROR")
	}
	for metricName, val := range counterValues {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &val,
		})
	}

	return metricList, nil
}

func (ms *MetricService) GetMetric(ctx context.Context, metric *metrics.Metrics) error {
	if ms.metricStorage == nil {
		return errors.New("INTERNAL_SERVER_ERROR")
	}
	switch (metric.MType) {
	case "gauge":
		val, ok, err := ms.metricStorage.GetGaugeValue(ctx, metric.ID)
		if err != nil {
			return errors.New("INTERNAL_SERVER_ERROR")
		}
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Value = &val
	case "counter":
		val, ok, err := ms.metricStorage.GetCounterValue(ctx, metric.ID)
		if err != nil {
			return errors.New("INTERNAL_SERVER_ERROR")
		}
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Delta = &val
	default:
		return errors.New("INVALID_METRIC_TYPE")
	}
	return nil
}

func (ms *MetricService) Ping(ctx context.Context) bool {
	if ms.metricStorage != nil {
		return ms.metricStorage.Ping(ctx)
	}
	return false
}
