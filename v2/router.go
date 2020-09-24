package v2

import (
	"context"

	"github.com/pixeltopic/sayori/v2/utils"

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

// HasDefault binds a default DiscordGo event handler to the builder.
// It is useful when there is a handler that consumes something other than a MessageCreate event.
//
// It returns a function that will remove the handler when executed.
func (r *Router) HasDefault(h interface{}) func() {
	return r.addHandler(h)
}

// HasOnceDefault binds a default DiscordGo event handler to the builder.
// It is useful when there is a handler that consumes something other than a MessageCreate event.
// The added handler will be removed upon the first execution.
//
// It returns a function that will remove the handler when executed.
func (r *Router) HasOnceDefault(h interface{}) func() {
	return r.addHandlerOnce(h)
}

func makeHandlerForDgo(route Route) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := context.Background()

		// finds deepest subroute and executes its handler with an accumulated context
		createHandlerFunc(&route)(utils.WithSes(utils.WithMsg(ctx, m.Message), s))
	}
}

// Has binds a Route to the Router.
func (r *Router) Has(route Route) func() {
	return r.addHandler(makeHandlerForDgo(route))
}

// HasOnce binds binds a Route to the Router, but the route will only fire at most once.
func (r *Router) HasOnce(route Route) func() {
	return r.addHandlerOnce(makeHandlerForDgo(route))
}

func (r *Router) addHandler(h interface{}) func() {
	if r.S != nil && h != nil {
		return r.S.AddHandler(h)
	}
	return nil
}

func (r *Router) addHandlerOnce(h interface{}) func() {
	if r.S != nil && h != nil {
		return r.S.AddHandlerOnce(h)
	}
	return nil
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
