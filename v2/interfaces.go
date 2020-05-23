package v2

import "context"

type (
	// CmdParser parses the content of a Discord message into a string slice.
	//
	// Optionally implemented by Commander
	CmdParser interface {
		Parse(string) ([]string, error)
	}

	// handlerFunc executes when a root route is invoked.
	// A root route is any route that is added to the router via Has
	handlerFunc func(ctx context.Context)

	// Middlewarer allows execution of a handler before Handle is executed.
	//
	// Do accepts a context and returns an error. If error is nil, will execute the next Middlewarer or Handle.
	// Otherwise, it will enter the Resolve function.
	//
	// If context is mutated within the Middlewarer, it will propagate to future handlers. For this reason, it is encouraged
	// to treat context as read-only.
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

	// Commander is used by a route to handle Discord's Message Create events.
	// https://discord.com/developers/docs/topics/gateway#message-create
	//
	// Can optionally implement CmdParser, but is not required.
	//
	// Handle is where a command's business logic should belong.
	//
	// Resolve is where an error in ctx.Err can be handled, along with any other necessary cleanup.
	// It is run if (custom) parsing fails, middleware fails, or if Handle fails.
	Commander interface {
		Handle(ctx context.Context) error
		Resolve(ctx context.Context)
	}
)
