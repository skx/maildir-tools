//
// Show maildirs beneath the given root.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/subcommands"
)

type maildirCmd struct {

	// Show only the short maildir names
	short bool

	// Show only folders with unread mail?
	unreadOnly bool

	// The root directory to our maildir hierarchy
	prefix string
}

//
// Glue
//
func (*maildirCmd) Name() string     { return "maildirs" }
func (*maildirCmd) Synopsis() string { return "Show maildir folders beneath the root." }
func (*maildirCmd) Usage() string {
	return `maildirs :
  Show maildir folders beneath the given root directory, recursively.
`
}

//
// Flag setup
//
func (p *maildirCmd) SetFlags(f *flag.FlagSet) {

	prefix := os.Getenv("HOME") + "/Maildir/"

	f.BoolVar(&p.unreadOnly, "unread", false, "Show only folders containing unread messages.")
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.BoolVar(&p.short, "short", false, "Show only the short-name.")
}

//
// Find and display the folders
//
func (p *maildirCmd) showMaildirs() {

	//
	// Directories we've found.
	//
	maildirs := []string{}

	//
	// Subdirectories we care about
	//
	dirs := []string{"cur", "new", "tmp"}

	//
	// Find maildirs
	//
	_ = filepath.Walk(p.prefix, func(path string, f os.FileInfo, err error) error {
		//
		// Ignore non-directories
		//
		switch mode := f.Mode(); {
		case mode.IsDir():
			// nop
		default:
			return nil
		}

		//
		// Look for `new/`, `cur/`, and `tmp/` subdirectoires
		// beneath the given directory.
		//
		for _, dir := range dirs {
			path := filepath.Join(path, dir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return nil
			}
		}

		//
		// Found a positive result
		//
		maildirs = append(maildirs, path)
		return nil
	})

	//
	// Sorted results are nicer.
	//
	sort.Strings(maildirs)

	//
	// Show the results
	//
	for _, ent := range maildirs {

		if p.short {
			ent = ent[len(p.prefix):]
		}
		fmt.Printf("%s\n", ent)
	}
}

//
// Entry-point.
//
func (p *maildirCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.showMaildirs()
	return subcommands.ExitSuccess
}
