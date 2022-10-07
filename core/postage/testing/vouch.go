package testing

import (
	crand "crypto/rand"
	"io"

	"github.com/redesblock/mop/core/postage"
)

const signatureSize = 65

// MustNewSignature will create a new random signature (65 byte slice). Panics
// on errors.
func MustNewSignature() []byte {
	sig := make([]byte, signatureSize)
	_, err := io.ReadFull(crand.Reader, sig)
	if err != nil {
		panic(err)
	}
	return sig
}

// MustNewVouch will generate a postage vouch with random data. Panics on
// errors.
func MustNewVouch() *postage.Vouch {
	return postage.NewVouch(MustNewID(), MustNewID()[:8], MustNewID()[:8], MustNewSignature())
}
