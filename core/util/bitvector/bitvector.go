// Package bitvector provides functionality of a
// simple bit vector implementation.
package bitvector

import (
	"errors"
)

var errInvalidLength = errors.New("invalid length")

// BitVector is a convenience object for manipulating and representing bit vectors
type BitVector struct {
	len int
	b   []byte
}

// New creates a new bit vector with the given length
func New(l int) (*BitVector, error) {
	return NewFromBytes(make([]byte, l/8+1), l)
}

// NewFromBytes creates a bit vector from the passed byte slice.
//
// Leftmost bit in byte slice becomes leftmost bit in bit vector
func NewFromBytes(b []byte, l int) (*BitVector, error) {
	if l <= 0 {
		return nil, errInvalidLength
	}
	if len(b)*8 < l {
		return nil, errInvalidLength
	}
	return &BitVector{
		len: l,
		b:   b,
	}, nil
}

// Get gets the corresponding bit, counted from left to right
func (bv *BitVector) Get(i int) bool {
	bi := i / 8
	return bv.b[bi]&(0x1<<uint(i%8)) != 0
}

// Set sets the bit corresponding to the index in the bitvector, counted from left to right
func (bv *BitVector) Set(i int) {
	bi := i / 8
	if !bv.Get(i) {
		bv.b[bi] ^= 0x1 << uint8(i%8)
	}
}

// Bytes retrieves the underlying bytes of the bitvector
func (bv *BitVector) Bytes() []byte {
	return bv.b
}
