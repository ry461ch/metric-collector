package fileworker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	memstorage "github.com/ry461ch/metric-collector/internal/storage/memory"
)

func TestBase(t *testing.T) {
	mReadStorage := memstorage.New()
	mReadStorage.Initialize(context.TODO())

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

	mWriteStorage := memstorage.New()
	mWriteStorage.Initialize(context.TODO())
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

type InvalidStorage struct{}

func (is *InvalidStorage) ExtractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	return nil, errors.New("")
}
func (is *InvalidStorage) SaveMetrics(ctx context.Context, metricList []metrics.Metric) error {
	return errors.New("")
}
func (is *InvalidStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	return errors.New("")
}

func TestInvalid(t *testing.T) {
	mInvalidStorage := &InvalidStorage{}

	filePath := "/tmp/invalid.json"

	fileReadWorker := New(filePath, mInvalidStorage)
	err := fileReadWorker.ImportToFile(context.TODO())
	assert.Error(t, err, "Expected error")

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

	file, _ := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	data, _ := json.Marshal(metricList)
	file.Write(data)

	fileWriteWorker := New(filePath, mInvalidStorage)
	err = fileWriteWorker.ExportFromFile(context.TODO())
	assert.Error(t, err, "Expected error")

	file.Write([]byte("invalid"))
	err = fileWriteWorker.ExportFromFile(context.TODO())
	assert.Error(t, err, "Expected error")
}
