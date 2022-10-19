package file_test

import (
	"testing"

	"github.com/redesblock/mop/core/keystore/file"
	"github.com/redesblock/mop/core/keystore/test"
)

func TestService(t *testing.T) {
	dir := t.TempDir()

	test.Service(t, file.New(dir))
}
