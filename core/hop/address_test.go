package hop_test

import (
	"testing"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/hop"

	ma "github.com/multiformats/go-multiaddr"
)

func TesthopAddress(t *testing.T) {
	node1ma, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/7070/p2p/16Uiu2HAkx8ULY8cTXhdVAcMmLcH9AsTKz6uBQ7DPLKRjMLgBVYkA")
	if err != nil {
		t.Fatal(err)
	}

	privateKey1, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	overlay := crypto.NewOverlayAddress(privateKey1.PublicKey, 3)
	signer1 := crypto.NewDefaultSigner(privateKey1)

	hopAddress, err := hop.NewAddress(signer1, node1ma, overlay, 3)
	if err != nil {
		t.Fatal(err)
	}

	hopAddress2, err := hop.ParseAddress(node1ma.Bytes(), overlay.Bytes(), hopAddress.Signature, 3)
	if err != nil {
		t.Fatal(err)
	}

	if !hopAddress.Equal(hopAddress2) {
		t.Fatalf("got %s expected %s", hopAddress2, hopAddress)
	}

	bytes, err := hopAddress.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var newhop hop.Address
	if err := newhop.UnmarshalJSON(bytes); err != nil {
		t.Fatal(err)
	}

	if !newhop.Equal(hopAddress) {
		t.Fatalf("got %s expected %s", newhop, hopAddress)
	}
}
