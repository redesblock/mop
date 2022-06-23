package mock

import (
	"context"

	"github.com/redesblock/hop/core/swarm"
)

// Steward represents steward.Interface mock.
type Steward struct {
	addr swarm.Address
}

// Reupload implements steward.Interface Reupload method.
// The given address is recorded.
func (s *Steward) Reupload(_ context.Context, addr swarm.Address) error {
	s.addr = addr
	return nil
}

// IsRetrievable implements steward.Interface IsRetrievable method.
// The method always returns true.
func (s *Steward) IsRetrievable(_ context.Context, addr swarm.Address) (bool, error) {
	return addr.Equal(s.addr), nil
}

// LastAddress returns the last address given to the Reupload method call.
func (s *Steward) LastAddress() swarm.Address {
	return s.addr
}
