package mock

import (
	"context"

	"github.com/redesblock/mop/core/flock"
)

// Steward represents steward.Interface mock.
type Steward struct {
	addr flock.Address
}

// Reupload implements steward.Interface Reupload method.
// The given address is recorded.
func (s *Steward) Reupload(_ context.Context, addr flock.Address) error {
	s.addr = addr
	return nil
}

// IsRetrievable implements steward.Interface IsRetrievable method.
// The method always returns true.
func (s *Steward) IsRetrievable(_ context.Context, _ flock.Address) (bool, error) {
	return true, nil
}

// LastAddress returns the last address given to the Reupload method call.
func (s *Steward) LastAddress() flock.Address {
	return s.addr
}
