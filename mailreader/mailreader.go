// Package mailreader allows reading a single email from disk to
// return the body and header-values by name.
//
// If we want to be fast we use the golang mail.* package, which
// is great for retrieving (& decoding) header-values.  However it
// doesn't provide easy access to the message body for complex
// cases of multipart/alternative and similar MIME messages.
//
// To handle the case of accessing the body of an email in a
// reliable fashion we embed Enmime and use that if we need the
// body.   Because opening and parsing a message two times is
// grossly inefficient the caller needs to use the appropriate
// constructor and know what they want to acces.
//
// Header-only access?  Fast?  Use New().
//
// Need the body?  Use NewEnmime.
package mailreader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/mail"
	"os"

	"github.com/jhillyerd/enmime"
)

// Email holds the state for this message-object.
type Email struct {

	// Filename holds the name of the file we're reading.
	Filename string

	// Message holds the mail message - if we're using
	// the golang parser (which we do for message-indexes
	// as it is faster).
	Message *mail.Message

	// Enmime holds the enmime handle to the message,
	// which we use when we want access to the body
	// attachments (TODO), etc.
	Enmime *enmime.Envelope

	// Use enmime?
	_enmime bool
}

// New creates a new mail-reading object which will use the
// golang mail-package to parse the message.
//
// Using this function will allow you access to the header-values
// easily, but not the message-body.
func New(file string) (*Email, error) {
	x := &Email{Filename: file}

	var content []byte
	var err error

	content, err = ioutil.ReadFile(file)
	if err != nil {
		return x, err
	}

	x.Message, err = mail.ReadMessage(bytes.NewReader(content))
	if err != nil {
		return x, err
	}

	return x, nil
}

// NewEnmime creates a new mail-reading object which uses the enmime
// library.
//
// Using this method is required if you wish to read the body of an
// email in a form suitable for rendering.  It is slower than the
// golang-native approach which is why the user must opt-into it.
func NewEnmime(file string) (*Email, error) {
	x := &Email{Filename: file, _enmime: true}

	var err error
	var f *os.File

	f, err = os.Open(file)
	if err != nil {
		return x, fmt.Errorf("failed to open %s - %s", file, err.Error())
	}

	// Parse message body with enmime.
	x.Enmime, err = enmime.ReadEnvelope(f)
	if err != nil {
		return x, err
	}

	// Ensure we don't leak.
	f.Close()

	return x, nil
}

// Header returns the value of the given header from within our message.
//
// Header values are RFC2047-decoded.
func (m *Email) Header(name string) string {

	// Split handling.
	if m._enmime {
		return m.Enmime.GetHeader(name)
	}

	// Get the header using the native-method.
	value := m.Message.Header.Get(name)

	// GO 1.5 does not decode headers, but this may change in
	// future releases...
	decoded, err := (&mime.WordDecoder{}).DecodeHeader(value)
	if err != nil || len(decoded) == 0 {
		return value
	}
	return decoded
}

// Body returns the body of an email message, in a useful format.
// That means that if a 'text/plain' part is present it will be
// returned, otherwise we'll use 'text/html'.  If neither part
// is present then the raw body will be returned.
func (m *Email) Body() string {

	if m._enmime {
		// Try to work out what to return
		if len(m.Enmime.Text) > 0 {
			return m.Enmime.Text
		} else if len(m.Enmime.HTML) > 0 {
			return m.Enmime.HTML
		}

	}

	// TODO - open the file.  Read until we hit
	// a newline, then return the contents.
	return "No body available.  Sorry!"
}
