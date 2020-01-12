//
// This command implements our (trivial) user-interface.
//
// It allows displaying a list of Maildir folders,
// a list of messages, and finally a single message.
//
// It is obviously very rough and ready, but
// despite that it seems to function reasonably
// well, albeit slowly due to lack of caching.
//
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/subcommands"
)

// uiCmd holds our state.
type uiCmd struct {

	// Current mode.
	mode string

	// List for maildir-entries.
	maildirList *widgets.List

	// The actual maildirs we have found, which are being displayed.
	maildirs []Maildir

	// List for message-entries.
	messageList *widgets.List

	// The actual messages in the currently-selected maildir,
	// which are being displayed.
	messages []SingleMessage

	// List for displaying a single message.
	emailList *widgets.List

	// List for displaying help
	helpList *widgets.List

	// Prefix for our maildir hierarchy
	prefix string
}

// getMaildirs returns ALL maildirs
func (p *uiCmd) getMaildirs() {

	helper := &maildirsCmd{prefix: p.prefix,
		format: "#{06unread}/#{06total} - #{name}"}
	p.maildirs = helper.GetMaildirs()

}

// getMessages gets the messages in the currently selected maildir
func (p *uiCmd) getMessages() {

	// Get the selected folder
	curMaildir := p.maildirs[p.maildirList.SelectedRow].Path

	// The messages are empty now
	p.messages = []SingleMessage{}

	// No directory?  That's a bug really
	if curMaildir == "" {
		return
	}

	// Get the messages via our helper.
	//
	// TODO/Gross/Hack/FIXME
	//
	helper := &messagesCmd{}
	var err error
	p.messages, err = helper.GetMessages(curMaildir, "[#{06index}/#{06total} [#{4flags}] #{subject}")
	if err != nil {
		ui.Close()
		panic(err)
	}
}

// getMessage returns the content of a single
// email.
func (p *uiCmd) getMessage() []string {

	// Get the message
	selectedMsg := p.messageList.SelectedRow

	// avoid empty reading
	if selectedMsg < 0 {
		return []string{"empty", "message"}
	}

	// The file on-disk
	file := p.messages[selectedMsg].Path

	// Get the output
	helper := &messageCmd{}
	out, err := helper.GetMessage(file)
	if err != nil {
		return ([]string{err.Error()})
	}

	return strings.Split(out, "\n")
}

// deleteCurrentMessage is called in either the message-list mode,
// or in the message-view mode.  It deletes the current message
// in either case.
func (p *uiCmd) deleteCurrentMessage() {

	// Delete from the message-index
	if p.mode == "messages" {

		// Get the selected message
		i := p.messageList.SelectedRow

		// If we don't have one, something is weird.
		if i < 0 {
			return
		}

		// Get the file on-disk
		file := p.messages[i].Path

		// Delete it
		os.Remove(file)

		// Refresh the message-list
		// Update our list of messages
		p.getMessages()

		// Reset our UI
		p.messageList.Rows = []string{}

		for _, r := range p.messages {
			p.messageList.Rows = append(p.messageList.Rows, r.Rendered)
		}

		// Restore the index
		p.messageList.SelectedRow = i
		if i > len(p.messageList.Rows)-1 {
			p.messageList.SelectedRow--
		}
		return
	}

	// TODO - delete when viewing a single message.
}

