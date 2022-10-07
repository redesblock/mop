package feeds

import (
	"context"
	"encoding/binary"

	"github.com/redesblock/mop/core/cac"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/soc"
	"github.com/redesblock/mop/core/storage"
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
func NewPutter(putter storage.Putter, signer crypto.Signer, topic []byte) (*Putter, error) {
	owner, err := signer.EthereumAddress()
	if err != nil {
		return nil, err
	}
	feed := New(topic, owner)
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
	s := soc.New(id, cac)
	ch, err := s.Sign(u.signer)
	if err != nil {
		return err
	}
	_, err = u.putter.Put(ctx, storage.ModePutUpload, ch)
	return err
}

func toChunk(at uint64, payload []byte) (flock.Chunk, error) {
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, at)
	return cac.New(append(ts, payload...))
}
