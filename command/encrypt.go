package command

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/internal"
)

const (
	EnvSecretKey = "KEYRING_SECRET_KEY"
)

type EncryptCommand struct {
	ui cli.Ui
}

func (ec EncryptCommand) Synopsis() string {
	return "Encrypts a plain text input to a cipher text."
}

func (ec EncryptCommand) Help() string {
	helpText := `
Usage: keying transit encrypt [options] <plaintext>

  This encrypts a plain text input into a base64 encoded ciphertext.

  Options:

    -secret-key=<string>
      The secret key to use for encryption.

      If not provided here, the '%s' environment
      variable will be used.

      The secret key should be an AES key,
      either 16, 24, or 32 bytes long to select
      AES-128, AES-192, or AES-256.
`

	return fmt.Sprintf(helpText, EnvSecretKey)
}

func (ec *EncryptCommand) Run(args []string) int {

	var secretKey string

	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	fs.StringVar(&secretKey, "secret-key", "", "The secret key used to encrypt ciphertext")

	if err := fs.Parse(args); err != nil {
		ec.ui.Error("failed to parse flags")
		return 1
	}

	if secretKey == "" {
		secretKey = os.Getenv(EnvSecretKey)

		if secretKey == "" {
			ec.ui.Error("missing secret key")
			return 1
		}
	}

	if len(fs.Args()) != 1 {
		ec.ui.Error("expected only one argument")
		return 1
	}

	plaintext := fs.Arg(0)

	gcm, err := internal.AESFromKey([]byte(secretKey))

	if err != nil {
		ec.ui.Error(fmt.Sprintf("failed to generate GCM. %s", err.Error()))
		return 1
	}

	overhead := gcm.NonceSize() + gcm.Overhead()

	if len(plaintext) > math.MaxInt-overhead {
		ec.ui.Error("plaintext is too large")
		return 1
	}

	capacity := len(plaintext) + overhead

	size := gcm.NonceSize()

	out := make([]byte, size, capacity)

	nonce := out[0:gcm.NonceSize()]

	n, err := rand.Read(nonce)

	if err != nil {
		ec.ui.Error(fmt.Sprintf("failed to read random bytes to nonce %s", err.Error()))
		return 1
	}

	if n != len(nonce) {
		ec.ui.Error("unable to read enough random bytes to fill gcm nonce")
		return 1
	}

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		ec.ui.Error(fmt.Sprintf("failed to create nonce. %s", err.Error()))
		return 1
	}

	cipherBytes := gcm.Seal(out, nonce, []byte(plaintext), nil)

	ec.ui.Info(fmt.Sprintf("ciphertext: %s", base64.StdEncoding.EncodeToString(cipherBytes)))

	return 0
}
