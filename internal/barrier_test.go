package internal

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/woodrufj4/keyring-practice/backend"
	"github.com/woodrufj4/keyring-practice/backend/bbolt"
)

const (
	DefaultTestKeyringPath = "keyring-test.db"
)

func TestBarrierInit(t *testing.T) {

	if testing.Short() {
		t.Skip("to slow for testing.Short. IO operations")
	}

	fileBackend := setupBackend(t)

	t.Cleanup(func() {

		if shutdownErr := fileBackend.Cleanup(context.Background()); shutdownErr != nil {
			t.Errorf("failed to cleanly shutdown the backend: %s", shutdownErr.Error())
		}

		if removeErr := os.Remove(DefaultTestKeyringPath); removeErr != nil {
			t.Errorf("failed to remove backend artifact: %s", removeErr.Error())
		}

	})

	barrier, err := NewBarrier(fileBackend)

	if err != nil {
		t.Fatalf("failed to instantiate barrier: %s", err.Error())
	}

	initialKey, err := barrier.GenerateKey()

	if err != nil {
		t.Fatalf("failed to generate random token: %s", err.Error())
	}

	err = barrier.Initialize(context.Background(), string(initialKey))

	if err != nil {
		t.Fatalf("failed to initialize barrier: %s", err.Error())
	}

	if !barrier.Initialized() {
		t.Fatalf("expected barrier to be properly initialized")
	}

}

func TestBarrierReInit(t *testing.T) {

	if testing.Short() {
		t.Skip("to slow for testing.Short. IO operations")
	}

	fileBackend := setupBackend(t)

	t.Cleanup(func() {

		if shutdownErr := fileBackend.Cleanup(context.Background()); shutdownErr != nil {
			t.Errorf("failed to cleanly shutdown the backend: %s", shutdownErr.Error())
		}

		if removeErr := os.Remove(DefaultTestKeyringPath); removeErr != nil {
			t.Errorf("failed to remove backend artifact: %s", removeErr.Error())
		}

	})

	barrier, err := NewBarrier(fileBackend)

	if err != nil {
		t.Fatalf("failed to instantiate barrier: %s", err.Error())
	}

	initialKey, err := barrier.GenerateKey()

	if err != nil {
		t.Fatalf("failed to generate random token: %s", err.Error())
	}

	err = barrier.Initialize(context.Background(), string(initialKey))

	if err != nil {
		t.Fatalf("failed to initialize barrier: %s", err.Error())
	}

	if !barrier.Initialized() {
		t.Fatalf("expected barrier to be properly initialized")
	}

	keyring1, err := barrier.Keyring()
	if err != nil {
		t.Fatalf("failed to retrieve barrier 1 keyring: %s", err.Error())
	}

	// This is simulating a new barrier after initial persistance
	barrier2, err := NewBarrier(fileBackend)

	if err != nil {
		t.Fatalf("failed to instantiate second test barrier: %s", err.Error())
	}

	err = barrier2.Initialize(context.Background(), string(initialKey))

	if err != nil {
		t.Fatalf("failed to initialize second test barrier: %s", err.Error())
	}

	if !barrier2.Initialized() {
		t.Fatalf("expected barrier 2 to be property initialized")
	}

	keyring2, err := barrier2.Keyring()
	if err != nil {
		t.Fatalf("failed to retrieve barrier 2 keyring: %s", err.Error())
	}

	if !bytes.Equal(keyring1.RootKey(), keyring2.RootKey()) {
		t.Fatalf("expected initial and persisted root keys to withstand encryption and decryption")
	}

	if keyring1.ActiveTerm() != keyring2.ActiveTerm() {
		t.Fatalf("expected the persisted active term to be %d, but got %d", keyring1.ActiveTerm(), keyring2.ActiveTerm())
	}

}

func TestBarrierExists(t *testing.T) {

	if testing.Short() {
		t.Skip("to slow for testing.Short. IO operations")
	}

	fileBackend := setupBackend(t)

	t.Cleanup(func() {

		if shutdownErr := fileBackend.Cleanup(context.Background()); shutdownErr != nil {
			t.Errorf("failed to cleanly shutdown the backend: %s", shutdownErr.Error())
		}

		if removeErr := os.Remove(DefaultTestKeyringPath); removeErr != nil {
			t.Errorf("failed to remove backend artifact: %s", removeErr.Error())
		}

	})

	barrier, err := NewBarrier(fileBackend)

	if err != nil {
		t.Fatalf("failed to instantiate barrier: %s", err.Error())
	}

	initialKey, err := barrier.GenerateKey()

	if err != nil {
		t.Fatalf("failed to generate random token: %s", err.Error())
	}

	err = barrier.Initialize(context.Background(), string(initialKey))

	if err != nil {
		t.Fatalf("failed to initialize barrier: %s", err.Error())
	}

	if !barrier.Initialized() {
		t.Fatalf("expected barrier to be properly initialized")
	}

	exists, err := barrier.KeyringPersisted(context.Background())

	if err != nil {
		t.Fatalf("failed to validate if keyring is persistedL: %s", err.Error())
	}

	if !exists {
		t.Fatalf("expected keyring to report it exists")
	}

}

func setupBackend(t *testing.T) backend.Backend {
	t.Helper()

	config := bbolt.DefaultConfig()
	config.Path = DefaultTestKeyringPath
	fileBackend, err := bbolt.NewBoltBackend(config)

	if err != nil {
		t.Fatalf("failed to instantiate backend: %s", err.Error())
	}

	defaultContext, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = fileBackend.Setup(defaultContext)

	if err != nil {
		t.Fatalf("failed to setup backend: %s", err.Error())
	}

	return fileBackend
}
