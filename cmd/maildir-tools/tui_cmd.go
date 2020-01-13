//
// TODO
//  1. Setup a stack of past modes.
//  2. Bind global `q` to that.
//  3. Bind `Q` to quit, globally.
//  4. When changing to a folder with only one message the onChange
//     handler is not called.  Set selection to first item manually
//  5. Add "/" to search forward.
//  6. Add vi-like keys to global state.
//      j -> 	p.app.QueueUpdateDraw(func() { sendkey( "down") } )
//     etc.
//  7. Add prompt function for listbox searching
//  8. Update helper
//  9. Commit.
//
package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/google/subcommands"
	"github.com/rivo/tview"
)

// tuiCmd holds our state.
type tuiCmd struct {

	// app holds the pointer to our application
	app *tview.Application

	// Current mode - not really used yet.
	mode string

	// The actual maildirs we have found, which we want to display
	maildirs []Maildir

	// List for displaying maildir-entries.
	maildirList *tview.List

	// Currently selected maildir
	curMaildir string

	// List for message-entries.
	messageList *tview.List

	// The actual messages in the currently-selected maildir,
	// which are being displayed.
	messages []SingleMessage

	// Path to current message
	curMessage string

	// List for displaying a single message.
	emailList *tview.List

	// List for displaying help
	helpList *tview.List

	// Prefix for our maildir hierarchy
	prefix string
}

// getMaildirs returns ALL maildirs
func (p *tuiCmd) getMaildirs() {

	helper := &maildirsCmd{prefix: p.prefix,
		format: "#{06unread}/#{06total} - #{name}"}
	p.maildirs = helper.GetMaildirs()

}

// getMessages gets the messages in the currently selected maildir
func (p *tuiCmd) getMessages() {

	// The messages are empty now
	p.messages = []SingleMessage{}

	// Get the messages via our helper.
	//
	// TODO/Gross/Hack/FIXME
	//
	helper := &messagesCmd{}
	var err error
	p.messages, err = helper.GetMessages(p.curMaildir, "[#{06index}/#{06total} [#{4flags}] #{subject}")
	if err != nil {
		panic(err)
	}
}

// getMessage returns the content of a single email.
func (p *tuiCmd) getMessage() []string {

	// The file on-disk
	file := p.curMessage

	// Get the output
	helper := &messageCmd{}
	out, err := helper.GetMessage(file)
	if err != nil {
		return ([]string{"Failed to read " + file + " " + err.Error()})
	}

	return strings.Split(out, "\n")
}

func (p *tuiCmd) SetMode(mode string) {

	// Update our current mode
	p.mode = mode

	if mode == "maildir" {

		// get the initial lines for the maildir view
		p.getMaildirs()

		// Empty the list
		p.maildirList.Clear()

		// Add each (rendered) maildir
		for _, r := range p.maildirs {
			p.maildirList.AddItem(r.Rendered, r.Path, 0,
				func() {

					// Change the mode
					p.SetMode("messages")

					// Change the view
					p.app.SetRoot(p.messageList, true).
						SetFocus(p.messageList)
				}).
				SetChangedFunc(func(index int, rendered string, path string, shorcut rune) {
					p.curMaildir = path
				})
		}

		return
	}

	if mode == "messages" {

		// get the messages we want to display
		p.getMessages()

		// Empty the list
		p.messageList.Clear()

		// Add each (rendered) item
		for _, r := range p.messages {
			p.messageList.AddItem(r.Rendered, r.Path, 0,
				func() {

					// Change the mode
					p.SetMode("message")

					// Change the view
					p.app.SetRoot(p.messageList, true).
						SetFocus(p.messageList)
				}).
				SetChangedFunc(func(index int, rendered string, path string, shorcut rune) {
					p.curMessage = path
				})
		}

		return
	}

	if mode == "message" {

		// get the message we want to display
		txt := p.getMessage()

		// Empty the list
		p.messageList.Clear()

		// Add each (rendered) item
		for _, r := range txt {
			p.messageList.AddItem(r, "", 0, nil)
		}

		return
	}

	if mode == "help" {
	}
}

// TUI sets up our user-interface, and handles the execution of the
// main-loop.
//
// All our code is built around a set of list-views, although we only
// display one at a time.
func (p *tuiCmd) TUI() {

	// Create the applicaiton
	p.app = tview.NewApplication()

	// Create & populate our maildir-list
	p.maildirList = tview.NewList()
	p.maildirList.ShowSecondaryText(false)
	p.maildirList.SetWrapAround(true)
	p.maildirList.SetHighlightFullLine(true)

	// q quits
	p.maildirList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune('q') {
			p.app.Stop()
		}
		return event
	})

	// Listbox to hold our message list.
	p.messageList = tview.NewList()
	p.messageList.ShowSecondaryText(false)
	p.messageList.SetWrapAround(true)
	p.messageList.SetHighlightFullLine(true)

	// message-list: q -> maildirs
	p.messageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune('q') {
			p.SetMode("maildir")
			return nil
		}
		return event
	})

	// Listbox to hold the contents of a single email
	p.emailList = tview.NewList()
	p.emailList.ShowSecondaryText(false)
	p.emailList.SetWrapAround(true)
	p.emailList.SetHighlightFullLine(true)

	// email-view: q -> index
	p.messageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune('q') {
			p.SetMode("messages")
			return nil
		}
		return event
	})

	// Listbox to hold the help
	p.helpList = tview.NewList()
	p.helpList.ShowSecondaryText(false)
	p.helpList.SetWrapAround(true)
	p.helpList.SetHighlightFullLine(true)

	// Default mode
	p.app.QueueUpdateDraw(func() {
		p.SetMode("maildir")
	})

	if err := p.app.SetRoot(p.maildirList, true).SetFocus(p.maildirList).Run(); err != nil {
		panic(err)
	}

}

//
// Glue
//
func (*tuiCmd) Name() string     { return "tui" }
func (*tuiCmd) Synopsis() string { return "Show our text-based user-interface." }
func (*tuiCmd) Usage() string {
	return `tui :
  Show our text-based user-interface.
`
}

//
// Flag setup
//
func (p *tuiCmd) SetFlags(f *flag.FlagSet) {
	prefix := os.Getenv("HOME") + "/Maildir/"
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
}

//
// Entry-point.
//
func (p *tuiCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.TUI()
	return subcommands.ExitSuccess
}
