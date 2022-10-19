package cmd_test

import (
	"bytes"
	"testing"

	ver "github.com/redesblock/mop"
	"github.com/redesblock/mop/cmd/mop/cmd"
)

func TestVersionCmd(t *testing.T) {
	var outputBuf bytes.Buffer
	if err := newCommand(t,
		cmd.WithArgs("version"),
		cmd.WithOutput(&outputBuf),
	).Execute(); err != nil {
		t.Fatal(err)
	}

	want := ver.Version + "\n"
	got := outputBuf.String()
	if got != want {
		t.Errorf("got output %q, want %q", got, want)
	}
}
