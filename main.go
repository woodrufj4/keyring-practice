package main

import (
	"os"

	"github.com/woodrufj4/keyring-practice/command"
)

func main() {

	os.Exit(command.Run(os.Args[1:]))

}
