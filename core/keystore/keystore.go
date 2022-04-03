package keystore

import (
	"crypto/ecdsa"
	"errors"
)

var ErrInvalidPassword = errors.New("invalid password")

type Service interface {
	Key(name, password string) (k *ecdsa.PrivateKey, created bool, err error)
}
