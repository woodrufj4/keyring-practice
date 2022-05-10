package command

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
)

func CommandFactory() map[string]cli.CommandFactory {

	coloredUI := cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}

	commands := map[string]cli.CommandFactory{
		"encrypt": func() (cli.Command, error) {
			return &EncryptCommand{
				ui: &coloredUI,
			}, nil
		},
		"decrypt": func() (cli.Command, error) {
			return &DecryptCommand{
				ui: &coloredUI,
			}, nil
		},
		"init": func() (cli.Command, error) {
			return &InitCommand{
				ui: &coloredUI,
			}, nil
		},
	}
	return commands

}

func Run(args []string) int {

	cli := cli.CLI{
		Name:         "keyring",
		Version:      "0.1.0",
		Args:         args,
		Commands:     CommandFactory(),
		Autocomplete: true,
	}

	exitcode, err := cli.Run()

	if err != nil {
		log.Fatalf("failed to execute command. Error: %s\n", err.Error())
	}

	return exitcode
}
