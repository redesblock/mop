package retrieval

import (
	"sync"

	"github.com/redesblock/mop/core/flock"
)

type skipPeers struct {
	overdraftAddresses []flock.Address
	addresses          []flock.Address
	mu                 sync.Mutex
}

func newSkipPeers() *skipPeers {
	return &skipPeers{}
}

func (s *skipPeers) All() []flock.Address {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append(append(s.addresses[:0:0], s.addresses...), s.overdraftAddresses...)
}

func (s *skipPeers) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overdraftAddresses = []flock.Address{}
}

func (s *skipPeers) Add(address flock.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.addresses {
		if a.Equal(address) {
			return
		}
	}

	s.addresses = append(s.addresses, address)
}

func (s *skipPeers) AddOverdraft(address flock.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.overdraftAddresses {
		if a.Equal(address) {
			return
		}
	}

	s.overdraftAddresses = append(s.overdraftAddresses, address)
}

// Saturated function returns whether all skipped entries a permanently skipped for this skiplist
// Temporary entries are stored in the overdraftAddresses slice of the skiplist, so if that is empty, the function returns true
func (s *skipPeers) Saturated() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.overdraftAddresses) <= 0
}
