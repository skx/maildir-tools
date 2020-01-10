package formatter

import (
	"testing"
)

func TestBasic(t *testing.T) {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "DAY_PART":
			return "morning"
		case "NAME":
			return "Gopher"

		}
		return ""
	}

	out := Expand("Good #{DAY_PART}, #{NAME}!", mapper)

	if out != "Good morning, Gopher!" {
		t.Errorf("Got unexpected output:" + out)
	}
}

func TestPadding(t *testing.T) {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "FLAGS":
			return "SRF"
		}
		return ""
	}

	out := Expand("#{06FLAGS}", mapper)
	if out != "000SRF" {
		t.Errorf("Got unexpected output:" + out)
	}
}

func TestTruncation(t *testing.T) {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "NAME":
			return "STEVE KEMP"
		}
		return ""
	}

	out := Expand("#{04NAME}", mapper)
	if out != "STEV" {
		t.Errorf("Got unexpected output:" + out)
	}
}
