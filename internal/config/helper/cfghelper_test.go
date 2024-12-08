package cfghelper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Cfg struct {
	Test string `json:"test"`
}

func TestBase(t *testing.T) {
	cfgPath := "/tmp/cfg.test"

	cfgFile, _ := os.OpenFile(cfgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer cfgFile.Close()
	cfgFile.Write([]byte("{\"test\": \"test\"}"))

	cfg := &Cfg{}
	ParseCfgFile(cfgPath, cfg)
	assert.Equal(t, "test", cfg.Test, "Parse broken")
}
