package main

import (
	"os"

	"github.com/k4lipso/gokill/cmd/gokillctl/commands"
	"github.com/k4lipso/gokill/internal"
)

func main() {

	if err := commands.RootCmd.Execute(); err != nil {
		internal.Log.Error(err.Error())
		os.Exit(1)
	}
}
