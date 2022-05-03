package command

import (
	"log"

	"github.com/mitchellh/cli"
)

func CommandFactory() map[string]cli.CommandFactory {

	commands := map[string]cli.CommandFactory{}
	return commands

}

func Run(args []string) int {

	cli := cli.CLI{
		Name:     "Keyring Practice",
		Version:  "0.1.0",
		Args:     args,
		Commands: CommandFactory(),
	}

	exitcode, err := cli.Run()

	if err != nil {
		log.Fatalf("failed to execute command. Error: %s\n", err.Error())
	}

	return exitcode
}
