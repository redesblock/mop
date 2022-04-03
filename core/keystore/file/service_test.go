package file_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/redesblock/hop/core/keystore/file"
	"github.com/redesblock/hop/core/keystore/test"
)

func TestService(t *testing.T) {
	dir, err := ioutil.TempDir("", "hop-keystore-file-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	test.Service(t, file.New(dir))
}
