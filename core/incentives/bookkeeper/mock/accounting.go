// Package mock provides a mock implementation for the
// bookkeeper interface.
package mock

import (
	"context"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/bookkeeper"
)

// Service is the mock Accounting service.
type Service struct {
	lock                    sync.Mutex
	balances                map[string]*big.Int
	prepareDebitFunc        func(peer cluster.Address, price uint64) (bookkeeper.Action, error)
	prepareCreditFunc       func(peer cluster.Address, price uint64, originated bool) (bookkeeper.Action, error)
	balanceFunc             func(cluster.Address) (*big.Int, error)
	shadowBalanceFunc       func(cluster.Address) (*big.Int, error)
	balancesFunc            func() (map[string]*big.Int, error)
	compensatedBalanceFunc  func(cluster.Address) (*big.Int, error)
	compensatedBalancesFunc func() (map[string]*big.Int, error)
	peerAccountingFunc      func() (map[string]bookkeeper.PeerInfo, error)
	balanceSurplusFunc      func(cluster.Address) (*big.Int, error)
}

type debitAction struct {
	accounting *Service
	price      *big.Int
	peer       cluster.Address
	applied    bool
}

type creditAction struct {
	accounting *Service
	price      *big.Int
	peer       cluster.Address
	applied    bool
}

// WithDebitFunc sets the mock Debit function
func WithPrepareDebitFunc(f func(peer cluster.Address, price uint64) (bookkeeper.Action, error)) Option {
	return optionFunc(func(s *Service) {
		s.prepareDebitFunc = f
	})
}

// WithDebitFunc sets the mock Debit function
func WithPrepareCreditFunc(f func(peer cluster.Address, price uint64, originated bool) (bookkeeper.Action, error)) Option {
	return optionFunc(func(s *Service) {
		s.prepareCreditFunc = f
	})
}

// WithBalanceFunc sets the mock Balance function
func WithBalanceFunc(f func(cluster.Address) (*big.Int, error)) Option {
	return optionFunc(func(s *Service) {
		s.balanceFunc = f
	})
}

// WithBalancesFunc sets the mock Balances function
func WithBalancesFunc(f func() (map[string]*big.Int, error)) Option {
	return optionFunc(func(s *Service) {
		s.balancesFunc = f
	})
}

// WithCompensatedBalanceFunc sets the mock Balance function
func WithCompensatedBalanceFunc(f func(cluster.Address) (*big.Int, error)) Option {
	return optionFunc(func(s *Service) {
		s.compensatedBalanceFunc = f
	})
}

// WithCompensatedBalancesFunc sets the mock Balances function
func WithCompensatedBalancesFunc(f func() (map[string]*big.Int, error)) Option {
	return optionFunc(func(s *Service) {
		s.compensatedBalancesFunc = f
	})
}

// WithBalanceSurplusFunc sets the mock SurplusBalance function
func WithBalanceSurplusFunc(f func(cluster.Address) (*big.Int, error)) Option {
	return optionFunc(func(s *Service) {
		s.balanceSurplusFunc = f
	})
}

func WithPeerAccountingFunc(f func() (map[string]bookkeeper.PeerInfo, error)) Option {
	return optionFunc(func(s *Service) {
		s.peerAccountingFunc = f
	})
}

// NewAccounting creates the mock bookkeeper implementation
func NewAccounting(opts ...Option) *Service {
	mock := new(Service)
	mock.balances = make(map[string]*big.Int)
	for _, o := range opts {
		o.apply(mock)
	}
	return mock
}

func (s *Service) MakeCreditAction(peer cluster.Address, price uint64) bookkeeper.Action {
	return &creditAction{
		accounting: s,
		price:      new(big.Int).SetUint64(price),
		peer:       peer,
		applied:    false,
	}
}

