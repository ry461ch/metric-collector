package serverconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

func TestBase(t *testing.T) {
	cfg := New()
	assert.Equal(t, cfg.Addr, netaddr.NetAddress{Host: "localhost", Port: 8080})
	assert.Equal(t, cfg.LogLevel, "INFO")
	assert.Equal(t, cfg.StoreInterval, int64(10))
}
