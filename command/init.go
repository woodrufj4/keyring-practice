package command

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/backend"
	"github.com/woodrufj4/keyring-practice/backend/bbolt"
	"github.com/woodrufj4/keyring-practice/internal"
)

const (
	coreKeyringPath      = "core/keyring"
	coreKeyringCipherKey = "keyringCipher"
)

type InitCommand struct {
	ui cli.Ui
}

func (ic InitCommand) Synopsis() string {
	return "Initializes the keyring and datastore"
}

func (ic InitCommand) Help() string {
	return `
Usage: keyring init [options]

  Initializes the keyring and the datastore.

  Options:

    -path=<string>
      The file path where your secrets will be persisted to disc.
`
}

// Run initializes the keyring if it hasn't already been initialized.
//
// If not already initialized, this generates a master key,
// along with an initial keyring. It then encrypts and persist the
// keyring.
func (ic *InitCommand) Run(args []string) int {

	var dbPath string

	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.StringVar(&dbPath, "path", "", "The file path to the local datastore")
	err := fs.Parse(args)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to parse args: %s", err.Error()))
		return 1
	}

	// Check if there is an already initialized / persisted keyring
	backendConfig := bbolt.DefaultConfig()

	if dbPath != "" {
		backendConfig.Path = dbPath
	}

	fileBackend, err := bbolt.NewBoltBackend(backendConfig)
	defaultCtx := context.Background()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to instantiate backend: %s", err.Error()))
		return 1
	}

	if err := fileBackend.Setup(defaultCtx); err != nil {
		ic.ui.Error(fmt.Sprintf("failed to setup backend: %s", err.Error()))
		return 1
	}

	defer fileBackend.Cleanup(defaultCtx)

	// check if keyring exists
	existingEntries, err := fileBackend.Get(defaultCtx, coreKeyringPath)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to validate if backend is initialized: %s", err.Error()))
		return 1
	}

	if existingEntries != nil {
		// keyring is already initialized
		ic.ui.Info("keyring has already been initialized")
		return 0
	}

	// Firstly, generate the master key
	keyring, err := internal.InitNewKeyRing()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to initialize keyring: %s", err.Error()))
		return 1
	}

	gcm, err := internal.AESFromKey(keyring.RootKey())

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate GCM: %s", err.Error()))
		return 1
	}

	keyringBytes, err := keyring.Serialize()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to serialize keyring: %s", err.Error()))
		return 1
	}

	keyringCipher, err := internal.Encrypt(gcm, 0, keyringBytes)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to encrypt keyring: %s", err.Error()))
		return 1
	}

	// encrypt and perist initial keyring
	entries := []*backend.BackendEntry{
		{
			Key:   coreKeyringCipherKey,
			Value: keyringCipher,
		},
	}

	err = fileBackend.Put(defaultCtx, coreKeyringPath, entries)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to persist keying: %s", err.Error()))
		return 1
	}

	// display root token to user
	msg := `Keyring initialized!
This is the one and only time the root token will be displayed!

root token: %s
`

	ic.ui.Warn(fmt.Sprintf(msg, base64.StdEncoding.EncodeToString(keyring.RootKey())))

	return 0
}
