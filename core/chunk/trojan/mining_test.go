package trojan_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/redesblock/mop/core/chunk/trojan"
	"github.com/redesblock/mop/core/crypto"
)

func newTargets(length, depth int) trojan.Targets {
	targets := make([]trojan.Target, length)
	for i := 0; i < length; i++ {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(i))
		targets[i] = trojan.Target(buf[:depth])
	}
	return trojan.Targets(targets)
}

func BenchmarkWrap(b *testing.B) {
	cases := []struct {
		length int
		depth  int
	}{
		{1, 1},
		{256, 2},
		{8, 1},
		{256, 1},
		{16, 2},
		{64, 2},
		{256, 3},
		{4096, 3},
		{16384, 3},
	}
	topic := trojan.NewTopic("topic")
	msg := []byte("this is my scariest")
	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		b.Fatal(err)
	}
	pubkey := &key.PublicKey
	ctx := context.Background()
	for _, c := range cases {
		name := fmt.Sprintf("length:%d,depth:%d", c.length, c.depth)
		b.Run(name, func(b *testing.B) {
			targets := newTargets(c.length, c.depth)
			for i := 0; i < b.N; i++ {
				if _, err := trojan.Wrap(ctx, topic, msg, pubkey, targets); err != nil {
					b.Fatal(err)
				}
			}
		})
	}

}
