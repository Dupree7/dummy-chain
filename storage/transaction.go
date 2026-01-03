package storage

import (
	"bytes"
	"dummy-chain/common/types"
	"encoding/gob"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
)

func (b *BadgerDb) SetTransaction(transaction *types.Transaction) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(*transaction); err != nil {
		return err
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(getTransactionKey(transaction.Hash), buf.Bytes())
	}); err != nil {
		return err
	}
	return nil
}

func (b *BadgerDb) GetTransaction(hash ecommon.Hash) (*types.Transaction, error) {
	var decodedTransaction types.Transaction
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getTransactionKey(hash))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if errDecode := gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedTransaction); errDecode != nil {
			return errDecode
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &decodedTransaction, nil
}

func (b *BadgerDb) DeleteTransaction(hash ecommon.Hash) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(getTransactionKey(hash))
	})
}
