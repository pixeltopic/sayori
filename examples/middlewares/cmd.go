package main

import (
	"strings"

	"github.com/pixeltopic/sayori"
)

// AdminCmd is a privileged command only admins can use.
type AdminCmd struct{}

// Match examines the first token to see if it matches a valid alias
func (c *AdminCmd) Match(toks sayori.Toks) (string, bool) {
	alias, ok := toks.Get(0)
	if !ok {
		return "", false
	}
	alias = strings.ToLower(alias)

	for _, validAlias := range []string{"which"} {
		if alias == validAlias { // do not use HasPrefix, or something like `e!ec ho` will pass despite not matching any alias
			return alias, true
		}
	}
	return "", false
}

// Handle handles the echo command
func (c *AdminCmd) Handle(ctx sayori.Context) error {

	_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "You are privileged!")

	return nil
}

// Resolve handles any errors
func (c *AdminCmd) Resolve(ctx sayori.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Session.ChannelMessageSend(
			ctx.Message.ChannelID, ctx.Err.Error())
	}
}
