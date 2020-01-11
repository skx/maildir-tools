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

func TestName(t *testing.T) {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "NAME1":
			return "\"Steve\"<steve@steve.fi>"
		case "NAME2":
			return "\"Steve Kemp\"    <steve@steve.fi>"
		}
		return ""
	}

	out := Expand("#{NAME1.name}", mapper)
	if out != "Steve" {
		t.Errorf("Got unexpected output:" + out)
	}
	out = Expand("#{NAME2.name}", mapper)
	if out != "Steve Kemp" {
		t.Errorf("Got unexpected output:" + out)
	}
}

func TestEmail(t *testing.T) {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "NAME":
			return "\"Steve\"<steve@steve.fi>"
		case "TEST":
			return "<steve@steve.fi>"
		case "TEST2":
			return "steve@steve.fi"
		}
		return ""
	}

	out := Expand("#{NAME.email}", mapper)
	if out != "<steve@steve.fi>" {
		t.Errorf("Got unexpected output:" + out)
	}
	out = Expand("#{TEST.email}", mapper)
	if out != "<steve@steve.fi>" {
		t.Errorf("Got unexpected output:" + out)
	}
	out = Expand("#{TEST2.email}", mapper)
	if out != "steve@steve.fi" {
		t.Errorf("Got unexpected output:" + out)
	}
}
