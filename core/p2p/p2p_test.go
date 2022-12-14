package p2p_test

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/redesblock/mop/core/p2p"
)

func TestNewClusterStreamName(t *testing.T) {
	want := "/cluster/hive/1.2.0/peers"
	got := p2p.NewClusterStreamName("hive", "1.2.0", "peers")

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestReachabilityStatus_String(t *testing.T) {
	mapping := map[string]string{
		p2p.ReachabilityStatusUnknown.String(): network.ReachabilityUnknown.String(),
		p2p.ReachabilityStatusPrivate.String(): network.ReachabilityPrivate.String(),
		p2p.ReachabilityStatusPublic.String():  network.ReachabilityPublic.String(),
	}
	for have, want := range mapping {
		if have != want {
			t.Fatalf("have reachability status string %q; want %q", have, want)
		}
	}
}
