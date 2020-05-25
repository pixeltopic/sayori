package v2

import (
	"context"

	"github.com/pixeltopic/sayori/v2/utils"

	"github.com/bwmarrin/discordgo"
)

// CmdContext is an aux structure for storing invocation values extracted from a given Context to reduce boilerplate.
//
// This need not be manually initialized; simply call CmdFromContext.
type CmdContext struct {
	Ses    *discordgo.Session
	Msg    *discordgo.Message
	Prefix string
	Alias  []string
	Args   []string
	Err    error
}

// CmdFromContext derives all Command invocation values from given Context.
func CmdFromContext(ctx context.Context) *CmdContext {
	return &CmdContext{
		Ses:    utils.GetSes(ctx),
		Msg:    utils.GetMsg(ctx),
		Prefix: utils.GetPrefix(ctx),
		Alias:  utils.GetAlias(ctx),
		Args:   utils.GetArgs(ctx),
		Err:    utils.GetErr(ctx),
	}
}
