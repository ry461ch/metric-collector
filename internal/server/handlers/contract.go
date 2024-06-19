package handlers

type storage interface {
	UpdateGaugeValue(key string, value float64)
	GetGaugeValue(key string) (float64, bool)
	UpdateCounterValue(key string, value int64)
	GetCounterValue(key string) (int64, bool)

	GetGaugeValues() map[string]float64
	GetCounterValues() map[string]int64
}
