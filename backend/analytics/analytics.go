package analytics

import (
	"errors"
)

var (
	ErrTooManyAddresses   = errors.New("request contains too many addresses")
	ErrNonExistentAddress = errors.New("address does not exist")
)
