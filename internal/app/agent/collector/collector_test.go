package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

func TestCollectMetric(t *testing.T) {
	metricsCollector := New(2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	metricChannel := metricsCollector.CollectMetricsGenerator(ctx)
	time.Sleep(time.Second)

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

func BenchmarkCollectMetric(b *testing.B) {
	metricChannel := make(chan metrics.Metric, 100)
	defer close(metricChannel)

	for i := 0; i < b.N; i++ {
		collectMetrics(context.TODO(), metricChannel)

		b.StopTimer()
		collectedMetrics := 0
		for shouldStop := false; !shouldStop; {
			select {
			case <-metricChannel:
				collectedMetrics++
			default:
				shouldStop = true
			}
		}
		b.StartTimer()
	}
}
