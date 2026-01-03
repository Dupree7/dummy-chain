package types

import (
	"dummy-chain/common"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address ecommon.Address
	Nonce   uint64
	Balance *big.Int
}

func (a *Account) ToInfo() AccountInfo {
	return AccountInfo{
		Address:    a.Address.String(),
		Nonce:      a.Nonce,
		Balance:    common.FormatBigInt(a.Balance),
		BalanceRaw: new(big.Int).Set(a.Balance),
	}
}

type AccountInfo struct {
	Address    string
	Nonce      uint64
	Balance    string
	BalanceRaw *big.Int
}
