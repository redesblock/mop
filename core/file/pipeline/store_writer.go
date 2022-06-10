package pipeline

import (
	"context"

	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
)

type storeWriter struct {
	l    storage.Putter
	mode storage.ModePut
	ctx  context.Context
	next chainWriter
}

// newStoreWriter returns a storeWriter. It just writes the given data
// to a given storage.Storer.
func newStoreWriter(ctx context.Context, l storage.Putter, mode storage.ModePut, next chainWriter) chainWriter {
	return &storeWriter{ctx: ctx, l: l, mode: mode, next: next}
}

func (w *storeWriter) chainWrite(p *pipeWriteArgs) error {
	tag := sctx.GetTag(w.ctx)
	if tag != nil {
		tag.Inc(tags.StateSplit)
	}
	c := swarm.NewChunk(swarm.NewAddress(p.ref), p.data)
	seen, err := w.l.Put(w.ctx, w.mode, c)
	if err != nil {
		return err
	}
	if tag != nil {
		tag.Inc(tags.StateStored)
		if seen[0] {
			tag.Inc(tags.StateSeen)
		}
	}
	if w.next == nil {
		return nil
	}

	return w.next.chainWrite(p)

}

func (w *storeWriter) sum() ([]byte, error) {
	return w.next.sum()
}
