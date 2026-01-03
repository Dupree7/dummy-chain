package storage

import (
	"bytes"
	"dummy-chain/common"
	"dummy-chain/common/types"
	"encoding/gob"
	"math/big"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

type BadgerDb struct {
	db *badger.DB
}

func NewBadgerDb(name string) (*BadgerDb, error) {
	dbDir := filepath.Join(common.DefaultDataDir(), common.DefaultStorageDir, name)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err = os.MkdirAll(dbDir, 0700); err != nil {
			return nil, err
		}
	}
	opts := badger.DefaultOptions(dbDir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDb{
		db: db,
	}, nil
}

// Start This method will simulate the interpretation of the genesis block
// It should add balances and init accounts to the first 100 indexes of the mnemonic
func (b *BadgerDb) Start() error {
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

	genesisBlock := &types.Block{
		Height:       0,
		Timestamp:    1767441600,
		PrevHash:     ecommon.Hash{},
		Validator:    ecommon.Address{},
		Signature:    []byte{},
		Transactions: []ecommon.Hash{},
	}

	for i := uint32(0); i < 5; i++ {
		addressKey, errChild := changeKey.NewChildKey(i)
		if errChild != nil {
			return errChild
		}

		privateKey, errPrivateKey := crypto.ToECDSA(addressKey.Key)
		if errPrivateKey != nil {
			return errPrivateKey
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		common.GlobalLogger.Debugf("%d. Address: %s", i, address.Hex())

		amount := new(big.Int).Mul(common.OneCoin, common.Big10000)
		bigI := new(big.Int).SetUint64(uint64(i))
		account := &types.Account{
			Address: address,
			Nonce:   0,
			Balance: new(big.Int).Add(amount, bigI),
		}

		if errSet := b.SetAccount(account); errSet != nil {
			return errSet
		}

		tx := &types.Transaction{
			BlockHeight: 0,
			From:        ecommon.Address{},
			Nonce:       0,
			To:          account.Address,
			Value:       new(big.Int).Set(account.Balance),
			Signature:   []byte{},
		}
		tx.Hash = tx.GetHash()
		if errSet := b.SetTransaction(tx); errSet != nil {
			return errSet
		}

		genesisBlock.Transactions = append(genesisBlock.Transactions, tx.Hash)
	}

	genesisBlock.Hash = genesisBlock.GetHash()
	common.GlobalLogger.Debugf("Genesis %s", genesisBlock.String())
	if errSet := b.SetBlock(genesisBlock); errSet != nil {
		return errSet
	}

	return nil
}

func (b *BadgerDb) SetHeight(height uint64) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(height); err != nil {
		return err
	}

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(getHeightKey(), buf.Bytes())
	}); err != nil {
		return err
	}
	return nil
}

func (b *BadgerDb) GetHeight() (uint64, error) {
	var decodedHeight uint64
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getHeightKey())
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if errDecode := gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedHeight); errDecode != nil {
			return errDecode
		}
		return nil
	}); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}
	}
	return decodedHeight, nil
}
