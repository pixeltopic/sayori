package v2

import (
	"github.com/bwmarrin/discordgo"
)

// Context of an incoming discordgo.MessageCreate event
type Context struct {
	Ses    *discordgo.Session
	Msg    *discordgo.Message
	Prefix *string
	Alias  []string
	Toks   []string
	Err    error
}

// NewContext returns a new Context.
func NewContext() *Context {
	return &Context{
		Ses:    nil,
		Msg:    nil,
		Prefix: nil,
		Alias:  []string{},
		Toks:   []string{},
		Err:    nil,
	}
}
