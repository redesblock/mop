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
		issuersMap: make(map[string]*postage.StampIssuer),
	}
	for _, v := range o {
		v.apply(m)
	}

	return m
}

// WithAcceptAll sets the mock to return a new BatchIssuer on every
// call to GetStampIssuer.
func WithAcceptAll() Option {
	return optionFunc(func(m *mockPostage) { m.acceptAll = true })
}

func WithIssuer(s *postage.StampIssuer) Option {
	return optionFunc(func(m *mockPostage) {
		m.issuersMap = map[string]*postage.StampIssuer{string(s.ID()): s}
	})
}

type mockPostage struct {
	issuersMap map[string]*postage.StampIssuer
	issuerLock sync.Mutex
	acceptAll  bool
}

func (m *mockPostage) Add(s *postage.StampIssuer) error {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	m.issuersMap[string(s.ID())] = s
	return nil
}

func (m *mockPostage) StampIssuers() []*postage.StampIssuer {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	issuers := []*postage.StampIssuer{}
	for _, v := range m.issuersMap {
		issuers = append(issuers, v)
	}
	return issuers
}

func (m *mockPostage) GetStampIssuer(id []byte) (*postage.StampIssuer, error) {
	if m.acceptAll {
		return postage.NewStampIssuer("test fallback", "test identity", id, big.NewInt(3), 24, 6, 1000, true), nil
	}

	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	i, exists := m.issuersMap[string(id)]
	if !exists {
		return nil, errors.New("stampissuer not found")
	}
	return i, nil
}

func (m *mockPostage) IssuerUsable(_ *postage.StampIssuer) bool {
	return true
}

func (m *mockPostage) HandleCreate(_ *postage.Batch) error { return nil }

func (m *mockPostage) HandleTopUp(_ []byte, _ *big.Int) {}

func (m *mockPostage) HandleDepthIncrease(_ []byte, _ uint8, _ *big.Int) {}

func (m *mockPostage) Close() error {
	return nil
}
