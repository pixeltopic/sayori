package sayori

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
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
		Match(toks Toks) (string, bool)
	}

	// Event is an event that does not require a prefix or alias to handle
	Event interface {
		// Handle is where a command's business logic should belong.
		Handle(ctx Context) error
		// Catch is where an error in `ctx.Err` should be handled if non-nil.
		Catch(ctx Context)
	}
)

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
//
// if it fails, will return false with an empty string.
//
// if prefix is an empty string, will return the command as-is.
//
// if command is an empty string and prefix is not empty, will return false.
func (r *Router) trimPrefix(command, prefix string) (string, bool) {
	var c string
	if prefix == "" {
		return command, true
	}
	if command == "" {
		return command, false
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

			toks := newToks(cmd)
			if alias, ok = c.Match(toks); !ok {
				return
			}

			ctx := NewContext()
			ctx.Session = s
			ctx.Alias = alias
			ctx.Prefix = prefix
			ctx.Message = ev.Message
			ctx.Toks = toks

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
