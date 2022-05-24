package internal

import (
	"errors"
	"os"
	"strings"

	"github.com/woodrufj4/keyring-practice/backend"
)

const (
	DefaultEnvRootToken = "KEYRING_ROOT_TOKEN"
)

var (
	ErrorUnsupportedBackendType = errors.New("unsupported backend type")
)

type GeneralConfig struct {
	RootToken string         `json:"rootToken"`
	Backend   *BackendConfig `json:"backend"`
}

func (gc *GeneralConfig) Merge(b *GeneralConfig) *GeneralConfig {

	result := *gc

	if b.RootToken != "" {
		result.RootToken = b.RootToken
	}

	if b.Backend != nil {
		if result.Backend != nil {
			result.Backend = result.Backend.Merge(b.Backend)
		} else {
			result.Backend = b.Backend
		}
	}

	return &result
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		RootToken: os.Getenv(DefaultEnvRootToken),
		Backend: &BackendConfig{
			Type:    string(backend.FileBackend),
			Options: map[string]interface{}{},
		},
	}
}

// BackendConfig is the general configuration given the type
// of backend.
type BackendConfig struct {
	Type    string      `json:"type"`
	Options interface{} `json:"options"`
}

func (bc *BackendConfig) Merge(b *BackendConfig) *BackendConfig {

	result := *bc

	if b.Type != "" {
		result.Type = b.Type
	}

	if b.Options != nil {
		result.Options = b.Options
	}

	return &result
}

func (bc *BackendConfig) GetBackendType() (backend.BackendType, error) {
	switch strings.TrimSpace(strings.ToLower(bc.Type)) {

	case string(backend.FileBackend):
		return backend.FileBackend, nil

	default:
		return "", ErrorUnsupportedBackendType
	}
}
