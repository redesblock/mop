package hive_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ma "github.com/multiformats/go-multiaddr"

	ab "github.com/redesblock/mop/core/address"
	mop "github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/cluster/test"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p/protobuf"
	"github.com/redesblock/mop/core/p2p/streamtest"
	"github.com/redesblock/mop/core/protocol/hive"
	"github.com/redesblock/mop/core/protocol/hive/pb"
	"github.com/redesblock/mop/core/storer/statestore/mock"
)

var (
	tx    = common.HexToHash("0x2").Bytes()
	block = common.HexToHash("0x1").Bytes()
)

func TestHandlerRateLimit(t *testing.T) {
	logger := log.Noop
	statestore := mock.NewStateStore()
	addressbook := ab.New(statestore)
	networkID := uint64(1)

	addressbookclean := ab.New(mock.NewStateStore())

	// new recorder for handling Ping
	streamer := streamtest.New()
	// create a hive server that handles the incoming stream
	server, _ := hive.New(streamer, addressbookclean, networkID, false, true, logger)

	serverAddress := test.RandomAddress()

	// setup the stream recorder to record stream data
	serverRecorder := streamtest.New(
		streamtest.WithProtocols(server.Protocol()),
		streamtest.WithBaseAddr(serverAddress),
	)

	peers := make([]cluster.Address, hive.LimitBurst+1)
	for i := range peers {

		underlay, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/" + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		pk, err := crypto.GenerateSecp256k1Key()
		if err != nil {
			t.Fatal(err)
		}
		signer := crypto.NewDefaultSigner(pk)
		overlay, err := crypto.NewOverlayAddress(pk.PublicKey, networkID, block)
		if err != nil {
			t.Fatal(err)
		}
		mopAddress, err := mop.NewAddress(signer, underlay, overlay, networkID, tx)
		if err != nil {
			t.Fatal(err)
		}

		err = addressbook.Put(mopAddress.Overlay, *mopAddress)
		if err != nil {
			t.Fatal(err)
		}
		peers[i] = mopAddress.Overlay
	}

	// create a hive client that will do broadcast
	client, _ := hive.New(serverRecorder, addressbook, networkID, false, true, logger)
	err := client.BroadcastPeers(context.Background(), serverAddress, peers...)
	if err != nil {
		t.Fatal(err)
	}

	rec, err := serverRecorder.Records(serverAddress, "hive", "1.0.0", "peers")
	if err != nil {
		t.Fatal(err)
	}

	lastRec := rec[len(rec)-1]

	if lastRec.Err() != nil {
		t.Fatal("want nil error")
	}
}

