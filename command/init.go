package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/internal"
)

const (
	coreKeyringPath = "core/keyring"
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

    -directory=<string>
      The directory where your secrets will be persisted to disc.
`
}

// Run initializes the keyring if it hasn't already been initialized.
//
// If not already initialized, this generates a master key,
// along with an initial keyring. It then encrypts and persist the
// keyring.
func (ic *InitCommand) Run(args []string) int {

	// @TODO: Check if there is an already initialized / persisted keyring

	// Firstly, generate the master key
	keyring, err := internal.InitNewKeyRing()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to initialize keyring: %s", err.Error()))
		return 1
	}

	_, err = internal.AESFromTerm(keyring.ActiveTerm(), keyring)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate GCM: %s", err.Error()))
		return 1
	}

	// @TODO: encrypt and perist initial keyring

	// return success
	ic.ui.Info("Initialized keyring")
	return 0
}
