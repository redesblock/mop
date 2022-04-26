package hop

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/swarm"

	ma "github.com/multiformats/go-multiaddr"
)

var ErrInvalidAddress = errors.New("invalid address")

// Address represents the hop address in swarm.
// It consists of a peers underlay (physical) address, overlay (topology) address and signature.
// Signature is used to verify the `Overlay/Underlay` pair, as it is based on `underlay|networkID`, signed with the public key of Overlay address
type Address struct {
	Underlay  ma.Multiaddr
	Overlay   swarm.Address
	Signature []byte
}

type addressJSON struct {
	Overlay   string `json:"overlay"`
	Underlay  string `json:"underlay"`
	Signature string `json:"signature"`
}

func NewAddress(signer crypto.Signer, underlay ma.Multiaddr, overlay swarm.Address, networkID uint64) (*Address, error) {
	networkIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(networkIDBytes, networkID)
	signature, err := signer.Sign(append(underlay.Bytes(), networkIDBytes...))
	if err != nil {
		return nil, err
	}

	return &Address{
		Underlay:  underlay,
		Overlay:   overlay,
		Signature: signature,
	}, nil
}

func ParseAddress(underlay, overlay, signature []byte, networkID uint64) (*Address, error) {
	networkIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(networkIDBytes, networkID)
	recoveredPK, err := crypto.Recover(signature, append(underlay, networkIDBytes...))
	if err != nil {
		return nil, ErrInvalidAddress
	}

	recoveredOverlay := crypto.NewOverlayAddress(*recoveredPK, networkID)
	if !bytes.Equal(recoveredOverlay.Bytes(), overlay) {
		return nil, ErrInvalidAddress
	}

	multiUnderlay, err := ma.NewMultiaddrBytes(underlay)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	return &Address{
		Underlay:  multiUnderlay,
		Overlay:   swarm.NewAddress(overlay),
		Signature: signature,
	}, nil
}

func (a *Address) Equal(b *Address) bool {
	return a.Overlay.Equal(b.Overlay) && a.Underlay.Equal(b.Underlay) && bytes.Equal(a.Signature, b.Signature)
}

func (p *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(&addressJSON{
		Overlay:   p.Overlay.String(),
		Underlay:  p.Underlay.String(),
		Signature: base64.StdEncoding.EncodeToString(p.Signature),
	})
}

func (p *Address) UnmarshalJSON(b []byte) error {
	v := &addressJSON{}
	err := json.Unmarshal(b, v)
	if err != nil {
		return err
	}

	a, err := swarm.ParseHexAddress(v.Overlay)
	if err != nil {
		return err
	}

	p.Overlay = a

	m, err := ma.NewMultiaddr(v.Underlay)
	if err != nil {
		return err
	}

	p.Underlay = m
	p.Signature, err = base64.StdEncoding.DecodeString(v.Signature)
	return err
}
