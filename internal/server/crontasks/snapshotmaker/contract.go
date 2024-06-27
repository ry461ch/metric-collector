package snapshotmaker

type storage interface {
	UpdateGaugeValue(key string, value float64)
	UpdateCounterValue(key string, value int64)

	GetGaugeValues() map[string]float64
	GetCounterValues() map[string]int64
}
