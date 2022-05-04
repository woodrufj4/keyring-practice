package internal

import (
	"bytes"
	"testing"
)

func TestKeyRing(t *testing.T) {

	k := NewKeyRing()

	if k.ActiveTerm() != 0 {
		t.Fatalf("expected currently active term to be 0, but was %d", k.ActiveTerm())
	}

	if key := k.ActiveKey(); key != nil {
		t.Fatalf("expected not have have any currently active keys. current active key:\n%#v", key)
	}
}

func TestRootKey(t *testing.T) {

	k := NewKeyRing()

	rootKey, err := GenerateKey()

	if err != nil {
		t.Fatalf("failed to generate random key. Error: %s", err.Error())
	}

	if err := k.SetRootKey(rootKey); err != nil {
		t.Fatalf("failed to set root key. Error: %s", err.Error())
	}

	if !bytes.Equal(k.RootKey(), rootKey) {
		t.Fatalf("expected to retrieve the same root key as what was set")
	}

}

func TestKeyRingAddKey(t *testing.T) {

	k := NewKeyRing()

	key := &Key{
		Term:    1,
		Value:   []byte("some-value"),
		Version: 1,
	}

	err := k.AddKey(key)

	if err != nil {
		t.Fatalf("failed to add key to keyring. Error: %s", err.Error())
	}

	if k.ActiveTerm() != key.Term {
		t.Fatalf("expected the active term to be %d, but was %d", key.Term, k.ActiveTerm())
	}

}

func TestKeyRingRemoveKey(t *testing.T) {
	k := NewKeyRing()

	key := &Key{
		Term:    1,
		Value:   []byte("some-value"),
		Version: 1,
	}

	err := k.AddKey(key)

	if err != nil {
		t.Fatalf("failed to add key to keyring. Error: %s", err.Error())
	}

	if err := k.RemoveKey(key.Term); err == nil {
		t.Fatalf("expected an error when attempting to remove currently active key")
	}

	key2 := &Key{
		Term:    2,
		Value:   []byte("some-other-value"),
		Version: 1,
	}

	err = k.AddKey(key2)

	if err != nil {
		t.Fatalf("failed to add key to keyring. Error: %s", err.Error())
	}

	if err := k.RemoveKey(key.Term); err != nil {
		t.Fatalf("failed to remove non-active key. Error: %s", err.Error())
	}

}
