package v2

import (
	"fmt"
	"strings"
)

// TODO: remove useless funcs, move some logic into helper funcs, test subcommands, write tests for different combinations of interfaces/aliases/middlewares/etc

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

// Route wraps a command and contains its
// handler logic, aliases, and subroutes
type Route struct {
	c Commander
	p Prefixer

	aliases     []string
	handler     HandlerFunc // handler func is only executed on the top level if a route is not a subroute. if subroute, directly accesses handle/resolve
	subroutes   []*Route
	middlewares []Middlewarer
}

// HasAlias returns whether the given string is an alias or not.
// Not case sensitive.
//
// Since Events do not have an alias, if the route has no aliases it will default to true.
func (r *Route) HasAlias(a string) bool {
	if len(r.aliases) == 0 {
		return true
	}
	a = strings.ToLower(a)
	for _, alias := range r.aliases {
		if strings.ToLower(alias) == a {
			return true
		}
	}
	return false
}

// Find a subroute alias in this route's direct children.
// If alias is not found, will return nil
func (r *Route) Find(subAlias string) *Route {
	for _, sub := range r.subroutes {
		if sub.HasAlias(subAlias) {
			return sub
		}
	}
	return nil
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
// Identifiers must not have any whitespace.
func (r *Route) On(aliases ...string) *Route {
	r.aliases = append(r.aliases, aliases...)
	return r
}

// Has binds subroutes to the current route.
// Subroutes are executed sequentially,
// assuming the current route handler succeeds
func (r *Route) Has(subroutes ...*Route) *Route {
	r.subroutes = append(r.subroutes, subroutes...)
	return r
}

// Use adds middlewares to the route.
func (r *Route) Use(middlewares ...Middlewarer) *Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Do composes a commander implementation into a CtxHandler
//
// if cmd is nil, will simply update ctx and
// proceed to execute subhandlers if present
func (r *Route) Do(cmd Commander) *Route {
	r.c = cmd
	r.handler = r.createHandlerFunc()
	//r.handler = r.makeCtxHandler(cmd)
	return r
}

// findDeepestHandler finds the deepest route matching the subcommand and executes its handler
// with an accumulated context
//
// ctx must contain the session and message.
func (r *Route) createHandlerFunc() HandlerFunc {

	if r.c == nil {
		return nil
	}

	return func(ctx *Context) {
		var (
			alias string
			ok    bool
			cmd   = ctx.Msg.Content
		)

		// only do this if we are at a top level command
		// must compare with nil because Events have empty string prefixes
		if ctx.Prefix == nil {
			prefix := r.getGuildPrefix(ctx.Msg.GuildID)
			ctx.Prefix = &prefix
			if cmd, ok = trimPrefix(cmd, *ctx.Prefix); !ok {
				return
			}
		}

		fmt.Println("@@@ cmd:", cmd)

		// if parse fails, execute resolver for root level subroute
		//if c != nil {
		//	defer c.Resolve(ctx)
		//}

		args, err := handleParse(r.c, cmd)
		if err != nil {
			ctx.Err = err
			r.c.Resolve(ctx)
			return
		}

		fmt.Println("@@@ args:", args)

		route := r
		for depth := 0; len(args) > 0; depth++ {
			fmt.Printf("depth %d; args := %s\n", depth, args)
			//fmt.Println("@@@ args:", args)

			alias = args[0]
			if route.HasAlias(alias) {
				ctx.Alias = append(ctx.Alias, alias)
			} else {
				if depth == 0 {
					//ctx.Args = append(ctx.Args, args...)
					//ctx.Err = errors.New("no matching alias")
					return
				}
				ctx.Args = append(ctx.Args, args[1:]...)
				break
			}
			args = args[1:]
			if len(args) > 0 {
				subroute := route.Find(args[0])
				if subroute == nil {
					ctx.Args = append(ctx.Args, args...)
					break
				} else {
					fmt.Println("subroute found")
					route = subroute
				}
			}
		}

		if ctx.Err = handleMiddlewares(ctx, route.middlewares); ctx.Err != nil {
			route.c.Resolve(ctx)
			return
		}

		ctx.Err = route.c.Handle(ctx)

		route.c.Resolve(ctx)

	}
}

// makeCtxHandler creates a MsgHandler given a command
func (r *Route) makeCtxHandler(c Commander) HandlerFunc {
	//var (
	//	alias string
	//	ok    bool
	//)

	return func(ctx *Context) {
		//cmd := ctx.Msg.Content
		//
		//// only do this if we are at a top level command
		//// must compare with nil because Events have empty string prefixes
		//if ctx.Prefix == nil {
		//	prefix := r.getGuildPrefix(ctx.Msg.GuildID)
		//	ctx.Prefix = &prefix
		//	if cmd, ok = trimPrefix(cmd, *ctx.Prefix); !ok {
		//		return
		//	}
		//}

		//if c != nil {
		//	defer c.Resolve(ctx)
		//}

		//args, err := handleParse(c, cmd)
		//if err != nil {
		//	ctx.Err = err
		//	return
		//}

		//	if len(r.aliases) > 0 && len(args) >= 1 {
		//		if alias, ok = matchAlias(r.aliases, args[0]); !ok {
		//			return
		//		}
		//		// finish initializing ctx
		//		ctx.Alias = append(ctx.Alias, alias)
		//		ctx.Args = args[1:]
		//	} else if len(r.aliases) > 0 && len(args) == 0 {
		//		return
		//	}
		//
		//	if ctx.Err = handleMiddlewares(ctx, r.middlewares); ctx.Err != nil {
		//		return
		//	}
		//
		//	if c != nil {
		//		ctx.Err = c.Handle(ctx)
		//	}
		//
		//	for _, sub := range r.subroutes {
		//		sub.handler(CopyContext(ctx))
		//	}
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
