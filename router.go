package sayori

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
	// Args is a set of args parsed from the command
	Args map[string]interface{}

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

		Parse(fullcommand string) Args
	}

	// Event is an event that does not require a prefix or alias to handle
	Event interface {
		Handle(ctx Context)
	}
)

// Context contains data relating to the command invocation context
type Context struct {
	Session *discordgo.Session
	Message *discordgo.Message
	Prefix  string
	Alias   string
	Args    Args
}

// SessionRouter maps commands to handlers.
type SessionRouter struct {
	session *discordgo.Session
	p       Prefixer
}

// NewSessionRouter returns a new SessionRouter.
func NewSessionRouter(dg *discordgo.Session, p Prefixer) *SessionRouter {
	return &SessionRouter{
		session: dg,
		p:       p,
	}
}

// AddHandler calls discordgo.Session.AddHandler
func (r *SessionRouter) AddHandler(handler interface{}) {
	r.session.AddHandler(handler)
}

// AddEvent registers a MessageCreate event handler that does not require an alias or prefix
func (r *SessionRouter) AddEvent(e Event, rules EventHandlerRule) {
	r.session.AddHandler(func(s *discordgo.Session, i interface{}) {
		ctx := Context{
			Session: r.session,
		}
		switch v := i.(type) {
		case *discordgo.MessageCreate:
			// TODO: parse rules here

			ctx.Message = v.Message
			ctx.Args = nil

			e.Handle(ctx)
		default:
		}
	})
}

// AddCommand registers a command with an optional ruleset argument.
func (r *SessionRouter) AddCommand(c Command, rules EventHandlerRule) {
	r.session.AddHandler(func(s *discordgo.Session, i interface{}) {
		ctx := Context{
			Session: r.session,
		}
		switch v := i.(type) {
		case *discordgo.MessageCreate:
			// TODO: parse ruleset here

			command := v.Message.Content

			// Check fullcommand for prefix; retrieves custom prefix or uses default.
			// empty prefix with ok=true should ALWAYS pass.

			prefix, ok := r.p.Load(v.GuildID)
			if !ok {
				prefix = r.p.Default()
			}
			if !strings.HasPrefix(command, prefix) {
				return
			}
			ctx.Prefix = prefix
			command = strings.TrimPrefix(command, ctx.Prefix)

			if alias, ok := c.Match(command); ok {
				ctx.Alias = alias
			} else {
				return
			}

			ctx.Message = v.Message
			ctx.Args = c.Parse(command)

			c.Handle(ctx)
		default:
		}
	})
}
