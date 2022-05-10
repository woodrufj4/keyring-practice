package internal

import (
	"crypto/aes"
	"crypto/cipher"
)

// AESFromKey generates a block cipher from the provided key.
func AESFromKey(key []byte) (cipher.AEAD, error) {

	aesBlock, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesBlock)

	if err != nil {
		return nil, err
	}

	return gcm, nil
}
