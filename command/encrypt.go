package command

import "github.com/mitchellh/cli"

type EncryptCommand struct {
	ui cli.Ui
}

func (ec EncryptCommand) Synopsis() string {
	return "Encrypts a plain text input to a cipher text."
}

func (ec EncryptCommand) Help() string {
	return `
Usage: keying encrypt [options] <plaintext>

  This encrypts a plain text input into a ciphertext.
`
}

func (ec *EncryptCommand) Run(args []string) int {
	return 0
}
