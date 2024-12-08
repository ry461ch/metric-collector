package cfghelper

import (
	"encoding/json"
	"os"
)

func ParseCfgFile(path string, cfg any) error {
	cfgData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(cfgData, cfg)
	if err != nil {
		return err
	}
	return nil
}
