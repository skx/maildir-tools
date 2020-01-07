package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/subcommands"
)

//
// This is a trivial toy that shows maildir-lists
//
func showUI() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	selected := -1
	l := widgets.NewList()
	l.Title = "Maildirs"
	l.Rows = []string{}

	// get the lines
	out, err := exec.Command("./maildir-utils", "maildirs", "-format=${name}").Output()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		ui.Close()
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	l.Rows = []string{}
	for _, r := range strings.Split(string(out), "\n") {
		l.Rows = append(l.Rows, r)
	}

	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, 80, 25)

	ui.Render(l)

	previousKey := ""
	uiEvents := ui.PollEvents()
	run := true

	for run {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			run = false
		case "j", "<Down>":
			l.ScrollDown()
		case "k", "<Up>":
			l.ScrollUp()
		case "J", "<C-d>":
			l.ScrollHalfPageDown()
		case "K", "<C-u>":
			l.ScrollHalfPageUp()
		case "<PageDown>", "<C-f>":
			l.ScrollPageDown()
		case "<PageUp>", "<C-b>":
			l.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				l.ScrollTop()
			}
		case "<Home>":
			l.ScrollTop()
		case "<Enter>", "<C-j>":
			selected = l.SelectedRow
			run = false
		case "G", "<End>":
			l.ScrollBottom()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(l)
	}

	ui.Close()
	if selected != -1 {
		fmt.Printf("Selected row %d: %s\n", selected, l.Rows[selected])
	}
}

type uiCmd struct {
	verbose bool
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
}

//
// Entry-point.
//
func (p *uiCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	showUI()
	return subcommands.ExitSuccess
}
