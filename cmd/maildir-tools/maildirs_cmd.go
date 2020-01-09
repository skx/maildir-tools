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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/subcommands"
)

type maildirsCmd struct {

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

	f.BoolVar(&p.short, "short", false, "Show only the short-name.")
	f.BoolVar(&p.unreadOnly, "unread", false, "Show only folders containing unread messages.")
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.StringVar(&p.format, "format", "${06unread}/${06total} - ${name}", "The format string to display.")
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

	// The results we'll return
	var results []Maildir

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

		// Ignore the new/cur/tmp directories we'll terminate with
		for _, dir := range dirs {
			if strings.HasSuffix(path, dir) {
				return nil
			}
		}

		//
		// Ignore non-directories. (We might might find Dovecot
		// index-files, etc.)
		//
		mode := f.Mode()
		if !mode.IsDir() {
			return nil
		}

		//
		// Look for `new/`, `cur/`, and `tmp/` subdirectoires
		// beneath the given directory.
		//
		// If any are missing then this is not a maildir.
		//
		for _, dir := range dirs {
			path = filepath.Join(path, dir)
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

		r := regexp.MustCompile("^([0-9]+)(.*)$")

		//
		// Helper for expanding our format-string
		//
		mapper := func(field string) string {

			padding := ""
			match := r.FindStringSubmatch(field)
			if len(match) > 0 {
				padding = match[1]
				field = match[2]
			}

			ret := ""

			switch field {
			case "name":
				ret = ent
			case "total":
				ret = fmt.Sprintf("%d", messagesInMaildir(name))
			case "unread":
				if unreadCount < 0 {
					unreadCount = unreadMessagesInMaildir(name)
				}
				ret = fmt.Sprintf("%d", unreadCount)
			default:
				ret = "Unknown variable " + field
			}

			if padding != "" {

				// padding character
				char := " "
				if padding[0] == byte('0') {
					char = "0"
				}

				// size we need to pad to
				size, _ := strconv.Atoi(padding)
				for len(ret) < size {
					ret = char + ret
				}
			}
			return ret

		}

		// Are we only showing folders with unread messages?
		// If so continue unless this qualifies
		if p.unreadOnly && unreadCount < 1 {
			continue
		}

		results = append(results, Maildir{Path: name, Rendered: os.Expand(p.format, mapper)})
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
