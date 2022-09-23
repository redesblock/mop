package mock_test

import (
	"testing"

	"github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/statestore/test"
	"github.com/redesblock/mop/core/storage"
)

func TestMockStateStore(t *testing.T) {
	test.Run(t, func(t *testing.T) storage.StateStorer {
		return mock.NewStateStore()
	})
}
