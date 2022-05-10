package command

import (
	"encoding/base64"
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
	masterKeyBytes, err := internal.GenerateKey()

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate master key: %s", err.Error()))
		return 1
	}

	ic.ui.Info(fmt.Sprintf("master key: %s", base64.StdEncoding.EncodeToString(masterKeyBytes)))

	keyring := internal.NewKeyRing()

	if err := keyring.SetRootKey(masterKeyBytes); err != nil {
		ic.ui.Error(fmt.Sprintf("failed to set master key on keyring: %s", err.Error()))
		return 1
	}

	firstKey, err := internal.GenerateKey()
	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate initial encryption key: %s", err.Error()))
		return 1
	}

	err = keyring.AddKey(&internal.Key{
		Term:    1,
		Value:   firstKey,
		Version: 1,
	})

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to set initial encryption key on keyring: %s", err.Error()))
		return 1
	}

	_, err = internal.AESFromKey(masterKeyBytes)

	if err != nil {
		ic.ui.Error(fmt.Sprintf("failed to generate GCM: %s", err.Error()))
		return 1
	}

	// @TODO: encrypt and perist initial keyring

	// return success
	ic.ui.Info("Initialized keyring")
	return 0
}