func TestBroadcastPeers(t *testing.T) {
	logger := log.Noop
	rand.Seed(time.Now().UnixNano())
	statestore := mock.NewStateStore()
	addressbook := ab.New(statestore)
	networkID := uint64(1)

	// populate all expected and needed random resources for 2 full batches
	// tests cases that uses fewer resources can use sub-slices of this data
	var mopAddresses []mop.Address
	var overlays []cluster.Address
	var wantMsgs []pb.Peers

	for i := 0; i < 2; i++ {
		wantMsgs = append(wantMsgs, pb.Peers{Peers: []*pb.MopAddress{}})
	}

	for i := 0; i < 2*hive.MaxBatchSize; i++ {
		base := "/ip4/127.0.0.1/udp/"
		if i == 2*hive.MaxBatchSize-1 {
			base = "/ip4/1.1.1.1/udp/" // The last underlay has public address.
		}
		underlay, err := ma.NewMultiaddr(base + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		pk, err := crypto.GenerateSecp256k1Key()
		if err != nil {
			t.Fatal(err)
		}
		signer := crypto.NewDefaultSigner(pk)
		overlay, err := crypto.NewOverlayAddress(pk.PublicKey, networkID, block)
		if err != nil {
			t.Fatal(err)
		}
		mopAddress, err := mop.NewAddress(signer, underlay, overlay, networkID, tx)
		if err != nil {
			t.Fatal(err)
		}

		mopAddresses = append(mopAddresses, *mopAddress)
		overlays = append(overlays, mopAddress.Overlay)
		err = addressbook.Put(mopAddress.Overlay, *mopAddress)
		if err != nil {
			t.Fatal(err)
		}

		wantMsgs[i/hive.MaxBatchSize].Peers = append(wantMsgs[i/hive.MaxBatchSize].Peers, &pb.MopAddress{
			Overlay:     mopAddresses[i].Overlay.Bytes(),
			Underlay:    mopAddresses[i].Underlay.Bytes(),
			Signature:   mopAddresses[i].Signature,
			Transaction: tx,
		})
	}

	testCases := map[string]struct {
		addresee          cluster.Address
		peers             []cluster.Address
		wantMsgs          []pb.Peers
		wantOverlays      []cluster.Address
		wantMopAddresses  []mop.Address
		allowPrivateCIDRs bool
		pingErr           func(addr ma.Multiaddr) (time.Duration, error)
	}{
		"OK - single record": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             []cluster.Address{overlays[0]},
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers[:1]}},
			wantOverlays:      []cluster.Address{overlays[0]},
			wantMopAddresses:  []mop.Address{mopAddresses[0]},
			allowPrivateCIDRs: true,
		},
		"OK - single batch - multiple records": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             overlays[:15],
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers[:15]}},
			wantOverlays:      overlays[:15],
			wantMopAddresses:  mopAddresses[:15],
			allowPrivateCIDRs: true,
		},
		"OK - single batch - max number of records": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             overlays[:hive.MaxBatchSize],
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers[:hive.MaxBatchSize]}},
			wantOverlays:      overlays[:hive.MaxBatchSize],
			wantMopAddresses:  mopAddresses[:hive.MaxBatchSize],
			allowPrivateCIDRs: true,
		},
		"OK - multiple batches": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             overlays[:hive.MaxBatchSize+10],
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers}, {Peers: wantMsgs[1].Peers[:10]}},
			wantOverlays:      overlays[:hive.MaxBatchSize+10],
			wantMopAddresses:  mopAddresses[:hive.MaxBatchSize+10],
			allowPrivateCIDRs: true,
		},
		"OK - multiple batches - max number of records": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             overlays[:2*hive.MaxBatchSize],
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers}, {Peers: wantMsgs[1].Peers}},
			wantOverlays:      overlays[:2*hive.MaxBatchSize],
			wantMopAddresses:  mopAddresses[:2*hive.MaxBatchSize],
			allowPrivateCIDRs: true,
		},
		"OK - single batch - skip ping failures": {
			addresee:          cluster.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:             overlays[:15],
			wantMsgs:          []pb.Peers{{Peers: wantMsgs[0].Peers[:15]}},
			wantOverlays:      overlays[:10],
			wantMopAddresses:  mopAddresses[:10],
			allowPrivateCIDRs: true,
			pingErr: func(addr ma.Multiaddr) (rtt time.Duration, err error) {
				for _, v := range mopAddresses[10:15] {
					if v.Underlay.Equal(addr) {
						return rtt, errors.New("ping failure")
					}
				}
				return rtt, nil
			},
		},
		"Ok - don't advertise private CIDRs": {
			addresee:          overlays[len(overlays)-1],
			peers:             overlays[:15],
			wantMsgs:          []pb.Peers{{}},
			wantOverlays:      nil,
			wantMopAddresses:  nil,
			allowPrivateCIDRs: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			addressbookclean := ab.New(mock.NewStateStore())

			// new recorder for handling Ping
			var streamer *streamtest.Recorder
			if tc.pingErr != nil {
				streamer = streamtest.New(streamtest.WithPingErr(tc.pingErr))
			} else {
				streamer = streamtest.New()
			}
			// create a hive server that handles the incoming stream
			server, _ := hive.New(streamer, addressbookclean, networkID, false, true, logger)

			// setup the stream recorder to record stream data
			recorder := streamtest.New(
				streamtest.WithProtocols(server.Protocol()),
			)

			// create a hive client that will do broadcast
			client, _ := hive.New(recorder, addressbook, networkID, false, tc.allowPrivateCIDRs, logger)
			if err := client.BroadcastPeers(context.Background(), tc.addresee, tc.peers...); err != nil {
				t.Fatal(err)
			}

			// get a record for this stream
			records, err := recorder.Records(tc.addresee, "hive", "1.0.0", "peers")
			if err != nil {
				t.Fatal(err)
			}
			if l := len(records); l != len(tc.wantMsgs) {
				t.Fatalf("got %v records, want %v", l, len(tc.wantMsgs))
			}

			// there is a one record per batch (wantMsg)
			for i, record := range records {
				messages, err := readAndAssertPeersMsgs(record.In(), 1)
				if err != nil {
					t.Fatal(err)
				}

				if fmt.Sprint(messages[0]) != fmt.Sprint(tc.wantMsgs[i]) {
					t.Errorf("Messages got %v, want %v", messages, tc.wantMsgs)
				}
			}

			expectOverlaysEventually(t, addressbookclean, tc.wantOverlays)
			expectMopAddresessEventually(t, addressbookclean, tc.wantMopAddresses)
		})
	}
}

