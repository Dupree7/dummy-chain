package common

import "github.com/pkg/errors"

var (
	ErrDataDirUsed     = errors.New("dataDir already used by another process")
	DataDirInUseErrNos = map[uint]bool{11: true, 32: true, 35: true}

	ErrNotFound             = errors.New("not found")
	ErrAddressNotValid      = errors.New("Address is not valid")
	ErrNotEnoughBalanceUser = errors.New("Not enough balance")
)
