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
	Toks struct {
		Toks []string
		Raw  string
	}

	// Prefixer identifies the prefix based on the guildID.
	Prefixer interface {
		// Load fetches a prefix that matches the guildID.
		// returns the prefix mapped to the guildID with an `ok` bool.
		Load(guildID string) (string, bool)
		// Default returns the default prefix
		Default() string
	}

	// Parseable represents an entity that can be parsed.
	// Required for Command but optional for Event.
	Parseable interface {
		// Parse is where Toks will be parsed into Args.
		// if an error is non-nil, will immediately be handled by `Catch(ctx Context)`
		Parse(toks Toks) (Args, error)
	}

	// Command is used to handle a command
	Command interface {
		Event
		Parseable

		// Match is where a prefix-less command will be matched with a given alias.
		// returns an alias parsed from the command with an `ok` bool.
		// if false, will immediately terminate the handler execution.
		Match(cmd string) (string, bool)
	}

	// Event is an event that does not require a prefix or alias to handle
	Event interface {
		// Handle is where a command's business logic should belong.
		Handle(ctx Context) error
		// Catch is where an error in `ctx.Err` should be handled if non-nil.
		Catch(ctx Context)
	}
)

// NewArgs makes a new instance of Args for storing key-argument mappings
func NewArgs() Args {
	return make(Args)
}

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

// Will defines a handler in the router which should satisfy Event or Command interface. Is an alias for Has.
//
// If Rule is nil, will ignore.
//
// If `h` does not satisfy `Event` or `Command`, will not consider Rule regardless if nil or not.
func (r *Router) Will(h interface{}, rule *Rule) {
	r.Has(h, rule)
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
		newHandler = r.makeMsgEvent(v, rule)
	default:
		newHandler = h
	}
	r.session.AddHandler(newHandler)
}

// makeMsgEvent registers a MessageCreate event handler that does not require an alias or prefix
func (r *Router) makeMsgEvent(e Event, rule *Rule) func(*discordgo.Session, interface{}) {
	return func(s *discordgo.Session, i interface{}) {

		switch ev := i.(type) {
		case *discordgo.MessageCreate:
			ctx := NewContext()
			ctx.Session = s
			ctx.Message = ev.Message
			ctx.Toks = newToks(ev.Message.Content)

			if p, ok := e.(Parseable); ok {
				args, err := p.Parse(ctx.Toks)
				if err != nil {
					ctx.Err = err
					defer e.Catch(ctx)
					return
				}
				ctx.Args = args
			}

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

			args, err := c.Parse(ctx.Toks)
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
