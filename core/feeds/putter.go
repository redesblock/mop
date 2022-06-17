package feeds

import (
	"context"
	"encoding/binary"

	"github.com/redesblock/hop/core/bmtpool"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/soc"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

// Updater is the generic interface f
type Updater interface {
	Update(ctx context.Context, at int64, payload []byte) error
	Feed() *Feed
}

// Putter encapsulates a chunk store putter and a Feed to store feed updates
type Putter struct {
	putter storage.Putter
	signer crypto.Signer
	*Feed
}

// NewPutter constructs a feed Putter
func NewPutter(putter storage.Putter, signer crypto.Signer, topic string) (*Putter, error) {
	owner, err := signer.EthereumAddress()
	if err != nil {
		return nil, err
	}
	feed, err := New(topic, owner)
	if err != nil {
		return nil, err
	}
	return &Putter{putter, signer, feed}, nil
}

// Put pushes an update to the feed through the chunk stores
func (u *Putter) Put(ctx context.Context, i Index, at int64, payload []byte) error {
	id, err := u.Feed.Update(i).Id()
	if err != nil {
		return err
	}
	cac, err := toChunk(uint64(at), payload)
	if err != nil {
		return err
	}
	ch, err := soc.NewChunk(id, cac, u.signer)
	if err != nil {
		return err
	}
	_, err = u.putter.Put(ctx, storage.ModePutUpload, ch)
	return err
}

func toChunk(at uint64, payload []byte) (swarm.Chunk, error) {
	hasher := bmtpool.Get()
	defer bmtpool.Put(hasher)

	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, at)
	content := append(ts, payload...)
	_, err := hasher.Write(content)
	if err != nil {
		return nil, err
	}
	span := make([]byte, 8)
	binary.LittleEndian.PutUint64(span, uint64(len(content)))
	err = hasher.SetSpanBytes(span)
	if err != nil {
		return nil, err
	}
	return swarm.NewChunk(swarm.NewAddress(hasher.Sum(nil)), append(append([]byte{}, span...), content...)), nil
}
