//
// Show messages in the given directory
//

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/subcommands"
)

type messageCmd struct {

	// The prefix to our maildir hierarchy
	prefix string

	// The format-string to use for displaying messages
	format string
}

//
// Glue
//
func (*messageCmd) Name() string     { return "messages" }
func (*messageCmd) Synopsis() string { return "Show the messages in the given directory." }
func (*messageCmd) Usage() string {
	return `messages :
  Show the messages in the specified maildir folder.
`
}

//
// Flag setup
//
func (p *messageCmd) SetFlags(f *flag.FlagSet) {
	prefix := os.Getenv("HOME") + "/Maildir/"

	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.StringVar(&p.format, "format", "[${flags}] ${subject}", "Specify the format-string to use for the message-display")
}

//
// Show the messages in the given folder
//
func (p *messageCmd) showMessages(path string) {

	//
	// Ensure the maildir exists
	//
	prefixes := []string{path, filepath.Join(p.prefix, path)}

	//
	// Found it yet?
	//
	found := false

	//
	// Test both the complete path, and the directory relative
	// to our prefix-root.
	//
	for _, possible := range prefixes {
		if _, err := os.Stat(possible); !os.IsNotExist(err) {
			found = true
			path = possible
		}
	}
	if !found {
		fmt.Printf("maildir '%s' wasn't found\n", path)
		return
	}

	//
	// Get the files
	//
	var files []string

	dirs := []string{"cur", "new", "tmp"}

	for _, dir := range dirs {

		prefix := filepath.Join(path, dir)

		_ = filepath.Walk(prefix, func(path string, f os.FileInfo, err error) error {
			switch mode := f.Mode(); {
			case mode.IsRegular():
				// nop
			default:
				return nil
			}

			files = append(files, path)

			return nil
		})
	}

	//
	// For each file - parse the email message and output a summary.
	//
	for _, msg := range files {

		//
		// Open it
		//
		file, err := os.Open(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %s: %s\n",
				msg, err.Error())
			continue
		}

		//
		// Create a mail-object.
		//
		m, err := mail.ReadMessage(bufio.NewReader(file))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file %s: %s\n",
				msg, err.Error())
			continue
		}

		//
		// Ensure we don't leak.
		//
		file.Close()

		header := m.Header

		headerMapper := func(field string) string {
			switch field {
			case "flags":
				//
				flags := ""

				// get the flags
				i := strings.Index(msg, ":2,")
				if i > 0 {
					flags = msg[i+3:]
				}

				// Add on a fake (N)ew flag
				if strings.Contains(msg, "/new/") {
					flags += "N"
				}

				s := strings.Split(flags, "")
				sort.Strings(s)
				return (strings.Join(s, ""))

			case "file":
				return msg
			default:
				return (header.Get(field))
			}
		}

		fmt.Println(os.Expand(p.format, headerMapper))
	}
}

//
// Entry-point.
//
func (p *messageCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	for _, path := range f.Args() {
		p.showMessages(path)
	}
	return subcommands.ExitSuccess
}
