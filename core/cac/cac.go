package cac

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/redesblock/mop/core/bmtpool"
	"github.com/redesblock/mop/core/flock"
)

var (
	errTooShortChunkData = errors.New("short chunk data")
	errTooLargeChunkData = errors.New("data too large")
)

// New creates a new content address chunk by initializing a span and appending the data to it.
func New(data []byte) (flock.Chunk, error) {
	dataLength := len(data)
	if dataLength > flock.ChunkSize {
		return nil, errTooLargeChunkData
	}

	if dataLength == 0 {
		return nil, errTooShortChunkData
	}

	span := make([]byte, flock.SpanSize)
	binary.LittleEndian.PutUint64(span, uint64(dataLength))
	return newWithSpan(data, span)
}

// NewWithDataSpan creates a new chunk assuming that the span precedes the actual data.
func NewWithDataSpan(data []byte) (flock.Chunk, error) {
	dataLength := len(data)
	if dataLength > flock.ChunkSize+flock.SpanSize {
		return nil, errTooLargeChunkData
	}

	if dataLength < flock.SpanSize {
		return nil, errTooShortChunkData
	}
	return newWithSpan(data[flock.SpanSize:], data[:flock.SpanSize])
}

// newWithSpan creates a new chunk prepending the given span to the data.
func newWithSpan(data, span []byte) (flock.Chunk, error) {
	h := hasher(data)
	hash, err := h(span)
	if err != nil {
		return nil, err
	}

	cdata := make([]byte, len(data)+len(span))
	copy(cdata[:flock.SpanSize], span)
	copy(cdata[flock.SpanSize:], data)
	return flock.NewChunk(flock.NewAddress(hash), cdata), nil
}

// hasher is a helper function to hash a given data based on the given span.
func hasher(data []byte) func([]byte) ([]byte, error) {
	return func(span []byte) ([]byte, error) {
		hasher := bmtpool.Get()
		defer bmtpool.Put(hasher)

		hasher.SetHeader(span)
		if _, err := hasher.Write(data); err != nil {
			return nil, err
		}
		return hasher.Hash(nil)
	}
}

// Valid checks whether the given chunk is a valid content-addressed chunk.
func Valid(c flock.Chunk) bool {
	data := c.Data()
	if len(data) < flock.SpanSize {
		return false
	}

	if len(data) > flock.ChunkSize+flock.SpanSize {
		return false
	}

	h := hasher(data[flock.SpanSize:])
	hash, _ := h(data[:flock.SpanSize])
	return bytes.Equal(hash, c.Address().Bytes())
}
