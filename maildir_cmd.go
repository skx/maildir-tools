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

	// Format string for output
	format string

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

	f.BoolVar(&p.short, "short", false, "Show only the short-name.")
	f.BoolVar(&p.unreadOnly, "unread", false, "Show only folders containing unread messages.")
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.StringVar(&p.format, "format", "${unread}/${total} - ${name}", "The format string to display.")
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

		//
		// Copy the name, in case we truncate it
		//
		name := ent

		if p.short {
			ent = ent[len(p.prefix):]
		}

		//
		// Unread-cache
		//
		unreadCount := -1

		if p.unreadOnly {
			unreadCount = unreadMessagesInMaildir(name)
		}

		//
		// Helper for expanding our format-string
		//
		mapper := func(field string) string {
			switch field {
			case "name":
				return ent
			case "total":
				return fmt.Sprintf("%d", messagesInMaildir(name))
			case "unread":
				if unreadCount < 0 {
					unreadCount = unreadMessagesInMaildir(name)
				}
				return fmt.Sprintf("%d", unreadCount)
			default:
				return "Unknown variable " + field
			}
		}

		// Are we only showing folders with unread messages?
		// If so continue unless this qualifies
		if p.unreadOnly && unreadCount < 1 {
			continue
		}

		// Show the output
		fmt.Println(os.Expand(p.format, mapper))

	}
}

//
// Entry-point.
//
func (p *maildirCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.showMaildirs()
	return subcommands.ExitSuccess
}
