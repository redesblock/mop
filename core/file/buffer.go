package file

import (
	"io"

	"github.com/redesblock/hop/core/swarm"
)

const (
	maxBufferSize = swarm.ChunkSize * 2
)

// ChunkPipe ensures that only the last read is smaller than the chunk size,
// regardless of size of individual writes.
type ChunkPipe struct {
	io.ReadCloser
	writer io.WriteCloser
	data   []byte
	cursor int
}

// Creates a new ChunkPipe
func NewChunkPipe() io.ReadWriteCloser {
	r, w := io.Pipe()
	return &ChunkPipe{
		ReadCloser: r,
		writer:     w,
		data:       make([]byte, maxBufferSize),
	}
}

// Read implements io.Reader
func (c *ChunkPipe) Read(b []byte) (int, error) {
	return c.ReadCloser.Read(b)
}

// Writer implements io.Writer
func (c *ChunkPipe) Write(b []byte) (int, error) {
	copy(c.data[c.cursor:], b)
	c.cursor += len(b)
	if c.cursor >= swarm.ChunkSize {
		_, err := c.writer.Write(c.data[:swarm.ChunkSize])
		if err != nil {
			return len(b), err
		}
		c.cursor -= swarm.ChunkSize
		copy(c.data, c.data[swarm.ChunkSize:])
	}
	return len(b), nil
}

// Closer implements io.Closer
func (c *ChunkPipe) Close() error {
	if c.cursor > 0 {
		_, err := c.writer.Write(c.data[:c.cursor])
		if err != nil {
			return err
		}
	}
	return c.writer.Close()
}
