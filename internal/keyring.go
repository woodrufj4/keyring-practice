package internal

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"encoding/json"
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

type EncodedKeyring struct {
	MasterKey []byte
	Keys      []*Key
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

// InitNewKeyRing generates an initialized keyring with a master key and an
// initial encryption key.
func InitNewKeyRing() (*Keyring, error) {

	masterKey, err := GenerateKey()

	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %s", err.Error())
	}

	firstKey, err := GenerateKey()

	if err != nil {
		return nil, fmt.Errorf("failed to generate initial encryption key: %s", err.Error())
	}

	keyRing := &Keyring{
		keys:       make(map[uint32]*Key, 0),
		activeTerm: 0,
	}

	err = keyRing.SetRootKey(masterKey)

	if err != nil {
		return nil, fmt.Errorf("failed to generate initial encryption key: %s", err.Error())
	}

	err = keyRing.AddKey(&Key{
		Term:    1,
		Value:   firstKey,
		Version: 1,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to set initial encryption key on keyring: %s", err.Error())
	}

	return keyRing, nil

}

func DeserializeKeyring(buf []byte) (*Keyring, error) {

	var enc EncodedKeyring

	err := json.Unmarshal(buf, &enc)

	if err != nil {
		return nil, err
	}

	keyring := NewKeyRing()

	err = keyring.SetRootKey(enc.MasterKey)

	if err != nil {
		return nil, err
	}

	for _, key := range enc.Keys {
		keyring.keys[key.Term] = key

		if key.Term > keyring.activeTerm {
			keyring.activeTerm = key.Term
		}
	}

	return keyring, nil
}

func (k *Keyring) Serialize() ([]byte, error) {

	enc := &EncodedKeyring{
		MasterKey: k.RootKey(),
	}

	for _, key := range k.keys {
		enc.Keys = append(enc.Keys, key)
	}

	return json.Marshal(enc)
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