// showUI handles state-transitions and displays
//
// All our code is built around a set of list-views,
// although we only build one at a time.
func (p *uiCmd) showUI() {

	// setup our terminal
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
		return
	}

	// Get the dimensions so we can scale our UI
	width, height := ui.TerminalDimensions()

	// Listbox to hold our maildir entries
	p.maildirList = widgets.NewList()
	p.maildirList.Title = "Maildir Entries"
	p.maildirList.Rows = []string{}

	// Listbox to hold our message list.
	p.messageList = widgets.NewList()
	p.messageList.Title = "Messages"
	p.messageList.Rows = []string{}

	// Listbox to hold the contents of a single email
	p.emailList = widgets.NewList()
	p.emailList.Title = "Email"
	p.emailList.Rows = []string{}

	// Listbox to hold the help
	p.helpList = widgets.NewList()
	p.helpList.Title = "Help"
	p.helpList.Rows = []string{}

	p.helpList.Rows = strings.Split(
		`

This is a simple console-based email client.

This UI is primarily a demo, to prove to the author that the idea of
a simple set of primitives could be useful.

Navigation with the keyboard is the same in all modes:

Key                | Action
-------------------+--------------------------------------
q                  | Return to the previous mode
Q                  | Quit
j, Down            | Scroll down
k, Up              | Scroll up
Ctrl-d             | Scroll down half a page
Ctrl-u             | Scroll up half a page
Ctrl-f, PageDown   | Scroll down a page
Ctrl-b, PageUp     | Scroll up a page
gg, <, HOME        | Go to top of list
gG, >, END         | Go to end of list
w                  | Toggle line-wrap
Ctrl-j, Ret, Space | Select

Message-View Mode has two more keys

Key | Action
----+--------------------------
J   | Select the next message.
K   | Select the previous message.

Press 'q' to exit this help window.

Steve
--
`, "\n")

	// Default mode
	p.mode = "maildir"

	// get the initial lines for the maildir view
	p.getMaildirs()
	for _, r := range p.maildirs {
		p.maildirList.Rows = append(p.maildirList.Rows, r.Rendered)
	}

	// Process each known-widget
	tmp := []*widgets.List{p.maildirList,
		p.messageList,
		p.emailList,
		p.helpList}

	// Set the text-style, the size, and colours
	for _, entry := range tmp {
		entry.TextStyle = ui.NewStyle(ui.ColorGreen)
		entry.WrapText = false
		entry.SetRect(0, 0, width, height)
	}

	// Render the starting state
	ui.Render(p.maildirList)

	// We allow some multi-key inputs, which requires
	// keeping track of the previously typed character.
	previousKey := ""

	// Poll for events, until we should stop 'run'ning.
	uiEvents := ui.PollEvents()
	run := true

	// Since we have three modes, but each of them
	// uses the same widget-type to display their
	// output we can just keep a pointer to the one
	// visible right now - and use that.
	widget := p.maildirList

	// Loop forever
	for run {

		// Handle events - mostly this means
		// responding to keyboard events
		e := <-uiEvents

		switch e.ID {

		// "?" shows help
		case "?":
			p.mode = "help"
			widget = p.helpList

		// Q quits in all modes.
		case "Q", "<C-c>":
			run = false

			// q moves back to the previous mode
		case "q":

			// maildir -> exit
			if p.mode == "maildir" {
				run = false
			}
			// index -> maildir
			if p.mode == "messages" || p.mode == "help" {
				p.mode = "maildir"
				widget = p.maildirList

				p.getMaildirs()
				p.maildirList.Rows = []string{}
				for _, r := range p.maildirs {
					p.maildirList.Rows = append(p.maildirList.Rows, r.Rendered)
				}
			}
			// message -> index
			if p.mode == "message" {
				p.mode = "messages"
				widget = p.messageList

				// Update our list of messages
				p.getMessages()

				// Reset our UI
				p.messageList.Rows = []string{}

				for _, r := range p.messages {
					p.messageList.Rows = append(p.messageList.Rows, r.Rendered)
				}
			}
		case "d":
			if p.mode == "messages" ||
				p.mode == "message" {
				p.deleteCurrentMessage()
			}
		case "j", "<Down>":
			widget.ScrollDown()
		case "k", "<Up>":
			widget.ScrollUp()
		case "<C-d>":
			widget.ScrollHalfPageDown()
		case "<C-u>":
			widget.ScrollHalfPageUp()
		case "<PageDown>", "<C-f>":
			widget.ScrollPageDown()
		case "<PageUp>", "<C-b>":
			widget.ScrollPageUp()

			// gg -> start of list
		case "g":
			if previousKey == "g" {
				widget.ScrollTop()
				widget.ScrollDown()
			}
		case "w":
			widget.WrapText = !widget.WrapText
		case "<", "<Home>":
			widget.ScrollTop()
			widget.ScrollDown()
		case ">", "<End>":
			widget.ScrollBottom()
			widget.ScrollUp()
		case "<Enter>", "<C-j>", "<Space>":

			// index -> message
			if p.mode == "messages" {

				// Change to the message-view mode
				p.mode = "message"
				widget = p.emailList

				// Get the message body
				lines := p.getMessage()
				p.emailList.Rows = []string{}

				// Update the UI with it
				p.emailList.Rows = append(p.emailList.Rows, lines...)
				p.emailList.SelectedRow = 0
			}

			// maildir -> index
			if p.mode == "maildir" {

				// Get folder to view.
				offset := p.maildirList.SelectedRow
				folder := p.maildirs[offset].Path

				p.mode = "messages"
				widget = p.messageList

				// Update our list of messages
				p.getMessages()

				// Reset our UI
				p.messageList.Rows = []string{}

				for _, r := range p.messages {
					p.messageList.Rows = append(p.messageList.Rows, r.Rendered)
				}
				widget.Title = "messages:" + folder
				p.messageList.SelectedRow = 0
			}

			// gG to the end of the list
		case "G":
			if previousKey == "g" {
				widget.ScrollBottom()
				widget.ScrollUp()
			}
		case "J":
			if p.mode == "message" {
				offset := p.messageList.SelectedRow
				if offset < len(p.messageList.Rows)-1 {
					p.messageList.SelectedRow++
					// Get the message body
					lines := p.getMessage()
					p.emailList.Rows = []string{}

					// Update the UI with it
					p.emailList.Rows = append(p.emailList.Rows, lines...)
					p.emailList.SelectedRow = 0
				}
			}

		case "K":
			if p.mode == "message" {
				offset := p.messageList.SelectedRow
				if offset > 0 {
					p.messageList.SelectedRow--

					// Get the message body
					lines := p.getMessage()
					p.emailList.Rows = []string{}

					// Update the UI with it
					p.emailList.Rows = append(p.emailList.Rows, lines...)
					p.emailList.SelectedRow = 0

				}

			}
		}
		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(widget)
	}

	// All done
	ui.Close()
}

//
// Glue
//
func (*uiCmd) Name() string     { return "ui" }
func (*uiCmd) Synopsis() string { return "Show our user-interface." }
func (*uiCmd) Usage() string {
	return `ui :
  Show our user-interface.
  This will let you choose a maildir.
`
}

//
// Flag setup
//
func (p *uiCmd) SetFlags(f *flag.FlagSet) {
	prefix := os.Getenv("HOME") + "/Maildir/"
	f.StringVar(&p.prefix, "prefix", prefix, "The prefix directory.")
}

//
// Entry-point.
//
func (p *uiCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.showUI()
	return subcommands.ExitSuccess
}
