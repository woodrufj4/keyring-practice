package internal

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/woodrufj4/keyring-practice/backend"
)

const (

	// keyringPath is the location of the keyring data. This is encrypted
	// by the root key.
	keyringPath      = "core/keyring"
	keyringPrefix    = "core/"
	keyringCipherKey = "keyringCipher"
)

var (

	// ErrKeyringNotFound is returned if the keyring isn't found
	// within the backend persistance.
	ErrKeyringNotFound         = errors.New("keyring not found in backend")
	ErrBarrerInvalidBackend    = errors.New("keyring backend is invalid")
	ErrKeyringCipherKeyInvalid = errors.New("keyring cipher key is invalid")

	// ErrorKeyringNotSet is returned if an operation is being performed on
	// a non initialized barrier.
	// No operation is expected to succeed before initializing
	ErrKeyringNotSet = errors.New("keyring is not setup")
)

type Barrier struct {
	initialized bool
	backend     backend.Backend
	keyring     *Keyring
	sync        sync.RWMutex
}

// NewBarrier instantiates a new barrier
func NewBarrier(backend backend.Backend) (*Barrier, error) {

	if backend == nil {
		return nil, ErrBarrerInvalidBackend
	}

	return &Barrier{
		initialized: false,
		backend:     backend,
	}, nil

}

// Initialize sets up the barrier an initializes a keyring.
//
// If the barrier has already been initialized, this does nothting.
func (b *Barrier) Initialize(ctx context.Context, rootKey string) error {

	if b.initialized {
		return nil
	}

	if err := b.loadKeyring(ctx, rootKey); err != nil {

		if err == ErrKeyringNotFound {

			// generate a new keyring
			if err := b.initKeyring(rootKey); err != nil {
				return err
			}

			// persist the new keyring
			if err := b.persistKeyring(ctx); err != nil {
				return err
			}

		} else {
			return err
		}
	}

	b.initialized = true

	return nil
}

func (b *Barrier) KeyringPersisted(ctx context.Context) (bool, error) {

	if b.backend == nil {
		return false, ErrBarrerInvalidBackend
	}

	// Check if there is a keyring in the backend
	entries, err := b.backend.Get(ctx, keyringPath)

	if err != nil {
		return false, err
	}

	if entries == nil {
		return false, nil
	}

	if entries[0].Key != keyringCipherKey {
		return false, ErrKeyringCipherKeyInvalid
	}

	return true, nil

}

func (b *Barrier) Initialized() bool {
	return b.initialized
}

func (b *Barrier) GenerateKey() ([]byte, error) {

	// Generate a 256bit key
	buf := make([]byte, 2*aes.BlockSize)
	_, err := rand.Read(buf)

	return buf, err
}

