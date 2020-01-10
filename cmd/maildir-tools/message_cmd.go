// Show a single message.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"text/template"

	"github.com/google/subcommands"
	"github.com/skx/maildir-tools/mailreader"
)

type messageCmd struct {
}

//
// Glue
//
func (*messageCmd) Name() string     { return "message" }
func (*messageCmd) Synopsis() string { return "Show a message." }
func (*messageCmd) Usage() string {
	return `message :
  Show a single formatted message.
`
}

//
// Flag setup
//
func (p *messageCmd) SetFlags(f *flag.FlagSet) {
}

// Show the specified email
func (p *messageCmd) GetMessage(path string) (string, error) {

	//
	// We'll format the message with a template
	//
	tmpl := `To: {{.To}}
From: {{.From}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Body}}`

	// This is the structure we'll use to populate that
	// template with.
	type Message struct {
		To      string
		From    string
		Subject string
		Date    string
		Body    string
	}

	//
	// Parse the message - note that this is gross.
	//
	helper, err := mailreader.NewEnmime(path)
	if err != nil {
		return "", err
	}

	//
	// Populate the instance
	//
	var data Message
	data.Subject = helper.Header("Subject")
	data.To = helper.Header("To")
	data.From = helper.Header("From")
	data.Date = helper.Header("Date")
	data.Body = helper.Body()

	// Render.
	var out bytes.Buffer
	t := template.Must(template.New("view.tmpl").Parse(tmpl))
	err = t.Execute(&out, data)

	return out.String(), err
}

// Entry-point.
func (p *messageCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	for _, path := range f.Args() {
		out, err := p.GetMessage(path)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		} else {
			fmt.Println(out)
		}
	}

	return subcommands.ExitSuccess
}
