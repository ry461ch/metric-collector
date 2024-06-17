package memstorage

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func (ms *MemStorage) UpdateGaugeValue(key string, value float64) {
	if ms.gauge == nil {
		ms.gauge = map[string]float64{}
	}
	ms.gauge[key] = value
}

func (ms *MemStorage) GetGaugeValue(key string) (float64, bool) {
	if ms.gauge == nil {
		return 0, false
	}
	val, ok := ms.gauge[key]
	return val, ok
}

func (ms *MemStorage) UpdateCounterValue(key string, value int64) {
	if ms.counter == nil {
		ms.counter = map[string]int64{}
	}
	ms.counter[key] += value
}

func (ms *MemStorage) GetCounterValue(key string) (int64, bool) {
	if ms.counter == nil {
		return 0, false
	}
	val, ok := ms.counter[key]
	return val, ok
}

func (ms *MemStorage) GetGaugeValues() map[string]float64 {
	if ms.gauge == nil {
		return map[string]float64{}
	}
	return ms.gauge
}

func (ms *MemStorage) GetCounterValues() map[string]int64 {
	if ms.counter == nil {
		return map[string]int64{}
	}
	return ms.counter
}
