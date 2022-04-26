package handshake_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/hop"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/p2p/libp2p/internal/handshake"
	"github.com/redesblock/hop/core/p2p/libp2p/internal/handshake/mock"
	"github.com/redesblock/hop/core/p2p/libp2p/internal/handshake/pb"
	"github.com/redesblock/hop/core/p2p/protobuf"

	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func TestHandshake(t *testing.T) {
	logger := logging.New(ioutil.Discard, 0)
	networkID := uint64(3)
	node1ma, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/7070/p2p/16Uiu2HAkx8ULY8cTXhdVAcMmLcH9AsTKz6uBQ7DPLKRjMLgBVYkA")
	if err != nil {
		t.Fatal(err)
	}
	node2ma, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/7070/p2p/16Uiu2HAkx8ULY8cTXhdVAcMmLcH9AsTKz6uBQ7DPLKRjMLgBVYkS")
	if err != nil {
		t.Fatal(err)
	}

	node2AddrInfo, err := libp2ppeer.AddrInfoFromP2pAddr(node2ma)
	if err != nil {
		t.Fatal(err)
	}

	privateKey1, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	privateKey2, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	signer1 := crypto.NewDefaultSigner(privateKey1)
	signer2 := crypto.NewDefaultSigner(privateKey2)
	node1HopAddress, err := hop.NewAddress(signer1, node1ma, crypto.NewOverlayAddress(privateKey1.PublicKey, networkID), networkID)
	if err != nil {
		t.Fatal(err)
	}
	node2HopAddress, err := hop.NewAddress(signer2, node2ma, crypto.NewOverlayAddress(privateKey2.PublicKey, networkID), networkID)
	if err != nil {
		t.Fatal(err)
	}

	node1Info := handshake.Info{
		HopAddress: node1HopAddress,
		Light:      false,
	}
	node2Info := handshake.Info{
		HopAddress: node2HopAddress,
		Light:      false,
	}

	handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Handshake - OK", func(t *testing.T) {
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w, r := protobuf.NewWriterAndReader(stream2)
		if err := w.WriteMsg(&pb.SynAck{
			Syn: &pb.Syn{
				HopAddress: &pb.HopAddress{
					Underlay:  node2HopAddress.Underlay.Bytes(),
					Overlay:   node2HopAddress.Overlay.Bytes(),
					Signature: node2HopAddress.Signature,
				},
				Light:     node2Info.Light,
				NetworkID: networkID,
			},
			Ack: &pb.Ack{HopAddress: &pb.HopAddress{
				Underlay:  node1HopAddress.Underlay.Bytes(),
				Overlay:   node1HopAddress.Overlay.Bytes(),
				Signature: node1HopAddress.Signature,
			}},
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handshake(stream1)
		if err != nil {
			t.Fatal(err)
		}

		testInfo(t, *res, node2Info)
		if err := r.ReadMsg(&pb.Ack{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Handshake - Syn write error", func(t *testing.T) {
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("write syn message: %w", testErr)
		stream := &mock.Stream{}
		stream.SetWriteErr(testErr, 0)
		res, err := handshakeService.Handshake(stream)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handshake returned non-nil res")
		}
	})

	t.Run("Handshake - Syn read error", func(t *testing.T) {
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("read synack message: %w", testErr)
		stream := mock.NewStream(nil, &bytes.Buffer{})
		stream.SetReadErr(testErr, 0)
		res, err := handshakeService.Handshake(stream)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handshake returned non-nil res")
		}
	})

	t.Run("Handshake - ack write error", func(t *testing.T) {
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("write ack message: %w", testErr)
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream1.SetWriteErr(testErr, 1)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.SynAck{
			Syn: &pb.Syn{
				HopAddress: &pb.HopAddress{
					Underlay:  node2HopAddress.Underlay.Bytes(),
					Overlay:   node2HopAddress.Overlay.Bytes(),
					Signature: node2HopAddress.Signature,
				},
				Light:     node2Info.Light,
				NetworkID: networkID,
			},
			Ack: &pb.Ack{HopAddress: &pb.HopAddress{
				Underlay:  node1HopAddress.Underlay.Bytes(),
				Overlay:   node1HopAddress.Overlay.Bytes(),
				Signature: node1HopAddress.Signature,
			}},
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handshake(stream1)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handshake returned non-nil res")
		}
	})

	t.Run("Handshake - networkID mismatch", func(t *testing.T) {
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.SynAck{
			Syn: &pb.Syn{
				HopAddress: &pb.HopAddress{
					Underlay:  node2HopAddress.Underlay.Bytes(),
					Overlay:   node2HopAddress.Overlay.Bytes(),
					Signature: node2HopAddress.Signature,
				},
				NetworkID: 5,
				Light:     node2Info.Light,
			},
			Ack: &pb.Ack{HopAddress: &pb.HopAddress{
				Underlay:  node1HopAddress.Underlay.Bytes(),
				Overlay:   node1HopAddress.Overlay.Bytes(),
				Signature: node1HopAddress.Signature,
			}},
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handshake(stream1)
		if res != nil {
			t.Fatal("res should be nil")
		}

		if err != handshake.ErrNetworkIDIncompatible {
			t.Fatalf("expected %s, got %s", handshake.ErrNetworkIDIncompatible, err)
		}
	})

	t.Run("Handshake - invalid ack", func(t *testing.T) {
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.SynAck{
			Syn: &pb.Syn{
				HopAddress: &pb.HopAddress{
					Underlay:  node2HopAddress.Underlay.Bytes(),
					Overlay:   node2HopAddress.Overlay.Bytes(),
					Signature: node2HopAddress.Signature,
				},
				Light:     node2Info.Light,
				NetworkID: networkID,
			},
			Ack: &pb.Ack{HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			}},
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handshake(stream1)
		if res != nil {
			t.Fatal("res should be nil")
		}

		if err != handshake.ErrInvalidAck {
			t.Fatalf("expected %s, got %s", handshake.ErrInvalidAck, err)
		}
	})

	t.Run("Handshake - invalid signature", func(t *testing.T) {
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.SynAck{
			Syn: &pb.Syn{
				HopAddress: &pb.HopAddress{
					Underlay:  node2HopAddress.Underlay.Bytes(),
					Overlay:   []byte("wrong signature"),
					Signature: node2HopAddress.Signature,
				},
				Light:     node2Info.Light,
				NetworkID: networkID,
			},
			Ack: &pb.Ack{HopAddress: &pb.HopAddress{
				Underlay:  node1HopAddress.Underlay.Bytes(),
				Overlay:   node1HopAddress.Overlay.Bytes(),
				Signature: node1HopAddress.Signature,
			}},
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handshake(stream1)
		if res != nil {
			t.Fatal("res should be nil")
		}

		if err != handshake.ErrInvalidHopAddress {
			t.Fatalf("expected %s, got %s", handshake.ErrInvalidHopAddress, err)
		}
	})

	t.Run("Handle - OK", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			Light:     node2Info.Light,
			NetworkID: networkID,
		}); err != nil {
			t.Fatal(err)
		}

		if err := w.WriteMsg(&pb.Ack{HopAddress: &pb.HopAddress{
			Underlay:  node1HopAddress.Underlay.Bytes(),
			Overlay:   node1HopAddress.Overlay.Bytes(),
			Signature: node1HopAddress.Signature,
		}}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream1, node2AddrInfo.ID)
		if err != nil {
			t.Fatal(err)
		}

		testInfo(t, *res, node2Info)

		_, r := protobuf.NewWriterAndReader(stream2)
		var got pb.SynAck
		if err := r.ReadMsg(&got); err != nil {
			t.Fatal(err)
		}

		HopAddress, err := hop.ParseAddress(got.Syn.HopAddress.Underlay, got.Syn.HopAddress.Overlay, got.Syn.HopAddress.Signature, got.Syn.NetworkID)
		if err != nil {
			t.Fatal(err)
		}

		testInfo(t, node1Info, handshake.Info{
			HopAddress: HopAddress,
			Light:      got.Syn.Light,
		})
	})

	t.Run("Handle - read error ", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("read syn message: %w", testErr)
		stream := &mock.Stream{}
		stream.SetReadErr(testErr, 0)
		res, err := handshakeService.Handle(stream, node2AddrInfo.ID)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handle returned non-nil res")
		}
	})

	t.Run("Handle - write error ", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("write synack message: %w", testErr)
		var buffer bytes.Buffer
		stream := mock.NewStream(&buffer, &buffer)
		stream.SetWriteErr(testErr, 1)
		w := protobuf.NewWriter(stream)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			Light:     node2Info.Light,
			NetworkID: networkID,
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream, node2AddrInfo.ID)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handshake returned non-nil res")
		}
	})

	t.Run("Handle - ack read error ", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		testErr := errors.New("test error")
		expectedErr := fmt.Errorf("read ack message: %w", testErr)
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)
		stream1.SetReadErr(testErr, 1)
		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			Light:     node2Info.Light,
			NetworkID: networkID,
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream1, node2AddrInfo.ID)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatal("expected:", expectedErr, "got:", err)
		}

		if res != nil {
			t.Fatal("handshake returned non-nil res")
		}
	})

	t.Run("Handle - networkID mismatch ", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			NetworkID: 5,
			Light:     node2Info.Light,
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream1, node2AddrInfo.ID)
		if res != nil {
			t.Fatal("res should be nil")
		}

		if err != handshake.ErrNetworkIDIncompatible {
			t.Fatalf("expected %s, got %s", handshake.ErrNetworkIDIncompatible, err)
		}
	})

	t.Run("Handle - duplicate handshake", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			Light:     node2Info.Light,
			NetworkID: networkID,
		}); err != nil {
			t.Fatal(err)
		}

		if err := w.WriteMsg(&pb.Ack{HopAddress: &pb.HopAddress{
			Underlay:  node1HopAddress.Underlay.Bytes(),
			Overlay:   node1HopAddress.Overlay.Bytes(),
			Signature: node1HopAddress.Signature,
		}}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream1, node2AddrInfo.ID)
		if err != nil {
			t.Fatal(err)
		}

		testInfo(t, *res, node2Info)

		_, r := protobuf.NewWriterAndReader(stream2)
		var got pb.SynAck
		if err := r.ReadMsg(&got); err != nil {
			t.Fatal(err)
		}

		HopAddress, err := hop.ParseAddress(got.Syn.HopAddress.Underlay, got.Syn.HopAddress.Overlay, got.Syn.HopAddress.Signature, got.Syn.NetworkID)
		if err != nil {
			t.Fatal(err)
		}

		testInfo(t, node1Info, handshake.Info{
			HopAddress: HopAddress,
			Light:      got.Syn.Light,
		})

		_, err = handshakeService.Handle(stream1, node2AddrInfo.ID)
		if err != handshake.ErrHandshakeDuplicate {
			t.Fatalf("expected %s, got %s", handshake.ErrHandshakeDuplicate, err)
		}
	})

	t.Run("Handle - invalid ack", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   node2HopAddress.Overlay.Bytes(),
				Signature: node2HopAddress.Signature,
			},
			Light:     node2Info.Light,
			NetworkID: networkID,
		}); err != nil {
			t.Fatal(err)
		}

		if err := w.WriteMsg(&pb.Ack{HopAddress: &pb.HopAddress{
			Underlay:  node2HopAddress.Underlay.Bytes(),
			Overlay:   node2HopAddress.Overlay.Bytes(),
			Signature: node2HopAddress.Signature,
		}}); err != nil {
			t.Fatal(err)
		}

		_, err = handshakeService.Handle(stream1, node2AddrInfo.ID)
		if err != handshake.ErrInvalidAck {
			t.Fatalf("expected %s, got %s", handshake.ErrInvalidAck, err)
		}
	})

	t.Run("Handle - invalid signature ", func(t *testing.T) {
		handshakeService, err := handshake.New(node1Info.HopAddress.Overlay, node1Info.HopAddress.Underlay, signer1, networkID, logger)
		if err != nil {
			t.Fatal(err)
		}
		var buffer1 bytes.Buffer
		var buffer2 bytes.Buffer
		stream1 := mock.NewStream(&buffer1, &buffer2)
		stream2 := mock.NewStream(&buffer2, &buffer1)

		w := protobuf.NewWriter(stream2)
		if err := w.WriteMsg(&pb.Syn{
			HopAddress: &pb.HopAddress{
				Underlay:  node2HopAddress.Underlay.Bytes(),
				Overlay:   []byte("wrong signature"),
				Signature: node2HopAddress.Signature,
			},
			NetworkID: networkID,
			Light:     node2Info.Light,
		}); err != nil {
			t.Fatal(err)
		}

		res, err := handshakeService.Handle(stream1, node2AddrInfo.ID)
		if res != nil {
			t.Fatal("res should be nil")
		}

		if err != handshake.ErrInvalidHopAddress {
			t.Fatalf("expected %s, got %s", handshake.ErrInvalidHopAddress, err)
		}
	})
}

// testInfo validates if two Info instances are equal.
func testInfo(t *testing.T, got, want handshake.Info) {
	t.Helper()
	if !got.HopAddress.Equal(want.HopAddress) || got.Light != want.Light {
		t.Fatalf("got info %+v, want %+v", got, want)
	}
}
