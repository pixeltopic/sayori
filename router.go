package sayori

import (
	"github.com/bwmarrin/discordgo"
)

// Context contains data relating to the command invocation context
type Context struct {
	Session *discordgo.Session
	Message *discordgo.Message
	Prefix  string
	Cmd     string
}

// SessionRouter maps commands to handlers.
type SessionRouter struct {
	session       *discordgo.Session
	defaultPrefix string
}

// NewSessionRouter returns a new SessionRouter.
func NewSessionRouter(dg *discordgo.Session, defaultPrefix string) *SessionRouter {
	return &SessionRouter{
		session:       dg,
		defaultPrefix: defaultPrefix,
	}
}

// On registers a handler under an alias.
func (r *SessionRouter) On(alias string /*ruleset EventHandlerRule,*/, handler func(Context)) {
	r.session.AddHandler(func(s *discordgo.Session, i interface{}) {
		ctx := Context{
			Session: r.session,
			Prefix:  r.defaultPrefix,
			Cmd:     alias,
		}
		mc, ok := i.(*discordgo.MessageCreate)
		if ok {
			ctx.Message = mc.Message
		}
		// select prefix here
		// parse ruleset here
		// parse alias set here
		handler(ctx)
	})
}
