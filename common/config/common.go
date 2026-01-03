package config

import (
	"dummy-chain/common"
	"path/filepath"
)

type BaseConfig struct {
	DataPath     string
	Mnemonic     string
	AccountIndex uint32
	Url          string
}

func (c *BaseConfig) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"DataPath":     c.DataPath,
		"Mnemonic":     "REDACTED",
		"AccountIndex": c.AccountIndex,
		"Url":          c.Url,
	}
}

func (c *BaseConfig) MakePathsAbsolute() error {
	if c.DataPath == "" {
		c.DataPath = common.DefaultDataDir()
	} else {
		absDataDir, err := filepath.Abs(c.DataPath)
		if err != nil {
			return err
		}
		c.DataPath = absDataDir
	}

	return nil
}
