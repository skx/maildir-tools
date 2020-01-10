//
// Show messages in the given Maildir folder.
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
	"github.com/skx/maildir-tools/finder"
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
	f.StringVar(&p.format, "format", "[${index}/${total} - ${flags}] ${subject}", "Specify the format-string to use for the message-display")
}

//
// Show the messages in the given folder
//
func (p *messagesCmd) GetMessages(path string, format string) ([]SingleMessage, error) {

	//
	// The messages we'll find
	//
	var messages []SingleMessage

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
		return messages, fmt.Errorf("maildir '%s' wasn't found", path)
	}

	finder := finder.New(p.prefix)
	files := finder.Messages(path)

	//
	// For each file - parse the email message and output a summary.
	//
	for index, msg := range files {

		//
		// Get the mail
		//
		mail := mailreader.New(msg)

		r := regexp.MustCompile("^([0-9]+)(.*)$")

		//
		// Expand the template-string
		//
		headerMapper := func(field string) string {

			padding := ""
			match := r.FindStringSubmatch(field)
			if len(match) > 0 {
				padding = match[1]
				field = match[2]
			}

			ret := ""

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
				ret = (strings.Join(s, ""))

			case "file":
				ret = msg
			case "index":
				ret = fmt.Sprintf("%d", index+1)
			case "total":
				ret = fmt.Sprintf("%d", len(files))
			default:
				ret, _ = mail.Header(field)
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

		messages = append(messages, SingleMessage{Path: msg,
			Rendered: os.Expand(format, headerMapper)})
	}

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
