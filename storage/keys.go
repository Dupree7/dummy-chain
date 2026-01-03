package storage

import (
	"dummy-chain/common"

	ecommon "github.com/ethereum/go-ethereum/common"
)

var (
	heightPrefix       = []byte{0}
	accountPrefix      = []byte{1}
	transactionPrefix  = []byte{2}
	blockPrefix        = []byte{3}
	heightToHashPrefix = []byte{10}
)

func getHeightKey() []byte {
	return heightPrefix
}

func getAccountKey(address ecommon.Address) []byte {
	return common.JoinBytes(accountPrefix, address.Bytes())
}

func getTransactionKey(hash ecommon.Hash) []byte {
	return common.JoinBytes(transactionPrefix, hash.Bytes())
}

func getBlockKey(hash ecommon.Hash) []byte {
	return common.JoinBytes(blockPrefix, hash.Bytes())
}

func getHeightToHashKey(height uint64) []byte {
	return common.JoinBytes(heightToHashPrefix, common.Uint64ToBytes(height))
}
