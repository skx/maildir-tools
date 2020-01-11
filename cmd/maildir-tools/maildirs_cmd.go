//
// Show maildirs beneath the given root.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/skx/maildir-tools/finder"
	"github.com/skx/maildir-tools/formatter"
)

type maildirsCmd struct {

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
func (*maildirsCmd) Name() string     { return "maildirs" }
func (*maildirsCmd) Synopsis() string { return "Show maildir folders beneath the root." }
func (*maildirsCmd) Usage() string {
	return `maildirs :
  Show maildir folders beneath the given root directory, recursively.
`
}

//
// Flag setup
//
func (p *maildirsCmd) SetFlags(f *flag.FlagSet) {

	prefix := os.Getenv("HOME") + "/Maildir/"

	f.BoolVar(&p.unreadOnly, "unread", false, "Show only folders containing unread messages.")
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.StringVar(&p.format, "format", "#{06unread}/#{06total} - #{name}", "The format string to display.")
}

// Maildir is the type of object we return from our main
// function.
type Maildir struct {

	// Path contains the complete path to the maildir.
	Path string

	// Rendered contains the maildir formated via the
	// supplied format-string.
	Rendered string
}

//
// Find and display the folders
//
func (p *maildirsCmd) GetMaildirs() []Maildir {

	//
	// The results we'll return
	//
	var results []Maildir

	//
	// Find the maildir entries beneath our
	// prefix directory.
	//
	finder := finder.New(p.prefix)
	maildirs := finder.Maildirs()

	//
	// Do we need to count files?
	//
	count := false
	if strings.Contains(p.format, "total}") ||
		strings.Contains(p.format, "unread}") ||
		p.unreadOnly {
		count = true
	}

	//
	// Build up the formatted results, according to
	// our formatted string.
	//
	for _, ent := range maildirs {

		//
		// Copy the name, in case we truncate it
		//
		name := ent

		//
		// Count of unread and total messages in the
		// maildir.  These might not be used.
		//
		unread := 0
		total := 0

		//
		// Count files if we're supposed to
		//
		if count {
			messages := finder.Messages(name)
			total = len(messages)

			for _, entry := range messages {
				if strings.Contains(entry, "/new/") {
					unread++
				}
			}
		}

		//
		// Helper for expanding our format-string
		//
		mapper := func(field string) string {

			ret := ""

			switch field {
			case "name":
				ret = ent
			case "shortname":
				ret = ent[len(p.prefix):]
			case "total":
				ret = fmt.Sprintf("%d", total)
			case "unread":
				ret = fmt.Sprintf("%d", unread)
			default:
				ret = "Unknown variable " + field
			}

			return ret
		}

		// Are we only showing folders with unread messages?
		// If so continue unless this qualifies
		if p.unreadOnly && unread < 1 {
			continue
		}

		results = append(results, Maildir{Path: name, Rendered: formatter.Expand(p.format, mapper)})
	}

	return results
}

//
// Entry-point.
//
func (p *maildirsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	maildirs := p.GetMaildirs()
	for _, ent := range maildirs {
		fmt.Println(ent.Rendered)
	}

	return subcommands.ExitSuccess
}
