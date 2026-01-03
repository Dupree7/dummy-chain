package app

import (
	"dummy-chain/common"
	config2 "dummy-chain/common/config"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func MakeConfig() (*config2.GlobalConfig, error) {
	cfg := config2.NewGlobalConfig()

	// 2. Load config file.
	err := readConfigFromFile(cfg)
	if err != nil {
		return nil, err
	}

	// 3. Make dir paths absolute
	if err := cfg.MakePathsAbsolute(); err != nil {
		return nil, err
	}

	// 4. Log config
	if j, err := json.MarshalIndent(cfg.AsMap(), "", "    "); err == nil {
		common.GlobalLogger.Info("Using the following node config: \n", string(j))
	}

	// 5. Write it so a default one is created after the first run
	if errWrite := config2.WriteConfig(cfg); errWrite != nil {
		return nil, errWrite
	}

	return cfg, nil
}

func readConfigFromFile(cfg *config2.GlobalConfig) error {
	// second read default settings
	dataPath := cfg.GetDataPath()
	configPath := filepath.Join(dataPath, cfg.DefaultNetworkConfigFileName())
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return err
	}

	if jsonConf, err := ioutil.ReadFile(configPath); err == nil {
		err = json.Unmarshal(jsonConf, &cfg)
		if err == nil {
			return nil
		}
		log.Print("GlobalConfig malformed: please check", "error", err)
		return err
	} else {
		log.Printf("Error when reading %s: - %s", configPath, err.Error())
	}
	return nil
}