// aesFromKey generates a block cipher from the provided key.
func (b *Barrier) aesFromKey(key []byte) (cipher.AEAD, error) {

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
func (b *Barrier) aesFromTerm(term uint32) (cipher.AEAD, error) {

	if !b.initialized {
		return nil, ErrKeyringNotSet
	}

	key := b.keyring.TermKey(term)

	if key == nil {
		return nil, fmt.Errorf("no decryption key available for term %d", term)
	}

	return b.aesFromKey(key.Value)

}

// persistKeyring encrypts and then stores the keyring within the backend.
func (b *Barrier) persistKeyring(ctx context.Context) error {

	if b.keyring == nil {
		return ErrKeyringNotSet
	}

	keyringBytes, err := b.keyring.Serialize()

	if err != nil {
		return fmt.Errorf("failed to serialize keyring: %s", err.Error())
	}

	gcm, err := b.aesFromKey(b.keyring.rootKey)

	if err != nil {
		return err
	}

	keyringCipher, err := b.encrypt(gcm, 0, keyringBytes)

	if err != nil {
		return fmt.Errorf("failed to encrypt keyring: %s", err.Error())
	}

	// encrypt and perist initial keyring
	entries := []*backend.BackendEntry{
		{
			Key:   keyringCipherKey,
			Value: keyringCipher,
		},
	}

	return b.backend.Put(ctx, keyringPath, entries)

}

// loadKeyring attempts to retrieve the encrypted keyring from the backend.
// This method decrypts the keyring with the provided root token.
func (b *Barrier) loadKeyring(ctx context.Context, rootKey string) error {

	if b.backend == nil {
		return ErrBarrerInvalidBackend
	}

	// Check if there is a keyring in the backend
	entries, err := b.backend.Get(ctx, keyringPath)

	if err != nil {
		return err
	}

	if entries == nil {
		return ErrKeyringNotFound
	}

	gcm, err := b.aesFromKey([]byte(rootKey))

	if err != nil {
		return err
	}

	if entries[0].Key != keyringCipherKey {
		return ErrKeyringCipherKeyInvalid
	}

	keyringBytes, err := b.decrypt(gcm, entries[0].Value)

	if err != nil {
		return fmt.Errorf("failed to decrypt keyring: %s", err.Error())
	}

	var encodedKeyring EncodedKeyring

	err = json.Unmarshal(keyringBytes, &encodedKeyring)

	if err != nil {
		return err
	}

	keyring := &Keyring{
		rootKey: encodedKeyring.MasterKey,
		keys:    make(map[uint32]*Key, 0),
	}

	for _, key := range encodedKeyring.Keys {

		keyring.keys[key.Term] = key

		if key.Term > keyring.activeTerm {
			keyring.activeTerm = key.Term
		}
	}

	b.keyring = keyring

	return nil
}

// initKeyring initializes a keyring and stores it within the barrier.
func (b *Barrier) initKeyring(rootKey string) error {

	firstKey, err := b.GenerateKey()

	if err != nil {
		return fmt.Errorf("failed to generate initial encryption key: %s", err.Error())
	}

	keyRing := &Keyring{
		keys:       make(map[uint32]*Key, 0),
		activeTerm: 0,
	}

	err = keyRing.SetRootKey([]byte(rootKey))

	if err != nil {
		return fmt.Errorf("failed to generate initial encryption key: %s", err.Error())
	}

	err = keyRing.AddKey(&Key{
		Term:    1,
		Value:   firstKey,
		Version: 1,
	})

	if err != nil {
		return fmt.Errorf("failed to set initial encryption key on keyring: %s", err.Error())
	}

	b.keyring = keyRing

	return nil
}

// Keyring provides a copy of the currently managed keyring.
func (b *Barrier) Keyring() (*Keyring, error) {

	b.sync.RLock()
	defer b.sync.RUnlock()

	if !b.initialized {
		return nil, ErrKeyringNotSet
	}

	// return a clone of the keyring
	return b.keyring.Clone(), nil
}

// Encrypt performs encryption and persistance of secrets
func (b *Barrier) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {

	gcm, err := b.aesFromTerm(b.keyring.activeTerm)

	if err != nil {
		return nil, err
	}

	return b.encrypt(gcm, b.keyring.activeTerm, plaintext)
}

// encrypt performs encryption on plain text
func (b *Barrier) encrypt(gcm cipher.AEAD, term uint32, plain []byte) ([]byte, error) {

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

func (b *Barrier) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {

	gcm, err := b.aesFromTerm(b.keyring.activeTerm)

	if err != nil {
		return nil, err
	}

	return b.decrypt(gcm, ciphertext)
}

// decrypt only perform decryptions on cipher texts.
func (b *Barrier) decrypt(gcm cipher.AEAD, cipher []byte) ([]byte, error) {

	if len(cipher) < gcm.NonceSize() {
		return nil, fmt.Errorf("length of ciphertext is invalid")
	}

	capacity := termSize + len(cipher) + gcm.NonceSize() + gcm.Overhead()

	out := make([]byte, 0, capacity)

	nonce := cipher[termSize : termSize+gcm.NonceSize()]

	raw := cipher[termSize+gcm.NonceSize():]

	return gcm.Open(out, nonce, raw, nil)
}

func (b *Barrier) Put(ctx context.Context, path string, entries []*backend.BackendEntry) error {
	return nil
}

func (b *Barrier) Get(ctx context.Context, path string) ([]*backend.BackendEntry, error) {
	return nil, nil
}

func (b *Barrier) List(ctx context.Context, pathPrefix string) ([]string, error) {
	return nil, nil
}

func (b *Barrier) Delete(ctx context.Context, path string) error {
	return nil
}
