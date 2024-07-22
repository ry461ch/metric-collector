package storage

import (
	"context"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type Storage interface {
	Initialize(ctx context.Context) error
	ExtractMetrics(ctx context.Context) ([]metrics.Metric, error)
	SaveMetrics(ctx context.Context, metricList []metrics.Metric) error
	GetMetric(ctx context.Context, metric *metrics.Metric) error
}

type ExternalStorage interface {
	Storage
	Ping(ctx context.Context) bool
	Close()
}
