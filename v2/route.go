package v2

import (
	"context"
	"strings"

	"github.com/pixeltopic/sayori/v2/utils"
)

var cmdParserDefault = strings.Fields

// handleParse checks if an event implements CmdParser; if it does, runs CmdParser. Else, runs default parser
func handleParse(h Handler, content string) ([]string, error) {
	var (
		p  CmdParser
		ok bool
	)
	if p, ok = h.(CmdParser); !ok {
		return cmdParserDefault(content), nil
	}

	return p.Parse(content)
}

// handleResolve returns the Resolve func if Handler implements Resolver. Otherwise returns a Resolve func stub.
func handleResolve(h Handler) func(ctx context.Context) {
	if r, ok := h.(Resolver); ok {
		return r.Resolve
	}
	return func(_ context.Context) {}
}

// handleMiddlewares runs each middleware in order until completion.
// Will abort on the first error returned by a middleware.
func handleMiddlewares(ctx context.Context, m []Middlewarer) error {
	for i := 0; i < len(m); i++ {
		if err := m[i].Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// trimPrefix accepts a command (with prefix attached) and attempts to return the command without the prefix.
//
// if it fails, will return false with an empty string.
//
// if prefix is an empty string, will return the command as-is.
//
// if command is an empty string and prefix is not empty, will return false.
func trimPrefix(command, prefix string) (string, bool) {
	var c string
	if prefix == "" && command != "" {
		return command, true
	}
	if command == "" {
		return command, false
	}
	if c = strings.TrimPrefix(command, prefix); c == command {
		return "", false
	}

	return c, len(c) != 0 // if command was "[prefix]" and it's trimmed into "" it should be false

}

// Route represents a command which consumes a DiscordGo MessageCreate event under the hood.
//
// Routes can be modified after they are added to the router, and are not goroutine safe.
type Route struct {
	h           Handler
	p           Prefixer
	aliases     []string
	subroutes   []*Route
	middlewares []Middlewarer
}

// IsDefault returns true if a route will always be executed when Discord produces a Message Create event
// with respect to the Prefixer (if provided).
//
// https://discord.com/developers/docs/topics/gateway#message-create
//
// A default route will only be executed if it is not a subroute. Any subroutes it might contain will be ignored.
func (r *Route) IsDefault() bool {
	return len(r.aliases) == 0
}

// HasAlias returns if the given string is a case-insensitive alias of the current route
func (r *Route) HasAlias(a string) bool {
	if r.IsDefault() {
		return false
	}
	a = strings.ToLower(a)
	for _, alias := range r.aliases {
		if strings.ToLower(alias) == a {
			return true
		}
	}
	return false
}

// Find the first subroute of this route matching the given subroute alias.
// Subroute match is prioritized by order of the subroute appends.
// If alias is not found, will return nil.
func (r *Route) Find(subAlias string) *Route {
	for _, sub := range r.subroutes {
		if sub.HasAlias(subAlias) {
			return sub
		}
	}
	return nil
}

// FindAllSubroutes the all applicable subroutes of this route matching the given subroute alias
// from its immediate children
func (r *Route) FindAllSubroutes(subAlias string) (routes []*Route) {
	for _, sub := range r.subroutes {
		if sub.HasAlias(subAlias) {
			routes = append(routes, sub)
		}
	}
	return
}

// getGuildPrefix returns guildID's custom prefix or if none,
// returns default prefix
func (r *Route) getGuildPrefix(guildID string) string {
	if r.p == nil {
		return ""
	}
	prefix, ok := r.p.Load(guildID)
	if !ok {
		prefix = r.p.Default()
	}
	return prefix
}

// On adds new identifiers for a Route.
// By default, aliases with whitespaces will not be matched unless a Commander also implements the CmdParser interface
func (r *Route) On(aliases ...string) *Route {
	r.aliases = append(r.aliases, aliases...)
	return r
}

// Has binds subroutes to the current route.
// Subroutes with duplicate aliases will be prioritized in order of which they were added.
func (r *Route) Has(subroutes ...*Route) *Route {
	r.subroutes = append(r.subroutes, subroutes...)
	return r
}

// Use adds middlewares to the route.
// Middlewares are executed in order of which they were added, and will always run immediately before the core handler.
// Middleware errors can be handled by Resolve.
func (r *Route) Use(middlewares ...Middlewarer) *Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Do execution of the provided Handler when there is a Message Create event.
// https://discord.com/developers/docs/topics/gateway#message-create
//
// A Commander can optionally implement CmdParser for custom parsing of message content.
// Handling errors and cleanup of other resources will be handled by Resolve if Resolver is implemented, otherwise skipped.
//
// No-ops if Handler is nil.
//
// If Do is called multiple times, the previous Do call will be overwritten.
func (r *Route) Do(h Handler) *Route {
	r.h = h
	return r
}

// NewRoute returns a new Route.
//
// If Prefixer is nil, the route's prefix will be assumed to be empty.
// The Prefixer will only be run if a Route is added to the Router as a root-level route, and not a subroute.
func NewRoute(p Prefixer) *Route {
	return &Route{
		p:           p,
		aliases:     []string{},
		subroutes:   []*Route{},
		middlewares: []Middlewarer{},
	}
}

// createHandlerFunc returns a HandlerFunc which is a wrapper around the given Commander.
// A handlerFunc is only executed on the on a root level route.
// If subroute, will only access its respective Handle and Resolve
//
// It handles prefix trimming, message parsing,
// and search for the deepest subroute matching a given ctx.
// ctx gets accumulated as the HandlerFunc executes.
//
// ctx must contain the session and message.
func createHandlerFunc(route *Route) handlerFunc {

	if route == nil {
		return nil
	}
	if route.h == nil {
		return nil
	}

	return func(ctx context.Context) {
		var (
			ok  bool
			msg = utils.GetMsg(ctx)
			cmd = msg.Content
		)

		prefix := route.getGuildPrefix(msg.GuildID)
		ctx = utils.WithPrefix(ctx, prefix)
		if cmd, ok = trimPrefix(cmd, prefix); !ok {
			return
		}

		args, err := handleParse(route.h, cmd)
		if err != nil {
			ctx = utils.WithErr(ctx, err)
			handleResolve(route.h)(ctx)
			return
		}

		route, depth := findRouteRecursive(route, args, 1)
		if route == nil {
			return
		}

		ctx = utils.WithAlias(ctx, args[:depth])
		ctx = utils.WithArgs(ctx, args[depth:])

		if err = handleMiddlewares(ctx, route.middlewares); err != nil {
			ctx = utils.WithErr(ctx, err)
			handleResolve(route.h)(ctx)
			return
		}

		err = route.h.Handle(ctx)
		if err != nil {
			ctx = utils.WithErr(ctx, err)
		}

		handleResolve(route.h)(ctx)

	}
}

// findRoute finds the deepest subroute and returns it along with the depth.
// a depth of zero means there is no route that matches provided args
// the value of depth is equal to the number of aliases.
func findRoute(route *Route, args []string) (*Route, int) {
	var depth int

	if len(args) == 0 || route == nil {
		return nil, depth
	}

	// no aliases means this is an event handler so immediately return
	if route.IsDefault() {
		return route, depth
	}

	if !route.HasAlias(args[0]) {
		return nil, depth
	}

	// note: depth is incremented even if depth < len(args) except on the initial entry
	// if a route with no aliases is found in this loop, will consider it not found and return last valid route
	for depth = 1; depth < len(args); depth++ {

		// finds a subroute matching the token from a given route; if no match returns nil
		// will not match subroutes that have no aliases.
		subroute := route.Find(args[depth])

		if subroute == nil {
			return route, depth
		}

		// we can keep looking deeper.
		route = subroute
	}

	return route, depth
}

// findRouteRecursive finds the deepest subroute and returns it along with the depth.
// a depth of zero means there is no route that matches provided args
// the value of depth is equal to the number of aliases.
//
// Unlike findRoute, it scans the ENTIRE command tree for the first valid match. Meaning if there are duplicate subroutes (with identical aliases)
// it will "merge" the subroutes' subroutes on a first-in-first-called basis
//
// initial depth MUST be 1
func findRouteRecursive(route *Route, args []string, depth int) (*Route, int) {
	if depth <= 0 {
		return nil, 0
	}

	if len(args) == 0 || route == nil {
		return nil, depth - 1
	}

	// no aliases means this is an event handler so immediately return
	if route.IsDefault() {
		return route, depth - 1
	}

	// more recent arg must be an alias of current route. this should only ever fail on a root route.
	if !route.HasAlias(args[depth-1]) {
		return nil, depth - 1
	}

	finalDepth := depth // finalDepth is a temp variable so depth does not get reassigned, invalidating subsequent iterations

	if depth < len(args) {
		for _, sr := range route.FindAllSubroutes(args[depth]) {
			newRoute, newDepth := findRouteRecursive(sr, args, depth+1)

			// depth check prevents shallower subroutes from overwriting a better match.
			// <= will prioritize most recently added subroutes while < will prioritize least recently added
			if finalDepth <= newDepth && newRoute != nil {
				route, finalDepth = newRoute, newDepth
			}
		}
	}

	return route, finalDepth
}
