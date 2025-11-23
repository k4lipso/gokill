package main

import (
	"os"

	"github.com/k4lipso/gokill/cmd/gokillctl/commands"
)

func main() {

	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
