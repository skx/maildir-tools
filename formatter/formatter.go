// Package formatter allows expanding variables inside template-strings.
//
// This is almost a direct port of `os.Expand`, however there is support
// for using field-sizes as well as different deliminators.
package formatter

import (
	"regexp"
	"strconv"
)

var (
	// Helper lets us find variables to expand.
	helper = regexp.MustCompile("^(.*?)#{([^}]+)}(.*)$")

	// Length specifies the field-length.
	length = regexp.MustCompile("^([0-9]+)(.*)$")
)

// Expand replaces ${var} or $var in the string based on the mapping function.
func Expand(format string, mapping func(string) string) string {

	out := ""

	match := helper.FindStringSubmatch(format)

	for len(match) > 0 {

		// Get the field-name we should interpolate.
		field := match[2]

		//
		// Look for a padding/truncation setup.
		//
		padding := ""
		pMatches := length.FindStringSubmatch(field)
		if len(pMatches) > 0 {
			padding = pMatches[1]
			field = pMatches[2]
		}

		// Add the prefix
		out += match[1]

		// Get the field-value, via the callback
		output := mapping(field)

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

	// Add on the format-string if it didn't match.
	return out + format
}
