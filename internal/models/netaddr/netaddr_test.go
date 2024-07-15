package netaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetAddrBase(t *testing.T) {
	testCases := []struct {
		testName     string
		input        string
		expectedAddr *NetAddress
	}{
		{testName: "default net addr", input: "", expectedAddr: &NetAddress{Host: "localhost", Port: 8080}},
		{testName: "default port", input: "1.2.3.4", expectedAddr: &NetAddress{Host: "1.2.3.4", Port: 8080}},
		{testName: "valid custom net addr", input: "1.2.3.4:1234", expectedAddr: &NetAddress{Host: "1.2.3.4", Port: 1234}},
		{testName: "too much :", input: "1.2.3.4:1234:1234", expectedAddr: nil},
		{testName: "invalid port", input: "1.2.3.4:invalid", expectedAddr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			resAddr := NetAddress{Host: "localhost", Port: 8080}
			err := resAddr.Set(tc.input)
			if err != nil {
				assert.Nil(t, tc.expectedAddr, "Invalid input was successfully parsed")
				return
			}
			assert.Equal(t, tc.expectedAddr.Host, resAddr.Host, "hosts mismatch")
			assert.Equal(t, tc.expectedAddr.Port, resAddr.Port, "ports mismatch")
		})
	}
}
