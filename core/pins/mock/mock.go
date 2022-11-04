package mock

import (
	"context"

	"github.com/redesblock/mop/core/cluster"
	pins "github.com/redesblock/mop/core/pins"
)

var _ pins.Interface = (*ServiceMock)(nil)

// NewServiceMock is a convenient constructor for creating ServiceMock.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{index: make(map[string]int)}
}

// ServiceMock represents a simple mock of pins.Interface.
// The implementation is not goroutine-safe.
type ServiceMock struct {
	index      map[string]int
	references []cluster.Address
}

// CreatePin implements pins.Interface CreatePin method.
func (sm *ServiceMock) CreatePin(_ context.Context, ref cluster.Address, _ bool) error {
	if _, ok := sm.index[ref.String()]; ok {
		return nil
	}
	sm.index[ref.String()] = len(sm.references)
	sm.references = append(sm.references, ref)
	return nil
}

// DeletePin implements pins.Interface DeletePin method.
func (sm *ServiceMock) DeletePin(_ context.Context, ref cluster.Address) error {
	i, ok := sm.index[ref.String()]
	if !ok {
		return nil
	}
	delete(sm.index, ref.String())
	sm.references = append(sm.references[:i], sm.references[i+1:]...)
	return nil
}

// HasPin implements pins.Interface HasPin method.
func (sm *ServiceMock) HasPin(ref cluster.Address) (bool, error) {
	_, ok := sm.index[ref.String()]
	return ok, nil
}

// Pins implements pins.Interface Pins method.
func (sm *ServiceMock) Pins() ([]cluster.Address, error) {
	return append([]cluster.Address(nil), sm.references...), nil
}
