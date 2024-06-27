package snapshotmaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/helpers/metricfilehelper"
	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
	"github.com/ry461ch/metric-collector/internal/server/logger"
)

func TestBase(t *testing.T) {
	slogger.TestInitialize()
	mReadStorage := memstorage.MemStorage{}
	mReadStorage.UpdateCounterValue("test_1", 6)
	mReadStorage.UpdateGaugeValue("test_2", 5.5)

	currentTime := time.Now()
	filepath := "/tmp/metrics-db.json"
	options := config.Options{StoreInterval: 5, FileStoragePath: filepath}
	timeState := TimeState{LastSnapshotTime: currentTime}
	snapshotMaker := SnapshotMaker{
		timeState: &timeState,
		options:   options,
		mStorage:  &mReadStorage,
	}

	snapshotMaker.runIteration()
	assert.Equal(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Сработал if, хотя еще не время")

	timeState.LastSnapshotTime = time.Now().Add(-time.Second * 6)
	snapshotMaker.runIteration()
	assert.NotEqual(t, currentTime, snapshotMaker.timeState.LastSnapshotTime, "Не сработал if, хотя должен был")

	mWriteStorage := memstorage.MemStorage{}
	metricfilehelper.SaveToStorage(filepath, &mWriteStorage)
	counterVal, _ := mWriteStorage.GetCounterValue("test_1")
	assert.Equal(t, int64(6), counterVal, "Несовпадают значения counter")
	gaugeVal, _ := mWriteStorage.GetGaugeValue("test_2")
	assert.Equal(t, float64(5.5), gaugeVal, "Несовпадают значения gauge")
}
