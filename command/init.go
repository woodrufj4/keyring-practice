package command

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/backend/bbolt"
	"github.com/woodrufj4/keyring-practice/internal"
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

	barrier, err := internal.NewBarrier(fileBackend)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to instantiate barrier: %s", err.Error()))
		return 1
	}

	exists, err := barrier.KeyringPersisted(defaultCtx)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to validate if barrier pre-exists: %s", err.Error()))
		return 1
	}

	if exists {
		ic.ui.Info("keyring already initialized")
		return 0
	}

	initialKey, err := barrier.GenerateKey()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate initial root token: %s", err.Error()))
		return 1
	}

	err = barrier.Initialize(defaultCtx, string(initialKey))

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to initialize barrier: %s", err.Error()))
		return 1
	}

	// display root token to user
	msg := `Keyring initialized!
This is the one and only time the root token will be displayed!

root token: %s
`

	ic.ui.Warn(fmt.Sprintf(msg, base64.StdEncoding.EncodeToString(initialKey)))

	return 0
}
