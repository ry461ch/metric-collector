package memstorage

import "context"

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}

func (ms *MemStorage) UpdateGaugeValue(ctx context.Context, key string, value float64) error {
	if ms.gauge == nil {
		ms.gauge = map[string]float64{}
	}
	ms.gauge[key] = value
	return nil
}

func (ms *MemStorage) GetGaugeValue(ctx context.Context, key string) (float64, bool, error) {
	if ms.gauge == nil {
		return 0, false, nil
	}
	val, ok := ms.gauge[key]
	return val, ok, nil
}

func (ms *MemStorage) UpdateCounterValue(ctx context.Context, key string, value int64) error {
	if ms.counter == nil {
		ms.counter = map[string]int64{}
	}
	ms.counter[key] += value
	return nil
}

func (ms *MemStorage) GetCounterValue(ctx context.Context, key string) (int64, bool, error) {
	if ms.counter == nil {
		return 0, false, nil
	}
	val, ok := ms.counter[key]
	return val, ok, nil
}

func (ms *MemStorage) GetGaugeValues(ctx context.Context) (map[string]float64, error) {
	if ms.gauge == nil {
		return map[string]float64{}, nil
	}
	return ms.gauge, nil
}

func (ms *MemStorage) GetCounterValues(ctx context.Context) (map[string]int64, error) {
	if ms.counter == nil {
		return map[string]int64{}, nil
	}
	return ms.counter, nil
}

func (ms *MemStorage) Ping(ctx context.Context) bool {
	return true
}

func (ms *MemStorage) Close() {}
