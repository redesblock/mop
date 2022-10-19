package mem_test

import (
	"testing"

	"github.com/redesblock/mop/core/keystore/mem"
	"github.com/redesblock/mop/core/keystore/test"
)

func TestService(t *testing.T) {
	test.Service(t, mem.New())
}
