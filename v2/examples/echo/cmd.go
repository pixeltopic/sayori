package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pixeltopic/sayori/v2/context"

	"github.com/bwmarrin/discordgo"
)

// Echo defines a simple EchoCmd.
type Echo struct{}

// Handle handles the echo command
func (*Echo) Handle(ctx *context.Context) error {

	if len(ctx.Args) == 0 {
		return errors.New("nothing to echo")
	}

	_, _ = ctx.Ses.ChannelMessageSend(
		ctx.Msg.ChannelID, "Echoing! "+strings.TrimSpace(strings.TrimPrefix(
			ctx.Msg.Content, *ctx.Prefix+ctx.Alias[len(ctx.Alias)-1])))

	return nil
}

// Resolve handles any errors
func (*Echo) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}

// EchoFmt defines a simple subcommand of EchoCmd.
type EchoFmt struct{}

// Handle handles the echo subcommand
func (*EchoFmt) Handle(ctx *context.Context) error {

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
func (*EchoFmt) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}

// EchoColor echoes with color!
type EchoColor struct{}

// Handle handles the echo subcommand
func (*EchoColor) Handle(ctx *context.Context) error {

	if len(ctx.Args) == 0 {
		return errors.New("nothing to color echo")
	}

	toTrim := *ctx.Prefix + strings.Join(ctx.Alias, " ")

	codeBlockWrap := "```"

	msgContent := fmt.Sprintf(`
%scss
"%s"
%s 
- %s#%s`,
		codeBlockWrap,
		strings.TrimSpace(strings.TrimPrefix(ctx.Msg.Content, toTrim)),
		codeBlockWrap,
		ctx.Msg.Author.Username,
		ctx.Msg.Author.Discriminator,
	)

	_, _ = ctx.Ses.ChannelMessageSendEmbed(
		ctx.Msg.ChannelID, &discordgo.MessageEmbed{
			Description: msgContent,
		})

	return nil
}

// Resolve handles any errors
func (*EchoColor) Resolve(ctx *context.Context) {
	if ctx.Err != nil {
		_, _ = ctx.Ses.ChannelMessageSend(ctx.Msg.ChannelID, ctx.Err.Error())
	}
}
