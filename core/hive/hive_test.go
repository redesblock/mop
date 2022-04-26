package hive_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	ma "github.com/multiformats/go-multiaddr"

	ab "github.com/redesblock/hop/core/addressbook"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/hive"
	"github.com/redesblock/hop/core/hive/pb"
	"github.com/redesblock/hop/core/hop"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p/protobuf"
	"github.com/redesblock/hop/core/p2p/streamtest"
	"github.com/redesblock/hop/core/statestore/mock"
	"github.com/redesblock/hop/core/swarm"
)

func TestBroadcastPeers(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	logger := logging.New(ioutil.Discard, 0)
	statestore := mock.NewStateStore()
	addressbook := ab.New(statestore)
	networkID := uint64(1)

	// populate all expected and needed random resources for 2 full batches
	// tests cases that uses fewer resources can use sub-slices of this data
	var hopAddresses []hop.Address
	var overlays []swarm.Address
	var wantMsgs []pb.Peers

	for i := 0; i < 2; i++ {
		wantMsgs = append(wantMsgs, pb.Peers{Peers: []*pb.HopAddress{}})
	}

	for i := 0; i < 2*hive.MaxBatchSize; i++ {
		underlay, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/" + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		pk, err := crypto.GenerateSecp256k1Key()
		if err != nil {
			t.Fatal(err)
		}
		signer := crypto.NewDefaultSigner(pk)
		overlay := crypto.NewOverlayAddress(pk.PublicKey, networkID)
		hopAddr, err := hop.NewAddress(signer, underlay, overlay, networkID)
		if err != nil {
			t.Fatal(err)
		}

		hopAddresses = append(hopAddresses, *hopAddr)
		overlays = append(overlays, hopAddr.Overlay)
		err = addressbook.Put(hopAddr.Overlay, *hopAddr)
		if err != nil {
			t.Fatal(err)
		}
		wantMsgs[i/hive.MaxBatchSize].Peers = append(wantMsgs[i/hive.MaxBatchSize].Peers, &pb.HopAddress{Overlay: hopAddresses[i].Overlay.Bytes(), Underlay: hopAddresses[i].Underlay.Bytes(), Signature: hopAddresses[i].Signature})
	}

	testCases := map[string]struct {
		addresee         swarm.Address
		peers            []swarm.Address
		wantMsgs         []pb.Peers
		wantOverlays     []swarm.Address
		wanthopAddresses []hop.Address
	}{
		"OK - single record": {
			addresee:         swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:            []swarm.Address{overlays[0]},
			wantMsgs:         []pb.Peers{{Peers: wantMsgs[0].Peers[:1]}},
			wantOverlays:     []swarm.Address{overlays[0]},
			wanthopAddresses: []hop.Address{hopAddresses[0]},
		},
		"OK - single batch - multiple records": {
			addresee:         swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:            overlays[:15],
			wantMsgs:         []pb.Peers{{Peers: wantMsgs[0].Peers[:15]}},
			wantOverlays:     overlays[:15],
			wanthopAddresses: hopAddresses[:15],
		},
		"OK - single batch - max number of records": {
			addresee:         swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:            overlays[:hive.MaxBatchSize],
			wantMsgs:         []pb.Peers{{Peers: wantMsgs[0].Peers[:hive.MaxBatchSize]}},
			wantOverlays:     overlays[:hive.MaxBatchSize],
			wanthopAddresses: hopAddresses[:hive.MaxBatchSize],
		},
		"OK - multiple batches": {
			addresee:         swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:            overlays[:hive.MaxBatchSize+10],
			wantMsgs:         []pb.Peers{{Peers: wantMsgs[0].Peers}, {Peers: wantMsgs[1].Peers[:10]}},
			wantOverlays:     overlays[:hive.MaxBatchSize+10],
			wanthopAddresses: hopAddresses[:hive.MaxBatchSize+10],
		},
		"OK - multiple batches - max number of records": {
			addresee:         swarm.MustParseHexAddress("ca1e9f3938cc1425c6061b96ad9eb93e134dfe8734ad490164ef20af9d1cf59c"),
			peers:            overlays[:2*hive.MaxBatchSize],
			wantMsgs:         []pb.Peers{{Peers: wantMsgs[0].Peers}, {Peers: wantMsgs[1].Peers}},
			wantOverlays:     overlays[:2*hive.MaxBatchSize],
			wanthopAddresses: hopAddresses[:2*hive.MaxBatchSize],
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			addressbookclean := ab.New(mock.NewStateStore())

			// create a hive server that handles the incoming stream
			server := hive.New(hive.Options{
				Logger:      logger,
				AddressBook: addressbookclean,
				NetworkID:   networkID,
			})

			// setup the stream recorder to record stream data
			recorder := streamtest.New(
				streamtest.WithProtocols(server.Protocol()),
			)

			// create a hive client that will do broadcast
			client := hive.New(hive.Options{
				Streamer:    recorder,
				Logger:      logger,
				AddressBook: addressbook,
				NetworkID:   networkID,
			})

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
			expecthopAddresessEventually(t, addressbookclean, tc.wanthopAddresses)
		})
	}
}

func expectOverlaysEventually(t *testing.T, exporter ab.Interface, wantOverlays []swarm.Address) {
	for i := 0; i < 100; i++ {
		var stringOverlays []string
		var stringWantOverlays []string
		o, err := exporter.Overlays()
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range o {
			stringOverlays = append(stringOverlays, k.String())
		}

		for _, k := range wantOverlays {
			stringWantOverlays = append(stringWantOverlays, k.String())
		}

		sort.Strings(stringOverlays)
		sort.Strings(stringWantOverlays)
		if reflect.DeepEqual(stringOverlays, stringWantOverlays) {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	o, err := exporter.Overlays()
	if err != nil {
		t.Fatal(err)
	}

	t.Errorf("Overlays got %v, want %v", o, wantOverlays)
}

func expecthopAddresessEventually(t *testing.T, exporter ab.Interface, wanthopAddresses []hop.Address) {
	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		addresses, err := exporter.Addresses()
		if err != nil {
			t.Fatal(err)
		}

		if len(addresses) != len(wanthopAddresses) {
			continue
		}

		for i, v := range addresses {
			if !v.Equal(&wanthopAddresses[i]) {
				continue
			}
		}

		return
	}

	m, err := exporter.Addresses()
	if err != nil {
		t.Fatal(err)
	}

	t.Errorf("hop addresses got %v, want %v", m, wanthopAddresses)
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
