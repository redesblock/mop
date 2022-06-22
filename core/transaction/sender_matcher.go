package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/swarm"
)

type Matcher struct {
	backend Backend
	signer  types.Signer
}

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionPending       = errors.New("transaction in pending status")
	ErrTransactionSenderInvalid = errors.New("invalid transaction sender")
)

func NewMatcher(backend Backend, signer types.Signer) *Matcher {
	return &Matcher{
		backend: backend,
		signer:  signer,
	}
}

func (m Matcher) Matches(ctx context.Context, tx []byte, networkID uint64, senderOverlay swarm.Address) (bool, error) {
	incomingTx := common.BytesToHash(tx)

	nTx, isPending, err := m.backend.TransactionByHash(ctx, incomingTx)
	if err != nil {
		return false, fmt.Errorf("%v: %w", err, ErrTransactionNotFound)
	}

	if isPending {
		return false, ErrTransactionPending
	}

	sender, err := types.Sender(m.signer, nTx)
	if err != nil {
		return false, fmt.Errorf("%v: %w", err, ErrTransactionSenderInvalid)
	}

	expectedRemoteHopAddress := crypto.NewOverlayFromEthereumAddress(sender.Bytes(), networkID)

	return expectedRemoteHopAddress.Equal(senderOverlay), nil
}
