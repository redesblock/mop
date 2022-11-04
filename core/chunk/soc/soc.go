// Package soc provides the single-owner chunk implementation
// and validator.
package soc

import (
	"bytes"
	"errors"

	"github.com/redesblock/mop/core/chunk/cac"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
)

var (
	errInvalidAddress = errors.New("soc: invalid address")
	errWrongChunkSize = errors.New("soc: chunk length is less than minimum")
)

// ID is a SOC identifier
type ID []byte

// SOC wraps a content-addressed chunk.
type SOC struct {
	id        ID
	owner     []byte // owner is the address in bytes of SOC owner.
	signature []byte
	chunk     cluster.Chunk // wrapped chunk.
}

// New creates a new SOC representation from arbitrary id and
// a content-addressed chunk.
func New(id ID, ch cluster.Chunk) *SOC {
	return &SOC{
		id:    id,
		chunk: ch,
	}
}

// NewSigned creates a single-owner chunk based on already signed data.
func NewSigned(id ID, ch cluster.Chunk, owner, sig []byte) (*SOC, error) {
	s := New(id, ch)
	if len(owner) != crypto.AddressSize {
		return nil, errInvalidAddress
	}
	s.owner = owner
	s.signature = sig
	return s, nil
}

// address returns the SOC chunk address.
func (s *SOC) address() (cluster.Address, error) {
	if len(s.owner) != crypto.AddressSize {
		return cluster.ZeroAddress, errInvalidAddress
	}
	return CreateAddress(s.id, s.owner)
}

// WrappedChunk returns the chunk wrapped by the SOC.
func (s *SOC) WrappedChunk() cluster.Chunk {
	return s.chunk
}

// Chunk returns the SOC chunk.
func (s *SOC) Chunk() (cluster.Chunk, error) {
	socAddress, err := s.address()
	if err != nil {
		return nil, err
	}
	return cluster.NewChunk(socAddress, s.toBytes()), nil
}

// toBytes is a helper function to convert the SOC data to bytes.
func (s *SOC) toBytes() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(s.id)
	buf.Write(s.signature)
	buf.Write(s.chunk.Data())
	return buf.Bytes()
}

// Sign signs a SOC using the given signer.
// It returns a signed SOC chunk ready for submission to the network.
func (s *SOC) Sign(signer crypto.Signer) (cluster.Chunk, error) {
	// create owner
	publicKey, err := signer.PublicKey()
	if err != nil {
		return nil, err
	}
	ownerAddressBytes, err := crypto.NewBSCAddress(*publicKey)
	if err != nil {
		return nil, err
	}
	if len(ownerAddressBytes) != crypto.AddressSize {
		return nil, errInvalidAddress
	}
	s.owner = ownerAddressBytes

	// generate the data to sign
	toSignBytes, err := hash(s.id, s.chunk.Address().Bytes())
	if err != nil {
		return nil, err
	}

	// sign the chunk
	signature, err := signer.Sign(toSignBytes)
	if err != nil {
		return nil, err
	}
	s.signature = signature

	return s.Chunk()
}

// FromChunk recreates a SOC representation from cluster.Chunk data.
func FromChunk(sch cluster.Chunk) (*SOC, error) {
	chunkData := sch.Data()
	if len(chunkData) < cluster.SocMinChunkSize {
		return nil, errWrongChunkSize
	}

	// add all the data fields to the SOC
	s := &SOC{}
	cursor := 0

	s.id = chunkData[cursor:cluster.HashSize]
	cursor += cluster.HashSize

	s.signature = chunkData[cursor : cursor+cluster.SocSignatureSize]
	cursor += cluster.SocSignatureSize

	ch, err := cac.NewWithDataSpan(chunkData[cursor:])
	if err != nil {
		return nil, err
	}

	toSignBytes, err := hash(s.id, ch.Address().Bytes())
	if err != nil {
		return nil, err
	}

	// recover owner information
	recoveredOwnerAddress, err := recoverAddress(s.signature, toSignBytes)
	if err != nil {
		return nil, err
	}
	if len(recoveredOwnerAddress) != crypto.AddressSize {
		return nil, errInvalidAddress
	}
	s.owner = recoveredOwnerAddress
	s.chunk = ch

	return s, nil
}

// CreateAddress creates a new SOC address from the id and
// the BNB Smart Chain of the owner.
func CreateAddress(id ID, owner []byte) (cluster.Address, error) {
	sum, err := hash(id, owner)
	if err != nil {
		return cluster.ZeroAddress, err
	}
	return cluster.NewAddress(sum), nil
}

// hash hashes the given values in order.
func hash(values ...[]byte) ([]byte, error) {
	h := cluster.NewHasher()
	for _, v := range values {
		_, err := h.Write(v)
		if err != nil {
			return nil, err
		}
	}
	return h.Sum(nil), nil
}

// recoverAddress returns the BNB Smart Chain of the owner of a SOC.
func recoverAddress(signature, digest []byte) ([]byte, error) {
	recoveredPublicKey, err := crypto.Recover(signature, digest)
	if err != nil {
		return nil, err
	}
	recoveredBSCAddress, err := crypto.NewBSCAddress(*recoveredPublicKey)
	if err != nil {
		return nil, err
	}
	return recoveredBSCAddress, nil
}
