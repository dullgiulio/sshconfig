package sshconfig

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const spaces = " \t"

// Section represents a Host configuration section, for example:
//
//     Host test
//         ConfigOption yes
//         OtherOption no
//
// Where the Host part is the Name, the options the Values in map.
type Section struct {
	Name   string
	Values map[string]string
}

// Creates a new empty Section
func NewSection(name string) *Section {
	return &Section{Name: name, Values: make(map[string]string)}
}

// Internal token representation with line number
type token struct {
	line int
	val  string
}

// Tokenize a single line, skipping spaces and comments
func parseLine(nline int, line string, tokens chan<- token) {
	var end int

	line = strings.TrimLeft(line, spaces)
	if len(line) < 1 || line[0] == '#' {
		return
	}

	for end = 0; end < len(line); {
		r, s := utf8.DecodeRuneInString(line[end:])
		if strings.ContainsRune(spaces, r) {
			break
		}
		end = end + s
	}
	if end > 0 {
		tokens <- token{nline, line[0:end]}
	} else {
		return
	}

	line = strings.Trim(line[end:], spaces)
	for end = 0; end < len(line); {
		r, s := utf8.DecodeRuneInString(line[end:])
		if r == '#' {
			break
		}
		end = end + s
	}

	if end > 0 {
		tokens <- token{nline, strings.Trim(line[0:end], spaces)}
	} else {
		tokens <- token{nline, ""}
	}
}

// Tokenize line by line the entire reader
func parseFile(f io.Reader, tokens chan<- token) {
	var nline int

	scanner := bufio.NewScanner(f)
	defer close(tokens)

	for scanner.Scan() {
		nline++
		parseLine(nline, scanner.Text(), tokens)
	}
}

// Consume tokens from the channel until the next
// Host section
func (s *Section) loadMap(tokens <-chan token) {
	var hasKey bool
	var key string

	for token := range tokens {
		if token.val == "Host" {
			return
		}
		if hasKey {
			s.Values[key] = token.val
			hasKey = false
		} else {
			key = token.val
			hasKey = true
		}
	}
}

// Parse parses a SSH config file and returns either an error or the
// slice of parsed Sections.
func Parse(r io.Reader) ([]*Section, error) {
	tokens := make(chan token)
	go parseFile(r, tokens)

	sections := make([]*Section, 0)

	token := <-tokens
	if token.val != "Host" {
		for range tokens {
		}
		return nil, fmt.Errorf("line %s: expected Host", token.line)
	}

	for token := range tokens {
		sec := NewSection(token.val)
		sec.loadMap(tokens)
		sections = append(sections, sec)
	}

	return sections, nil
}
