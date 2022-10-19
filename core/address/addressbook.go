package address

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/storer/storage"
)

const keyPrefix = "addressbook_entry_"

var _ Interface = (*store)(nil)

var ErrNotFound = errors.New("address: not found")

// Interface is the AddressBook interface.
type Interface interface {
	GetPutter
	Remover
	// Overlays returns a list of all overlay addresses saved in address.
	Overlays() ([]cluster.Address, error)
	// IterateOverlays exposes overlays in a form of an iterator.
	IterateOverlays(func(cluster.Address) (bool, error)) error
	// Addresses returns a list of all Address-es saved in address.
	Addresses() ([]Address, error)
}

type GetPutter interface {
	Getter
	Putter
}

type Getter interface {
	// Get returns pointer to saved Address for requested overlay address.
	Get(overlay cluster.Address) (addr *Address, err error)
}

type Putter interface {
	// Put saves relation between peer overlay address and Address address.
	Put(overlay cluster.Address, addr Address) (err error)
}

type Remover interface {
	// Remove removes overlay address.
	Remove(overlay cluster.Address) error
}

type store struct {
	store storage.StateStorer
}

// New creates new address for state storer.
func New(storer storage.StateStorer) Interface {
	return &store{
		store: storer,
	}
}

func (s *store) Get(overlay cluster.Address) (*Address, error) {
	key := keyPrefix + overlay.String()
	v := &Address{}
	err := s.store.Get(key, &v)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, ErrNotFound
		}

		return nil, err
	}
	return v, nil
}

func (s *store) Put(overlay cluster.Address, addr Address) (err error) {
	key := keyPrefix + overlay.String()
	return s.store.Put(key, &addr)
}

func (s *store) Remove(overlay cluster.Address) error {
	return s.store.Delete(keyPrefix + overlay.String())
}

func (s *store) IterateOverlays(cb func(cluster.Address) (bool, error)) error {
	return s.store.Iterate(keyPrefix, func(key, _ []byte) (stop bool, err error) {
		k := string(key)
		if !strings.HasPrefix(k, keyPrefix) {
			return true, nil
		}
		split := strings.SplitAfter(k, keyPrefix)
		if len(split) != 2 {
			return true, fmt.Errorf("invalid overlay key: %s", k)
		}
		addr, err := cluster.ParseHexAddress(split[1])
		if err != nil {
			return true, err
		}
		stop, err = cb(addr)
		if err != nil {
			return true, err
		}
		if stop {
			return true, nil
		}
		return false, nil
	})
}

func (s *store) Overlays() (overlays []cluster.Address, err error) {
	err = s.IterateOverlays(func(addr cluster.Address) (stop bool, err error) {
		overlays = append(overlays, addr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return overlays, nil
}

func (s *store) Addresses() (addresses []Address, err error) {
	err = s.store.Iterate(keyPrefix, func(_, value []byte) (stop bool, err error) {
		entry := &Address{}
		err = entry.UnmarshalJSON(value)
		if err != nil {
			return true, err
		}

		addresses = append(addresses, *entry)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return addresses, nil
}
