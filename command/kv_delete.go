package command

import "github.com/mitchellh/cli"

type KVDeleteCommand struct {
	ui cli.Ui
}

func (kv KVDeleteCommand) Synopsis() string {
	return "Deletes the key-value pairs from the data store at the provided path"
}

func (kv KVDeleteCommand) Help() string {
	helpText := `
Usage: keying delete [options] <path>

  Deletes all the key-value pairs given the provided path.

  Example:

    $ keyring delete secret/foo

  Options:

    -root-token=<string>
      The root token to use for encryption.

      If not provided here, the '%s' environment
      variable will be used.
`
	return helpText
}

func (kv *KVDeleteCommand) Run(args []string) int {
	return 0
}
