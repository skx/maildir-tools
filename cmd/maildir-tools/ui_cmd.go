//
// The `ui` sub-command presents a simple GUI using the primitives we've
// defined elsewhere.
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

// UIHistory stores UI history.
//
// When we change mode we record the current/previous mode, and selection
// offset, so that we can go back to that when we need to.
//
// The way this structure is used is suboptimal, because we store the
// index in the record we remove - rather than the index we restore to.
// See `PreviousMode` for details.
type UIHistory struct {

	// mode holds the mode we changed TO.
	mode string

	// offset holds the offset of the selected item in the list-view,
	// BEFORE we changed mode.
	offset int
}

// uiCmd holds the state of our TUI application.
type uiCmd struct {

	// app holds the pointer to our application.
	app *tview.Application

	// Keep track of our mode-transitions.
	//
	// Here we're treating our client as a modal-application,
	// where we're always in one of three main states ("maildir-list",
	// "message-list", or "message-display").  We keep track of our
	// previous mode whenever we change - which allows us to jump
	// back quickly, easily, and efficiently.
	//
	// Mostly this is overkill because we always have the same
	// entry and exit-states.  But for things like config-mode
	// (unimplemented) or help-display we will need to support a
	// more random mode-transition.
	//
	// TODO: Reconsider this, perhaps.  cc:lumail
	modeHistory []UIHistory

	// The actual maildirs we have found, and which we want to display.
	maildirs []Maildir

	// List for displaying maildir-entries.
	maildirList *tview.List

	// Currently selected maildir.
	//
	// TODO: Do we need this?  `maildirList` is global so we can
	// read from that on-demand.   cc:lumail
	//
	curMaildir string

	// List for message-entries.
	messageList *tview.List

	// The actual messages in the currently-selected maildir,
	// which are being displayed.
	messages []SingleMessage

	// Path to current message
	//
	// TODO: Do we need this?  `messageList` is global so we can
	// read from that on-demand.   cc:lumail
	//
	curMessage string

	// List for displaying a single message.
	emailList *tview.List

	// List for displaying help
	helpList *tview.List

	// Prefix for our maildir hierarchy
	prefix string

	// Searching - these shouldn't be global.
	// TODO: Remove these.
	searchText string
	inputField *tview.InputField
}

// getMaildirs returns ALL maildirs beneath our configured prefix-directory.
func (p *uiCmd) getMaildirs() {
	helper := &maildirsCmd{prefix: p.prefix, format: "[#{06unread}/#{06total}] #{name}"}
	p.maildirs = helper.GetMaildirs()
}

// getMessages gets all the messages in the currently selected maildir.
func (p *uiCmd) getMessages() {

	var err error

	// If we have no current-maildir AND the list is non-empty
	// then we use the first.  The change-handler doesn't run
	// for the first item highlighted by the tview UI
	if p.curMaildir == "" {
		if len(p.maildirs) > 0 {
			p.curMaildir = p.maildirs[0].Path
		}
	}

	// The messages are empty now
	p.messages = []SingleMessage{}

	// Get the messages via our helper.
	helper := &messagesCmd{}
	p.messages, err = helper.GetMessages(p.curMaildir, "[#{06index}/#{06total} [#{4flags}] #{subject}")

	// Failed to get messages?
	if err != nil {
		// TODO: Dialog
		panic(err)
	}
}

// getMessage returns the content of a single email.
func (p *uiCmd) getMessage() []string {

	// The file on-disk
	file := p.curMessage

	// If we have no current-message AND the list is non-empty
	// then we use the first.  The change-handler doesn't run
	// for the first item highlighted by the tview UI
	if file == "" {
		if len(p.messages) > 0 {
			file = p.messages[0].Path
		}
	}

	// Get the output
	helper := &messageCmd{}
	out, err := helper.GetMessage(file)
	if err != nil {
		// TODO: Dialog
		return ([]string{"Failed to read " + file + " " + err.Error()})
	}

	return strings.Split(out, "\n")
}

