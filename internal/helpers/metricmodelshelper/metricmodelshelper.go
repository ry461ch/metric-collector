package metricmodelshelper

import (
	"errors"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

func SaveMetrics(metricList []metrics.Metrics, mStorage storage) error {
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
			mStorage.UpdateGaugeValue(metric.ID, *metric.Value)
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC_VALUE")
			}
			mStorage.UpdateCounterValue(metric.ID, *metric.Delta)
		default:
			return errors.New("INVALID_METRIC_TYPE")
		}
	}
	return nil
}

func ExtractMetrics(mStorage storage) []metrics.Metrics {
	metricList := []metrics.Metrics{}
	for metricName, val := range mStorage.GetGaugeValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &val,
		})
	}
	for metricName, val := range mStorage.GetCounterValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &val,
		})
	}

	return metricList
}
