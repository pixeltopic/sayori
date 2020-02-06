package main

import (
	"errors"
	"strings"

	"github.com/pixeltopic/sayori"
)

// EchoCmd defines a simple EchoCmd.
type EchoCmd struct {
	aliases []string
}

// Match returns a matched alias if the bool is true
func (c *EchoCmd) Match(fullcommand string) (string, bool) {
	for i := range c.aliases {
		if strings.HasPrefix(fullcommand, c.aliases[i]) {
			return c.aliases[i], true
		}
	}
	return "", false
}

// Parse returns args found in the command
func (c *EchoCmd) Parse(fullcommand string) (sayori.Args, error) {
	sslice := strings.Fields(fullcommand)
	if len(sslice) < 2 {
		return nil, errors.New("not enough args to echo :(")
	}
	return sayori.Args{
		"alias":   sslice[0],
		"to-echo": strings.Join(sslice[1:], " "),
	}, nil
}

// Handle handles the echo command
func (c *EchoCmd) Handle(ctx sayori.Context) error {
	if msg, ok := ctx.Args.Load("to-echo"); ok {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "Echoing! "+msg.(string))
	}
	return nil
}

// Catch handles any errors
func (c *EchoCmd) Catch(ctx sayori.Context) {
	if ctx.Err != nil {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Err.Error())
	}
}
