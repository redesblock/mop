package postage

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/swarm"
)

// VouchSize is the number of bytes in the serialisation of a vouch
const (
	VouchSize   = 113
	IndexSize   = 8
	BucketDepth = 16
)

var (
	// ErrOwnerMismatch is the error given for invalid signatures.
	ErrOwnerMismatch = errors.New("owner mismatch")
	// ErrInvalidIndex the error given for invalid vouch index.
	ErrInvalidIndex = errors.New("invalid index")
	// ErrVouchInvalid is the error given if vouch cannot deserialise.
	ErrVouchInvalid = errors.New("invalid vouch")
	// ErrBucketMismatch is the error given if vouch index bucket verification fails.
	ErrBucketMismatch = errors.New("bucket mismatch")
)

var _ swarm.Vouch = (*Vouch)(nil)

// Vouch represents a postage vouch as attached to a chunk.
type Vouch struct {
	batchID   []byte // postage batch ID
	index     []byte // index of the batch
	timestamp []byte // to signal order when assigning the indexes to multiple chunks
	sig       []byte // common r[32]s[32]v[1]-style 65 byte ECDSA signature of batchID|index|address by owner or grantee
}

// NewVouch constructs a new vouch from a given batch ID, index and signatures.
func NewVouch(batchID, index, timestamp, sig []byte) *Vouch {
	return &Vouch{batchID, index, timestamp, sig}
}

// BatchID returns the batch ID of the vouch.
func (s *Vouch) BatchID() []byte {
	return s.batchID
}

// Index returns the within-batch index of the vouch.
func (s *Vouch) Index() []byte {
	return s.index
}

// Sig returns the signature of the vouch by the user
func (s *Vouch) Sig() []byte {
	return s.sig
}

// Timestamp returns the timestamp of the vouch
func (s *Vouch) Timestamp() []byte {
	return s.timestamp
}

// MarshalBinary gives the byte slice serialisation of a vouch:
// batchID[32]|index[8]|timestamp[8]|Signature[65].
func (s *Vouch) MarshalBinary() ([]byte, error) {
	buf := make([]byte, VouchSize)
	copy(buf, s.batchID)
	copy(buf[32:40], s.index)
	copy(buf[40:48], s.timestamp)
	copy(buf[48:], s.sig)
	return buf, nil
}

// UnmarshalBinary parses a serialised vouch into id and signature.
func (s *Vouch) UnmarshalBinary(buf []byte) error {
	if len(buf) != VouchSize {
		return ErrVouchInvalid
	}
	s.batchID = buf[:32]
	s.index = buf[32:40]
	s.timestamp = buf[40:48]
	s.sig = buf[48:]
	return nil
}

// toSignDigest creates a digest to represent the vouch which is to be signed by
// the owner.
func toSignDigest(addr, batchId, index, timestamp []byte) ([]byte, error) {
	h := swarm.NewHasher()
	_, err := h.Write(addr)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(batchId)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(index)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(timestamp)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

type ValidVouchFn func(chunk swarm.Chunk, vouchBytes []byte) (swarm.Chunk, error)

// ValidVouch returns a vouchvalidator function passed to protocols with chunk entrypoints.
func ValidVouch(batchStore Storer) ValidVouchFn {
	return func(chunk swarm.Chunk, vouchBytes []byte) (swarm.Chunk, error) {
		vouch := new(Vouch)
		err := vouch.UnmarshalBinary(vouchBytes)
		if err != nil {
			return nil, err
		}
		b, err := batchStore.Get(vouch.BatchID())
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("batchstore get: %v, %w", err, ErrNotFound)
			}
			return nil, err
		}
		if err = vouch.Valid(chunk.Address(), b.Owner, b.Depth, b.BucketDepth, b.Immutable); err != nil {
			return nil, err
		}
		return chunk.WithVouch(vouch).WithBatch(b.Radius, b.Depth, b.BucketDepth, b.Immutable), nil
	}
}

// Valid checks the validity of the postage vouch; in particular:
// - authenticity - check batch is valid on the blockchain
// - authorisation - the batch owner is the vouch signer
// the validity  check is only meaningful in its association of a chunk
// this chunk address needs to be given as argument
func (s *Vouch) Valid(chunkAddr swarm.Address, ownerAddr []byte, depth, bucketDepth uint8, immutable bool) error {
	toSign, err := toSignDigest(chunkAddr.Bytes(), s.batchID, s.index, s.timestamp)
	if err != nil {
		return err
	}
	signerPubkey, err := crypto.Recover(s.sig, toSign)
	if err != nil {
		return err
	}
	signerAddr, err := crypto.NewEthereumAddress(*signerPubkey)
	if err != nil {
		return err
	}
	bucket, index := bytesToIndex(s.index)
	if toBucket(bucketDepth, chunkAddr) != bucket {
		return ErrBucketMismatch
	}
	if index >= 1<<int(depth-bucketDepth) {
		return ErrInvalidIndex
	}
	if !bytes.Equal(signerAddr, ownerAddr) {
		return ErrOwnerMismatch
	}
	return nil
}
