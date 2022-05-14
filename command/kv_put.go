package command

import (
	"github.com/mitchellh/cli"
)

type KVPutCommand struct {
	ui cli.Ui
}

func (kv KVPutCommand) Synopsis() string {
	return "Stores and encrypts key-value pairs at the provided path"
}

func (kv KVPutCommand) Help() string {
	helpText := `
Usage: keying put [options] <path> [data]

  This stores and encrypts a key-value pairs at the given path.
  The path is used to lookup the key-value pairs during decryption.

  There can be any number of key-value pairs. For example:

    $ keyring put secret/foo bar=baz key=secret

  The data can also be consumed from a file on disk by prefixing with the "@"
  symbol. For example:

    $ keyring put secret/foo @data.json

  Options:

    -root-token=<string>
      The root token to use for encryption.

      If not provided here, the '%s' environment
      variable will be used.
`
	return helpText
}

func (kv *KVPutCommand) Run(args []string) int {
	return 0
}
