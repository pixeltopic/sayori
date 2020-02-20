package sayori

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// defaultFmtRule is the default format function to convert a failing Rule into an error string
func defaultFmtRule(r Rule) string {
	return fmt.Sprintf("rule id %d failed", r)
}

// Context contains data relating to the command invocation context
type Context struct {
	Rule
	Session    *discordgo.Session
	Message    *discordgo.Message
	Prefix     string
	Alias      string
	Args       Args
	Toks       Toks
	Err        error
	FmtRuleErr func(Rule) string // format a rule const into an error string
}

// ruleToErr converts an error string to a RuleError
func (c Context) ruleToErr(r Rule) error {
	return &RuleError{
		rule:   r,
		reason: c.FmtRuleErr(r),
	}
}

// NewContext returns an unpopulated context with defaults set
func NewContext() Context {
	return Context{
		Rule:       Rule(0),
		Session:    nil,
		Message:    nil,
		Prefix:     "",
		Alias:      "",
		Args:       nil,
		Toks:       Toks{},
		Err:        nil,
		FmtRuleErr: defaultFmtRule,
	}
}
