// Show a single message.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/google/subcommands"
	"github.com/jhillyerd/enmime"
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
	// Open it
	//
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s - %s", path, err.Error())
	}

	// Parse message body with enmime.
	env, err := enmime.ReadEnvelope(file)
	if err != nil {
		return "", err
	}

	//
	// Ensure we don't leak.
	//
	file.Close()

	//
	// Now format the message
	//
	tmpl := `To: {{.To}}
From: {{.From}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Body}}
`

	type Message struct {
		To          string
		From        string
		Subject     string
		Date        string
		Body        string
		Attachments []string
	}
	var data Message
	data.Subject = env.GetHeader("Subject")
	data.To = env.GetHeader("To")
	data.From = env.GetHeader("From")
	data.Date = env.GetHeader("Date")

	// Body should be text, might be HTML
	if len(env.Text) > 0 {
		data.Body = env.Text
	} else if len(env.HTML) > 0 {
		data.Body = env.HTML
	} else {
		data.Body = "No body"
	}

	for _, a := range env.Attachments {
		data.Attachments = append(data.Attachments, a.FileName)
	}

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
