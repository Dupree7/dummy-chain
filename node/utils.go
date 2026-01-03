package node

import (
	"dummy-chain/common"
	"dummy-chain/common/config"
	"os"
	"path/filepath"

	"github.com/prometheus/tsdb/fileutil"
	"go.uber.org/zap"
)

func (node *Node) Wait() {
	signalReceived := <-node.stopChan
	node.logger.Info("Received signal from Wait method: ", signalReceived)
}

func (node *Node) GetGlobalConfig() *config.GlobalConfig {
	return node.globalConfig
}

func (node *Node) openDataDir() error {
	if node.globalConfig.GetDataPath() == "" {
		return nil
	}

	if err := os.MkdirAll(node.globalConfig.GetDataPath(), 0700); err != nil {
		return err
	}
	node.logger.Info("successfully ensured dataPath exists", zap.String("data-path", node.globalConfig.GetDataPath()))

	// Lock the instance directory to prevent concurrent use by another instance as well as
	// accidental use of the instance directory as a database.
	if fileLock, _, err := fileutil.Flock(filepath.Join(node.globalConfig.GetDataPath(), ".lock")); err != nil {
		node.logger.Info("unable to acquire file-lock", zap.String("reason", err.Error()))
		return common.ConvertFileLockError(err)
	} else {
		node.dataDirLock = fileLock
	}

	node.logger.Info("successfully locked dataDir")
	return nil
}

func (node *Node) closeDataDir() {
	node.logger.Info("releasing dataDir lock ... ")
	// Release instance directory lock.
	if node.dataDirLock != nil {
		if err := node.dataDirLock.Release(); err != nil {
			node.logger.Error("can't release dataDir lock", zap.String("reason", err.Error()))
		}
		node.dataDirLock = nil
	}
}
