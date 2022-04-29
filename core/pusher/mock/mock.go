package mock

import (
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

func (m *MockPusher) SendChunk(uid uint32) error {
	ta, err := m.tag.Get(uid)
	if err != nil {
		return err
	}
	ta.Inc(tags.StateSent)

	return nil
}

func (m *MockPusher) RcvdReceipt(uid uint32) error {
	ta, err := m.tag.Get(uid)
	if err != nil {
		return err
	}
	ta.Inc(tags.StateSynced)

	return nil
}