// Debit is the mock function wrapper that calls the set implementation
func (s *Service) PrepareDebit(_ context.Context, peer cluster.Address, price uint64) (bookkeeper.Action, error) {
	if s.prepareDebitFunc != nil {
		return s.prepareDebitFunc(peer, price)
	}

	bigPrice := new(big.Int).SetUint64(price)
	return &debitAction{
		accounting: s,
		price:      bigPrice,
		peer:       peer,
		applied:    false,
	}, nil
}

func (s *Service) PrepareCredit(_ context.Context, peer cluster.Address, price uint64, originated bool) (bookkeeper.Action, error) {
	if s.prepareCreditFunc != nil {
		return s.prepareCreditFunc(peer, price, originated)
	}

	return s.MakeCreditAction(peer, price), nil
}

func (a *debitAction) Apply() error {
	a.accounting.lock.Lock()
	defer a.accounting.lock.Unlock()

	if bal, ok := a.accounting.balances[a.peer.String()]; ok {
		a.accounting.balances[a.peer.String()] = new(big.Int).Add(bal, new(big.Int).Set(a.price))
	} else {
		a.accounting.balances[a.peer.String()] = new(big.Int).Set(a.price)
	}

	return nil
}

func (a *creditAction) Cleanup() {}

func (a *creditAction) Apply() error {
	a.accounting.lock.Lock()
	defer a.accounting.lock.Unlock()

	if bal, ok := a.accounting.balances[a.peer.String()]; ok {
		a.accounting.balances[a.peer.String()] = new(big.Int).Sub(bal, new(big.Int).Set(a.price))
	} else {
		a.accounting.balances[a.peer.String()] = new(big.Int).Neg(a.price)
	}

	return nil
}

func (a *debitAction) Cleanup() {}

// Balance is the mock function wrapper that calls the set implementation
func (s *Service) Balance(peer cluster.Address) (*big.Int, error) {
	if s.balanceFunc != nil {
		return s.balanceFunc(peer)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if bal, ok := s.balances[peer.String()]; ok {
		return bal, nil
	} else {
		return big.NewInt(0), nil
	}
}

func (s *Service) ShadowBalance(peer cluster.Address) (*big.Int, error) {
	if s.shadowBalanceFunc != nil {
		return s.shadowBalanceFunc(peer)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if bal, ok := s.balances[peer.String()]; ok {
		return new(big.Int).Neg(bal), nil
	} else {
		return big.NewInt(0), nil
	}
}

// Balances is the mock function wrapper that calls the set implementation
func (s *Service) Balances() (map[string]*big.Int, error) {
	if s.balancesFunc != nil {
		return s.balancesFunc()
	}
	return s.balances, nil
}

// CompensatedBalance is the mock function wrapper that calls the set implementation
func (s *Service) CompensatedBalance(peer cluster.Address) (*big.Int, error) {
	if s.compensatedBalanceFunc != nil {
		return s.compensatedBalanceFunc(peer)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.balances[peer.String()], nil
}

// CompensatedBalances is the mock function wrapper that calls the set implementation
func (s *Service) CompensatedBalances() (map[string]*big.Int, error) {
	if s.compensatedBalancesFunc != nil {
		return s.compensatedBalancesFunc()
	}
	return s.balances, nil
}

func (s *Service) PeerAccounting() (map[string]bookkeeper.PeerInfo, error) {
	if s.peerAccountingFunc != nil {
		return s.peerAccountingFunc()
	}
	return map[string]bookkeeper.PeerInfo{}, nil
}

func (s *Service) Connect(peer cluster.Address, full bool) {

}

func (s *Service) Disconnect(peer cluster.Address) {

}

func (s *Service) SurplusBalance(peer cluster.Address) (*big.Int, error) {
	if s.balanceFunc != nil {
		return s.balanceSurplusFunc(peer)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	return big.NewInt(0), nil
}

// Option is the option passed to the mock bookkeeper service
type Option interface {
	apply(*Service)
}

type optionFunc func(*Service)

func (f optionFunc) apply(r *Service) { f(r) }
