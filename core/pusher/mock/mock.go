package mock

import (
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
)

type MockPusher struct {
	tag *tags.Tags
}

func NewMockPusher(tag *tags.Tags) *MockPusher {
	return &MockPusher{
		tag: tag,
	}
}

func (m *MockPusher) SendChunk(address swarm.Address) error {
	ta, err := m.tag.GetByAddress(address)
	if err != nil {
		return err
	}
	ta.Inc(tags.StateSent)

	return nil
}

func (m *MockPusher) RcvdReceipt(address swarm.Address) error {
	ta, err := m.tag.GetByAddress(address)
	if err != nil {
		return err
	}
	ta.Inc(tags.StateSynced)

	return nil
}
