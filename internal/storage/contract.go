package storage

import (
	"context"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type Storage interface {
	ExtractMetrics(ctx context.Context) ([]metrics.Metric, error)
	SaveMetrics(ctx context.Context, metricList []metrics.Metric) error
	GetMetric(ctx context.Context, metric *metrics.Metric) error
	Ping(ctx context.Context) bool
	Close()
}
