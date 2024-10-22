package server

import (
	"context"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

// Storage - интерфейс для хранилища метрик
type Storage interface {
	Initialize(ctx context.Context) error
	ExtractMetrics(ctx context.Context) ([]metrics.Metric, error)
	SaveMetrics(ctx context.Context, metricList []metrics.Metric) error
	GetMetric(ctx context.Context, metric *metrics.Metric) error
}

// ExternalStorage для удаленного хранилища метрик + функцональность доступности хранлища
type ExternalStorage interface {
	Storage
	Ping(ctx context.Context) bool
	Close()
}
