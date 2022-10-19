package mock

import (
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
)

type mockStamper struct{}

// NewStamper returns anew new mock stamper.
func NewStamper() voucher.Stamper {
	return &mockStamper{}
}

// Stamp implements the Stamper interface. It returns an empty voucher stamp.
func (mockStamper) Stamp(_ cluster.Address) (*voucher.Stamp, error) {
	return &voucher.Stamp{}, nil
}
