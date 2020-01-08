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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/google/subcommands"
	"github.com/jhillyerd/enmime"
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

	err error
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

///
// worker stuff
///

type job struct {
	index int
	total int
	path  string
}

func (p *messagesCmd) worker(jobs <-chan job, results chan<- SingleMessage) {
	for j := range jobs {

		var t SingleMessage

		msg := j.path
		index := j.index
		total := j.total

		file, err := os.Open(msg)
		if err != nil {
			t.err = err
			results <- t
			continue
		}

		// Parse message body with enmime.
		env, err := enmime.ReadEnvelope(file)
		if err != nil {
			t.err = err
			results <- t
			continue
		}

		//
		// Ensure we don't leak.
		//
		file.Close()

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
				ret = fmt.Sprintf("%d", total)
			default:
				ret = env.GetHeader(field)
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

		ret := SingleMessage{Path: j.path,
			Rendered: os.Expand(p.format, headerMapper)}

		// do some work
		results <- ret
	}
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

	//
	// Build the list of message filenames here.
	//
	var files []string

	//
	// Directories we examine beneath the maildir
	//
	dirs := []string{"cur", "new"}

	//
	// For each subdirectory
	//
	for _, dir := range dirs {

		// Build up the complete-path
		prefix := filepath.Join(path, dir)

		// Now record all files
		_ = filepath.Walk(prefix, func(path string, f os.FileInfo, err error) error {
			switch mode := f.Mode(); {
			case mode.IsRegular():
				files = append(files, path)
			}
			return nil
		})
	}

	//
	// We now have a bunch of filenames we need to read/parse
	// do it in parallel.
	//
	// ugh
	//
	jobs := make(chan job, len(files))
	results := make(chan SingleMessage, len(files))

	//
	// spin up workers and use a sync.WaitGroup to indicate completion
	//
	// We'll assume two jobs for each CPU.  Yeah.
	//
	var wg sync.WaitGroup
	wg.Add(2 * runtime.NumCPU())
	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			p.worker(jobs, results)
		}()
	}

	//
	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	//
	go func() {
		defer close(results)
		wg.Wait()
	}()

	//
	// Send each of our jobs.
	//
	go func() {
		defer close(jobs)
		for i, e := range files {
			j := job{index: i, total: len(files), path: e}
			jobs <- j
		}
	}()

	// read all the results
	c := 0
	for r := range results {
		fmt.Printf("%s\n", r.Rendered)
		c++
		messages = append(messages, r)
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
