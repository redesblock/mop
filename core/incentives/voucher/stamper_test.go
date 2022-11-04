package voucher_test

import (
	crand "crypto/rand"
	"errors"
	"io"
	"math/big"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/incentives/voucher"
)

// TestStamperStamping tests if the stamp created by the stamper is valid.
func TestStamperStamping(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	owner, err := crypto.NewBSCAddress(privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	signer := crypto.NewDefaultSigner(privKey)
	createStamp := func(t *testing.T, stamper voucher.Stamper) (cluster.Address, *voucher.Stamp) {
		t.Helper()
		h := make([]byte, 32)
		_, err = io.ReadFull(crand.Reader, h)
		if err != nil {
			t.Fatal(err)
		}
		chunkAddr := cluster.NewAddress(h)
		stamp, err := stamper.Stamp(chunkAddr)
		if err != nil {
			t.Fatal(err)
		}
		return chunkAddr, stamp
	}

	// tests a valid stamp
	t.Run("valid stamp", func(t *testing.T) {
		st := newTestStampIssuer(t, 1000)
		stamper := voucher.NewStamper(st, signer)
		chunkAddr, stamp := createStamp(t, stamper)
		if err := stamp.Valid(chunkAddr, owner, 12, 8, true); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// tests that Stamps returns with voucher.ErrBucketMismatch
	t.Run("bucket mismatch", func(t *testing.T) {
		st := newTestStampIssuer(t, 1000)
		stamper := voucher.NewStamper(st, signer)
		chunkAddr, stamp := createStamp(t, stamper)
		a := chunkAddr.Bytes()
		a[0] ^= 0xff
		if err := stamp.Valid(cluster.NewAddress(a), owner, 12, 8, true); !errors.Is(err, voucher.ErrBucketMismatch) {
			t.Fatalf("expected ErrBucketMismatch, got %v", err)
		}
	})

	// tests that Stamps returns with voucher.ErrInvalidIndex
	t.Run("invalid index", func(t *testing.T) {
		st := newTestStampIssuer(t, 1000)
		stamper := voucher.NewStamper(st, signer)
		// issue 1 stamp
		chunkAddr, _ := createStamp(t, stamper)
		// issue another 15
		// collision depth is 8, committed batch depth is 12, bucket volume 2^4
		for i := 0; i < 14; i++ {
			_, err = stamper.Stamp(chunkAddr)
			if err != nil {
				t.Fatalf("error adding stamp at step %d: %v", i, err)
			}
		}
		stamp, err := stamper.Stamp(chunkAddr)
		if err != nil {
			t.Fatalf("error adding last stamp: %v", err)
		}
		if err := stamp.Valid(chunkAddr, owner, 11, 8, true); !errors.Is(err, voucher.ErrInvalidIndex) {
			t.Fatalf("expected ErrInvalidIndex, got %v", err)
		}
	})

	// tests that Stamps returns with voucher.ErrBucketFull iff
	// issuer has the corresponding collision bucket filled]
	t.Run("bucket full", func(t *testing.T) {
		st := voucher.NewStampIssuer("", "", newTestStampIssuer(t, 1000).ID(), big.NewInt(3), 12, 8, 1000, true)
		stamper := voucher.NewStamper(st, signer)
		// issue 1 stamp
		chunkAddr, _ := createStamp(t, stamper)
		// issue another 15
		// collision depth is 8, committed batch depth is 12, bucket volume 2^4
		for i := 0; i < 15; i++ {
			_, err = stamper.Stamp(chunkAddr)
			if err != nil {
				t.Fatalf("error adding stamp at step %d: %v", i, err)
			}
		}
		// the bucket should now be full, not allowing a stamp for the  pivot chunk
		if _, err = stamper.Stamp(chunkAddr); !errors.Is(err, voucher.ErrBucketFull) {
			t.Fatalf("expected ErrBucketFull, got %v", err)
		}
	})

	// tests return with ErrOwnerMismatch
	t.Run("owner mismatch", func(t *testing.T) {
		owner[0] ^= 0xff // bitflip the owner first byte, this case must come last!
		st := newTestStampIssuer(t, 1000)
		stamper := voucher.NewStamper(st, signer)
		chunkAddr, stamp := createStamp(t, stamper)
		if err := stamp.Valid(chunkAddr, owner, 12, 8, true); !errors.Is(err, voucher.ErrOwnerMismatch) {
			t.Fatalf("expected ErrOwnerMismatch, got %v", err)
		}
	})

}
