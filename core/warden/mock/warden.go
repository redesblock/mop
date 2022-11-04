package mock

import (
	"context"

	"github.com/redesblock/mop/core/cluster"
)

// Steward represents warden.Interface mock.
type Steward struct {
	addr cluster.Address
}

// Reupload implements warden.Interface Reupload method.
// The given address is recorded.
func (s *Steward) Reupload(_ context.Context, addr cluster.Address) error {
	s.addr = addr
	return nil
}

// IsRetrievable implements warden.Interface IsRetrievable method.
// The method always returns true.
func (s *Steward) IsRetrievable(_ context.Context, addr cluster.Address) (bool, error) {
	return addr.Equal(s.addr), nil
}

// LastAddress returns the last address given to the Reupload method call.
func (s *Steward) LastAddress() cluster.Address {
	return s.addr
}
