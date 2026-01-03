package types

import (
	"bytes"
	"dummy-chain/common"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Block struct {
	Hash         ecommon.Hash
	Height       uint64
	Timestamp    int64
	PrevHash     ecommon.Hash
	Validator    ecommon.Address
	Signature    []byte
	Transactions []ecommon.Hash
}

func (b *Block) GetHash() ecommon.Hash {
	buf := new(bytes.Buffer)
	buf.Write(common.Uint64ToBytes(b.Height))
	buf.Write(common.Uint64ToBytes(uint64(b.Timestamp)))
	buf.Write(b.PrevHash.Bytes())
	buf.Write(b.Validator.Bytes())

	for _, txHash := range b.Transactions {
		buf.Write(txHash.Bytes())
	}

	return crypto.Keccak256Hash(buf.Bytes())
}

func (b *Block) ToInfo() BlockInfo {
	return BlockInfo{
		Hash:         b.Hash.String(),
		Height:       b.Height,
		Timestamp:    b.Timestamp,
		PrevHash:     b.PrevHash.String(),
		Validator:    b.Validator.String(),
		Signature:    base64.StdEncoding.EncodeToString(b.Signature),
		Transactions: make([]TransactionInfo, 0),
	}
}

func (b *Block) String() string {
	var txs strings.Builder
	for i, tx := range b.Transactions {
		txs.WriteString(fmt.Sprintf("\n\t\t%d: %s", i, tx.String()))
	}

	return fmt.Sprintf(`Block:
	Hash:      %s
	Height:    %d
	Timestamp: %s (%d)
	PrevHash:  %s
	Validator: %s
	Signature: %s
	Transactions (%d):%s
`,
		b.Hash.String(),
		b.Height,
		time.Unix(b.Timestamp, 0).Format(time.DateTime),
		b.Timestamp,
		b.PrevHash.String(),
		b.Validator.String(),
		base64.StdEncoding.EncodeToString(b.Signature),
		len(b.Transactions),
		txs.String(),
	)
}

type BlockInfo struct {
	Hash         string
	Height       uint64
	Timestamp    int64
	PrevHash     string
	Validator    string
	Signature    string
	Transactions []TransactionInfo
}
