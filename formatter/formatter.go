// Package formatter allows expanding variables inside template-strings.
//
// This is almost a direct port of `os.Expand`, however there is support
// for using field-sizes as well as different deliminators.
package formatter

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Helper lets us find variables to expand.
	helper = regexp.MustCompile("^(.*?)#{([^}]+)}(.*)$")

	// Length specifies the field-length.
	length = regexp.MustCompile("^([0-9]+)(.*)$")

	// Mail gets an email address from a field
	emailRE = regexp.MustCompile(".*?(<.*>)$")

	// Mail gets an email address from a field
	nameRE = regexp.MustCompile("^\"(.*)\".*<.*>$")
)

// Expand replaces ${var} or $var in the string based on the mapping function.
func Expand(format string, mapping func(string) string) string {

	// For email we allow "to.name" or "#{to.email}" to
	// return just the part of the matching field.
	//
	// That goes for Cc too, and all other fields.
	name := false
	email := false

	out := ""

	match := helper.FindStringSubmatch(format)

	for len(match) > 0 {

		// Get the field-name we should interpolate.
		field := match[2]

		// Look for a padding/truncation setup.
		padding := ""
		pMatches := length.FindStringSubmatch(field)
		if len(pMatches) > 0 {
			padding = pMatches[1]
			field = pMatches[2]
		}

		// Email-specific modifiers?
		if strings.HasSuffix(field, ".name") {
			name = true
			field = strings.TrimSuffix(field, ".name")
		}
		if strings.HasSuffix(field, ".email") {
			email = true
			field = strings.TrimSuffix(field, ".email")
		}

		// Add the prefix
		out += match[1]

		// Get the field-value, via the callback
		output := mapping(field)

		if email {
			eMatches := emailRE.FindStringSubmatch(output)
			if len(eMatches) == 2 {
				output = eMatches[1]
			}
		}
		if name {
			nMatches := nameRE.FindStringSubmatch(output)
			if len(nMatches) == 2 {
				output = nMatches[1]
			}
		}

		if padding != "" {

			// padding character
			char := " "
			if padding[0] == byte('0') {
				char = "0"
			}

			// size we need to pad to
			size, _ := strconv.Atoi(padding)
			for len(output) < size {
				output = char + output
			}

			// or truncate to
			if len(output) > size {
				output = output[:size]
			}
		}

		// Add the value
		out += output

		// Move on.
		format = match[3]
		match = helper.FindStringSubmatch(format)
	}

	// Add on the format-string if it didn't match,
	// or any trailing suffix if it did.
	return out + format
}
