package p2p_test

import (
	"testing"

	"github.com/redesblock/hop/core/p2p"
)

func TestNewSwarmStreamName(t *testing.T) {
	want := "/swarm/hive/1.2.0/peers"
	got := p2p.NewSwarmStreamName("hive", "1.2.0", "peers")

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
