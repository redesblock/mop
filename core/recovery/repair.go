package recovery

import (
	"context"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/pss"
	"github.com/redesblock/hop/core/pushsync"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

const (
	// RecoveryTopicText is the string used to construct the recovery topic.
	RecoveryTopicText = "RECOVERY"
)

var (
	// RecoveryTopic is the topic used for repairing globally pinned chunks.
	RecoveryTopic = pss.NewTopic(RecoveryTopicText)
)

// RecoveryHook defines code to be executed upon failing to retrieve chunks.
type RecoveryHook func(chunkAddress swarm.Address, targets pss.Targets) error

// NewRecoveryHook returns a new RecoveryHook with the sender function defined.
func NewRecoveryHook(pssSender pss.Sender) RecoveryHook {
	privk := crypto.Secp256k1PrivateKeyFromBytes([]byte(RecoveryTopicText))
	recipient := privk.PublicKey
	return func(chunkAddress swarm.Address, targets pss.Targets) error {
		payload := chunkAddress
		ctx := context.Background()
		err := pssSender.Send(ctx, RecoveryTopic, payload.Bytes(), &recipient, targets)
		return err
	}
}

// NewRepairHandler creates a repair function to re-upload globally pinned chunks to the network with the given store.
func NewRepairHandler(s storage.Storer, logger logging.Logger, pushSyncer pushsync.PushSyncer) pss.Handler {
	return func(ctx context.Context, m []byte) {
		chAddr := m

		// check if the chunk exists in the local store and proceed.
		// otherwise the Get will trigger a unnecessary network retrieve
		exists, err := s.Has(ctx, swarm.NewAddress(chAddr))
		if err != nil {
			return
		}
		if !exists {
			return
		}

		// retrieve the chunk from the local store
		ch, err := s.Get(ctx, storage.ModeGetRequest, swarm.NewAddress(chAddr))
		if err != nil {
			logger.Tracef("chunk repair: error while getting chunk for repairing: %v", err)
			return
		}

		// push the chunk using push sync so that it reaches it destination in network
		_, err = pushSyncer.PushChunkToClosest(ctx, ch)
		if err != nil {
			logger.Tracef("chunk repair: error while sending chunk or receiving receipt: %v", err)
			return
		}
	}
}
