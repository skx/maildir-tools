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

	// List for maildir-entries
	maildirList *widgets.List

	// The actual maildirs we have
	maildirs []Maildir

	// The currently selected maildir
	curMaildir string

	// List for message-entries
	messageList *widgets.List

	// The actual messages in the directory
	messages []SingleMessage

	// List for a single message
	emailList *widgets.List

	// Prefix for maildirs
	prefix string
}

// getMaildirs returns ALL maildirs
func (p *uiCmd) getMaildirs() {

	helper := &maildirsCmd{prefix: p.prefix,
		format: "${06unread}/${06total} - ${name}"}
	p.maildirs = helper.GetMaildirs()

}

// getMessages gets the messages in the currently selected maildir
func (p *uiCmd) getMessages() {

	// The messages are empty now
	p.messages = []SingleMessage{}

	// No directory?  That's a bug really
	if p.curMaildir == "" {
		return
	}

	// Get the messages via our helper.
	//
	// TODO/Gross/Hack/FIXME
	//
	helper := &messagesCmd{}
	var err error
	p.messages, err = helper.GetMessages(p.curMaildir, "[${06index}/${06total} [${4flags}] ${subject}")
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

	// Default mode
	p.mode = "maildir"

	// get the initial lines for the maildir view
	p.getMaildirs()
	for _, r := range p.maildirs {
		p.maildirList.Rows = append(p.maildirList.Rows, r.Rendered)
	}

	// Set colours for maildir list
	p.maildirList.TextStyle = ui.NewStyle(ui.ColorYellow)
	p.maildirList.WrapText = false
	p.maildirList.SetRect(0, 0, width, height)

	// Set the colours for the message list
	p.messageList.TextStyle = ui.NewStyle(ui.ColorYellow)
	p.messageList.WrapText = false
	p.messageList.SetRect(0, 0, width, height)

	// Set the colours for our message-view
	p.emailList.TextStyle = ui.NewStyle(ui.ColorWhite)
	p.emailList.WrapText = false
	p.emailList.SetRect(0, 0, width, height)

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
			if p.mode == "messages" {
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
				p.curMaildir = p.maildirs[offset].Path

				p.mode = "messages"
				widget = p.messageList

				// Update our list of messages
				p.getMessages()

				// Reset our UI
				p.messageList.Rows = []string{}

				for _, r := range p.messages {
					p.messageList.Rows = append(p.messageList.Rows, r.Rendered)
				}
				widget.Title = "messages:" + p.curMaildir
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
