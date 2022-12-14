// Package tags provides the implementation for
// upload progress tracking.
package tags

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/storer/storage"
)

// loggerName is the tree path name of the logger for this package.
const loggerName = "tags"

const (
	maxPage      = 1000 // hard limit of page size
	tagKeyPrefix = "tags_"
)

var (
	TagUidFunc  = rand.Uint32
	ErrNotFound = errors.New("tag not found")
)

// Tags hold tag information indexed by a unique random uint32
type Tags struct {
	tags       *sync.Map
	stateStore storage.StateStorer
	logger     log.Logger
	rand       *rand.Rand
	randM      sync.Mutex
}

// NewTags creates a tags object
func NewTags(stateStore storage.StateStorer, logger log.Logger) *Tags {

	return &Tags{
		tags:       &sync.Map{},
		stateStore: stateStore,
		logger:     logger.WithName(loggerName).Register(),
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (ts *Tags) TagUidFunc() uint32 {
	ts.randM.Lock()
	defer ts.randM.Unlock()
	return ts.rand.Uint32()
}

// Create creates a new tag, stores it by a not yet in use UID and returns it
func (ts *Tags) Create(total int64) (*Tag, error) {

	exists := true

	var uid uint32

	for exists {
		uid = ts.TagUidFunc()
		if _, loaded := ts.tags.Load(uid); !loaded {
			exists = false
		}
	}

	t := NewTag(context.Background(), uid, total, nil, ts.stateStore, ts.logger)

	if _, loaded := ts.tags.LoadOrStore(t.Uid, t); loaded {
		return nil, errExists
	}

	return t, nil
}

// All returns all existing tags in Tags' chainsync.Map
// Note that tags are returned in no particular order
func (ts *Tags) All() (t []*Tag) {
	ts.tags.Range(func(k, v interface{}) bool {
		t = append(t, v.(*Tag))

		return true
	})

	return t
}

// Get returns the underlying tag for the uid or an error if not found
func (ts *Tags) Get(uid uint32) (*Tag, error) {
	t, ok := ts.tags.Load(uid)
	if !ok {
		// see if the tag is present in the store
		// if yes, load it in to the memory
		ta, err := ts.getTagFromStore(uid)
		if err != nil {
			return nil, ErrNotFound
		}
		ts.tags.LoadOrStore(ta.Uid, ta)
		return ta, nil
	}
	return t.(*Tag), nil
}

// GetByAddress returns the latest underlying tag for the address or an error if not found
func (ts *Tags) GetByAddress(address cluster.Address) (*Tag, error) {
	var t *Tag
	var lastTime time.Time
	ts.tags.Range(func(key interface{}, value interface{}) bool {
		rcvdTag := value.(*Tag)
		if rcvdTag.Address.Equal(address) && rcvdTag.StartedAt.After(lastTime) {
			t = rcvdTag
			lastTime = rcvdTag.StartedAt
		}
		return true
	})

	if t == nil {
		return nil, ErrNotFound
	}
	return t, nil
}

// Range exposes chainsync.Map's iterator
func (ts *Tags) Range(fn func(k, v interface{}) bool) {
	ts.tags.Range(fn)
}

func (ts *Tags) Delete(k interface{}) {
	ts.tags.Delete(k)

	// k is a uint32, try to create the tag key and remove
	// from statestore
	if uid, ok := k.(uint32); ok && uid != 0 {
		key := tagKey(uid)
		_ = ts.stateStore.Delete(key)
	}
}

func (ts *Tags) MarshalJSON() (out []byte, err error) {
	m := make(map[string]*Tag)
	ts.Range(func(k, v interface{}) bool {
		val := v.(*Tag)

		// don't persist tags which were already done
		if !val.Done(StateSynced) {
			key := strconv.FormatInt(int64(k.(uint32)), 10)
			m[key] = val
		}
		return true
	})
	return json.Marshal(m)
}

func (ts *Tags) UnmarshalJSON(value []byte) error {
	m := make(map[string]*Tag)
	err := json.Unmarshal(value, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		key, err := strconv.ParseUint(k, 10, 32)
		if err != nil {
			return err
		}

		// prevent a condition where a chunk was sent before shutdown
		// and the node was turned off before the receipt was received
		v.Sent = v.Synced

		ts.tags.Store(key, v)
	}

	return err
}

func (ts *Tags) ListAll(ctx context.Context, offset, limit int) (t []*Tag, err error) {
	if limit > maxPage {
		limit = maxPage
	}

	// range chainsync.Map first
	allTags := ts.All()
	sort.Slice(allTags, func(i, j int) bool { return allTags[i].Uid < allTags[j].Uid })
	for _, tag := range allTags {
		if offset > 0 {
			offset--
			continue
		}

		t = append(t, tag)

		limit--

		if limit == 0 {
			break
		}
	}

	if limit == 0 {
		return
	}

	// and then from statestore
	err = ts.stateStore.Iterate(tagKeyPrefix, func(key, value []byte) (stop bool, err error) {
		if offset > 0 {
			offset--
			return false, nil
		}

		var ta *Tag
		ta, err = decodeTagValueFromStore(value)
		if err != nil {
			return true, err
		}

		if _, ok := ts.tags.Load(ta.Uid); ok {
			// tag was already returned from chainsync.Map
			return false, nil
		}

		t = append(t, ta)

		limit--

		if limit == 0 {
			return true, nil
		}

		return false, nil
	})

	return t, err
}

func decodeTagValueFromStore(value []byte) (*Tag, error) {
	var data []byte
	err := json.Unmarshal(value, &data)
	if err != nil {
		return nil, err
	}
	var ta Tag
	err = ta.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &ta, nil
}

// getTagFromStore get a given tag from the state store.
func (ts *Tags) getTagFromStore(uid uint32) (*Tag, error) {
	key := tagKey(uid)
	var data []byte
	err := ts.stateStore.Get(key, &data)
	if err != nil {
		return nil, err
	}
	var ta Tag
	err = ta.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &ta, nil
}

// Close is called when the node goes down. This is when all the tags in memory is persisted.
func (ts *Tags) Close() (err error) {
	loggerV1 := ts.logger.V(1).Register()
	// store all the tags in memory
	tags := ts.All()
	for _, t := range tags {
		loggerV1.Debug("updating tag", "tag_uid", t.Uid)
		err := t.saveTag()
		if err != nil {
			return err
		}
	}
	return nil
}

func tagKey(uid uint32) string {
	return tagKeyPrefix + strconv.FormatInt(int64(uid), 10)
}
