package snapshotmaker

import (
	"testing"
	"time"
	"context"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func TestBase(t *testing.T) {
	logging.Initialize("DEBUG")
	mReadStorage := memstorage.MemStorage{}
	mReadService := metricservice.New(&mReadStorage)
	mReadStorage.UpdateCounterValue(context.TODO(), "test_1", 6)
	mReadStorage.UpdateGaugeValue(context.TODO(), "test_2", 5.5)

	currentTime := time.Now()
	filepath := "/tmp/metrics-db.json"
	config := config.Config{StoreInterval: 5, FileStoragePath: filepath}
	timeState := TimeState{LastSnapshotTime: currentTime}
	fileWorker := fileworker.New(filepath, mReadService)
	snapshotMaker := New(&timeState, &config, fileWorker)

	snapshotMaker.runIteration(context.TODO())
	assert.Equal(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Сработал if, хотя еще не время")

	timeState.LastSnapshotTime = time.Now().Add(-time.Second * 6)
	snapshotMaker.runIteration(context.TODO())
	assert.NotEqual(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Не сработал if, хотя должен был")

	mWriteStorage := memstorage.MemStorage{}
	mWriteService := metricservice.New(&mWriteStorage)
	fileWriteWorker := fileworker.New(filepath, mWriteService)
	fileWriteWorker.ExportFromFile(context.TODO())

	counterVal, _, _ := mWriteStorage.GetCounterValue(context.TODO(), "test_1")
	assert.Equal(t, int64(6), counterVal, "Несовпадают значения counter")
	gaugeVal, _, _ := mWriteStorage.GetGaugeValue(context.TODO(), "test_2")
	assert.Equal(t, float64(5.5), gaugeVal, "Несовпадают значения gauge")
}
