package command

import "github.com/mitchellh/cli"

type DecryptCommand struct {
	ui cli.Ui
}

func (ec DecryptCommand) Synopsis() string {
	return "Decrypts a ciphertext into base64 encoded plain text"
}

func (ec DecryptCommand) Help() string {
	return `
Usage: keying decrypt [options] <ciphertext>

  This decrypts a ciphertext into base64 encoded text.
`
}

func (ec *DecryptCommand) Run(args []string) int {
	return 0
}
