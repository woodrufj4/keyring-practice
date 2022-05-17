package internal

import (
	"testing"

	"github.com/woodrufj4/keyring-practice/backend"
)

func TestConfig_UnsupportedBackendType(t *testing.T) {

	config := BackendConfig{
		Type: "something unsupported",
	}

	_, err := config.GetBackendType()

	if err != ErrorUnsupportedBackendType {
		t.Fatalf("expected unsupported backend error, but got %s", err.Error())
	}

}

func TestConfig_SupportedBackendType(t *testing.T) {

	config := BackendConfig{
		Type: string(backend.FileBackend),
	}

	backendType, err := config.GetBackendType()

	if err != nil {
		t.Fatalf("failed to convert string to backend type: %s", err.Error())
	}

	if backendType != backend.FileBackend {
		t.Fatalf("failed to map a string to a valid backend type: %s", err.Error())
	}

}
