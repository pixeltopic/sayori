package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pixeltopic/sayori/v2/context"

	"github.com/bwmarrin/discordgo"
)

// EchoCmd defines a simple EchoCmd.
type EchoCmd struct{}

// Handle handles the echo command
func (c *EchoCmd) Handle(ctx *context.Context) error {

	if len(ctx.Args) == 0 {
		return errors.New("nothing to echo")
	}

	_, _ = ctx.Ses.ChannelMessageSend(
		ctx.Msg.ChannelID, "Echoing! "+strings.TrimSpace(strings.TrimPrefix(
			ctx.Msg.Content, *ctx.Prefix+ctx.Alias[len(ctx.Alias)-1])))

	return nil
}

// Resolve handles any errors
func (c *EchoCmd) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}

// EchoSubCmd defines a simple subcommand of EchoCmd.
type EchoSubCmd struct{}

// Aliases returns Aliases bound to EchoSubCmd
func (c *EchoSubCmd) Aliases() []string {
	return []string{"fmt", "f"}
}

// Handle handles the echo subcommand
func (c *EchoSubCmd) Handle(ctx *context.Context) error {

	if len(ctx.Args) == 0 {
		return errors.New("nothing to format echo")
	}

	toTrim := *ctx.Prefix + strings.Join(ctx.Alias, " ")

	_, _ = ctx.Ses.ChannelMessageSendEmbed(
		ctx.Msg.ChannelID, &discordgo.MessageEmbed{
			Description: fmt.Sprintf(`"%s" - %s#%s`, strings.TrimSpace(strings.TrimPrefix(
				ctx.Msg.Content, toTrim)), ctx.Msg.Author.Username, ctx.Msg.Author.Discriminator),
		})

	return nil
}

// Resolve handles any errors
func (c *EchoSubCmd) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}
