package bmt

import (
	"errors"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file/pipeline"
	bmtpool "github.com/redesblock/mop/core/util/bmt"
)

var (
	errInvalidData = errors.New("bmt: invalid data")
)

type bmtWriter struct {
	next pipeline.ChainWriter
}

// NewBmtWriter returns a new bmtWriter. Partial writes are not supported.
// Note: branching factor is the BMT branching factor, not the merkle trie branching factor.
func NewBmtWriter(next pipeline.ChainWriter) pipeline.ChainWriter {
	return &bmtWriter{
		next: next,
	}
}

// ChainWrite writes data in chain. It assumes span has been prepended to the data.
// The span can be encrypted or unencrypted.
func (w *bmtWriter) ChainWrite(p *pipeline.PipeWriteArgs) error {
	if len(p.Data) < cluster.SpanSize {
		return errInvalidData
	}
	hasher := bmtpool.Get()
	hasher.SetHeader(p.Data[:cluster.SpanSize])
	_, err := hasher.Write(p.Data[cluster.SpanSize:])
	if err != nil {
		bmtpool.Put(hasher)
		return err
	}
	p.Ref, err = hasher.Hash(nil)
	bmtpool.Put(hasher)
	if err != nil {
		return err
	}

	return w.next.ChainWrite(p)
}

// sum calls the next writer for the cryptographic sum
func (w *bmtWriter) Sum() ([]byte, error) {
	return w.next.Sum()
}
