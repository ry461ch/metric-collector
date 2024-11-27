package agentconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

func TestBase(t *testing.T) {
	cfg := New()
	assert.Equal(t, cfg.Addr, netaddr.NetAddress{Host: "localhost", Port: 8080})
	assert.Equal(t, cfg.PollIntervalSec, int64(2))
	assert.Equal(t, cfg.ReportIntervalSec, int64(10))
}
