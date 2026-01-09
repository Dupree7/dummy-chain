package node

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"dummy-chain/common"
	"dummy-chain/common/config"
	"dummy-chain/common/types"
	"dummy-chain/metadata"
	"dummy-chain/rpc"
	"dummy-chain/storage"
	"encoding/base64"
	"encoding/gob"
	"math/big"
	"math/rand/v2"
	"os"
	"reflect"
	"sync"
	"time"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	memPool   *rpc.MemoryPool
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
		memPool:      rpc.NewMemoryPool(),
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
		node.rpcServer, err = rpc.NewServer(node.storage, node.memPool)
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

func (node *Node) Start(shouldSync bool) error {
	node.lock.Lock()
	defer node.lock.Unlock()

	if errStart := node.storage.Start(); errStart != nil {
		return errStart
	}

	if metadata.Role == common.ValidatorRole {
		go func() {
			if errStart := node.rpcServer.Start(); errStart != nil {
				node.logger.Debugf("rpc server start error: %v", errStart)
			}
		}()

		go node.CreateBlocks(context.Background())
	}

	if shouldSync && metadata.Role == common.ClientRole {
		if errStart := node.Sync(); errStart != nil {
			return errStart
		}

		go node.FetchBlocks(context.Background())
	}

	/// Print own address
	mnemonic := "margin bounce nominee submit pupil duty bird daughter hotel onion wave write"

	seed := bip39.NewSeed(mnemonic, "")
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return err
	}

	purposeKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		return err
	}
	coinTypeKey, err := purposeKey.NewChildKey(bip32.FirstHardenedChild + 60)
	if err != nil {
		return err
	}
	accountKey, err := coinTypeKey.NewChildKey(bip32.FirstHardenedChild)
	if err != nil {
		return err
	}
	changeKey, err := accountKey.NewChildKey(0)
	if err != nil {
		return err
	}

	addressKey, errChild := changeKey.NewChildKey(node.globalConfig.AccountIndex)
	if errChild != nil {
		return errChild
	}

	privateKey, errPrivateKey := crypto.ToECDSA(addressKey.Key)
	if errPrivateKey != nil {
		return errPrivateKey
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	common.GlobalLogger.Debugf("%d. Address: %s", node.globalConfig.AccountIndex, address.Hex())

	account, err := node.storage.GetAccount(address)
	if err != nil {
		return err
	}
	common.GlobalLogger.Debugf("My balance is: %s and my nonce is: %d",
		common.FormatBigInt(account.Balance), account.Nonce)

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

// Every 5 seconds we take all the transaction in the mempool, execute them and then create the block
func (node *Node) CreateBlocks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentBlockHeight, err := node.storage.GetHeight()
			if err != nil {
				node.logger.Debugf("Failed to get current block height: %s", err.Error())
				continue
			}
			prevBlock, err := node.storage.GetBlockByHeight(currentBlockHeight)
			if err != nil {
				node.logger.Debugf("Failed to get previous block: %s", err.Error())
				continue
			}

			block := &types.Block{
				ChainId:   common.DummyChainId,
				Height:    currentBlockHeight + 1,
				Timestamp: time.Now().Unix(),
				PrevHash:  prevBlock.Hash,
				Validator: *node.address,
			}

			txs := node.memPool.GetMemPool()

			// Only for testing purposes
			//if len(txs) == 0 {
			//	txs = node.GenerateRandomTestTransactions()
			//}

			var goodTxs []*types.Transaction
			if len(txs) > 0 {
				goodTxs = node.VerifyTransactions(txs)
				for _, tx := range goodTxs {
					tx.BlockHeight = block.Height
					block.Transactions = append(block.Transactions, tx.Hash)
				}
			}

			block.Hash = block.GetHash()
			block.Signature, err = crypto.Sign(block.Hash[:], node.privateKey)
			if err != nil {
				node.logger.Debugf("Failed to sign block: %s", err.Error())
				continue
			}

			if errSet := node.storage.SetBlock(block, goodTxs); errSet != nil {
				node.logger.Debugf("Failed to set block: %s", errSet.Error())
				continue
			} else {
				// Only now we remove the transactions from the mempool
				node.memPool.RemoveTxs(txs)
			}

			node.logger.Debugf("Created a new block: %s", block.String())
		}
	}
}

