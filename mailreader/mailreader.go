// Package mailreader allows reading a single email from disk.
package mailreader

import (
	"bytes"
	"io/ioutil"
	"net/mail"
)

// Email holds the state for this message-object.
type Email struct {

	// Filename holds the name of the file we're reading
	Filename string

	// Message holds the mail message
	Message *mail.Message
}

// New creates a new mail-reading object
func New(file string) *Email {
	return &Email{Filename: file, Message: nil}
}

func (m *Email) readMessage() error {
	var content []byte
	var err error

	content, err = ioutil.ReadFile(m.Filename)
	if err != nil {
		return err
	}

	m.Message, err = mail.ReadMessage(bytes.NewReader(content))
	if err != nil {
		return err
	}
	return nil
}

// Header returns the value of the given message
func (m *Email) Header(name string) (string, error) {
	if m.Message == nil {
		err := m.readMessage()
		if err != nil {
			return "", err
		}
	}
	return m.Message.Header.Get(name), nil
}
