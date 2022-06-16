package backendmock

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/hop/core/settlement/swap/chequebook"
)

type backendMock struct {
	codeAt             func(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	sendTransaction    func(ctx context.Context, tx *types.Transaction) error
	suggestGasPrice    func(ctx context.Context) (*big.Int, error)
	estimateGas        func(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	transactionReceipt func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	pendingNonceAt     func(ctx context.Context, account common.Address) (uint64, error)
	transactionByHash  func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
}

func (m *backendMock) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	if m.codeAt != nil {
		return m.codeAt(ctx, contract, blockNumber)
	}
	return nil, errors.New("Error")
}

func (*backendMock) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return nil, errors.New("Error")
}

func (*backendMock) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return nil, errors.New("Error")
}

func (m *backendMock) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	if m.pendingNonceAt != nil {
		return m.pendingNonceAt(ctx, account)
	}
	return 0, errors.New("Error")
}

func (m *backendMock) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if m.suggestGasPrice != nil {
		return m.suggestGasPrice(ctx)
	}
	return nil, errors.New("Error")
}

func (m *backendMock) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	if m.estimateGas != nil {
		return m.estimateGas(ctx, call)
	}
	return 0, errors.New("Error")
}

func (m *backendMock) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if m.sendTransaction != nil {
		return m.sendTransaction(ctx, tx)
	}
	return errors.New("Error")
}

func (*backendMock) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return nil, errors.New("Error")
}

func (*backendMock) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return nil, errors.New("Error")
}

func (m *backendMock) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if m.transactionReceipt != nil {
		return m.transactionReceipt(ctx, txHash)
	}
	return nil, errors.New("Error")
}

func (m *backendMock) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	if m.transactionByHash != nil {
		return m.transactionByHash(ctx, hash)
	}
	return nil, false, errors.New("Error")
}

func New(opts ...Option) chequebook.Backend {
	mock := new(backendMock)
	for _, o := range opts {
		o.apply(mock)
	}
	return mock
}

// Option is the option passed to the mock Chequebook service
type Option interface {
	apply(*backendMock)
}

type optionFunc func(*backendMock)

func (f optionFunc) apply(r *backendMock) { f(r) }

func WithCodeAtFunc(f func(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)) Option {
	return optionFunc(func(s *backendMock) {
		s.codeAt = f
	})
}

func WithPendingNonceAtFunc(f func(ctx context.Context, account common.Address) (uint64, error)) Option {
	return optionFunc(func(s *backendMock) {
		s.pendingNonceAt = f
	})
}

func WithSuggestGasPriceFunc(f func(ctx context.Context) (*big.Int, error)) Option {
	return optionFunc(func(s *backendMock) {
		s.suggestGasPrice = f
	})
}

func WithEstimateGasFunc(f func(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)) Option {
	return optionFunc(func(s *backendMock) {
		s.estimateGas = f
	})
}

func WithTransactionReceiptFunc(f func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)) Option {
	return optionFunc(func(s *backendMock) {
		s.transactionReceipt = f
	})
}

func WithTransactionByHashFunc(f func(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)) Option {
	return optionFunc(func(s *backendMock) {
		s.transactionByHash = f
	})
}

func WithSendTransactionFunc(f func(ctx context.Context, tx *types.Transaction) error) Option {
	return optionFunc(func(s *backendMock) {
		s.sendTransaction = f
	})
}
