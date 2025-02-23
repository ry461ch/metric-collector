package metrics

// Main struct for metric representation
type Metric struct {
	ID    string   `json:"id"`                         // имя метрики
	MType string   `json:"type" enums:"counter,gauge"` // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"`            // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"`            // значение метрики в случае передачи gauge
}
