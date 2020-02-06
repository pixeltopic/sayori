package sayori

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

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

// Get retrieves the token matching the index
func (t Toks) Get(i int) (string, bool) {
	l := len(t)
	if i >= l || i < 0 {
		return "", false
	}
	return t[i], true
}

func newToks(command string) Toks {
	return Toks(strings.Fields(command))
}

// Context contains data relating to the command invocation context
type Context struct {
	Session *discordgo.Session
	Message *discordgo.Message
	Prefix  string
	Alias   string
	Args    Args
	Toks    Toks
	Err     error
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

// AddHandler calls discordgo.Session.AddHandler
func (r *Router) AddHandler(handler interface{}) {
	r.session.AddHandler(handler)
}

// Has defines a handler in the router which should satisfy Event or Command interface
func (r *Router) Has(h interface{}, rules EventHandlerRule) {
	switch v := h.(type) {
	case Command:
		r.command(v, rules)
	case Event:
		r.msgEvent(v, rules)
	default:

	}
}

// event registers a MessageCreate event handler that does not require an alias or prefix
func (r *Router) msgEvent(e Event, rules EventHandlerRule) {
	r.session.AddHandler(func(s *discordgo.Session, i interface{}) {
		ctx := Context{
			Session: r.session,
		}
		switch v := i.(type) {
		case *discordgo.MessageCreate:
			// TODO: parse rules here

			ctx.Message = v.Message
			ctx.Args = nil

			ctx.Err = e.Handle(ctx)

			defer e.Catch(ctx)

		default:
		}
	})
}

// command registers a command with an optional ruleset argument.
func (r *Router) command(c Command, rules EventHandlerRule) {

	var (
		prefix, alias, cmd string
		ok                 bool
	)

	r.session.AddHandler(func(s *discordgo.Session, i interface{}) {
		switch v := i.(type) {
		case *discordgo.MessageCreate:
			// TODO: parse ruleset here

			cmd = v.Message.Content
			prefix = r.getGuildPrefix(v.Message.GuildID)

			if cmd, ok = r.trimPrefix(cmd, prefix); !ok {
				return
			}

			if alias, ok = c.Match(cmd); !ok {
				return
			}

			ctx := Context{
				Session: r.session,
				Alias:   alias,
				Prefix:  prefix,
				Message: v.Message,
				Toks:    newToks(cmd),
			}

			args, err := c.Parse(cmd)
			if err != nil {
				ctx.Err = err
				defer c.Catch(ctx)
				return
			}
			ctx.Args = args
			ctx.Err = c.Handle(ctx)

			defer c.Catch(ctx)

		default:
		}
	})
}