func (node *Node) Sync() error {
	for {
		currentHeight, err := node.storage.GetHeight()
		if err != nil {
			return err
		}

		list, err := node.rpcClient.GetBlocksInterval(currentHeight+1, currentHeight+10)
		if err != nil {
			return err
		} else if list.Count == 0 {
			break
		}
		node.logger.Debugf("Syncing block height: %d. Synced %d blocks", currentHeight, list.Count)
		for _, blockInfo := range list.Blocks {
			block, errTo := blockInfo.ToBlock()
			if errTo != nil {
				return errTo
			}
			//node.logger.Debugf("Adding block: %s", block.Hash.String())
			txs := make([]*types.Transaction, 0)
			for _, txInfo := range blockInfo.Transactions {
				tx, errTo := txInfo.ToTransaction()
				if errTo != nil {
					return errTo
				}
				txs = append(txs, tx)
			}
			if errSet := node.storage.SetBlock(block, txs); errSet != nil {
				return errSet
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (node *Node) FetchBlocks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			errSync := node.Sync()
			if errSync != nil {
				node.logger.Debugf("Failed to sync blocks: %s", errSync.Error())
			}
		}
	}
}

// VerifyTransactions
// It will simulate  the execution of all transactions
// Return a map of bad transaction to ignore them when adding to the block
func (node *Node) VerifyTransactions(txs []*types.Transaction) []*types.Transaction {
	virtualBalances := make(map[ecommon.Address]*big.Int)
	virtualNonce := make(map[ecommon.Address]uint64)
	goodTxs := make([]*types.Transaction, 0)

	for _, tx := range txs {
		//node.logger.Debugf("Verifying transaction: %s", tx.String())

		// Check signature
		if pubKeyBytes, err := crypto.Ecrecover(tx.Hash.Bytes(), tx.Signature); err != nil {
			node.logger.Debugf("Failed to recover tx signature: %s", err.Error())
			continue
		} else {
			if recoveredPubKey, err := crypto.UnmarshalPubkey(pubKeyBytes); err != nil {
				node.logger.Debugf("Failed to unmarshal pubkey: %s", err.Error())
				continue
			} else {
				recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)
				if !reflect.DeepEqual(recoveredAddress, tx.From) {
					node.logger.Debugf("Signer %s is not the sender %s", recoveredAddress, tx.From)
					continue
				}
			}
		}

		// Check that sender is different than the receiver
		if reflect.DeepEqual(tx.From, tx.To) {
			node.logger.Debugf("Transaction from %s and to %s are equal", tx.From, tx.To)
			continue
		}

		if _, ok := virtualBalances[tx.From]; !ok {
			accountInfo, err := node.storage.GetAccount(tx.From)
			if err != nil {
				node.logger.Debugf("Failed to get account info from storage: %s", err.Error())
				continue
			}

			virtualBalances[tx.From] = new(big.Int).Set(accountInfo.Balance)
			virtualNonce[tx.From] = accountInfo.Nonce
		}
		fromBalance := virtualBalances[tx.From]

		if _, ok := virtualBalances[tx.To]; !ok {
			accountInfo, err := node.storage.GetAccount(tx.To)
			if err != nil {
				node.logger.Debugf("Failed to get account info from storage: %s", err.Error())
				continue
			}

			virtualBalances[tx.To] = new(big.Int).Set(accountInfo.Balance)
		}
		toBalance := virtualBalances[tx.To]

		// Check if the sender has enough balance
		if fromBalance.Cmp(tx.Value) < 0 {
			node.logger.Debugf("From balnace %s is less than tx value %s", fromBalance, tx.Value)
			continue
		}
		fromBalance.Sub(fromBalance, tx.Value)
		toBalance.Add(toBalance, tx.Value)

		// Check nonce continuity
		if virtualNonce[tx.From] != tx.Nonce {
			node.logger.Debugf("Transaction nonce %d is different from virtual nonce %d", tx.Nonce, virtualNonce[tx.From])
			continue
		}
		virtualNonce[tx.From] += 1

		goodTxs = append(goodTxs, tx)
	}
	return goodTxs
}

func (node *Node) GenerateRandomTestTransactions() []*types.Transaction {
	txs := make([]*types.Transaction, 0)

	n := rand.N(4)
	account, err := node.storage.GetAccount(*node.address)
	if err != nil {
		node.logger.Debugf("Failed to get account info from storage: %s", err.Error())
		return txs
	}
	balance := new(big.Int).Set(account.Balance)
	node.logger.Debugf("Current balance: %s", common.FormatBigInt(balance))
	nonce := account.Nonce

	for i := 0; i < n; i++ {
		value := new(big.Int).Set(common.OneCoin)
		value.Mul(value, big.NewInt(int64(i+1)))
		if value.Cmp(balance) > 0 {
			return txs
		}

		tx := &types.Transaction{
			From:  *node.address,
			Nonce: nonce,
			To:    common.DummyAddress,
			Value: new(big.Int).Set(value),
		}
		tx.Hash = tx.GetHash()
		tx.Signature, err = crypto.Sign(tx.Hash[:], node.privateKey)
		if err != nil {
			node.logger.Debugf("Failed to sign tx: %s", err.Error())
			continue
		}

		balance.Sub(balance, value)
		nonce += 1

		txs = append(txs, tx)
	}
	return txs
}

func (node *Node) SendTransaction(to string, value string) error {
	valueBig, err := common.ParseToBigInt(value)
	if err != nil {
		return err
	}

	account, err := node.storage.GetAccount(*node.address)
	if err != nil {
		return err
	}

	if valueBig.Cmp(account.Balance) > 0 {
		return common.ErrNotEnoughBalanceUser
	}

	tx := &types.Transaction{
		BlockHeight: 0,
		From:        *node.address,
		Nonce:       account.Nonce,
		To:          ecommon.HexToAddress(to),
		Value:       new(big.Int).Set(valueBig),
		Signature:   []byte{},
	}

	tx.Hash = tx.GetHash()
	tx.Signature, err = crypto.Sign(tx.Hash[:], node.privateKey)
	if err != nil {
		return err
	}
	var txBuf bytes.Buffer
	if err := gob.NewEncoder(&txBuf).Encode(*tx); err != nil {
		return err
	}
	base64Tx := base64.StdEncoding.EncodeToString(txBuf.Bytes())
	return node.rpcClient.SendTransaction(base64Tx)
}
