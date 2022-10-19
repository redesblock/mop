package settlement

import (
	"errors"
	"math/big"

	"github.com/redesblock/mop/core/cluster"
)

var (
	ErrPeerNoSettlements = errors.New("no settlements for peer")
)

// Interface is the interface used by Accounting to trigger settlement
type Interface interface {
	// TotalSent returns the total amount sent to a peer
	TotalSent(peer cluster.Address) (totalSent *big.Int, err error)
	// TotalReceived returns the total amount received from a peer
	TotalReceived(peer cluster.Address) (totalSent *big.Int, err error)
	// SettlementsSent returns sent settlements for each individual known peer
	SettlementsSent() (map[string]*big.Int, error)
	// SettlementsReceived returns received settlements for each individual known peer
	SettlementsReceived() (map[string]*big.Int, error)
}

type Accounting interface {
	PeerDebt(peer cluster.Address) (*big.Int, error)
	NotifyPaymentReceived(peer cluster.Address, amount *big.Int) error
	NotifyPaymentSent(peer cluster.Address, amount *big.Int, receivedError error)
	NotifyRefreshmentReceived(peer cluster.Address, amount *big.Int, timestamp int64) error
	NotifyRefreshmentSent(peer cluster.Address, attemptedAmount, amount *big.Int, timestamp, interval int64, receivedError error)
	Connect(peer cluster.Address, fullNode bool)
	Disconnect(peer cluster.Address)
}
