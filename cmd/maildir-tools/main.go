package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

//
// Setup our sub-commands and use them.
//
func main() {

	// Defaults
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	// Our commands
	subcommands.Register(&maildirsCmd{}, "")
	subcommands.Register(&messagesCmd{}, "")
	subcommands.Register(&messageCmd{}, "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&uiCmd{}, "")
	subcommands.Register(&tuiCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
