package sayori

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
	// Prefixer identifies the prefix based on the guildID before a `Command` execution and removes the prefix of the command string if matched.
	//
	// `Load` fetches a prefix that matches the `guildID` and returns the prefix mapped to the `guildID` with an `ok` bool.
	//
	// `Default` returns the default prefix
	Prefixer interface {
		Load(guildID string) (string, bool)
		Default() string
	}

	// Parseable represents an entity that can be parsed. It is implemented by `Command` but optional for `Event`
	//
	// `Parse` is where `Toks` will be parsed into `Args`. If an error is non-nil, will immediately be handled by `Catch(ctx Context)`
	Parseable interface {
		Parse(toks Toks) (Args, error)
	}

	// Command is used to handle a command which will only be run on a `*discordgo.MessageCreate` event.
	// Encapsulates the implementation of both `Event` and `Parseable`.
	//
	// `Match` is where a command with a trimmed prefix will be matched an alias. It returns an alias parsed from the command with an `ok` bool.
	// If `ok` is false, the Command will immediately be terminated.
	Command interface {
		Event
		Parseable
		Match(toks Toks) (string, bool)
	}

	// Event is used to handle a `*discordgo.MessageCreate` event.
	// Only contains the core functions required to implement a `Command`, thus does not require a prefix or alias to be parsed.
	// Can optionally implement `Parseable`, but is not required.
	//
	// `Handle` is where a command's business logic should belong.
	//
	// `Catch` is where an error in `ctx.Err` should be handled if non-nil.
	Event interface {
		Handle(ctx Context) error
		Catch(ctx Context)
	}
)

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

// Command binds a `Command` implementation to the builder.
func (r *Router) Command(c Command) *Builder {
	b := &Builder{}
	return b.command(c)
}

// Event binds an `Event` implementation to the builder.
func (r *Router) Event(e Event) *Builder {
	b := &Builder{}
	return b.event(e)
}

// HandleDefault binds a default discordgo event handler to the builder.
func (r *Router) HandleDefault(h interface{}) *Builder {
	b := &Builder{
		handler: h,
	}
	return b
}

// getGuildPrefix returns guildID's custom prefix or if none, returns default prefix
func (r *Router) getGuildPrefix(guildID string) string {
	prefix, ok := r.p.Load(guildID)
	if !ok {
		prefix = r.p.Default()
	}
	return prefix
}

// trimPrefix accepts a command (with prefix attached) and attempts to return the command without the prefix.
//
// if it fails, will return false with an empty string.
//
// if prefix is an empty string, will return the command as-is.
//
// if command is an empty string and prefix is not empty, will return false.
func (r *Router) trimPrefix(command, prefix string) (string, bool) {
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

// Has binds a `Builder` to the router which should implement `Event` or `Command`
// with any desired `Filter` to control when the handler fires.
//
// `Filter` has consts defined in the package that start with the prefix `Messages*`
func (r *Router) Has(b *Builder) {
	var newHandler interface{}

	switch v := b.handler.(type) {
	case Command:
		newHandler = r.makeCommand(v, b.filter)
	case Event:
		newHandler = r.makeEvent(v, b.filter)
	default:
		newHandler = b.handler
	}
	r.addHandler(newHandler)
}

// HasOnce binds a `Builder` to the router which should implement `Event` or `Command`
// with any desired `Filter` to control when the handler fires.
//
// `Filter` has consts defined in the package that start with the prefix `Messages*`
//
// `b` will only fire at most once.
func (r *Router) HasOnce(b *Builder) {
	var newHandler interface{}

	switch v := b.handler.(type) {
	case Command:
		newHandler = r.makeCommand(v, b.filter)
	case Event:
		newHandler = r.makeEvent(v, b.filter)
	default:
		newHandler = b.handler
	}
	r.addHandlerOnce(newHandler)
}

func (r *Router) addHandler(h interface{}) {
	if r.session != nil && h != nil {
		r.session.AddHandler(h)
	}
}

func (r *Router) addHandlerOnce(h interface{}) {
	if r.session != nil && h != nil {
		r.session.AddHandlerOnce(h)
	}
}

// makeEvent registers a MessageCreate event handler that does not require an alias or prefix
func (r *Router) makeEvent(e Event, f Filter) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := NewContext()
		ctx.Session = s
		ctx.Message = m.Message
		ctx.Toks = NewToks(m.Message.Content)

		if p, ok := e.(Parseable); ok {
			args, err := p.Parse(ctx.Toks)
			if err != nil {
				ctx.Err = err
				defer e.Catch(ctx)
				return
			}
			ctx.Args = args
		}

		ctx.Filter = f
		if ok, failedFilters := f.allow(ctx); ok {
			ctx.Err = e.Handle(ctx)
		} else {
			ctx.Err = ctx.filterToErr(failedFilters)
		}

		defer e.Catch(ctx)
	}
}

// makeCommand registers a command
func (r *Router) makeCommand(c Command, f Filter) func(*discordgo.Session, *discordgo.MessageCreate) {
	var (
		prefix, alias, cmd string
		ok                 bool
	)

	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		cmd = m.Message.Content
		prefix = r.getGuildPrefix(m.Message.GuildID)

		if cmd, ok = r.trimPrefix(cmd, prefix); !ok {
			return
		}

		toks := NewToks(cmd)
		if alias, ok = c.Match(toks); !ok {
			return
		}

		ctx := NewContext()
		ctx.Session = s
		ctx.Alias = alias
		ctx.Prefix = prefix
		ctx.Message = m.Message
		ctx.Toks = toks

		args, err := c.Parse(ctx.Toks)
		if err != nil {
			ctx.Err = err
			defer c.Catch(ctx)
			return
		}
		ctx.Args = args
		ctx.Filter = f

		if ok, failedFilters := f.allow(ctx); ok {
			ctx.Err = c.Handle(ctx)
		} else {
			ctx.Err = ctx.filterToErr(failedFilters)
		}

		defer c.Catch(ctx)
	}
}
