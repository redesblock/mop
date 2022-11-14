package mock

import (
	"bytes"
	"errors"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/incentives/voucher"
)

type optionFunc func(*mockVoucher)

// Option is an option passed to a mock voucher Service.
type Option interface {
	apply(*mockVoucher)
}

func (f optionFunc) apply(r *mockVoucher) { f(r) }

// New creates a new mock voucher service.
func New(o ...Option) voucher.Service {
	m := &mockVoucher{
		issuersMap: make(map[string]*voucher.StampIssuer),
	}
	for _, v := range o {
		v.apply(m)
	}

	return m
}

// WithAcceptAll sets the mock to return a new BatchIssuer on every
// call to GetStampIssuer.
func WithAcceptAll() Option {
	return optionFunc(func(m *mockVoucher) { m.acceptAll = true })
}

func WithIssuer(s *voucher.StampIssuer) Option {
	return optionFunc(func(m *mockVoucher) {
		m.issuersMap = map[string]*voucher.StampIssuer{string(s.ID()): s}
	})
}

type mockVoucher struct {
	issuersMap map[string]*voucher.StampIssuer
	issuerLock sync.Mutex
	acceptAll  bool
}

func (m *mockVoucher) SetExpired() error {
	return nil
}

func (m *mockVoucher) HandleStampExpiry(id []byte) {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	for _, v := range m.issuersMap {
		if bytes.Equal(id, v.ID()) {
			v.SetExpired(true)
		}
	}
}

func (m *mockVoucher) Add(s *voucher.StampIssuer) error {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	m.issuersMap[string(s.ID())] = s
	return nil
}

func (m *mockVoucher) StampIssuers() []*voucher.StampIssuer {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	issuers := []*voucher.StampIssuer{}
	for _, v := range m.issuersMap {
		issuers = append(issuers, v)
	}
	return issuers
}

func (m *mockVoucher) GetStampIssuer(id []byte) (*voucher.StampIssuer, error) {
	if m.acceptAll {
		return voucher.NewStampIssuer("test fallback", "test identity", id, big.NewInt(3), 24, 6, 1000, true), nil
	}

	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	i, exists := m.issuersMap[string(id)]
	if !exists {
		return nil, errors.New("stampissuer not found")
	}
	return i, nil
}

func (m *mockVoucher) IssuerUsable(_ *voucher.StampIssuer) bool {
	return true
}

func (m *mockVoucher) HandleCreate(_ *voucher.Batch, _ *big.Int) error { return nil }

func (m *mockVoucher) HandleTopUp(batchID []byte, amount *big.Int) {
	m.issuerLock.Lock()
	defer m.issuerLock.Unlock()

	for _, v := range m.issuersMap {
		if bytes.Equal(batchID, v.ID()) {
			v.Amount().Add(v.Amount(), amount)
		}
	}
}

func (m *mockVoucher) HandleDepthIncrease(_ []byte, _ uint8) {}

func (m *mockVoucher) Close() error {
	return nil
}
