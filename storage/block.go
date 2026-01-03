package storage

import (
	"bytes"
	"dummy-chain/common/types"
	"encoding/gob"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
)

func (b *BadgerDb) SetBlock(block *types.Block) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(*block); err != nil {
		return err
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(getBlockKey(block.Hash), buf.Bytes())
	}); err != nil {
		return err
	}

	// Also link the height to the hash
	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(getHeightToHashKey(block.Height), block.Hash.Bytes())
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
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getBlockKey(*hash))
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

func (b *BadgerDb) DeleteBlock(hash ecommon.Hash) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(getBlockKey(hash))
	})
}
