package postage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/redesblock/mop/core/storage"
)

const (
	postagePrefix = "postage"
	// blockThreshold is used to allow threshold no of blocks to be synced before a
	// batch is usable.
	blockThreshold = 10
)

var (
	// ErrNotFound is the error returned when issuer with given batch ID does not exist.
	ErrNotFound = errors.New("not found")
	// ErrNotUsable is the error returned when issuer with given batch ID is not usable.
	ErrNotUsable = errors.New("not usable")
)

// Service is the postage service interface.
type Service interface {
	Add(*StampIssuer) error
	StampIssuers() []*StampIssuer
	GetStampIssuer([]byte) (*StampIssuer, error)
	IssuerUsable(*StampIssuer) bool
	BatchEventListener
	io.Closer
}

// service handles postage batches
// stores the active batches.
type service struct {
	lock         sync.Mutex
	store        storage.StateStorer
	postageStore Storer
	chainID      int64
	issuers      []*StampIssuer
}

// NewService constructs a new Service.
func NewService(store storage.StateStorer, postageStore Storer, chainID int64) (Service, error) {
	s := &service{
		store:        store,
		postageStore: postageStore,
		chainID:      chainID,
	}

	n := 0
	if err := s.store.Iterate(s.key(), func(_, _ []byte) (stop bool, err error) {
		n++
		return false, nil
	}); err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		st := &StampIssuer{}
		err := s.store.Get(s.keyForIndex(i), st)
		if err != nil {
			return nil, err
		}
		_ = s.add(st)
	}
	return s, nil
}

// Add adds a stamp issuer to the active issuers.
func (ps *service) Add(st *StampIssuer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.add(st) {
		if err := ps.store.Put(ps.keyForIndex(len(ps.issuers)), st); err != nil {
			return err
		}
	}

	return nil
}

// add adds a stamp issuer to the active issuers and returns false if it is already present.
func (ps *service) add(st *StampIssuer) bool {

	for _, v := range ps.issuers {
		if bytes.Equal(st.data.BatchID, v.data.BatchID) {
			return false
		}
	}
	ps.issuers = append(ps.issuers, st)

	return true
}

// HandleCreate implements the BatchEventListener interface. This is fired on receiving
// a batch creation event from the blockchain listener to ensure that if a stamp
// issuer was not created initially, we will create it here.
func (ps *service) HandleCreate(b *Batch) error {
	return ps.Add(NewStampIssuer(
		"recovered",
		string(b.Owner),
		b.ID,
		b.Value,
		b.Depth,
		b.BucketDepth,
		b.Start,
		b.Immutable,
	))
}

// HandleTopUp implements the BatchEventListener interface. This is fired on receiving
// a batch topup event from the blockchain to update stampissuer details
func (ps *service) HandleTopUp(batchID []byte, newValue *big.Int) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, v := range ps.issuers {
		if bytes.Equal(batchID, v.data.BatchID) {
			if newValue.Cmp(v.data.BatchAmount) > 0 {
				v.data.BatchAmount = newValue
			}
			return
		}
	}
}

func (ps *service) HandleDepthIncrease(batchID []byte, newDepth uint8, normalisedBalance *big.Int) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, v := range ps.issuers {
		if bytes.Equal(batchID, v.data.BatchID) {
			if newDepth > v.data.BatchDepth {
				v.data.BatchDepth = newDepth
				v.data.BatchAmount = normalisedBalance
			}
			return
		}
	}
}

// StampIssuers returns the currently active stamp issuers.
func (ps *service) StampIssuers() []*StampIssuer {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	return ps.issuers
}

func (ps *service) IssuerUsable(st *StampIssuer) bool {
	cs := ps.postageStore.GetChainState()

	// this checks atleast threshold blocks are seen on the blockchain after
	// the batch creation, before we start using a stamp issuer. The threshold
	// is meant to allow enough time for upstream peers to see the batch and
	// hence validate the stamps issued
	if cs.Block < st.data.BlockNumber || (cs.Block-st.data.BlockNumber) < blockThreshold {
		return false
	}
	return true
}

// GetStampIssuer finds a stamp issuer by batch ID.
func (ps *service) GetStampIssuer(batchID []byte) (*StampIssuer, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	for _, st := range ps.issuers {
		if bytes.Equal(batchID, st.data.BatchID) {
			if !ps.IssuerUsable(st) {
				return nil, ErrNotUsable
			}
			return st, nil
		}
	}
	return nil, ErrNotFound
}

// Close saves all the active stamp issuers to statestore.
func (ps *service) Close() error {
	for i, st := range ps.issuers {
		if err := ps.store.Put(ps.keyForIndex(i), st); err != nil {
			return err
		}
	}
	return nil
}

// keyForIndex returns the statestore key for an issuer
func (ps *service) keyForIndex(i int) string {
	return fmt.Sprintf(ps.key()+"%d", i)
}

// key returns the statestore base key for an issuer
// to disambiguate batches on different chains, chainID is part of the key
func (ps *service) key() string {
	return fmt.Sprintf(postagePrefix+"%d", ps.chainID)
}
