// Package loadsave provides lightweight persistence abstraction
// for manifest operations.
package loadsave

import (
	"bytes"
	"context"
	"errors"

	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/joiner"
	"github.com/redesblock/hop/core/file/pipeline"
	"github.com/redesblock/hop/core/file/pipeline/builder"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

var readonlyLoadsaveError = errors.New("readonly manifest loadsaver")

type PutGetter interface {
	storage.Putter
	storage.Getter
}

// loadSave is needed for manifest operations and provides
// simple wrapping over load and save operations using file
// package abstractions. use with caution since Loader will
// load all of the subtrie of a given hash in memory.
type loadSave struct {
	storer     PutGetter
	pipelineFn func() pipeline.Interface
}

// New returns a new read-write load-saver.
func New(storer PutGetter, pipelineFn func() pipeline.Interface) file.LoadSaver {
	return &loadSave{
		storer:     storer,
		pipelineFn: pipelineFn,
	}
}

// NewReadonly returns a new read-only load-saver
// which will error on write.
func NewReadonly(storer PutGetter) file.LoadSaver {
	return &loadSave{
		storer: storer,
	}
}

func (ls *loadSave) Load(ctx context.Context, ref []byte) ([]byte, error) {
	j, _, err := joiner.New(ctx, ls.storer, swarm.NewAddress(ref))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ls *loadSave) Save(ctx context.Context, data []byte) ([]byte, error) {
	if ls.pipelineFn == nil {
		return nil, readonlyLoadsaveError
	}

	pipe := ls.pipelineFn()
	address, err := builder.FeedPipeline(ctx, pipe, bytes.NewReader(data))
	if err != nil {
		return swarm.ZeroAddress.Bytes(), err
	}

	return address.Bytes(), nil
}
