package entry

import (
	"errors"

	"github.com/redesblock/hop/core/collection"
	"github.com/redesblock/hop/core/swarm"
)

var (
	_                  = collection.Entry(&Entry{})
	serializedDataSize = swarm.SectionSize * 2
)

// Entry provides addition of metadata to a data reference.
// Implements collection.Entry.
type Entry struct {
	reference swarm.Address
	metadata  swarm.Address
}

// New creates a new Entry.
func New(reference, metadata swarm.Address) *Entry {
	return &Entry{
		reference: reference,
		metadata:  metadata,
	}
}

// Reference implements collection.Entry
func (e *Entry) Reference() swarm.Address {
	return e.reference
}

// Metadata implements collection.Entry
func (e *Entry) Metadata() swarm.Address {
	return e.metadata
}

// MarshalBinary implements encoding.BinaryMarshaler
func (e *Entry) MarshalBinary() ([]byte, error) {
	br := e.reference.Bytes()
	bm := e.metadata.Bytes()
	b := append(br, bm...)
	return b, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (e *Entry) UnmarshalBinary(b []byte) error {
	if len(b) != serializedDataSize {
		return errors.New("invalid data length")
	}
	e.reference = swarm.NewAddress(b[:swarm.SectionSize])
	e.metadata = swarm.NewAddress(b[swarm.SectionSize:])
	return nil
}
