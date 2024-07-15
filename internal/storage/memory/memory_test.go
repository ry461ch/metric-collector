package memstorage

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

func TestGauge(t *testing.T) {
	storage := NewMemStorage(context.TODO())

	mValue := 10.0
	metricList := []metrics.Metric{
		{
			ID: "test",
			MType: "gauge",
			Value: &mValue,
		},
	}
	storage.SaveMetrics(context.TODO(), metricList)
	mNewValue := 12.0
	metricList = []metrics.Metric{
		{
			ID: "test",
			MType: "gauge",
			Value: &mNewValue,
		},
	}
	storage.SaveMetrics(context.TODO(), metricList)

	searchMetric := metrics.Metric{
		ID: "test",
		MType: "gauge",
	}
	storage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, float64(12.0), *searchMetric.Value, "неверно обновляется gauge метрика")
	
	notExistsMetric := metrics.Metric{
		ID: "unknown",
		MType: "gauge",
	}
	err := storage.GetMetric(context.TODO(), &notExistsMetric)

	assert.Error(t, err)
}

func TestCounter(t *testing.T) {
	storage := NewMemStorage(context.TODO())

	mValue := int64(10)
	metricList := []metrics.Metric{
		{
			ID: "test",
			MType: "counter",
			Delta: &mValue,
		},
	}
	storage.SaveMetrics(context.TODO(), metricList)
	mNewValue := int64(12)
	metricList = []metrics.Metric{
		{
			ID: "test",
			MType: "counter",
			Delta: &mNewValue,
		},
	}
	storage.SaveMetrics(context.TODO(), metricList)

	searchMetric := metrics.Metric{
		ID: "test",
		MType: "counter",
	}
	storage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, int64(22), *searchMetric.Delta, "неверно обновляется counter метрика")
	
	notExistsMetric := metrics.Metric{
		ID: "unknown",
		MType: "counter",
	}
	err := storage.GetMetric(context.TODO(), &notExistsMetric)

	assert.Error(t, err)
}

func TestExtractAll(t *testing.T) {
	storage := NewMemStorage(context.TODO())

	mCounterValue := int64(10)
	mGaugeValue := float64(12.0)
	metricList := []metrics.Metric{
		{
			ID: "test",
			MType: "counter",
			Delta: &mCounterValue,
		},
		{
			ID: "test",
			MType: "gauge",
			Value: &mGaugeValue,
		},
	}
	storage.SaveMetrics(context.TODO(), metricList)

	resultMetrics, _ := storage.ExtractMetrics(context.TODO())

	assert.Equal(t, 2, len(resultMetrics), "Кол-во метрик не совпадает с ожидаемым")
}
