package postage_test

import (
	crand "crypto/rand"
	"errors"
	"io"
	"math/big"
	"testing"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/swarm"
)

// TestVoucherVouching tests if the vouch created by the voucher is valid.
func TestVoucherVouching(t *testing.T) {
	privKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	owner, err := crypto.NewEthereumAddress(privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	signer := crypto.NewDefaultSigner(privKey)
	createVouch := func(t *testing.T, voucher postage.Voucher) (swarm.Address, *postage.Vouch) {
		t.Helper()
		h := make([]byte, 32)
		_, err = io.ReadFull(crand.Reader, h)
		if err != nil {
			t.Fatal(err)
		}
		chunkAddr := swarm.NewAddress(h)
		vouch, err := voucher.Vouch(chunkAddr)
		if err != nil {
			t.Fatal(err)
		}
		return chunkAddr, vouch
	}

	// tests a valid vouch
	t.Run("valid vouch", func(t *testing.T) {
		st := newTestVouchIssuer(t, 1000)
		voucher := postage.NewVoucher(st, signer)
		chunkAddr, vouch := createVouch(t, voucher)
		if err := vouch.Valid(chunkAddr, owner, 12, 8, true); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// tests that Vouchs returns with postage.ErrBucketMismatch
	t.Run("bucket mismatch", func(t *testing.T) {
		st := newTestVouchIssuer(t, 1000)
		voucher := postage.NewVoucher(st, signer)
		chunkAddr, vouch := createVouch(t, voucher)
		a := chunkAddr.Bytes()
		a[0] ^= 0xff
		if err := vouch.Valid(swarm.NewAddress(a), owner, 12, 8, true); !errors.Is(err, postage.ErrBucketMismatch) {
			t.Fatalf("expected ErrBucketMismatch, got %v", err)
		}
	})

	// tests that Vouchs returns with postage.ErrInvalidIndex
	t.Run("invalid index", func(t *testing.T) {
		st := newTestVouchIssuer(t, 1000)
		voucher := postage.NewVoucher(st, signer)
		// issue 1 vouch
		chunkAddr, _ := createVouch(t, voucher)
		// issue another 15
		// collision depth is 8, committed batch depth is 12, bucket volume 2^4
		for i := 0; i < 14; i++ {
			_, err = voucher.Vouch(chunkAddr)
			if err != nil {
				t.Fatalf("error adding vouch at step %d: %v", i, err)
			}
		}
		vouch, err := voucher.Vouch(chunkAddr)
		if err != nil {
			t.Fatalf("error adding last vouch: %v", err)
		}
		if err := vouch.Valid(chunkAddr, owner, 11, 8, true); !errors.Is(err, postage.ErrInvalidIndex) {
			t.Fatalf("expected ErrInvalidIndex, got %v", err)
		}
	})

	// tests that Vouchs returns with postage.ErrBucketFull iff
	// issuer has the corresponding collision bucket filled]
	t.Run("bucket full", func(t *testing.T) {
		st := postage.NewVouchIssuer("", "", newTestVouchIssuer(t, 1000).ID(), big.NewInt(3), 12, 8, 1000, true)
		voucher := postage.NewVoucher(st, signer)
		// issue 1 vouch
		chunkAddr, _ := createVouch(t, voucher)
		// issue another 15
		// collision depth is 8, committed batch depth is 12, bucket volume 2^4
		for i := 0; i < 15; i++ {
			_, err = voucher.Vouch(chunkAddr)
			if err != nil {
				t.Fatalf("error adding vouch at step %d: %v", i, err)
			}
		}
		// the bucket should now be full, not allowing a vouch for the  pivot chunk
		if _, err = voucher.Vouch(chunkAddr); !errors.Is(err, postage.ErrBucketFull) {
			t.Fatalf("expected ErrBucketFull, got %v", err)
		}
	})

	// tests return with ErrOwnerMismatch
	t.Run("owner mismatch", func(t *testing.T) {
		owner[0] ^= 0xff // bitflip the owner first byte, this case must come last!
		st := newTestVouchIssuer(t, 1000)
		voucher := postage.NewVoucher(st, signer)
		chunkAddr, vouch := createVouch(t, voucher)
		if err := vouch.Valid(chunkAddr, owner, 12, 8, true); !errors.Is(err, postage.ErrOwnerMismatch) {
			t.Fatalf("expected ErrOwnerMismatch, got %v", err)
		}
	})

}
