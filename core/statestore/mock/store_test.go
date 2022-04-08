package mock_test

import (
	"testing"

	"github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/statestore/test"
	"github.com/redesblock/hop/core/storage"
)

func TestMockStateStore(t *testing.T) {
	test.Run(t, func(t *testing.T) (storage.StateStorer, func()) {
		return mock.NewStateStore(), func() {}
	})
}
