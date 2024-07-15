package fileworker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func TestBase(t *testing.T) {
	mReadStorage := memstorage.MemStorage{}
	mReadService := metricservice.New(&mReadStorage)
	mReadStorage.UpdateCounterValue("test_1", 6)
	mReadStorage.UpdateGaugeValue("test_2", 5.5)

	filePath := "/tmp/metric_file_helper.json"
	fileReadWorker := New(filePath, mReadService)
	fileReadWorker.ImportToFile()

	mWriteStorage := memstorage.MemStorage{}
	mWriteService := metricservice.New(&mWriteStorage)
	fileWriteWorker := New(filePath, mWriteService)
	fileWriteWorker.ExportFromFile()

	counterVal, _ := mWriteStorage.GetCounterValue("test_1")
	assert.Equal(t, int64(6), counterVal, "counter not equal")
	gaugeVal, _ := mWriteStorage.GetGaugeValue("test_2")
	assert.Equal(t, float64(5.5), gaugeVal, "gauge not equal")
}
