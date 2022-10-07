package postage_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/postage/batchstore/mock"
	postagetesting "github.com/redesblock/mop/core/postage/testing"
	chunktesting "github.com/redesblock/mop/core/storage/testing"
)

// TestVouchMarshalling tests the idempotence  of binary marshal/unmarshals for Vouchs.
func TestVouchMarshalling(t *testing.T) {
	sExp := postagetesting.MustNewVouch()
	buf, _ := sExp.MarshalBinary()
	if len(buf) != postage.VouchSize {
		t.Fatalf("invalid length for serialised vouch. expected %d, got  %d", postage.VouchSize, len(buf))
	}
	s := postage.NewVouch(nil, nil, nil, nil)
	if err := s.UnmarshalBinary(buf); err != nil {
		t.Fatalf("unexpected error unmarshalling vouch: %v", err)
	}
	compareVouchs(t, sExp, s)
}

func compareVouchs(t *testing.T, s1, s2 *postage.Vouch) {
	if !bytes.Equal(s1.BatchID(), s2.BatchID()) {
		t.Fatalf("id mismatch, expected %x, got %x", s1.BatchID(), s2.BatchID())
	}
	if !bytes.Equal(s1.Index(), s2.Index()) {
		t.Fatalf("index mismatch, expected %x, got %x", s1.Index(), s2.Index())
	}
	if !bytes.Equal(s1.Timestamp(), s2.Timestamp()) {
		t.Fatalf("timestamp mismatch, expected %x, got %x", s1.Index(), s2.Index())
	}
	if !bytes.Equal(s1.Sig(), s2.Sig()) {
		t.Fatalf("sig mismatch, expected %x, got %x", s1.Sig(), s2.Sig())
	}
}

// TestVouchIndexMarshalling tests the idempotence of vouch index serialisation.
func TestVouchIndexMarshalling(t *testing.T) {
	var (
		expBucket uint32 = 11789
		expIndex  uint32 = 199999
	)
	index := postage.IndexToBytes(expBucket, expIndex)
	bucket, idx := postage.BytesToIndex(index)
	if bucket != expBucket {
		t.Fatalf("bucket mismatch. want %d, got %d", expBucket, bucket)
	}
	if idx != expIndex {
		t.Fatalf("index mismatch. want %d, got %d", expIndex, idx)
	}
}

func TestValidVouch(t *testing.T) {

	privKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	owner, err := crypto.NewEthereumAddress(privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	b := postagetesting.MustNewBatch(postagetesting.WithOwner(owner))
	bs := mock.New(mock.WithBatch(b))
	signer := crypto.NewDefaultSigner(privKey)
	issuer := postage.NewVouchIssuer("label", "keyID", b.ID, big.NewInt(3), b.Depth, b.BucketDepth, 1000, true)
	voucher := postage.NewVoucher(issuer, signer)

	// this creates a chunk with a mocked vouch. ValidVouch will override this
	// vouch on execution
	ch := chunktesting.GenerateTestRandomChunk()

	st, err := voucher.Vouch(ch.Address())
	if err != nil {
		t.Fatal(err)
	}
	stBytes, err := st.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// ensure the chunk doesnt have the batch details filled before we validate vouch
	if ch.Depth() == b.Depth || ch.BucketDepth() == b.BucketDepth {
		t.Fatal("expected chunk to not have correct depth and bucket depth at start")
	}

	ch, err = postage.ValidVouch(bs)(ch, stBytes)
	if err != nil {
		t.Fatal(err)
	}

	compareVouchs(t, st, ch.Vouch().(*postage.Vouch))

	if ch.Depth() != b.Depth {
		t.Fatalf("invalid batch depth added on chunk exp %d got %d", b.Depth, ch.Depth())
	}
	if ch.BucketDepth() != b.BucketDepth {
		t.Fatalf("invalid bucket depth added on chunk exp %d got %d", b.BucketDepth, ch.BucketDepth())
	}
	if ch.Immutable() != b.Immutable {
		t.Fatalf("invalid batch immutablility added on chunk exp %t got %t", b.Immutable, ch.Immutable())
	}
}
