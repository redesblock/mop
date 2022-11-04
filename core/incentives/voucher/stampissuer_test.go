package voucher_test

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/redesblock/mop/core/incentives/voucher"
)

// TestStampIssuerMarshalling tests the idempotence  of binary marshal/unmarshal.
func TestStampIssuerMarshalling(t *testing.T) {
	st := newTestStampIssuer(t, 1000)
	buf, err := st.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	st0 := &voucher.StampIssuer{}
	err = st0.UnmarshalBinary(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(st, st0) {
		t.Fatalf("unmarshal(marshal(StampIssuer)) != StampIssuer \n%v\n%v", st, st0)
	}
}

func newTestStampIssuer(t *testing.T, block uint64) *voucher.StampIssuer {
	t.Helper()
	id := make([]byte, 32)
	_, err := io.ReadFull(crand.Reader, id)
	if err != nil {
		t.Fatal(err)
	}
	return voucher.NewStampIssuer("label", "keyID", id, big.NewInt(3), 16, 8, block, true)
}
