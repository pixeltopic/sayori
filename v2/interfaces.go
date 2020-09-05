package v2

import "context"

type (
	// CmdParser parses the content of a Discord message into a string slice.
	//
	// Optionally implemented by Handler
	CmdParser interface {
		Parse(string) ([]string, error)
	}

	// handlerFunc executes when a root route is invoked.
	// A root route is any route that is added to the router via Has
	//
	// Not to be confused with Handler. handlerFunc calls the proper Handler given a command.
	handlerFunc func(ctx context.Context)

	// Middlewarer allows execution of a handler before Handle is executed.
	//
	// Do accepts a context and returns an error. If error is nil, will execute the next Middlewarer or Handle.
	// Otherwise, it will enter the Resolve function (if implemented)
	//
	// Context mutated from within a middleware will only persist within scope.
	Middlewarer interface {
		Do(ctx context.Context) error
	}

	// Prefixer identifies the prefix based on the guildID and removes the prefix of the command string if matched.
	//
	// Load fetches a prefix that matches the guildID and returns the prefix mapped to the guildID with an ok bool.
	//
	// Default returns the default prefix
	Prefixer interface {
		Load(guildID string) (string, bool)
		Default() string
	}

	// Handler is bound to a route and will be called when handling Discord's Message Create events.
	// https://discord.com/developers/docs/topics/gateway#message-create
	//
	// Can optionally implement CmdParser and Resolver, but is not required.
	//
	// If Handler does not implement Resolver, any returned error will be ignored.
	Handler interface {
		Handle(ctx context.Context) error
	}

	// Resolver is an optional interface that can be satisfied by a command.
	// It is used for handling any errors returned from Handler.
	//
	// Optionally implemented by Handler
	Resolver interface {
		Resolve(ctx context.Context)
	}
)
