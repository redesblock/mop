package testing

import (
	"testing"

	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
)

// MockSOC defines a mocked SOC with exported fields for easy testing.
type MockSOC struct {
	ID           soc.ID
	Owner        []byte
	Signature    []byte
	WrappedChunk cluster.Chunk
}

// Address returns the SOC address of the mocked SOC.
func (ms MockSOC) Address() cluster.Address {
	addr, _ := soc.CreateAddress(ms.ID, ms.Owner)
	return addr
}

// Chunk returns the SOC chunk of the mocked SOC.
func (ms MockSOC) Chunk() cluster.Chunk {
	return cluster.NewChunk(ms.Address(), append(ms.ID, append(ms.Signature, ms.WrappedChunk.Data()...)...))
}

// GenerateMockSOC generates a valid mocked SOC from given data.
func GenerateMockSOC(t *testing.T, data []byte) *MockSOC {
	t.Helper()

	privKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}
	signer := crypto.NewDefaultSigner(privKey)
	owner, err := signer.BSCAddress()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cac.New(data)
	if err != nil {
		t.Fatal(err)
	}

	id := make([]byte, cluster.HashSize)
	hasher := cluster.NewHasher()
	_, err = hasher.Write(append(id, ch.Address().Bytes()...))
	if err != nil {
		t.Fatal(err)
	}

	signature, err := signer.Sign(hasher.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}

	return &MockSOC{
		ID:           id,
		Owner:        owner.Bytes(),
		Signature:    signature,
		WrappedChunk: ch,
	}
}
