package cac

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/redesblock/mop/core/cluster"
	bmtpool "github.com/redesblock/mop/core/util/bmt"
)

var (
	errTooShortChunkData = errors.New("short chunk data")
	errTooLargeChunkData = errors.New("data too large")
)

// New creates a new content address chunk by initializing a span and appending the data to it.
func New(data []byte) (cluster.Chunk, error) {
	dataLength := len(data)
	if dataLength > cluster.ChunkSize {
		return nil, errTooLargeChunkData
	}

	if dataLength == 0 {
		return nil, errTooShortChunkData
	}

	span := make([]byte, cluster.SpanSize)
	binary.LittleEndian.PutUint64(span, uint64(dataLength))
	return newWithSpan(data, span)
}

// NewWithDataSpan creates a new chunk assuming that the span precedes the actual data.
func NewWithDataSpan(data []byte) (cluster.Chunk, error) {
	dataLength := len(data)
	if dataLength > cluster.ChunkSize+cluster.SpanSize {
		return nil, errTooLargeChunkData
	}

	if dataLength < cluster.SpanSize {
		return nil, errTooShortChunkData
	}
	return newWithSpan(data[cluster.SpanSize:], data[:cluster.SpanSize])
}

// newWithSpan creates a new chunk prepending the given span to the data.
func newWithSpan(data, span []byte) (cluster.Chunk, error) {
	h := hasher(data)
	hash, err := h(span)
	if err != nil {
		return nil, err
	}

	cdata := make([]byte, len(data)+len(span))
	copy(cdata[:cluster.SpanSize], span)
	copy(cdata[cluster.SpanSize:], data)
	return cluster.NewChunk(cluster.NewAddress(hash), cdata), nil
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
func Valid(c cluster.Chunk) bool {
	data := c.Data()
	if len(data) < cluster.SpanSize {
		return false
	}

	if len(data) > cluster.ChunkSize+cluster.SpanSize {
		return false
	}

	h := hasher(data[cluster.SpanSize:])
	hash, _ := h(data[:cluster.SpanSize])
	return bytes.Equal(hash, c.Address().Bytes())
}
