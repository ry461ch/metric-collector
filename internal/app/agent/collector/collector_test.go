package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

func TestCollectMetric(t *testing.T) {
	metricChannel := make(chan metrics.Metric, 100)
	defer close(metricChannel)

	collectMetrics(context.TODO(), metricChannel)

	collectedMetrics := 0
	for {
		select {
		case <-metricChannel:
			collectedMetrics++
		default:
			assert.LessOrEqual(t, 32, collectedMetrics, "Несовпадает количество отслеживаемых метрик")
			return
		}
	}
}
