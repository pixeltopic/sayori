package v2

import "github.com/pixeltopic/sayori/v2/context"

type (
	// CmdParser parses the content of a Discord message into a string slice.
	//
	// Optionally implemented by Commander
	CmdParser interface {
		Parse(string) ([]string, error)
	}

	// HandlerFunc handles the command given a Context.
	HandlerFunc func(ctx *context.Context)

	// Middlewarer allows a custom handler to determine if a message should be routed to the Command or Event handler.
	//
	// Do accepts a context and returns an error. If error is nil, will execute the next middleware or the Command or Event handler.
	// Otherwise, it will renter the Resolve function.
	//
	// If context is mutated within the middleware, it will propagate to future handlers.
	Middlewarer interface {
		Do(ctx *context.Context) error
	}

	// Prefixer identifies the prefix based on the guildID before a Command execution and removes the prefix of the command string if matched.
	//
	// Load fetches a prefix that matches the guildID and returns the prefix mapped to the guildID with an ok bool.
	//
	// Default returns the default prefix
	Prefixer interface {
		Load(guildID string) (string, bool)
		Default() string
	}

	// Commander is used to handle a command which will only be run on a *discordgo.MessageCreate event.
	// Can optionally implement CmdParser, but is not required.
	//
	// Handle is where a command's business logic should belong.
	//
	// Resolve is where an error in ctx.Err can be handled, along with any other necessary cleanup. It will always be the last function run.
	Commander interface {
		Handle(ctx *context.Context) error
		Resolve(ctx *context.Context)
	}
)
