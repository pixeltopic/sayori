package v2

import (
	"github.com/bwmarrin/discordgo"
)

// Context of an incoming discordgo.MessageCreate event.
// Ses and Msg should be read only.
type Context struct {
	Ses    *discordgo.Session
	Msg    *discordgo.Message
	Prefix *string
	Alias  []string
	Args   []string
	Err    error
}

// CopyContext clones a Context. Useful for passing contexts into subroutes.
// Err is not copied over, and prefix will still point to the same string as it should not change.
func CopyContext(ctx *Context) *Context {

	cp := NewContext()

	cp.Alias = append(cp.Alias, ctx.Alias...)
	cp.Args = append(cp.Args, ctx.Args...)

	cp.Prefix = ctx.Prefix

	return cp
}

// NewContext returns a new Context.
func NewContext() *Context {
	return &Context{
		Ses:    nil,
		Msg:    nil,
		Prefix: nil,
		Alias:  []string{},
		Args:   []string{},
		Err:    nil,
	}
}
