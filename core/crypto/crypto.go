package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/redesblock/mop/core/cluster"
	"golang.org/x/crypto/sha3"
)

// RecoverFunc is a function to recover the public key from a signature
type RecoverFunc func(signature, data []byte) (*ecdsa.PublicKey, error)

var ErrBadHashLength = errors.New("wrong block hash length")

const (
	AddressSize = 20
)

// NewOverlayAddress constructs a Cluster Address from ECDSA public key.
func NewOverlayAddress(p ecdsa.PublicKey, networkID uint64, nonce []byte) (cluster.Address, error) {

	ethAddr, err := NewBSCAddress(p)
	if err != nil {
		return cluster.ZeroAddress, err
	}

	if len(nonce) != 32 {
		return cluster.ZeroAddress, ErrBadHashLength
	}

	return NewOverlayFromBSCAddress(ethAddr, networkID, nonce)
}

// NewOverlayFromBSCAddress constructs a Cluster Address for an BNB Smart Chain address.
func NewOverlayFromBSCAddress(ethAddr []byte, networkID uint64, nonce []byte) (cluster.Address, error) {
	netIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(netIDBytes, networkID)
	data := append(ethAddr, netIDBytes...)
	data = append(data, nonce...)
	h, err := LegacyKeccak256(data)
	if err != nil {
		return cluster.ZeroAddress, err
	}
	return cluster.NewAddress(h[:]), nil
}

// GenerateSecp256k1Key generates an ECDSA private key using
// secp256k1 elliptic curve.
func GenerateSecp256k1Key() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(btcec.S256(), rand.Reader)
}

// EncodeSecp256k1PrivateKey encodes raw ECDSA private key.
func EncodeSecp256k1PrivateKey(k *ecdsa.PrivateKey) []byte {
	return (*btcec.PrivateKey)(k).Serialize()
}

// EncodeSecp256k1PublicKey encodes raw ECDSA public key in a 33-byte compressed format.
func EncodeSecp256k1PublicKey(k *ecdsa.PublicKey) []byte {
	return (*btcec.PublicKey)(k).SerializeCompressed()
}

// DecodeSecp256k1PrivateKey decodes raw ECDSA private key.
func DecodeSecp256k1PrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	if l := len(data); l != btcec.PrivKeyBytesLen {
		return nil, fmt.Errorf("secp256k1 data size %d expected %d", l, btcec.PrivKeyBytesLen)
	}
	privk, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	return (*ecdsa.PrivateKey)(privk), nil
}

// Secp256k1PrivateKeyFromBytes returns an ECDSA private key based on
// the byte slice.
func Secp256k1PrivateKeyFromBytes(data []byte) *ecdsa.PrivateKey {
	privk, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	return (*ecdsa.PrivateKey)(privk)
}

// NewBSCAddress returns a binary representation of BNB Smart Chain address.
// This function is based on github.com/ethereum/go-ethereum/crypto.PubkeyToAddress.
func NewBSCAddress(p ecdsa.PublicKey) ([]byte, error) {
	if p.X == nil || p.Y == nil {
		return nil, errors.New("invalid public key")
	}
	pubBytes := elliptic.Marshal(btcec.S256(), p.X, p.Y)
	pubHash, err := LegacyKeccak256(pubBytes[1:])
	if err != nil {
		return nil, err
	}
	return pubHash[12:], err
}

func LegacyKeccak256(data []byte) ([]byte, error) {
	var err error
	hasher := sha3.NewLegacyKeccak256()
	_, err = hasher.Write(data)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), err
}
