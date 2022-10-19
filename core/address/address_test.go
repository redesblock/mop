package address_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	mop "github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/crypto"

	ma "github.com/multiformats/go-multiaddr"
)

func TestMopAddress(t *testing.T) {
	node1ma, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1634/p2p/16Uiu2HAkx8ULY8cTXhdVAcMmLcH9AsTKz6uBQ7DPLKRjMLgBVYkA")
	if err != nil {
		t.Fatal(err)
	}

	nonce := common.HexToHash("0x2").Bytes()

	privateKey1, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	overlay, err := crypto.NewOverlayAddress(privateKey1.PublicKey, 3, nonce)
	if err != nil {
		t.Fatal(err)
	}
	signer1 := crypto.NewDefaultSigner(privateKey1)

	mopAddress, err := mop.NewAddress(signer1, node1ma, overlay, 3, nonce)
	if err != nil {
		t.Fatal(err)
	}

	mopAddress2, err := mop.ParseAddress(node1ma.Bytes(), overlay.Bytes(), mopAddress.Signature, nonce, true, 3)
	if err != nil {
		t.Fatal(err)
	}

	if !mopAddress.Equal(mopAddress2) {
		t.Fatalf("got %s expected %s", mopAddress2, mopAddress)
	}

	bytes, err := mopAddress.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var newmop mop.Address
	if err := newmop.UnmarshalJSON(bytes); err != nil {
		t.Fatal(err)
	}

	if !newmop.Equal(mopAddress) {
		t.Fatalf("got %s expected %s", newmop, mopAddress)
	}
}
