// Show a single message.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/google/subcommands"
	"github.com/skx/maildir-tools/mailreader"
)

var (
	defaultTemplate = `To: {{.To}}
From: {{.From}}
Cc: {{.Cc}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Body}}`
)

// messageCmd holds our state
type messageCmd struct {

	// Template points to a template to use for rendering the mail.
	template string

	// If this flag is true we just dump our template
	dumpTemplate bool
}

//
// Glue
//
func (*messageCmd) Name() string     { return "message" }
func (*messageCmd) Synopsis() string { return "Show a message." }
func (*messageCmd) Usage() string {
	return `message :
  Show a single formatted message.  By default an internal template
 will be used, but you may specify the filename of a Golang text/template
 file to use for rendering if you wish.
`
}

//
// Flag setup
//
func (p *messageCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.template, "template", "", "Specify the path to a golang text/template file to use for message-rendering")
	f.BoolVar(&p.dumpTemplate, "dump-template", false, "Dump the default template")
}

// Show the specified email, with the appropriate template
func (p *messageCmd) GetMessage(path string) (string, error) {

	// Load the default template
	tmpl := defaultTemplate

	if p.template != "" {
		content, err := ioutil.ReadFile(p.template)
		if err != nil {
			return "", err
		}
		tmpl = string(content)
	}

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

	if p.dumpTemplate {
		fmt.Println(defaultTemplate)
		return subcommands.ExitSuccess
	}

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
