package fileworker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func TestBase(t *testing.T) {
	mReadStorage := memstorage.NewMemStorage(context.TODO())

	mCounterValue := int64(10)
	mGaugeValue := float64(12.0)
	metricList := []metrics.Metric{
		{
			ID:    "test",
			MType: "counter",
			Delta: &mCounterValue,
		},
		{
			ID:    "test",
			MType: "gauge",
			Value: &mGaugeValue,
		},
	}
	mReadStorage.SaveMetrics(context.TODO(), metricList)

	filePath := "/tmp/metric_file_helper.json"
	fileReadWorker := New(filePath, mReadStorage)
	fileReadWorker.ImportToFile(context.TODO())

	mWriteStorage := memstorage.NewMemStorage(context.TODO())
	fileWriteWorker := New(filePath, mWriteStorage)
	fileWriteWorker.ExportFromFile(context.TODO())

	mSearchGauge := metrics.Metric{
		ID:    "test",
		MType: "gauge",
	}
	mSearchCounter := metrics.Metric{
		ID:    "test",
		MType: "counter",
	}
	mWriteStorage.GetMetric(context.TODO(), &mSearchCounter)
	assert.Equal(t, int64(10), *mSearchCounter.Delta, "counter not equal")
	mWriteStorage.GetMetric(context.TODO(), &mSearchGauge)
	assert.Equal(t, float64(12.0), *mSearchGauge.Value, "gauge not equal")
}
