package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/mitchellh/cli"
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

  Backend Options:

    -backend-type=<string>
      The type of backend to use.
      Currently, only the 'file' type backend is supported,
      and is also the default. 


    File Backend Options:

      -filepath=<string>
        The file path where your secrets will be persisted to disc.
`
}

// Run initializes the keyring if it hasn't already been initialized.
//
// If not already initialized, this generates a master key,
// along with an initial keyring. It then encrypts and persist the
// keyring.
func (ic *InitCommand) Run(args []string) int {

	defaultCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	config, _, err := internal.ReadConfig(args)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to read config: %s", err.Error()))
		return 1
	}

	initBackend, err := internal.SetupBackend(defaultCtx, config)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to configure backend: %s", err.Error()))
		return 1
	}

	defer func() {
		cleanupErr := initBackend.Cleanup(defaultCtx)
		if cleanupErr != nil {
			ic.ui.Warn(fmt.Sprintf("failed to gracefully shutdown backend: %s", cleanupErr.Error()))
		}
	}()

	barrier, err := internal.NewBarrier(initBackend)

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
