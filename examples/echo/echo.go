package main

import (
	"errors"
	"strings"

	"github.com/pixeltopic/sayori"
)

// EchoCmd defines a simple EchoCmd.
type EchoCmd struct{}

// Match attempts to find an alias contained within a prefix-less command
func (c *EchoCmd) Match(toks sayori.Toks) (string, bool) {
	alias, ok := toks.Get(0)
	if !ok {
		return "", false
	}
	alias = strings.ToLower(alias)

	for _, validAlias := range []string{"e", "echo"} {
		if alias == validAlias { // do not use HasPrefix, or something like `e!ec ho` will pass despite not matching any alias
			return alias, true
		}
	}
	return "", false
}

// Parse parses the command tokens and
// generates a mapping of arguments to keys
func (c *EchoCmd) Parse(toks sayori.Toks) (sayori.Args, error) {
	if toks.Len() < 2 {
		return nil, errors.New("not enough args to echo :(")
	}
	alias, _ := toks.Get(0)

	args := sayori.NewArgs()

	args.Store("alias", alias)
	args.Store("to-echo", strings.Join(toks.Iter()[1:], " "))

	return args, nil
}

// Handle handles the echo command
func (c *EchoCmd) Handle(ctx sayori.Context) error {
	if msg, ok := ctx.Args.Load("to-echo"); ok {
		ctx.Session.ChannelMessageSend(
			ctx.Message.ChannelID, "Echoing! "+msg.(string))
	}
	return nil
}

// Resolve handles any errors
func (c *EchoCmd) Resolve(ctx sayori.Context) {
	if ctx.Err != nil {
		ctx.Session.ChannelMessageSend(
			ctx.Message.ChannelID, ctx.Err.Error())
	}
}
