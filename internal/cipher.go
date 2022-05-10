package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
)

const (
	termSize = 5
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

// AESFromTerm generates a block cipher from an available key within the
// keyring, based on the key term.
func AESFromTerm(term uint32, keyRing *Keyring) (cipher.AEAD, error) {

	if keyRing == nil {
		return nil, fmt.Errorf("missing keyring")
	}

	key := keyRing.TermKey(term)

	if key == nil {
		return nil, fmt.Errorf("no decryption key available for term %d", term)
	}

	return AESFromKey(key.Value)

}

// Encrypt encrypts the plain text along with the key term to a cipher text.
//
// The key term is "baked" into the cipher text. This allows for decrypt
// operations to lookup the key based off the key term to decrypt the cipher text.
func Encrypt(gcm cipher.AEAD, term uint32, plain []byte) ([]byte, error) {

	overhead := gcm.NonceSize() + gcm.Overhead()

	if len(plain) > math.MaxInt-overhead {
		return nil, fmt.Errorf("plaintext is too large")
	}

	capacity := termSize + len(plain) + overhead

	size := termSize + gcm.NonceSize()

	out := make([]byte, size, capacity)

	// Add the key term to the dst output
	binary.BigEndian.PutUint32(out[:termSize], term)

	nonce := out[termSize : termSize+gcm.NonceSize()]

	n, err := rand.Read(nonce)

	if err != nil {
		return nil, err
	}

	if n != len(nonce) {
		return nil, fmt.Errorf("unable to read enough random bytes to fill gcm nonce")
	}

	return gcm.Seal(out, nonce, plain, nil), nil
}

// EncrytTracked encrypts a plain text into a cipher using the active term
// on the keyring.
func EncryptTracked(keyRing *Keyring, plain []byte) ([]byte, error) {

	gcm, err := AESFromTerm(keyRing.ActiveTerm(), keyRing)

	if err != nil {
		return nil, err
	}

	return Encrypt(gcm, keyRing.ActiveTerm(), plain)

}

func Decrypt(gcm cipher.AEAD, cipher []byte) ([]byte, error) {

	if len(cipher) < gcm.NonceSize() {
		return nil, fmt.Errorf("length of ciphertext is invalid")
	}

	capacity := termSize + len(cipher) + gcm.NonceSize() + gcm.Overhead()

	out := make([]byte, 0, capacity)

	nonce := cipher[termSize : termSize+gcm.NonceSize()]

	raw := cipher[termSize+gcm.NonceSize():]

	return gcm.Open(out, nonce, raw, nil)
}

// DecryptTracked decrypts the cipher text based on an existing key
// within the keyring.
func DecryptTracked(keyRing *Keyring, cipher []byte) ([]byte, error) {

	term := binary.BigEndian.Uint32(cipher[:termSize])

	gcm, err := AESFromTerm(term, keyRing)

	if err != nil {
		return nil, err
	}

	return Decrypt(gcm, cipher)

}
