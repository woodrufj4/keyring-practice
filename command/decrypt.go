package command

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/internal"
)

type DecryptCommand struct {
	ui cli.Ui
}

func (dc DecryptCommand) Synopsis() string {
	return "Decrypts a ciphertext into base64 encoded plain text"
}

func (dc DecryptCommand) Help() string {
	helpText := `
Usage: keying transit decrypt [options] <base64ciphertext>

  This decrypts a base64 encoded ciphertext into plain text.

  Options:

    -secret-key=<string>
      The secret key to use for decryption. This must be the same
      key that was used to encrypt the plaintext value.

      If not provided here, the '%s' environment
      variable will be used.
`

	return fmt.Sprintf(helpText, EnvRootToken)
}

func (dc *DecryptCommand) Run(args []string) int {

	var secretKey string

	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	fs.StringVar(&secretKey, "secret-key", "", "The secret key used to decrypt ciphertext")

	if err := fs.Parse(args); err != nil {
		dc.ui.Error("failed to parse flags")
		return 1
	}

	if secretKey == "" {
		secretKey = os.Getenv(EnvRootToken)

		if secretKey == "" {
			dc.ui.Error("missing secret key")
			return 1
		}
	}

	if len(fs.Args()) != 1 {
		dc.ui.Error("expected only one argument")
		return 1
	}

	cipherText, err := base64.StdEncoding.DecodeString(fs.Arg(0))

	if err != nil {
		dc.ui.Error(fmt.Sprintf("Failed to decode ciphertext. %s", err.Error()))
		return 1
	}

	gcm, err := internal.AESFromKey([]byte(secretKey))

	if err != nil {
		dc.ui.Error(fmt.Sprintf("failed to generate GCM. %s", err.Error()))
		return 1
	}

	if len(cipherText) < gcm.NonceSize() {
		dc.ui.Error("length of ciphertext is invalid")
		return 1
	}

	capacity := len(cipherText) + gcm.NonceSize() + gcm.Overhead()

	out := make([]byte, 0, capacity)

	nonce := cipherText[0:gcm.NonceSize()]

	raw := cipherText[gcm.NonceSize():]

	plaintext, err := gcm.Open(out, nonce, raw, nil)

	if err != nil {
		dc.ui.Error(fmt.Sprintf("failed to decrypt ciphertext. %s", err.Error()))
		return 1
	}

	dc.ui.Output(fmt.Sprintf("plaintext: %s", plaintext))

	return 0
}
