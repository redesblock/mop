package trojan_test

import (
	"bytes"
	"context"
	"github.com/redesblock/mop/core/chunk/trojan"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
)

func TestWrap(t *testing.T) {
	topic := trojan.NewTopic("topic")
	msg := []byte("some payload")
	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := &key.PublicKey
	depth := 1
	targets := newTargets(4, depth)

	chunk, err := trojan.Wrap(context.Background(), topic, msg, pubkey, targets)
	if err != nil {
		t.Fatal(err)
	}

	contains := trojan.Contains(targets, chunk.Address().Bytes()[0:depth])
	if !contains {
		t.Fatal("trojan address was expected to match one of the targets with prefix")
	}

	if len(chunk.Data()) != cluster.ChunkWithSpanSize {
		t.Fatalf("expected trojan data size to be %d, was %d", cluster.ChunkWithSpanSize, len(chunk.Data()))
	}
}

func TestUnwrap(t *testing.T) {
	topic := trojan.NewTopic("topic")
	msg := []byte("some payload")
	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := &key.PublicKey
	depth := 1
	targets := newTargets(4, depth)

	chunk, err := trojan.Wrap(context.Background(), topic, msg, pubkey, targets)
	if err != nil {
		t.Fatal(err)
	}

	topic1 := trojan.NewTopic("topic-1")
	topic2 := trojan.NewTopic("topic-2")

	unwrapTopic, unwrapMsg, err := trojan.Unwrap(context.Background(), key, chunk, []trojan.Topic{topic1, topic2, topic})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(msg, unwrapMsg) {
		t.Fatalf("message mismatch: expected %x, got %x", msg, unwrapMsg)
	}

	if !bytes.Equal(topic[:], unwrapTopic[:]) {
		t.Fatalf("topic mismatch: expected %x, got %x", topic[:], unwrapTopic[:])
	}
}

func TestUnwrapTopicEncrypted(t *testing.T) {
	topic := trojan.NewTopic("topic")
	msg := []byte("some payload")

	privk := crypto.Secp256k1PrivateKeyFromBytes(topic[:])
	pubkey := privk.PublicKey

	depth := 1
	targets := newTargets(4, depth)

	chunk, err := trojan.Wrap(context.Background(), topic, msg, &pubkey, targets)
	if err != nil {
		t.Fatal(err)
	}

	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	topic1 := trojan.NewTopic("topic-1")
	topic2 := trojan.NewTopic("topic-2")

	unwrapTopic, unwrapMsg, err := trojan.Unwrap(context.Background(), key, chunk, []trojan.Topic{topic1, topic2, topic})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(msg, unwrapMsg) {
		t.Fatalf("message mismatch: expected %x, got %x", msg, unwrapMsg)
	}

	if !bytes.Equal(topic[:], unwrapTopic[:]) {
		t.Fatalf("topic mismatch: expected %x, got %x", topic[:], unwrapTopic[:])
	}
}
