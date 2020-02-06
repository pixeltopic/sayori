package sayori

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// TODO:
// Test makeCommand and msgEvent by mocking messages (ensure ctx gets populated correctly) and test all other funcs

type (
	// Args is a set of args bound to identifiers that are parsed from the command
	Args map[string]interface{}

	// Toks are the tokens parsed from the command
	Toks []string

	// Prefixer identifies the prefix based on the serverID.
	Prefixer interface {
		// Load fetches a prefix that matches the serverID. Return args= (prefix, ok)
		Load(serverID string) (string, bool)
		// Default returns the default prefix
		Default() string
	}

	// Command is used to handle a command
	Command interface {
		Event

		// returns a parsed alias and ok bool if an alias was parsed successfully from the fullcommand
		// accepts a command str but does not include prefix
		Match(fullcommand string) (string, bool)

		Parse(fullcommand string) (Args, error)
	}

	// Event is an event that does not require a prefix or alias to handle
	Event interface {
		Handle(ctx Context) error
		Catch(ctx Context) // should handle errors attached in ctx.Err
	}
)

// Load loads a key from args
func (a Args) Load(key string) (interface{}, bool) {
	v, ok := a[key]
	return v, ok
}

// Store stores a key that maps to val in args
func (a Args) Store(key string, val interface{}) {
	a[key] = val
}

// Delete removes a key that maps to val in args, or if key does not exist, no-op
func (a Args) Delete(key string) {
	delete(a, key)
}

// Get retrieves the token matching the index
func (t Toks) Get(i int) (string, bool) {
	l := len(t)
	if i >= l || i < 0 {
		return "", false
	}
	return t[i], true
}

// newToks returns a slice of tokens split by whitespace
func newToks(s string) Toks {
	return Toks(strings.Fields(s))
}

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
		Rule:       NewRule(),
		Session:    nil,
		Message:    nil,
		Prefix:     "",
		Alias:      "",
		Args:       nil,
		Toks:       nil,
		Err:        nil,
		FmtRuleErr: defaultFmtRule,
	}
}

// Router maps commands to handlers.
type Router struct {
	session *discordgo.Session
	p       Prefixer
}

// New returns a new Router.
func New(dg *discordgo.Session, p Prefixer) *Router {
	return &Router{
		session: dg,
		p:       p,
	}
}

// getGuildPrefix returns guildID's custom prefix or if none, returns default prefix
func (r *Router) getGuildPrefix(guildID string) string {
	prefix, ok := r.p.Load(guildID)
	if !ok {
		prefix = r.p.Default()
	}
	return prefix
}

// trimPrefix accepts a command (with prefix attached) and attempts to return the command without the prefix.
// if it fails, will return false with an empty string.
// if prefix is an empty string, will return the command as-is.
func (r *Router) trimPrefix(command, prefix string) (string, bool) {
	var c string
	if prefix == "" {
		return command, true
	}
	if c = strings.TrimPrefix(command, prefix); c == command {
		return "", false
	}

	return c, true

}

// Has defines a handler in the router which should satisfy Event or Command interface
//
// If Rule is nil, will ignore.
//
// If `h` does not satisfy `Event` or `Command`, will not consider Rule regardless if nil or not.
func (r *Router) Has(h interface{}, rule *Rule) {
	var newHandler interface{}

	switch v := h.(type) {
	case Command:
		newHandler = r.makeCommand(v, rule)
	case Event:
		newHandler = r.msgEvent(v, rule)
	default:
		newHandler = h
	}
	r.session.AddHandler(newHandler)
}

// msgEvent registers a MessageCreate event handler that does not require an alias or prefix
func (r *Router) msgEvent(e Event, rule *Rule) func(*discordgo.Session, interface{}) {
	return func(s *discordgo.Session, i interface{}) {

		switch ev := i.(type) {
		case *discordgo.MessageCreate:
			ctx := NewContext()
			ctx.Session = s
			ctx.Message = ev.Message
			ctx.Toks = newToks(ev.Message.Content)

			if rule != nil {
				ctx.Rule = *rule
				if ok, failedRule := rule.allow(ctx); ok {
					ctx.Err = e.Handle(ctx)
				} else {
					ctx.Err = ctx.ruleToErr(failedRule)
				}
			} else {
				ctx.Err = e.Handle(ctx)
			}

			defer e.Catch(ctx)

		default:
		}
	}
}

// makeCommand registers a command with an optional rule argument.
func (r *Router) makeCommand(c Command, rule *Rule) func(*discordgo.Session, interface{}) {

	var (
		prefix, alias, cmd string
		ok                 bool
	)

	return func(s *discordgo.Session, i interface{}) {
		switch ev := i.(type) {
		case *discordgo.MessageCreate:

			cmd = ev.Message.Content
			prefix = r.getGuildPrefix(ev.Message.GuildID)

			if cmd, ok = r.trimPrefix(cmd, prefix); !ok {
				return
			}

			if alias, ok = c.Match(cmd); !ok {
				return
			}

			ctx := NewContext()
			ctx.Session = s
			ctx.Alias = alias
			ctx.Prefix = prefix
			ctx.Message = ev.Message
			ctx.Toks = newToks(cmd)

			args, err := c.Parse(cmd)
			if err != nil {
				ctx.Err = err
				defer c.Catch(ctx)
				return
			}
			ctx.Args = args

			if rule != nil {
				ctx.Rule = *rule
				if ok, failedRule := rule.allow(ctx); ok {
					ctx.Err = c.Handle(ctx)
				} else {
					ctx.Err = ctx.ruleToErr(failedRule)
				}
			} else {
				ctx.Err = c.Handle(ctx)
			}

			defer c.Catch(ctx)

		default:
		}
	}
}
