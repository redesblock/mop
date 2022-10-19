package voucher_test

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"testing"

	"github.com/redesblock/mop/core/incentives/voucher"
	pstoremock "github.com/redesblock/mop/core/incentives/voucher/batchstore/mock"
	vouchertesting "github.com/redesblock/mop/core/incentives/voucher/testing"
	storemock "github.com/redesblock/mop/core/storer/statestore/mock"
)

// TestSaveLoad tests the idempotence of saving and loading the voucher.Service
// with all the active stamp issuers.
func TestSaveLoad(t *testing.T) {
	store := storemock.NewStateStore()
	pstore := pstoremock.New()
	saved := func(id int64) voucher.Service {
		ps, err := voucher.NewService(store, pstore, id)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 16; i++ {
			err = ps.Add(newTestStampIssuer(t, 1000))
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := ps.Close(); err != nil {
			t.Fatal(err)
		}
		return ps
	}
	loaded := func(id int64) voucher.Service {
		ps, err := voucher.NewService(store, pstore, id)
		if err != nil {
			t.Fatal(err)
		}
		return ps
	}
	test := func(id int64) {
		psS := saved(id)
		psL := loaded(id)

		sMap := map[string]struct{}{}
		for _, s := range psS.StampIssuers() {
			sMap[string(s.ID())] = struct{}{}
		}

		for _, s := range psL.StampIssuers() {
			if _, ok := sMap[string(s.ID())]; !ok {
				t.Fatalf("mismatch between saved and loaded")
			}
		}
	}
	test(0)
	test(1)
}

func TestGetStampIssuer(t *testing.T) {
	store := storemock.NewStateStore()
	testChainState := vouchertesting.NewChainState()
	if testChainState.Block < uint64(voucher.BlockThreshold) {
		testChainState.Block += uint64(voucher.BlockThreshold + 1)
	}
	validBlockNumber := testChainState.Block - uint64(voucher.BlockThreshold+1)
	pstore := pstoremock.New(pstoremock.WithChainState(testChainState))
	ps, err := voucher.NewService(store, pstore, int64(0))
	if err != nil {
		t.Fatal(err)
	}
	ids := make([][]byte, 8)
	for i := range ids {
		id := make([]byte, 32)
		_, err := io.ReadFull(crand.Reader, id)
		if err != nil {
			t.Fatal(err)
		}
		ids[i] = id
		if i == 0 {
			continue
		}

		var shift uint64 = 0
		if i > 3 {
			shift = uint64(i)
		}
		err = ps.Add(voucher.NewStampIssuer(string(id), "", id, big.NewInt(3), 16, 8, validBlockNumber+shift, true))
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Run("found", func(t *testing.T) {
		for _, id := range ids[1:4] {
			st, err := ps.GetStampIssuer(id)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if st.Label() != string(id) {
				t.Fatalf("wrong issuer returned")
			}
		}
	})
	t.Run("not found", func(t *testing.T) {
		_, err := ps.GetStampIssuer(ids[0])
		if err != voucher.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
	t.Run("not usable", func(t *testing.T) {
		for _, id := range ids[4:] {
			_, err := ps.GetStampIssuer(id)
			if err != voucher.ErrNotUsable {
				t.Fatalf("expected ErrNotUsable, got %v", err)
			}
		}
	})
	t.Run("recovered", func(t *testing.T) {
		b := vouchertesting.MustNewBatch()
		b.Start = validBlockNumber
		testAmount := big.NewInt(1)
		err = ps.HandleCreate(b, testAmount)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		st, err := ps.GetStampIssuer(b.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if st.Label() != "recovered" {
			t.Fatal("wrong issuer returned")
		}
	})
	t.Run("topup", func(t *testing.T) {
		ps.HandleTopUp(ids[1], big.NewInt(10))
		_, err := ps.GetStampIssuer(ids[1])
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ps.StampIssuers()[0].Amount().Cmp(big.NewInt(13)) != 0 {
			t.Fatalf("expected amount %d got %d", 13, ps.StampIssuers()[0].Amount().Int64())
		}
	})
	t.Run("dilute", func(t *testing.T) {
		ps.HandleDepthIncrease(ids[2], 17)
		_, err := ps.GetStampIssuer(ids[2])
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ps.StampIssuers()[1].Amount().Cmp(big.NewInt(3)) != 0 {
			t.Fatalf("expected amount %d got %d", 3, ps.StampIssuers()[1].Amount().Int64())
		}
		if ps.StampIssuers()[1].Depth() != 17 {
			t.Fatalf("expected depth %d got %d", 17, ps.StampIssuers()[1].Depth())
		}
	})
}
