package command

import "github.com/mitchellh/cli"

type KVGetCommand struct {
	ui cli.Ui
}

func (kv KVGetCommand) Synopsis() string {
	return "Retrieves the key-value pairs at the provided path"
}

func (kv KVGetCommand) Help() string {
	helpText := `
Usage: keying get [options] <path>

  This retrieves and decrypts key-value pairs at given a path.

  Example:

    $ keyring get secret/foo

  Options:

    -root-token=<string>
      The root token to use for encryption.

      If not provided here, the '%s' environment
      variable will be used.
`
	return helpText
}

func (kv *KVGetCommand) Run(args []string) int {
	return 0
}
