package metricservice

import (
	"errors"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricService struct {
	metricStorage 	storage.Storage
}

func New(metricStorage storage.Storage) *MetricService {
	return &MetricService{metricStorage: metricStorage}
}

func (ms *MetricService) SaveMetrics(metricList []metrics.Metrics) error {
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
			ms.metricStorage.UpdateGaugeValue(metric.ID, *metric.Value)
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC_VALUE")
			}
			ms.metricStorage.UpdateCounterValue(metric.ID, *metric.Delta)
		default:
			return errors.New("INVALID_METRIC_TYPE")
		}
	}
	return nil
}

func (ms *MetricService) ExtractMetrics() []metrics.Metrics {
	metricList := []metrics.Metrics{}
	for metricName, val := range ms.metricStorage.GetGaugeValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &val,
		})
	}
	for metricName, val := range ms.metricStorage.GetCounterValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &val,
		})
	}

	return metricList
}

func (ms *MetricService) GetMetric(metric *metrics.Metrics) error {
	switch (metric.MType) {
	case "gauge":
		val, ok := ms.metricStorage.GetGaugeValue(metric.ID)
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Value = &val
	case "counter":
		val, ok := ms.metricStorage.GetCounterValue(metric.ID)
		if !ok {
			return errors.New("NOT_FOUND")
		}
		metric.Delta = &val
	default:
		return errors.New("INVALID_METRIC_TYPE")
	}
	return nil
}