// SetMode updates our global state to be one of:
//
//    maildirs | View a list of maildirs.
//    messages | View a list of messages.
//    message  | View a single message.
//    help     | Show our help.
//
// TODO:
//    config   |
//    compose  |
func (p *uiCmd) SetMode(mode string, record bool) {

	// If we're supposed to record our state-transition then
	// do so here.
	//
	// Basically we always record our mode-change, unless we're
	// reverting to a previous state (via `q`).  In that case
	// we reuse this method, but explicitly don't want to record
	// the state from which we returned - or we'd get a loop!
	//
	if record {

		var x UIHistory
		x.mode = mode
		x.offset = -1

		// If we're in a view then get the current list-offset
		focus := p.app.GetFocus()
		l, ok := focus.(*tview.List)
		if ok {
			x.offset = l.GetCurrentItem()
		}

		// Record the entry
		p.modeHistory = append(p.modeHistory, x)
	}

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
					p.SetMode("messages", true)

					// Change the view
					p.app.SetRoot(p.messageList, true)
				}).
				SetChangedFunc(func(index int, rendered string, path string, shorcut rune) {
					p.curMaildir = path
				})
		}

		// Update UI
		p.app.SetRoot(p.maildirList, true)
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
					p.SetMode("message", true)

					// Change the view
					p.app.SetRoot(p.messageList, true)
				}).
				SetChangedFunc(func(index int, rendered string, path string, shorcut rune) {
					p.curMessage = path
				})
		}

		// Update UI
		p.app.SetRoot(p.messageList, true)
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

		// Update UI
		p.app.SetRoot(p.emailList, true)
		return
	}

	if mode == "help" {
		p.helpList.Clear()
		p.helpList.AddItem("Help goes here", "", 0, nil)

		// Update UI
		p.app.SetRoot(p.helpList, true)
		return
	}

	// Can't happen?
	panic("unknown mode " + mode)
}

// Search is a function which will operate upon any `List`-based view.
//
// It will prompt for text, and select the next entry which matches that
// text.  The text is matched literally (albeit case-insensitively), rather
// than as a regular expression.
//
// TODO: Support regexp?  We'd have to implement the logic ourselves, but
// it wouldn't be hard.  Right now we use the List-specific helper from the
// UI-toolkit.
func (p *uiCmd) Search() {

	// Get the old UI element which had focus
	old := p.app.GetFocus()

	// Prompt for input
	p.inputField = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(40).
		SetDoneFunc(func(key tcell.Key) {

			// Highlight the old value
			p.app.SetRoot(old, true)

			// If this was a completed enter then do the search
			if key == tcell.KeyEnter {

				// Get the text - keep the previous
				// value if this is empty
				val := p.inputField.GetText()
				if len(val) > 0 {
					p.searchText = val
				}
				if len(p.searchText) < 1 {
					return
				}

				// Can we cast our (previous) UI-item
				// into a list?  If so do it
				l, ok := old.(*tview.List)
				if ok {

					//
					// Search.
					//
					inx := l.FindItems(p.searchText, p.searchText, false, true)

					//
					// We now want to find the "next"
					// match, handling wrap-arround.
					//
					cur := l.GetCurrentItem()
					max := l.GetItemCount()

					// Always search forward from
					// the next line.
					cur++
					if cur > max {
						cur = 0
					}

					tested := 0
					for tested < max {

						offset := cur + tested
						if offset >= max {
							offset -= max
						}

						// Grossly inefficient..
						for _, j := range inx {
							if j == offset {
								l.SetCurrentItem(j)
								return
							}
						}

						tested++
					}
				}
			}
		})

	// update ui
	p.app.SetRoot(p.inputField, true)

}

// Return to the previous mode, if possible, using our history-stack.
//
// This is a bit horrid.
func (p *uiCmd) PreviousMode() {

	// Default
	prev := "maildir"
	offset := -1

	// If we have history - remove the last entry.
	if len(p.modeHistory) > 0 {
		// the offset is on the last entry
		offset = p.modeHistory[len(p.modeHistory)-1].offset
		p.modeHistory = p.modeHistory[:len(p.modeHistory)-1]
	}

	// Now set the history to the previous one
	if len(p.modeHistory) > 0 {
		prev = p.modeHistory[len(p.modeHistory)-1].mode
	}

	// Set the mode now
	p.app.QueueUpdateDraw(func() {
		p.SetMode(prev, false)

		// If the current value is a list
		if offset != -1 {
			old := p.app.GetFocus()
			l, ok := old.(*tview.List)
			if ok {
				// set the old offset
				l.SetCurrentItem(offset)
			}
		}

	})
}

