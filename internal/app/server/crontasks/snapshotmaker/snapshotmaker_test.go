package snapshotmaker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func TestBase(t *testing.T) {
	logging.Initialize("DEBUG")
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

	currentTime := time.Now()
	filepath := "/tmp/metrics-db.json"
	config := config.Config{StoreInterval: 5, FileStoragePath: filepath}
	timeState := TimeState{LastSnapshotTime: currentTime}
	fileWorker := fileworker.New(filepath, mReadStorage)
	snapshotMaker := New(&timeState, &config, fileWorker)

	snapshotMaker.runIteration(context.TODO())
	assert.Equal(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Сработал if, хотя еще не время")

	timeState.LastSnapshotTime = time.Now().Add(-time.Second * 6)
	snapshotMaker.runIteration(context.TODO())
	assert.NotEqual(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Не сработал if, хотя должен был")

	mWriteStorage := memstorage.NewMemStorage(context.TODO())
	fileWriteWorker := fileworker.New(filepath, mWriteStorage)
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
