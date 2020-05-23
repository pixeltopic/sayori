package v2

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Router maps commands to handlers.
type Router struct {
	S *discordgo.Session
}

// New returns a new Router.
func New(s *discordgo.Session) *Router {
	return &Router{
		S: s,
	}
}

// HasDefault binds a default discordgo event handler to the builder.
func (r *Router) HasDefault(h interface{}) {
	r.addHandler(h)
}

// HasOnceDefault binds a default discordgo event handler to the builder.
func (r *Router) HasOnceDefault(h interface{}) {
	r.addHandlerOnce(h)
}

func makeHandlerForDgo(route *Route) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := context.Background()

		// finds deepest subroute and executes its handler with an accumulated context
		createHandlerFunc(route)(WithSes(WithMsg(ctx, m.Message), s))
	}
}

// Has binds a Route to the Router.
func (r *Router) Has(route *Route) {
	if route == nil {
		return
	}

	r.addHandler(makeHandlerForDgo(route))
}

// HasOnce binds binds a Route to the Router, but the route will only fire at most once.
func (r *Router) HasOnce(route *Route) {
	if route == nil {
		return
	}

	r.addHandlerOnce(makeHandlerForDgo(route))
}

func (r *Router) addHandler(h interface{}) {
	if r.S != nil && h != nil {
		r.S.AddHandler(h)
	}
}

func (r *Router) addHandlerOnce(h interface{}) {
	if r.S != nil && h != nil {
		r.S.AddHandlerOnce(h)
	}
}

// Open creates a websocket connection to Discord.
// See: https://discordapp.com/developers/docs/topics/gateway#connecting
func (r *Router) Open() error {
	return r.S.Open()
}

// Close closes a websocket and stops all listening/heartbeat goroutines.
func (r *Router) Close() error {
	return r.S.Close()
}
