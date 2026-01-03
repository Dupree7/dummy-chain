package rpc

import (
	"dummy-chain/common"
	"dummy-chain/common/types"
	"dummy-chain/storage"
	"fmt"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
)

type Service struct {
	storage *storage.BadgerDb
}

func NewService(db *storage.BadgerDb) *Service {
	return &Service{
		storage: db,
	}
}

// GetAccountInfo We assume that if we don't have an address locally, it means that it was never in a transaction so balance and nonce are 0
func (b *Service) GetAccountInfo(address ecommon.Address, reply *types.AccountInfo) error {
	account, err := b.storage.GetAccount(address)
	if err != nil {
		return err
	} else if account == nil {
		*reply = types.AccountInfo{
			Address:    address.String(),
			Nonce:      0,
			Balance:    "0",
			BalanceRaw: big.NewInt(0),
		}
		return nil
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
	fmt.Printf("block: %s\n", block.String())
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
