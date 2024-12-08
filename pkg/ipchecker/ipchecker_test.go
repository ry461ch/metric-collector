package ipchecker

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	ipchecker := New("127.0.0.1/30")
	reqIP := net.ParseIP("127.0.0.2")
	assert.True(t, ipchecker.Contains(&reqIP))

	invalidIP := net.ParseIP("127.0.1.1")
	assert.False(t, ipchecker.Contains(&invalidIP))
}

func TestInvalidCIDR(t *testing.T) {
	ipchecker := New("invalid")
	assert.Nil(t, ipchecker, "initiated")
}
