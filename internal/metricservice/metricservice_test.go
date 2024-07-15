package metricservice

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func TestBase(t *testing.T) {
	mReadStorage := memstorage.MemStorage{}
	mReadStorage.UpdateCounterValue(context.TODO(), "test_1", 6)
	mReadStorage.UpdateGaugeValue(context.TODO(), "test_2", 5.5)
	mReadService := MetricService{metricStorage: &mReadStorage}

	metricList, _ := mReadService.ExtractMetrics(context.TODO())

	mWriteStorage := memstorage.MemStorage{}
	mWriteService := MetricService{metricStorage: &mWriteStorage}
	mWriteService.SaveMetrics(context.Background(), metricList)

	counterVal, _, _ := mWriteStorage.GetCounterValue(context.TODO(), "test_1")
	assert.Equal(t, int64(6), counterVal, "counter not equal")
	gaugeVal, _, _ := mWriteStorage.GetGaugeValue(context.TODO(), "test_2")
	assert.Equal(t, float64(5.5), gaugeVal, "gauge not equal")
}
