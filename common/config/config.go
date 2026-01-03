package config

import (
	"dummy-chain/common"
	"encoding/json"
	"fmt"
)

type GlobalConfig struct {
	BaseConfig `json:"Base"`
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		BaseConfig: BaseConfig{
			DataPath:     common.DefaultDataDir(),
			Mnemonic:     "",
			AccountIndex: 0,
			Url:          "http://127.0.0.1:12345",
		},
	}
}

func (c *GlobalConfig) AsMap() map[string]interface{} {
	var result map[string]interface{}
	bytes, _ := json.Marshal(c)

	_ = json.Unmarshal(bytes, &result)

	result["Base"] = c.BaseConfig.AsMap()
	return result
}

func (c *GlobalConfig) GetDataPath() string {
	return c.DataPath
}

func (c *GlobalConfig) GetMnemonic() string {
	return c.Mnemonic
}

func (c *GlobalConfig) DefaultNetworkConfigFileName() string {
	return fmt.Sprintf("config.json")
}
