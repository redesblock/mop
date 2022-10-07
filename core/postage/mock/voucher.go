package mock

import (
	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/swarm"
)

type mockVoucher struct{}

// NewVoucher returns anew new mock voucher.
func NewVoucher() postage.Voucher {
	return &mockVoucher{}
}

// Vouch implements the Voucher interface. It returns an empty postage vouch.
func (mockVoucher) Vouch(_ swarm.Address) (*postage.Vouch, error) {
	return &postage.Vouch{}, nil
}
