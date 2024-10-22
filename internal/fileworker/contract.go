package fileworker

import (
	"context"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

// Хранилще метрик
type Storage interface {
	ExtractMetrics(ctx context.Context) ([]metrics.Metric, error)
	SaveMetrics(ctx context.Context, metricList []metrics.Metric) error
}
