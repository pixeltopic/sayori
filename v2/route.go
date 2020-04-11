package v2

import (
	"strings"
)

var cmdParserDefault = strings.Fields

func matchAlias(aliases []string, token string) (string, bool) {
	alias := strings.ToLower(token)

	for _, validAlias := range aliases {
		if alias == strings.ToLower(validAlias) { // do not use HasPrefix, or something like `e!ec ho` will pass despite not matching any alias
			return alias, true
		}
	}

	return "", false
}

// handleParse checks if an event implements Parseable; if it does, runs Parseable. Else, runs default parser
func handleParse(c Commander, content string) ([]string, error) {
	var (
		p  CmdParser
		ok bool
	)
	if p, ok = c.(CmdParser); !ok {
		return cmdParserDefault(content), nil
	}

	return p.Parse(content)
}

// handleMiddlewares runs each middleware in order until completion.
// Will abort on the first error returned by a middleware.
func handleMiddlewares(ctx *Context, m []Middlewarer) error {
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

// Route wraps a command and contains its handler logic, aliases, and subroutes
type Route struct {
	p   Prefixer
	cmd Commander

	aliases     []string
	handler     CtxHandler
	subroutes   []*Route
	middlewares []Middlewarer
}

// getGuildPrefix returns guildID's custom prefix or if none, returns default prefix
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

// On adds new identifiers for a Route
func (r *Route) On(aliases ...string) *Route {
	r.aliases = append(r.aliases, aliases...)
	return r
}

// Has binds subroutes to the current route.
// Subroutes are executed sequentially, assuming the current route handler succeeds
func (r *Route) Has(subroutes ...*Route) *Route {
	r.subroutes = append(r.subroutes, subroutes...)
	return r
}

// Use adds middlewares to the route.
func (r *Route) Use(middlewares ...Middlewarer) *Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Do composes a commander implementation into a MsgHandler
//
// if `cmd` is nil, will simply update ctx and proceed to execute subhandlers if present
func (r *Route) Do(cmd Commander) *Route {
	r.handler = r.makeCtxHandler(cmd)
	return r
}

// makeCtxHandler creates a MsgHandler given a command
func (r *Route) makeCtxHandler(c Commander) CtxHandler {
	var (
		alias string
		ok    bool
	)

	return func(ctx *Context) {
		cmd := ctx.Msg.Content

		// only do this if we are at a top level command
		// must compare with nil because Events have empty string prefixes
		if ctx.Prefix == nil {
			prefix := r.getGuildPrefix(ctx.Msg.GuildID)
			ctx.Prefix = &prefix
			if cmd, ok = trimPrefix(cmd, *ctx.Prefix); !ok {
				return
			}
		}

		if c != nil {
			defer c.Resolve(ctx)
		}

		args, err := handleParse(c, cmd)
		if err != nil {
			ctx.Err = err
			return
		}

		// TODO: find out if these aliases get updated if this is called before alias/middleware initialization
		if alias, ok = matchAlias(r.aliases, alias); !ok {
			return
		}

		// finish initializing ctx
		ctx.Alias = append(ctx.Alias, alias)
		ctx.Args = args[1:]

		if ctx.Err = handleMiddlewares(ctx, r.middlewares); ctx.Err != nil {
			return
		}

		if c != nil {
			ctx.Err = c.Handle(ctx)
		}

		for _, sub := range r.subroutes {
			sub.handler(ctx) // TODO: create a deep copy of ctx in case multiple subroutes get executed
		}
	}
}

// NewRoute returns a new Route.
func NewRoute(p Prefixer) *Route {
	return &Route{
		p:           p,
		aliases:     []string{},
		subroutes:   []*Route{},
		middlewares: []Middlewarer{},
	}
}
