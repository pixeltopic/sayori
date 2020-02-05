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

	// Aliaser returns a parsed alias and ok bool if an alias was parsed successfully from the fullcommand
	Aliaser interface {
		// accepts a command str but does not include prefix
		Match(fullcommand string) (string, bool)
	}

	// EventHandler can be used to handle a command
	EventHandler interface {
		// accepts a command str but does not include prefix
		Parse(fullcommand string) Args
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

// MessageHandler registers a message handler with an optional ruleset argument.
func (r *SessionRouter) MessageHandler(handler EventHandler, ruleset EventHandlerRule) {
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

			// Match alias. if no Aliaser interface is satisfied, skips alias and runs handler
			if a, ok := handler.(Aliaser); ok {
				if alias, ok := a.Match(command); ok {
					ctx.Alias = alias
				} else {
					return
				}
			}
			ctx.Message = v.Message
			ctx.Args = handler.Parse(command)

			handler.Handle(ctx)
		default:
		}
	})
}
