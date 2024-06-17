package metric_storage

type MetricStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func (storage *MetricStorage) UpdateGaugeValue(key string, value float64) {
	if storage.gauge == nil {
		storage.gauge = map[string]float64{}
	}
	storage.gauge[key] = value
}

func (storage *MetricStorage) GetGaugeValue(key string) (float64, bool) {
	if storage.gauge == nil {
		return 0, false
	}
	val, ok := storage.gauge[key]
	return val, ok
}

func (storage *MetricStorage) UpdateCounterValue(key string, value int64) {
	if storage.counter == nil {
		storage.counter = map[string]int64{}
	}
	storage.counter[key] += value
}

func (storage *MetricStorage) GetCounterValue(key string) (int64, bool) {
	if storage.counter == nil {
		return 0, false
	}
	val, ok := storage.counter[key]
	return val, ok
}

func (storage *MetricStorage) GetGaugeValues() map[string]float64 {
	if storage.gauge == nil {
		return map[string]float64{}
	}
	return storage.gauge
}

func (storage *MetricStorage) GetCounterValues() map[string]int64 {
	if storage.counter == nil {
		return map[string]int64{}
	}
	return storage.counter
}
