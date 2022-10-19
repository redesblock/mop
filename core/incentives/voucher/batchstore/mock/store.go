package mock

import (
	"bytes"
	"errors"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/incentives/voucher/batchstore"
	"github.com/redesblock/mop/core/storer/storage"
)

var _ voucher.Storer = (*BatchStore)(nil)

// BatchStore is a mock BatchStorer
type BatchStore struct {
	rs                *voucher.ReserveState
	cs                *voucher.ChainState
	id                []byte
	batch             *voucher.Batch
	getErr            error
	getErrDelayCnt    int
	updateErr         error
	updateErrDelayCnt int
	resetCallCount    int

	radiusSetter voucher.StorageRadiusSetter

	existsFn func([]byte) (bool, error)

	mtx sync.Mutex
}

func (bs *BatchStore) SetBatchExpiryHandler(eh voucher.BatchExpiryHandler) {}

// Option is an option passed to New.
type Option func(*BatchStore)

// New creates a new mock BatchStore
func New(opts ...Option) *BatchStore {
	bs := &BatchStore{}
	bs.cs = &voucher.ChainState{}
	bs.rs = &voucher.ReserveState{}
	for _, o := range opts {
		o(bs)
	}
	return bs
}

// WithReserveState will set the initial reservestate in the ChainStore mock.
func WithReserveState(rs *voucher.ReserveState) Option {
	return func(bs *BatchStore) {
		bs.rs = rs
	}
}

// WithChainState will set the initial chainstate in the ChainStore mock.
func WithChainState(cs *voucher.ChainState) Option {
	return func(bs *BatchStore) {
		bs.cs = cs
	}
}

// WithGetErr will set the get error returned by the ChainStore mock. The error
// will be returned on each subsequent call after delayCnt calls to Get have
// been made.
func WithGetErr(err error, delayCnt int) Option {
	return func(bs *BatchStore) {
		bs.getErr = err
		bs.getErrDelayCnt = delayCnt
	}
}

// WithUpdateErr will set the put error returned by the ChainStore mock.
// The error will be returned on each subsequent call after delayCnt
// calls to Update have been made.
func WithUpdateErr(err error, delayCnt int) Option {
	return func(bs *BatchStore) {
		bs.updateErr = err
		bs.updateErrDelayCnt = delayCnt
	}
}

// WithBatch will set batch to the one provided by user.
// This will be returned in the next Get.
func WithBatch(b *voucher.Batch) Option {
	return func(bs *BatchStore) {
		bs.batch = b
		bs.id = b.ID
	}
}

func WithExistsFunc(f func([]byte) (bool, error)) Option {
	return func(bs *BatchStore) {
		bs.existsFn = f
	}
}

func WithAcceptAllExistsFunc() Option {
	return func(bs *BatchStore) {
		bs.existsFn = func(_ []byte) (bool, error) {
			return true, nil
		}
	}
}

// Get mocks the Get method from the BatchStore.
func (bs *BatchStore) Get(id []byte) (*voucher.Batch, error) {
	if bs.getErr != nil {
		if bs.getErrDelayCnt == 0 {
			return nil, bs.getErr
		}
		bs.getErrDelayCnt--
	}
	exists, err := bs.Exists(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, storage.ErrNotFound
	}
	return bs.batch, nil
}

// Iterate mocks the Iterate method from the BatchStore
func (bs *BatchStore) Iterate(f func(*voucher.Batch) (bool, error)) error {
	if bs.batch == nil {
		return nil
	}
	_, err := f(bs.batch)
	return err
}

// Save mocks the Save method from the BatchStore.
func (bs *BatchStore) Save(batch *voucher.Batch) error {
	if bs.batch != nil {
		return errors.New("batch already taken")
	}
	bs.batch = batch
	bs.id = batch.ID
	return nil
}

// Update mocks the Update method from the BatchStore.
func (bs *BatchStore) Update(batch *voucher.Batch, newValue *big.Int, newDepth uint8) error {
	if bs.batch == nil || !bytes.Equal(batch.ID, bs.id) {
		return batchstore.ErrNotFound
	}
	if bs.updateErr != nil {
		if bs.updateErrDelayCnt == 0 {
			return bs.updateErr
		}
		bs.updateErrDelayCnt--
	}
	bs.batch = batch
	batch.Depth = newDepth
	batch.Value.Set(newValue)
	bs.id = batch.ID
	return nil
}

// GetChainState mocks the GetChainState method from the BatchStore
func (bs *BatchStore) GetChainState() *voucher.ChainState {
	return bs.cs
}

// PutChainState mocks the PutChainState method from the BatchStore
func (bs *BatchStore) PutChainState(cs *voucher.ChainState) error {
	if bs.updateErr != nil {
		if bs.updateErrDelayCnt == 0 {
			return bs.updateErr
		}
		bs.updateErrDelayCnt--
	}
	bs.cs = cs
	return nil
}

func (bs *BatchStore) GetReserveState() *voucher.ReserveState {
	bs.mtx.Lock()
	defer bs.mtx.Unlock()

	rs := new(voucher.ReserveState)
	if bs.rs != nil {
		rs.Radius = bs.rs.Radius
		rs.StorageRadius = bs.rs.StorageRadius
	}
	return rs
}

func (bs *BatchStore) SetStorageRadiusSetter(r voucher.StorageRadiusSetter) {
	bs.radiusSetter = r
}

func (bs *BatchStore) SetStorageRadius(f func(uint8) uint8) error {
	bs.mtx.Lock()
	defer bs.mtx.Unlock()

	bs.rs.StorageRadius = f(bs.rs.StorageRadius)
	if bs.radiusSetter != nil {
		bs.radiusSetter.SetStorageRadius(bs.rs.StorageRadius)
	}
	return nil
}

func (bs *BatchStore) Unreserve(_ voucher.UnreserveIteratorFn) error {
	panic("not implemented")
}

// Exists reports whether batch referenced by the give id exists.
func (bs *BatchStore) Exists(id []byte) (bool, error) {
	if bs.existsFn != nil {
		return bs.existsFn(id)
	}
	return bytes.Equal(bs.id, id), nil
}

func (bs *BatchStore) Reset() error {
	bs.resetCallCount++
	return nil
}

func (bs *BatchStore) ResetCalls() int {
	return bs.resetCallCount
}

type MockEventUpdater struct {
	inProgress bool
	err        error
}

func NewNotReady() *MockEventUpdater           { return &MockEventUpdater{inProgress: true} }
func NewWithError(err error) *MockEventUpdater { return &MockEventUpdater{inProgress: false, err: err} }

func (s *MockEventUpdater) GetSyncStatus() (isDone bool, err error) {
	return !s.inProgress, s.err
}