func expectOverlaysEventually(t *testing.T, exporter ab.Interface, wantOverlays []cluster.Address) {
	var (
		overlays []cluster.Address
		err      error
		isIn     = func(a cluster.Address, addrs []cluster.Address) bool {
			for _, v := range addrs {
				if a.Equal(v) {
					return true
				}
			}
			return false
		}
	)

	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		overlays, err = exporter.Overlays()
		if err != nil {
			t.Fatal(err)
		}

		if len(overlays) == len(wantOverlays) {
			break
		}
	}
	if len(overlays) != len(wantOverlays) {
		debug.PrintStack()
		t.Fatal("timed out waiting for overlays")
	}

	for _, v := range wantOverlays {
		if !isIn(v, overlays) {
			t.Errorf("overlay %s expected but not found", v.String())
		}
	}

	if t.Failed() {
		t.Errorf("overlays got %v, want %v", overlays, wantOverlays)
	}
}

func expectMopAddresessEventually(t *testing.T, exporter ab.Interface, wantMopAddresses []mop.Address) {
	var (
		addresses []mop.Address
		err       error

		isIn = func(a mop.Address, addrs []mop.Address) bool {
			for _, v := range addrs {
				if a.Equal(&v) {
					return true
				}
			}
			return false
		}
	)

	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		addresses, err = exporter.Addresses()
		if err != nil {
			t.Fatal(err)
		}

		if len(addresses) == len(wantMopAddresses) {
			break
		}
	}
	if len(addresses) != len(wantMopAddresses) {
		debug.PrintStack()
		t.Fatal("timed out waiting for mop addresses")
	}

	for _, v := range wantMopAddresses {
		if !isIn(v, addresses) {
			t.Errorf("address %s expected but not found", v.Overlay.String())
		}
	}

	if t.Failed() {
		t.Errorf("mop addresses got %v, want %v", addresses, wantMopAddresses)
	}
}

func readAndAssertPeersMsgs(in []byte, expectedLen int) ([]pb.Peers, error) {
	messages, err := protobuf.ReadMessages(
		bytes.NewReader(in),
		func() protobuf.Message {
			return new(pb.Peers)
		},
	)

	if err != nil {
		return nil, err
	}

	if len(messages) != expectedLen {
		return nil, fmt.Errorf("got %v messages, want %v", len(messages), expectedLen)
	}

	var peers []pb.Peers
	for _, m := range messages {
		peers = append(peers, *m.(*pb.Peers))
	}

	return peers, nil
}
