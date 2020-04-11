package v2

import (
	"github.com/bwmarrin/discordgo"
)

// Router maps commands to handlers.
type Router struct {
	*discordgo.Session
}

// New returns a new Router.
func New(s *discordgo.Session) *Router {
	return &Router{
		Session: s,
	}
}

// Has binds a Route to the Router.
func (r *Router) Has(route *Route) {

	handler := func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := NewContext()
		ctx.Msg = m.Message
		ctx.Ses = s

		route.handler(ctx)
	}

	r.addHandler(handler)
}

// HasOnce binds binds a Route to the Router, but the route will only fire at most once.
func (r *Router) HasOnce(route *Route) {

	handler := func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := NewContext()
		ctx.Msg = m.Message
		ctx.Ses = s

		route.handler(ctx)
	}

	r.addHandlerOnce(handler)
}

func (r *Router) addHandler(h interface{}) {
	if r.Session != nil && h != nil {
		r.AddHandler(h)
	}
}

func (r *Router) addHandlerOnce(h interface{}) {
	if r.Session != nil && h != nil {
		r.AddHandlerOnce(h)
	}
}
