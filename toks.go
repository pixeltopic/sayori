package sayori

import "strings"

// Toks are the tokens parsed from the command
type Toks struct {
	Toks []string
	Raw  string
}

// Len returns the amount of tokens found in the command
func (t Toks) Len() int {
	return len(t.Toks)
}

// Get retrieves the token matching the index
func (t Toks) Get(i int) (string, bool) {
	if t.Toks == nil {
		return "", false
	}
	l := t.Len()
	if i >= l || i < 0 {
		return "", false
	}
	return t.Toks[i], true
}

// newToks returns a slice of tokens split by whitespace
func newToks(s string) Toks {
	return Toks{
		Toks: strings.Fields(s),
		Raw:  s,
	}
}
