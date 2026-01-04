package rpc

import (
	"bytes"
	"dummy-chain/common"
	"dummy-chain/common/types"
	"dummy-chain/storage"
	"encoding/base64"
	"encoding/gob"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Service struct {
	storage *storage.BadgerDb
	memPool *MemoryPool
}

func NewService(db *storage.BadgerDb, memPool *MemoryPool) *Service {
	return &Service{
		storage: db,
		memPool: memPool,
	}
}

// GetAccountInfo We assume that if we don't have an address locally, it means that it was never in a transaction so balance and nonce are 0
func (b *Service) GetAccountInfo(address ecommon.Address, reply *types.AccountInfo) error {
	account, err := b.storage.GetAccount(address)
	if err != nil {
		return err
	}
	*reply = account.ToInfo()
	return nil
}

func (b *Service) GetTransactionByHash(hash ecommon.Hash, reply *types.TransactionInfo) error {
	tx, err := b.storage.GetTransaction(hash)
	if err != nil {
		return err
	} else if tx == nil {
		return common.ErrNotFound
	}

	*reply = tx.ToInfo()
	return nil
}

func (b *Service) GetCurrenBlockHeight(param *struct{}, reply *uint64) error {
	height, err := b.storage.GetHeight()
	if err != nil {
		return err
	}
	*reply = height
	return nil
}

func (b *Service) GetBlockByHash(hash ecommon.Hash, reply *types.BlockInfo) error {
	block, err := b.storage.GetBlockByHash(hash)
	if err != nil {
		return err
	} else if block == nil {
		return common.ErrNotFound
	}

	*reply = block.ToInfo()

	for _, txHash := range block.Transactions {
		tx, err := b.storage.GetTransaction(txHash)
		if err != nil {
			return err
		}
		reply.Transactions = append(reply.Transactions, tx.ToInfo())
	}

	return nil
}

func (b *Service) GetBlockByHeight(height uint64, reply *types.BlockInfo) error {
	block, err := b.storage.GetBlockByHeight(height)
	if err != nil {
		return err
	} else if block == nil {
		return common.ErrNotFound
	}
	*reply = block.ToInfo()

	for _, txHash := range block.Transactions {
		tx, err := b.storage.GetTransaction(txHash)
		if err != nil {
			return err
		}
		reply.Transactions = append(reply.Transactions, tx.ToInfo())
	}

	return nil
}

type BlockInterval struct {
	Left  uint64
	Right uint64
}

func (b *Service) GetBlocksInterval(interval BlockInterval, reply *types.BlockInfoList) error {
	*reply = types.BlockInfoList{}
	reply.Count = 0
	reply.Blocks = make([]types.BlockInfo, 0)

	for i := interval.Left; i <= interval.Right; i++ {
		block, err := b.storage.GetBlockByHeight(i)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) || errors.Is(err, badger.ErrKeyNotFound) {
				break
			}
			return err
		}
		blockInfo := block.ToInfo()
		for _, txHash := range block.Transactions {
			tx, err := b.storage.GetTransaction(txHash)
			if err != nil {
				return err
			}
			blockInfo.Transactions = append(blockInfo.Transactions, tx.ToInfo())
		}
		reply.Blocks = append(reply.Blocks, blockInfo)
		reply.Count += 1
	}

	return nil
}

func (b *Service) SendTransaction(base64Tx string, reply *bool) error {
	txBytes, err := base64.StdEncoding.DecodeString(base64Tx)
	if err != nil {
		return err
	}

	var transaction types.Transaction
	if errDecode := gob.NewDecoder(bytes.NewReader(txBytes)).Decode(&transaction); errDecode != nil {
		return errDecode
	}

	common.GlobalLogger.Debugf("Received transaction: %s", transaction.String())
	b.memPool.AddTransaction(&transaction)
	*reply = true
	return nil
}
