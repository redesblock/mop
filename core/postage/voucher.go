package postage

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/swarm"
)

var (
	// ErrBucketFull is the error when a collision bucket is full.
	ErrBucketFull = errors.New("bucket full")
)

// Voucher can issue vouches from the given address.
type Voucher interface {
	Vouch(swarm.Address) (*Vouch, error)
}

// voucher connects a vouchIssuer with a signer.
// A voucher is created for each upload session.
type voucher struct {
	issuer *VouchIssuer
	signer crypto.Signer
}

// NewVoucher constructs a Voucher.
func NewVoucher(st *VouchIssuer, signer crypto.Signer) Voucher {
	return &voucher{st, signer}
}

// Vouch takes chunk, see if the chunk can included in the batch and
// signs it with the owner of the batch of this Vouch issuer.
func (st *voucher) Vouch(addr swarm.Address) (*Vouch, error) {
	index, err := st.issuer.inc(addr)
	if err != nil {
		return nil, err
	}
	ts := timestamp()
	toSign, err := toSignDigest(addr.Bytes(), st.issuer.data.BatchID, index, ts)
	if err != nil {
		return nil, err
	}
	sig, err := st.signer.Sign(toSign)
	if err != nil {
		return nil, err
	}
	return NewVouch(st.issuer.data.BatchID, index, ts, sig), nil
}

func timestamp() []byte {
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().UnixNano()))
	return ts
}
