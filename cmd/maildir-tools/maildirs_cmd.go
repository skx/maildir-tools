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
	// Find the maildir entries beneath our prefix directory.
	//
	finder := finder.New(p.prefix)
	maildirs := finder.Maildirs()

	//
	// Do we need to count files?
	//
	// If we can avoid it that speeds things up :)
	//
	count := false
	if strings.Contains(p.format, "total}") || strings.Contains(p.format, "unread}") {
		count = true
	}

	//
	// Now we know how many results to expect.
	//
	results = make([]Maildir, len(maildirs))

	//
	// Build up the formatted results.
	//
	for index, ent := range maildirs {

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
			messages := finder.Messages(ent)
			total = len(messages)

			for _, entry := range messages {

				// TODO - Fix me
				//
				// A message is unread if EITHER
				//
				// A) it is in the new/ folder
				//
				// B) It does NOT have the `S`een flag.
				//
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

		//
		// Save the results
		//
		results[index] = Maildir{Path: ent, Rendered: formatter.Expand(p.format, mapper)}
	}

	return results
}

//
// Entry-point.
//
func (p *maildirsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Get all the maildirs we know about
	//
	maildirs := p.GetMaildirs()

	//
	// For each one, show the formatted output
	//
	for _, ent := range maildirs {
		fmt.Println(ent.Rendered)
	}

	//
	// All done.
	//
	return subcommands.ExitSuccess
}
