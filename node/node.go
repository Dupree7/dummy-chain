package node

import (
	"context"
	"crypto/ecdsa"
	"dummy-chain/common"
	"dummy-chain/common/config"
	"dummy-chain/metadata"
	"dummy-chain/rpc"
	"dummy-chain/storage"
	"os"
	"sync"
	"time"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"go.uber.org/zap"
)

type Node struct {
	globalConfig *config.GlobalConfig
	logger       *zap.SugaredLogger

	// Private key and address of the user
	privateKey *ecdsa.PrivateKey
	address    *ecommon.Address

	rpcServer *rpc.Server
	rpcClient *rpc.Client
	storage   *storage.BadgerDb

	// Channel to wait for termination notifications
	stopChan chan os.Signal
	lock     sync.RWMutex
	// Prevents concurrent use of instance directory
	dataDirLock fileutil.Releaser
}

func NewNode(globalConfig *config.GlobalConfig, logger *zap.Logger) (*Node, error) {
	var err error

	node := &Node{
		globalConfig: globalConfig,
		logger:       logger.Sugar(),
		stopChan:     make(chan os.Signal, 1),
	}

	if err = node.openDataDir(); err != nil {
		return nil, err
	}

	node.storage, err = storage.NewBadgerDb("chain")
	if err != nil {
		return nil, err
	}

	// Only the syncer is the server and the source of truth, other are clients
	// All RPCs will be sent to him
	if metadata.Role == common.ValidatorRole {
		node.rpcServer, err = rpc.NewServer(node.storage)
		if err != nil {
			return nil, err
		}
	} else {
		if globalConfig.AccountIndex == 0 {
			return nil, errors.New("Account index must be greater than 0")
		}
		node.rpcClient, err = rpc.NewClient(globalConfig.Url)
		if err != nil {
			return nil, err
		}
	}

	seed := bip39.NewSeed(globalConfig.GetMnemonic(), "")
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	node.privateKey, node.address, err = common.DeriveKey(masterKey, globalConfig.AccountIndex)

	return node, nil
}

func (node *Node) Start() error {
	node.lock.Lock()
	defer node.lock.Unlock()

	if errStart := node.storage.Start(); errStart != nil {
		return errStart
	}

	if metadata.Role == common.ValidatorRole {
		if errStart := node.rpcServer.Start(); errStart != nil {
			return errStart
		}
	}

	return nil
}

func (node *Node) Stop() error {
	node.lock.Lock()
	defer node.lock.Unlock()
	defer close(node.stopChan)
	node.logger.Info("stopping node ...")

	// Release instance directory lock.
	node.closeDataDir()
	return nil
}

func (node *Node) Sync(ctx context.Context) {
	// We first fetch all the remaining blocks
	//currentHeight, err := node.storage.GetHeight()
	//if err != nil {
	//	node.logger.Error("failed to get current height", err)
	//	return
	//}

	ticker := time.NewTicker(6 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

		}
	}
}
