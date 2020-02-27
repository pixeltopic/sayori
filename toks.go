package sayori

import (
	"strings"
)

// Toks are the tokens parsed from the command
type Toks struct {
	toks []string
	raw  string
}

// Raw returns the raw, untokenized string
func (t Toks) Raw() string {
	return t.raw
}

// Len returns the amount of tokens found in the command
func (t Toks) Len() int {
	return len(t.toks)
}

// Iter returns an iterable Toks.
func (t Toks) Iter() []string {
	cp := make([]string, len(t.toks))
	copy(cp, t.toks)
	return cp
}

// Set updates an existing token with a new value for all values of the Tok instance and returns an ok state.
//
// This is not concurrency-safe.
func (t Toks) Set(i int, tok string) bool {
	if t.toks == nil {
		return false
	}
	l := t.Len()
	if i >= l || i < 0 {
		return false
	}
	t.toks[i] = tok
	return true
}

// Get retrieves the token matching the index
func (t Toks) Get(i int) (string, bool) {
	if t.toks == nil {
		return "", false
	}
	l := t.Len()
	if i >= l || i < 0 {
		return "", false
	}
	return t.toks[i], true
}

// NewToks returns a slice of tokens split by whitespace
func NewToks(s string) Toks {
	return Toks{
		toks: strings.Fields(s),
		raw:  s,
	}
}
