package command

import "github.com/mitchellh/cli"

type TransitCommand struct {
	ui cli.Ui
}

func (t TransitCommand) Synopsis() string {
	return "Encrypts a key-value pairs at the provided path"
}

func (t TransitCommand) Help() string {
	helpText := `
Usage: keying transit [options] <subcommand>

  Performs transitive encrypt / decrypt operations without persisting
  the encrypted secrets to the keyring's backend.

  Please see individual subcommand help for detailed usage information.`
	return helpText
}

func (t TransitCommand) Run(args []string) int {
	return cli.RunResultHelp
}
