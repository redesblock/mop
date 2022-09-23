package file_test

import (
	"os"
	"testing"

	"github.com/redesblock/mop/core/keystore/file"
	"github.com/redesblock/mop/core/keystore/test"
)

func TestService(t *testing.T) {
	dir, err := os.MkdirTemp("", "mop-keystore-file-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	test.Service(t, file.New(dir))
}
