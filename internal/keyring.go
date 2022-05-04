package internal

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"time"
)

type Keyring struct {
	rootKey    []byte
	keys       map[uint32]*Key
	activeTerm uint32
}

type Key struct {
	Term        uint32
	Value       []byte
	Version     uint
	InstallTime time.Time
}

// GenerateKey is used to generate a new key vaule
func GenerateKey() ([]byte, error) {

	// Generate a 256bit key
	buf := make([]byte, 2*aes.BlockSize)
	_, err := rand.Read(buf)

	return buf, err
}

// NewKeyRing generates a new key ring.
func NewKeyRing() *Keyring {
	return &Keyring{
		keys:       make(map[uint32]*Key, 0),
		activeTerm: 0,
	}
}

// SetRootKey sets the root key.
func (k *Keyring) SetRootKey(key []byte) error {

	// Verify correct key size
	min, max := aes.BlockSize, 2*aes.BlockSize

	switch len(key) {
	case min, max:
		break
	default:
		return fmt.Errorf("key size must be %d or %d", min, max)
	}

	k.rootKey = key
	return nil
}

// RootKey returns the current value of the root key.
func (k *Keyring) RootKey() []byte {
	return k.rootKey
}

// AddKey adds a key to the key ring and makes the newly add key
// the active key
func (k *Keyring) AddKey(key *Key) error {

	// Ensure there is no conflict
	if exist, ok := k.keys[key.Term]; ok {

		if !bytes.Equal(key.Value, exist.Value) {
			return fmt.Errorf("conflicting key for term %d already installed", key.Term)
		}

		// Attempting to add the exact same key here... so, we'll just return 👍
		return nil
	}

	key.InstallTime = time.Now()
	k.keys[key.Term] = key

	// Update the active term if newer
	if key.Term > k.activeTerm {
		k.activeTerm = key.Term
	}

	return nil
}

// RemoveKey removes a key from the keyring... as long as the key being
// removed isn't the currently active key.
func (k *Keyring) RemoveKey(term uint32) error {

	// Ensure this is not the active key
	if term == k.activeTerm {
		return fmt.Errorf("cannot remove active key")
	}

	// Check if this term does not exist
	if _, ok := k.keys[term]; !ok {
		return nil
	}

	delete(k.keys, term)
	return nil

}

// ActiveTerm reports the currently active key term.
func (k *Keyring) ActiveTerm() uint32 {
	return k.activeTerm
}

// ActiveKey reports the currently active key.
func (k *Keyring) ActiveKey() *Key {
	return k.keys[k.activeTerm]
}

// TermKey retrieves the key given the key term.
func (k *Keyring) TermKey(term uint32) *Key {
	return k.keys[term]
}
