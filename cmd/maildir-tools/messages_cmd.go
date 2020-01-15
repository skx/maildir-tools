// Show messages in the given Maildir folder.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/skx/maildir-tools/finder"
	"github.com/skx/maildir-tools/formatter"
	"github.com/skx/maildir-tools/mailreader"
)

// messageCmd holds the state for this sub-command
type messagesCmd struct {

	// The prefix to our maildir hierarchy
	prefix string

	// The format-string to use for displaying messages
	format string
}

// SingleMessage holds the state for a single message
type SingleMessage struct {

	// Path contains the path to the file on-disk
	Path string

	// Rendered contains the rendered result of using
	// a format-string to output the message.
	Rendered string
}

//
// Glue
//
func (*messagesCmd) Name() string     { return "messages" }
func (*messagesCmd) Synopsis() string { return "Show the messages in the given directory." }
func (*messagesCmd) Usage() string {
	return `messages :
  Show the messages in the specified maildir folder.
`
}

//
// Flag setup
//
func (p *messagesCmd) SetFlags(f *flag.FlagSet) {
	prefix := os.Getenv("HOME") + "/Maildir/"

	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
	f.StringVar(&p.format, "format", "[#{index}/#{total} - #{5flags}] #{subject}", "Specify the format-string to use for the message-display")
}

// Find the absolute path to the given maildir folder
func (p *messagesCmd) getMaildirPath(input string) (string, error) {

	//
	// We look for the absolute path, then the path
	// beneath our root prefix.
	//
	prefixes := []string{input, filepath.Join(p.prefix, input)}

	//
	// Test each prefix.
	//
	for _, possible := range prefixes {

		// Found it?  Great
		if _, err := os.Stat(possible); !os.IsNotExist(err) {
			return possible, nil
		}
	}

	// No match
	return "", fmt.Errorf("maildir '%s' wasn't found", input)
}

// Get the messages in the given folder.
//
// If the path is absolute it will be used as-is, otherwise we'll hunt
// for it beneath our configured prefix.
func (p *messagesCmd) GetMessages(path string, format string) ([]SingleMessage, error) {

	//
	// The messages we'll find
	//
	var messages []SingleMessage

	//
	// Get the fully-qualified path to the given maildir
	// folder.
	//
	// e.g. "people-foo" becomes "/home/blah/Maildir/people-foo",
	// we need this path to find the messages from.
	//
	path, err := p.getMaildirPath(path)
	if err != nil {
		return messages, err
	}

	//
	// Helper for finding messages.
	//
	finder := finder.New(p.prefix)

	//
	// Find the messages
	//
	files := finder.Messages(path)

	//
	// We know how many messages to expect now.
	//
	messages = make([]SingleMessage, len(files))

	//
	// For each file - parse the email message and generate a summary.
	//
	for index, msg := range files {

		//
		// Read the mail, so we can access the data.
		//
		mail, err := mailreader.New(msg)
		if err != nil {
			return messages, err
		}

		//
		// Expand the template-string
		//
		headerMapper := func(field string) string {

			ret := ""

			switch field {
			case "unread_highlight":
				// Allow highlighting in the UI
				if strings.Contains(mail.Flags(), "N") {
					return "[red]"
				}
				return ""
			case "flags":
				ret = mail.Flags()
			case "file":
				ret = msg
			case "index":
				ret = fmt.Sprintf("%d", index+1)
			case "total":
				ret = fmt.Sprintf("%d", len(files))
			default:
				ret = mail.Header(field)
			}

			return ret
		}

		//
		// Record the entry.
		//
		messages[index] = SingleMessage{Path: msg,
			Rendered: formatter.Expand(format, headerMapper)}
	}

	//
	// All done.
	//
	return messages, nil
}

//
// Entry-point.
//
func (p *messagesCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	for _, path := range f.Args() {

		messages, err := p.GetMessages(path, p.format)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return subcommands.ExitFailure
		}
		for _, ent := range messages {
			fmt.Println(ent.Rendered)
		}
	}
	return subcommands.ExitSuccess
}
