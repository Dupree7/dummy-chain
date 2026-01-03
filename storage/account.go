package storage

import (
	"bytes"
	"dummy-chain/common/types"
	"encoding/gob"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
)

func (b *BadgerDb) SetAccount(account *types.Account) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(*account); err != nil {
		return err
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(getAccountKey(account.Address), buf.Bytes())
	}); err != nil {
		return err
	}
	return nil
}

func (b *BadgerDb) GetAccount(address ecommon.Address) (*types.Account, error) {
	var decodedAccount types.Account
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getAccountKey(address))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if errDecode := gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedAccount); errDecode != nil {
			return errDecode
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &decodedAccount, nil
}

func (b *BadgerDb) DeleteAccount(address ecommon.Address) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(getAccountKey(address))
	})
}
