package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteConfig(cfg *GlobalConfig) error {
	// second read default settings
	dataPath := cfg.GetDataPath()
	configPath := filepath.Join(dataPath, cfg.DefaultNetworkConfigFileName())

	configBytes, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(configPath, configBytes, 0644); err != nil {
		return err
	}
	return nil
}
