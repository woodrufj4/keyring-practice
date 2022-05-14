package command

import "github.com/mitchellh/cli"

type KVListCommand struct {
	ui cli.Ui
}

func (kv KVListCommand) Synopsis() string {
	return "Lists the stored paths given a path prefix"
}

func (kv KVListCommand) Help() string {
	helpText := `
Usage: keying list [options] <path-prefix>

  Lists the stored paths given a path prefix

  Example:

    $ keyring list secret

  Options:

    -root-token=<string>
      The root token to access the keyring.

      If not provided here, the '%s' environment
      variable will be used.
`
	return helpText
}

func (kv *KVListCommand) Run(args []string) int {
	return 0
}
