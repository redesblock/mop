package postage_test

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/redesblock/mop/core/postage"
)

// TestVouchIssuerMarshalling tests the idempotence  of binary marshal/unmarshal.
func TestVouchIssuerMarshalling(t *testing.T) {
	st := newTestVouchIssuer(t, 1000)
	buf, err := st.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	st0 := &postage.VouchIssuer{}
	err = st0.UnmarshalBinary(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(st, st0) {
		t.Fatalf("unmarshal(marshal(VouchIssuer)) != VouchIssuer \n%v\n%v", st, st0)
	}
}

func newTestVouchIssuer(t *testing.T, block uint64) *postage.VouchIssuer {
	t.Helper()
	id := make([]byte, 32)
	_, err := io.ReadFull(crand.Reader, id)
	if err != nil {
		t.Fatal(err)
	}
	return postage.NewVouchIssuer("label", "keyID", id, big.NewInt(3), 16, 8, block, true)
}
