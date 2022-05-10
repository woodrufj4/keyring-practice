package internal

import (
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {

	keyBytes, err := GenerateKey()

	if err != nil {
		t.Fatalf("failed to generate encryption key: %s", err.Error())
	}

	gcm, err := AESFromKey(keyBytes)

	if err != nil {
		t.Fatalf("failed to create block cipher: %s", err.Error())
	}

	plainBytes := []byte("something I want encrypted")

	cipher, err := Encrypt(gcm, 1, plainBytes)

	if err != nil {
		t.Fatalf("failed to encrypt plain text: %s", err.Error())
	}

	plain, err := Decrypt(gcm, cipher)

	if err != nil {
		t.Fatalf("failed to decrypt cipher text: %s", err.Error())
	}

	if !bytes.Equal(plain, plainBytes) {
		t.Fatalf("failed to properly decrypt cipher text. Wanted: %s, Got: %s", string(plainBytes), string(plain))
	}
}

func TestEncryptTracked(t *testing.T) {

	keyRing, err := InitNewKeyRing()

	if err != nil {
		t.Fatalf("failed to generate initialized keyring: %s", err.Error())
	}

	plainBytes := []byte("something else I want encrypted")

	cipher, err := EncryptTracked(keyRing, plainBytes)

	if err != nil {
		t.Fatalf("failed to encrypt plain text: %s", err.Error())
	}

	plain, err := DecryptTracked(keyRing, cipher)

	if err != nil {
		t.Fatalf("failed to decrypt cipher text: %s", err.Error())
	}

	if !bytes.Equal(plain, plainBytes) {
		t.Fatalf("failed to properly decrypt cipher text. Wanted: %s, Got: %s", string(plainBytes), string(plain))
	}
}
