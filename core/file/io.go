package file

import (
	"bytes"
	"errors"

	"io"
)

// simpleJoinerReadCloser wraps a byte slice in a io.ReadCloser implementation.
type simpleReadCloser struct {
	buffer io.Reader
	closed bool
}

func NewSimpleReadCloser(buffer []byte) io.ReadCloser {
	return &simpleReadCloser{
		buffer: bytes.NewBuffer(buffer),
	}
}

// Read implements io.Reader.
func (s *simpleReadCloser) Read(b []byte) (int, error) {
	if s.closed {
		return 0, errors.New("read on closed reader")
	}
	return s.buffer.Read(b)
}

// Close implements io.Closer.
func (s *simpleReadCloser) Close() error {
	if s.closed {
		return errors.New("close on already closed reader")
	}
	s.closed = true
	return nil
}
