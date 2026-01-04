package common

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	DummyChainId = 21
	BlockTime    = 5 * time.Second
	CoinDecimals = 18
	CoinSymbol   = "GO"

	DummyAddressStr   = "0x00000000000000000000000000000000DeaDBeef"
	DefaultStorageDir = "storage"
)

var (
	Big0      = big.NewInt(0)
	Big1      = big.NewInt(1)
	Big2      = big.NewInt(2)
	Big10     = big.NewInt(10)
	Big100    = big.NewInt(100)
	Big128    = big.NewInt(128)
	BigP128   = new(big.Int).Exp(Big2, Big128, nil)
	Big256    = big.NewInt(256)
	BigP256   = new(big.Int).Exp(Big2, Big256, nil)
	Big10000  = big.NewInt(10000)
	BigP256m1 = new(big.Int).Sub(BigP256, big.NewInt(1))

	OneCoin      = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(CoinDecimals)), nil)
	DummyAddress = common.HexToAddress(DummyAddressStr)

	ValidatorRole = "validator"
	ClientRole    = "client"
)
