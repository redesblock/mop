package mock

import (
	"github.com/redesblock/mop/core/chunk/encryption"
)

type chunkEncrypter struct {
	key []byte
}

func NewChunkEncrypter(key []byte) encryption.ChunkEncrypter { return &chunkEncrypter{key: key} }

func (c *chunkEncrypter) EncryptChunk(chunkData []byte) (encryption.Key, []byte, []byte, error) {
	enc := New(WithXOREncryption(c.key))
	encryptedSpan, err := enc.Encrypt(chunkData[:8])
	if err != nil {
		return nil, nil, nil, err
	}
	encryptedData, err := enc.Encrypt(chunkData[8:])
	if err != nil {
		return nil, nil, nil, err
	}
	return nil, encryptedSpan, encryptedData, nil
}
