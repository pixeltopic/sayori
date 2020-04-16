package context

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

// New returns a new Context.
func New() *Context {
	return &Context{
		Ses:    nil,
		Msg:    nil,
		Prefix: nil,
		Alias:  []string{},
		Args:   []string{},
		Err:    nil,
	}
}
