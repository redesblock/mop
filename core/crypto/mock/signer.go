package mock

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/crypto/eip712"
)

type signerMock struct {
	signTx        func(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	signTypedData func(*eip712.TypedData) ([]byte, error)
	bscAddress    func() (common.Address, error)
	signFunc      func([]byte) ([]byte, error)
}

func (m *signerMock) BSCAddress() (common.Address, error) {
	if m.bscAddress != nil {
		return m.bscAddress()
	}
	return common.Address{}, nil
}

func (m *signerMock) Sign(data []byte) ([]byte, error) {
	return m.signFunc(data)
}

func (m *signerMock) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return m.signTx(transaction, chainID)
}

func (*signerMock) PublicKey() (*ecdsa.PublicKey, error) {
	return nil, nil
}

func (m *signerMock) SignTypedData(d *eip712.TypedData) ([]byte, error) {
	return m.signTypedData(d)
}

func New(opts ...Option) crypto.Signer {
	mock := new(signerMock)
	for _, o := range opts {
		o.apply(mock)
	}
	return mock
}

// Option is the option passed to the mock Chequebook service
type Option interface {
	apply(*signerMock)
}

type optionFunc func(*signerMock)

func (f optionFunc) apply(r *signerMock) { f(r) }

func WithSignFunc(f func(data []byte) ([]byte, error)) Option {
	return optionFunc(func(s *signerMock) {
		s.signFunc = f
	})
}

func WithSignTxFunc(f func(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error)) Option {
	return optionFunc(func(s *signerMock) {
		s.signTx = f
	})
}

func WithSignTypedDataFunc(f func(*eip712.TypedData) ([]byte, error)) Option {
	return optionFunc(func(s *signerMock) {
		s.signTypedData = f
	})
}

func WithBSCAddressFunc(f func() (common.Address, error)) Option {
	return optionFunc(func(s *signerMock) {
		s.bscAddress = f
	})
}
