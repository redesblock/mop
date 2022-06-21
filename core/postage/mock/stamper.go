package mock

import (
	"github.com/redesblock/hop/core/postage"
	"github.com/redesblock/hop/core/swarm"
)

type mockStamper struct{}

// NewStamper returns anew new mock stamper.
func NewStamper() postage.Stamper {
	return &mockStamper{}
}

// Stamp implements the Stamper interface. It returns an empty postage stamp.
func (mockStamper) Stamp(_ swarm.Address) (*postage.Stamp, error) {
	return &postage.Stamp{}, nil
}
