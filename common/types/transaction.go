package types

import (
	"bytes"
	"dummy-chain/common"
	"encoding/base64"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Transaction struct {
	Hash        ecommon.Hash
	BlockHeight uint64
	From        ecommon.Address
	Nonce       uint64
	To          ecommon.Address
	Value       *big.Int
	Signature   []byte
}

func (tx *Transaction) GetHash() ecommon.Hash {
	buf := new(bytes.Buffer)
	buf.Write(tx.From.Bytes())
	buf.Write(common.Uint64ToBytes(tx.Nonce))
	buf.Write(tx.To.Bytes())
	buf.Write(tx.Value.Bytes())

	return crypto.Keccak256Hash(buf.Bytes())
}

func (tx *Transaction) ToInfo() TransactionInfo {
	return TransactionInfo{
		Hash:        tx.Hash.String(),
		BlockHeight: tx.BlockHeight,
		From:        tx.From.String(),
		Nonce:       tx.Nonce,
		To:          tx.To.String(),
		Value:       new(big.Int).Set(tx.Value),
		Signature:   base64.StdEncoding.EncodeToString(tx.Signature),
	}
}

type TransactionInfo struct {
	Hash        string
	BlockHeight uint64
	From        string
	Nonce       uint64
	To          string
	Value       *big.Int
	Signature   string
}
