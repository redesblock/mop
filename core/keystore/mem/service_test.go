package mem_test

import (
	"testing"

	"github.com/redesblock/hop/core/keystore/mem"
	"github.com/redesblock/hop/core/keystore/test"
)

func TestService(t *testing.T) {
	test.Service(t, mem.New())
}
