package storage

import (
	"bytes"
	"dummy-chain/common/types"
	"encoding/gob"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func (b *BadgerDb) SetBlock(block *types.Block, txs []*types.Transaction) error {
	// Atomic update
	if err := b.db.Update(func(txn *badger.Txn) error {
		// Store the block
		var blockBuf bytes.Buffer
		if err := gob.NewEncoder(&blockBuf).Encode(*block); err != nil {
			return err
		} else if err = txn.Set(getBlockKey(block.Hash), blockBuf.Bytes()); err != nil {
			return err
		}

		// Link the height to the hash
		if err := txn.Set(getHeightToHashKey(block.Height), block.Hash.Bytes()); err != nil {
			return err
		}

		var heightBuf bytes.Buffer
		if err := gob.NewEncoder(&heightBuf).Encode(block.Height); err != nil {
			return err
		} else if err = txn.Set(getHeightKey(), heightBuf.Bytes()); err != nil {
			return err
		}

		accountsCache := make(map[ecommon.Address]*types.Account)
		getAccount := func(address ecommon.Address) (*types.Account, error) {
			if acc, ok := accountsCache[address]; ok {
				return acc, nil
			}
			// Use the same txn
			item, err := txn.Get(getAccountKey(address))
			if errors.Is(err, badger.ErrKeyNotFound) {
				return &types.Account{
					Address: address,
					Nonce:   0,
					Balance: big.NewInt(0),
				}, nil
			} else if err != nil {
				return nil, err
			}

			var account types.Account
			data, err := item.ValueCopy(nil)
			if err != nil {
				return nil, err
			} else if err = gob.NewDecoder(bytes.NewReader(data)).Decode(&account); err != nil {
				return nil, err
			}
			return &account, err
		}

		// Store transactions and update accounts cache
		for _, tx := range txs {
			var txBuf bytes.Buffer
			if err := gob.NewEncoder(&txBuf).Encode(*tx); err != nil {
				return err
			} else if err = txn.Set(getTransactionKey(tx.Hash), txBuf.Bytes()); err != nil {
				return err
			}

			from, err := getAccount(tx.From)
			if err != nil {
				return err
			}
			from.Balance.Sub(from.Balance, tx.Value)
			from.Nonce = tx.Nonce + 1
			accountsCache[tx.From] = from

			to, err := getAccount(tx.To)
			if err != nil {
				return err
			}
			to.Balance.Add(to.Balance, tx.Value)
			accountsCache[tx.To] = to
		}

		// Store updated accounts
		for _, acc := range accountsCache {
			var accountBuf bytes.Buffer
			if err := gob.NewEncoder(&accountBuf).Encode(*acc); err != nil {
				return err
			} else if err = txn.Set(getAccountKey(acc.Address), accountBuf.Bytes()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (b *BadgerDb) GetBlockByHash(hash ecommon.Hash) (*types.Block, error) {
	var decodedBlock types.Block
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getBlockKey(hash))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if errDecode := gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedBlock); errDecode != nil {
			return errDecode
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &decodedBlock, nil
}

func (b *BadgerDb) GetBlockHashByHeight(height uint64) (*ecommon.Hash, error) {
	var decodedHash ecommon.Hash
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getHeightToHashKey(height))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		decodedHash = ecommon.BytesToHash(data)
		return nil
	}); err != nil {
		return nil, err
	}
	return &decodedHash, nil
}

func (b *BadgerDb) GetBlockByHeight(height uint64) (*types.Block, error) {
	hash, err := b.GetBlockHashByHeight(height)
	if err != nil {
		return nil, err
	}
	var decodedBlock types.Block
	if err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getBlockKey(*hash))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		} else if errDecode := gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedBlock); errDecode != nil {
			return errDecode
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &decodedBlock, nil
}

func (b *BadgerDb) DeleteBlock(hash ecommon.Hash) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(getBlockKey(hash))
	})
}
