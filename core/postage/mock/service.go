package mock

import (
	"errors"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/postage"
)

type optionFunc func(*mockPostage)

// Option is an option passed to a mock postage Service.
type Option interface {
	apply(*mockPostage)
}

func (f optionFunc) apply(r *mockPostage) { f(r) }

// New creates a new mock postage service.
func New(o ...Option) postage.Service {
	m := &mockPostage{
		issuersMap: make(map[string]*postage.VouchIssuer),
	}
	for _, v := range o {
		v.apply(m)
	}

	return m
}

// WithAcceptAll sets the mock to return a new BatchIssuer on every
// call to GetVouchIssuer.
func WithAcceptAll() Option {
	return optionFunc(func(m *mockPostage) { m.acceptAll = true })
}

func WithIssuer(s *postage.VouchIssuer) Option {
	return optionFunc(func(m *mockPostage) {
		m.issuersMap = map[string]*postage.VouchIssuer{string(s.ID()): s}
	})
}

type mockPostage struct {
	issuersMap map[string]*postage.VouchIssuer
	issuerLock sync.Mutex
	acceptAll  bool
}

func (m *mockPostage) Add(s *postage.VouchIssuer) error {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	m.issuersMap[string(s.ID())] = s
	return nil
}

func (m *mockPostage) VouchIssuers() []*postage.VouchIssuer {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	issuers := []*postage.VouchIssuer{}
	for _, v := range m.issuersMap {
		issuers = append(issuers, v)
	}
	return issuers
}

func (m *mockPostage) GetVouchIssuer(id []byte) (*postage.VouchIssuer, error) {
	if m.acceptAll {
		return postage.NewVouchIssuer("test fallback", "test identity", id, big.NewInt(3), 24, 6, 1000, true), nil
	}

	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	i, exists := m.issuersMap[string(id)]
	if !exists {
		return nil, errors.New("vouchIssuer not found")
	}
	return i, nil
}

func (m *mockPostage) IssuerUsable(_ *postage.VouchIssuer) bool {
	return true
}

func (m *mockPostage) HandleCreate(_ *postage.Batch) error { return nil }

func (m *mockPostage) HandleTopUp(_ []byte, _ *big.Int) {}

func (m *mockPostage) HandleDepthIncrease(_ []byte, _ uint8, _ *big.Int) {}

func (m *mockPostage) Close() error {
	return nil
}