// deleteSelectedMessage deletes the message under the point, in
// the list of messages.
func (p *uiCmd) deleteSelectedMessage() {

	// Get the current entry.
	selected := p.messageList.GetCurrentItem()

	// Delete the file
	path := p.messages[selected].Path
	os.Remove(path)

	// Reload messages - don't save history
	p.SetMode("messages", false)

	// Reset the selection
	if selected >= len(p.messages) {
		selected--
	}
	if selected < 0 {
		selected = 0
	}

	// If it is out-of-bounds, decrement
	p.messageList.SetCurrentItem(selected)
}

// TUI sets up our user-interface, and handles the execution of the
// main-loop.
//
// All our code is built around a set of list-views, although we only
// display one at a time.
func (p *uiCmd) TUI() {

	// Create the applicaiton
	p.app = tview.NewApplication()

	// Listbox to hold our maildir list.
	p.maildirList = tview.NewList()
	p.maildirList.ShowSecondaryText(false)
	p.maildirList.SetWrapAround(true)
	p.maildirList.SetHighlightFullLine(true)

	// Listbox to hold our message list.
	p.messageList = tview.NewList()
	p.messageList.ShowSecondaryText(false)
	p.messageList.SetWrapAround(true)
	p.messageList.SetHighlightFullLine(true)

	// specific binding for message-list
	p.messageList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == rune('d') {
			p.deleteSelectedMessage()
			return nil
		}
		return event
	})

	// Listbox to hold the contents of a single email.
	p.emailList = tview.NewList()
	p.emailList.ShowSecondaryText(false)
	p.emailList.SetWrapAround(true)
	p.emailList.SetHighlightFullLine(false)

	// Listbox to hold the help-text.
	p.helpList = tview.NewList()
	p.helpList.ShowSecondaryText(false)
	p.helpList.SetWrapAround(true)
	p.helpList.SetHighlightFullLine(true)

	// Setup some global keybindings.
	p.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		//
		// We setup global keybindings, but we actually don't
		// want these to be truely global.
		//
		// If they were then we get a horrid situation where the
		// user might press "/" to do a text-search, but then
		// sees failure when they try to search for "jenny".
		//
		// The leading "j" would get converted to a down-arrow,
		// which would then close the search-box.  Yes it took
		// me a while to appreciate this.
		//
		// Ignore ALL custom keybindings unless we're showing
		// a list-view.  That's a bit gross, but also works.
		//
		// FIXME/Hack?
		//
		focus := p.app.GetFocus()
		_, ok := focus.(*tview.List)
		if !ok {
			return event
		}

		// vi-emulation
		if event.Rune() == rune('j') {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		if event.Rune() == rune('k') {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		// steve-preferences
		if event.Rune() == rune('<') {
			return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
		}
		if event.Rune() == rune('>') ||
			event.Rune() == rune('*') {
			return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
		}

		// Q: Quit
		if event.Rune() == rune('Q') {
			p.app.Stop()
			return nil
		}

		// ?: Show help
		if event.Rune() == rune('?') {
			p.SetMode("help", true)
			return nil
		}

		// /: search for (literal) text
		if event.Rune() == rune('/') {
			p.Search()
			return nil
		}
		// q: Exit mode, and return to previous
		if event.Rune() == rune('q') {
			p.PreviousMode()
			return nil
		}
		return event
	})

	// Setup the default mode - we queue this to avoid issues
	p.app.QueueUpdateDraw(func() {
		p.SetMode("maildir", true)
	})

	// Run the mail UI event-loop.
	//
	// This runs until something calls `app.Stop()` or a
	// panic is received.
	if err := p.app.SetRoot(p.maildirList, true).Run(); err != nil {
		panic(err)
	}

}

//
// Glue
//
func (*uiCmd) Name() string     { return "tui" }
func (*uiCmd) Synopsis() string { return "Show our text-based user-interface." }
func (*uiCmd) Usage() string {
	return `tui :
  Show our text-based user-interface.
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

	// Run the TUI
	p.TUI()
	return subcommands.ExitSuccess
}
