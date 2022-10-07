package settlement

import (
	"errors"
	"math/big"

	"github.com/redesblock/mop/core/flock"
)

var (
	ErrPeerNoSettlements = errors.New("no settlements for peer")
)

// Interface is the interface used by Accounting to trigger settlement
type Interface interface {
	// TotalSent returns the total amount sent to a peer
	TotalSent(peer flock.Address) (totalSent *big.Int, err error)
	// TotalReceived returns the total amount received from a peer
	TotalReceived(peer flock.Address) (totalSent *big.Int, err error)
	// SettlementsSent returns sent settlements for each individual known peer
	SettlementsSent() (map[string]*big.Int, error)
	// SettlementsReceived returns received settlements for each individual known peer
	SettlementsReceived() (map[string]*big.Int, error)
}

type Accounting interface {
	PeerDebt(peer flock.Address) (*big.Int, error)
	NotifyPaymentReceived(peer flock.Address, amount *big.Int) error
	NotifyPaymentSent(peer flock.Address, amount *big.Int, receivedError error)
	NotifyRefreshmentReceived(peer flock.Address, amount *big.Int) error
	Connect(peer flock.Address)
	Disconnect(peer flock.Address)
}
