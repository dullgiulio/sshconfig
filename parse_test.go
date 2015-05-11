package sshconfig_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/dullgiulio/sshconfig"
)

func reader(s string) io.Reader {
	return bytes.NewReader([]byte(s))
}

func TestHostSection(t *testing.T) {
	hostSection := `# Some example Host section
Host local # Local SSH
	OptionOne do anything here
	OptionTwo 12345 # There is also a comment
	OptionWithoutValue
	OptionsWithComment # Comment
`
	sections, err := sshconfig.Parse(reader(hostSection))
	if err != nil {
		t.Fatal(err)
	}

	if len(sections) != 1 {
		t.Error("Unexpected number of sections")
	}

	s0 := sections[0]
	if s0.Name != "local" {
		t.Error("Unexpected section name")
	}
	if s0.Values["OptionOne"] != "do anything here" {
		t.Error("Unexpected option value (composite string)")
	}
	if s0.Values["OptionTwo"] != "12345" {
		t.Error("Unexpected option value (numeric)")
	}
	if v, ok := s0.Values["OptionWithoutValue"]; !ok || v != "" {
		t.Error("Unexpected value for option without value, or not parsed")
	}
	if v, ok := s0.Values["OptionsWithComment"]; !ok || v != "" {
		t.Error("Unexpected value for option without value, or not parsed")
	}
}
