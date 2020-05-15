package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pixeltopic/sayori/v2/context"

	"github.com/bwmarrin/discordgo"
)

// trimmer removes the prefix from a message and removes quotes if present
func trimmer(message, prefix string, alias []string) string {

	message = strings.Replace(message, prefix, "", 1)
	for _, a := range alias {
		message = strings.Replace(message, a, "", 1)
	}

	s := strings.TrimSpace(message)

	if len(s) >= 2 &&
		strings.HasPrefix(s, "\"") &&
		strings.HasSuffix(s, "\"") {

		return strings.TrimPrefix(strings.TrimSuffix(s, "\""), "\"")
	}

	return s

}

// Echo defines a simple EchoCmd.
type Echo struct{}

// Handle handles the echo command
func (*Echo) Handle(ctx *context.Context) error {

	if len(ctx.Args) == 0 {
		return errors.New("nothing to echo")
	}

	_, _ = ctx.Ses.ChannelMessageSend(
		ctx.Msg.ChannelID, "Echoing! "+trimmer(ctx.Msg.Content, ctx.Prefix, ctx.Alias))

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

	_, _ = ctx.Ses.ChannelMessageSendEmbed(
		ctx.Msg.ChannelID, &discordgo.MessageEmbed{
			Description: fmt.Sprintf(`"%s" - %s#%s`,
				trimmer(ctx.Msg.Content, ctx.Prefix, ctx.Alias),
				ctx.Msg.Author.Username,
				ctx.Msg.Author.Discriminator,
			),
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

	codeBlockWrap := "```"

	msgContent := fmt.Sprintf(`
%scss
"%s"
%s 
- %s#%s`,
		codeBlockWrap,
		trimmer(ctx.Msg.Content, ctx.Prefix, ctx.Alias),
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
