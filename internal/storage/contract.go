package storage

import "context"

type Storage interface {
	UpdateGaugeValue(ctx context.Context, key string, value float64) error
	GetGaugeValue(ctx context.Context, key string) (float64, bool, error)
	UpdateCounterValue(ctx context.Context, key string, value int64) error
	GetCounterValue(ctx context.Context, key string) (int64, bool, error)

	GetGaugeValues(ctx context.Context) (map[string]float64, error)
	GetCounterValues(ctx context.Context) (map[string]int64, error)
	Ping(ctx context.Context) bool
	Close()
}
