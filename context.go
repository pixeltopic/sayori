package sayori

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// defaultFmtFilterErr is the default format function to convert failing Filter(s) into an error string
func defaultFmtFilterErr(f Filter) string {
	return fmt.Sprintf("filter fail code '%d'", f)
}

// Context contains data relating to the command invocation context
type Context struct {
	Filter
	Session      *discordgo.Session
	Message      *discordgo.Message
	Prefix       string
	Alias        string
	Args         Args
	Toks         Toks
	Err          error
	FmtFilterErr func(Filter) string // format a Filter into an error string
}

// filterToErr converts an error string to a RuleError
func (c Context) filterToErr(f Filter) error {
	return &FilterError{
		f:      f,
		reason: c.FmtFilterErr(f),
	}
}

// NewContext returns an unpopulated context with defaults set
func NewContext() Context {
	return Context{
		Filter:       Filter(0),
		Session:      nil,
		Message:      nil,
		Prefix:       "",
		Alias:        "",
		Args:         nil,
		Toks:         Toks{},
		Err:          nil,
		FmtFilterErr: defaultFmtFilterErr,
	}
}
