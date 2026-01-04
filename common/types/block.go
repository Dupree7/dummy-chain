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
	ChainId      uint64
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
	buf.Write(common.Uint64ToBytes(b.ChainId))
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
		ChainId:      b.ChainId,
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

	return fmt.Sprintf(`
	Hash:      %s
	Height:    %d
	Timestamp: %s (%d)
	PrevHash:  %s
	Validator: %s
	ChainId:   %d
	Signature: %s
	Transactions (%d):%s
`,
		b.Hash.String(),
		b.Height,
		time.Unix(b.Timestamp, 0).Format(time.DateTime),
		b.Timestamp,
		b.PrevHash.String(),
		b.Validator.String(),
		b.ChainId,
		base64.StdEncoding.EncodeToString(b.Signature),
		len(b.Transactions),
		txs.String(),
	)
}

type BlockInfo struct {
	ChainId      uint64
	Hash         string
	Height       uint64
	Timestamp    int64
	PrevHash     string
	Validator    string
	Signature    string
	Transactions []TransactionInfo
}

func (bi *BlockInfo) ToBlock() (*Block, error) {
	sigBytes, err := base64.StdEncoding.DecodeString(bi.Signature)
	if err != nil {
		return nil, err
	}

	txHashes := make([]ecommon.Hash, len(bi.Transactions))
	for i, txInfo := range bi.Transactions {
		txHashes[i] = ecommon.HexToHash(txInfo.Hash)
	}

	return &Block{
		ChainId:      bi.ChainId,
		Hash:         ecommon.HexToHash(bi.Hash),
		Height:       bi.Height,
		Timestamp:    bi.Timestamp,
		PrevHash:     ecommon.HexToHash(bi.PrevHash),
		Validator:    ecommon.HexToAddress(bi.Validator),
		Signature:    sigBytes,
		Transactions: txHashes,
	}, nil
}

type BlockInfoList struct {
	Count  uint64
	Blocks []BlockInfo
}
