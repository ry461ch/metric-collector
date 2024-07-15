package memstorage

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"
)

func TestGaugeValues(t *testing.T) {
	storage := MemStorage{}

	storage.UpdateGaugeValue(context.TODO(),"test", 10.0)
	storage.UpdateGaugeValue(context.TODO(),"test", 12.0)

	val, _, _ := storage.GetGaugeValue(context.TODO(), "test")
	assert.Equal(t, float64(12.0), val, "неверно обновляется gauge метрика")

	storage.UpdateGaugeValue(context.TODO(),"test_2", 11.6)
	storage.UpdateGaugeValue(context.TODO(),"test_3", 13.5)

	_, ok, _ := storage.GetGaugeValue(context.TODO(),"unknown")
	assert.False(t, ok)

	expectedGaugeValues := map[string]float64{
		"test":   12.0,
		"test_2": 11.6,
		"test_3": 13.5,
	}

	gaugeValues, _ := storage.GetGaugeValues(context.TODO())
	assert.Equal(t, 3, len(gaugeValues), "Кол-во метрик типа gauge не совпадает с ожидаемым")
	for key, val := range gaugeValues {
		assert.Equal(t, float64(expectedGaugeValues[key]), val, "Значение метрики типа gauge не совпадает с ожидаемым")
	}
}

func TestCounterValues(t *testing.T) {
	storage := MemStorage{}

	storage.UpdateCounterValue(context.TODO(), "test", 10)
	storage.UpdateCounterValue(context.TODO(), "test", 12)

	val, _, _ := storage.GetCounterValue(context.TODO(), "test")
	assert.Equal(t, int64(22), val, "неверно обновляется counter метрика")

	storage.UpdateCounterValue(context.TODO(), "test_2", 11)
	storage.UpdateCounterValue(context.TODO(), "test_3", 13)

	_, ok, _ := storage.GetCounterValue(context.TODO(), "unknown")
	assert.False(t, ok)

	expectedCounterValues := map[string]int64{
		"test":   22,
		"test_2": 11,
		"test_3": 13,
	}

	counterValues, _ := storage.GetCounterValues(context.TODO())
	assert.Equal(t, 3, len(counterValues), "Кол-во метрик типа counter не совпадает с ожидаемым")
	for key, val := range counterValues {
		assert.Equal(t, int64(expectedCounterValues[key]), val, "Значение метрики типа counter не совпадает с ожидаемым")
	}
}
