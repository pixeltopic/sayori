package main

import (
	"errors"
	"fmt"
	"strings"

	sayori "github.com/pixeltopic/sayori/v2"

	"context"

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
func (*Echo) Handle(ctx context.Context) error {

	cmd := sayori.CmdFromContext(ctx)

	if len(cmd.Args) == 0 {
		return errors.New("nothing to echo")
	}

	_, _ = cmd.Ses.ChannelMessageSend(
		cmd.Msg.ChannelID, "Echoing! "+trimmer(cmd.Msg.Content, cmd.Prefix, cmd.Alias))

	return nil
}

// Resolve handles any errors
func (*Echo) Resolve(ctx context.Context) {
	cmd := sayori.CmdFromContext(ctx)

	if cmd.Err != nil {
		_, _ = cmd.Ses.ChannelMessageSend(cmd.Msg.ChannelID, cmd.Err.Error())
	}
}

// EchoFmt defines a simple subcommand of EchoCmd.
type EchoFmt struct{}

// Handle handles the echo subcommand
func (*EchoFmt) Handle(ctx context.Context) error {

	cmd := sayori.CmdFromContext(ctx)

	if len(cmd.Args) == 0 {
		return errors.New("nothing to format echo")
	}

	_, _ = cmd.Ses.ChannelMessageSendEmbed(
		cmd.Msg.ChannelID, &discordgo.MessageEmbed{
			Description: fmt.Sprintf(`"%s" - %s#%s`,
				trimmer(cmd.Msg.Content, cmd.Prefix, cmd.Alias),
				cmd.Msg.Author.Username,
				cmd.Msg.Author.Discriminator,
			),
		})

	return nil
}

// Resolve handles any errors
func (*EchoFmt) Resolve(ctx context.Context) {
	cmd := sayori.CmdFromContext(ctx)

	if cmd.Err != nil {
		_, _ = cmd.Ses.ChannelMessageSend(cmd.Msg.ChannelID, cmd.Err.Error())
	}
}

// EchoColor echoes with color!
type EchoColor struct{}

// Handle handles the echo subcommand
func (*EchoColor) Handle(ctx context.Context) error {

	cmd := sayori.CmdFromContext(ctx)

	if len(cmd.Args) == 0 {
		return errors.New("nothing to color echo")
	}

	codeBlockWrap := "```"

	msgContent := fmt.Sprintf(`
%scss
"%s"
%s 
- %s#%s`,
		codeBlockWrap,
		trimmer(cmd.Msg.Content, cmd.Prefix, cmd.Alias),
		codeBlockWrap,
		cmd.Msg.Author.Username,
		cmd.Msg.Author.Discriminator,
	)

	_, _ = cmd.Ses.ChannelMessageSendEmbed(
		cmd.Msg.ChannelID, &discordgo.MessageEmbed{
			Description: msgContent,
		})

	return nil
}

// Resolve handles any errors
func (*EchoColor) Resolve(ctx context.Context) {
	cmd := sayori.CmdFromContext(ctx)

	if cmd.Err != nil {
		_, _ = cmd.Ses.ChannelMessageSend(cmd.Msg.ChannelID, cmd.Err.Error())
	}
}
