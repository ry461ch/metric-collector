package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaugeValues(t *testing.T) {
    storage := MetricStorage{}

	storage.UpdateGaugeValue("test", 10.0)
	storage.UpdateGaugeValue("test", 12.0)

	assert.Equal(t, float64(12.0), storage.GetGaugeValue("test"), "неверно обновляется gauge метрика")

	storage.UpdateGaugeValue("test_2", 11.6)
	storage.UpdateGaugeValue("test_3", 13.5)

	expectedGaugeValues := map[string]float64{
		"test": 12.0,
		"test_2": 11.6,
		"test_3": 13.5,
	}

	gaugeValues := storage.GetGaugeValues()
	assert.Equal(t, 3, len(gaugeValues), "Кол-во метрик типа gauge не совпадает с ожидаемым")
	for key, val := range storage.GetGaugeValues() {
		assert.Equal(t, float64(expectedGaugeValues[key]), val, "Значение метрики типа gauge не совпадает с ожидаемым")
	}
}

func TestCounterValues(t *testing.T) {
    storage := MetricStorage{}

	storage.UpdateCounterValue("test", 10)
	storage.UpdateCounterValue("test", 12)

	assert.Equal(t, int64(22), storage.GetCounterValue("test"), "неверно обновляется counter метрика")

	storage.UpdateCounterValue("test_2", 11)
	storage.UpdateCounterValue("test_3", 13)

	expectedCounterValues := map[string]int64{
		"test": 22,
		"test_2": 11,
		"test_3": 13,
	}

	counterValues := storage.GetCounterValues()
	assert.Equal(t, 3, len(counterValues), "Кол-во метрик типа counter не совпадает с ожидаемым")
	for key, val := range storage.GetCounterValues() {
		assert.Equal(t, int64(expectedCounterValues[key]), val, "Значение метрики типа counter не совпадает с ожидаемым")
	}
}
