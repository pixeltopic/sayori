package main

import "github.com/pixeltopic/sayori"

// OnMsg will run on every message event.
type OnMsg struct {
}

// Handle will run on a MessageCreate event.
func (*OnMsg) Handle(ctx sayori.Context) error {
	if ctx.Message.Author.ID == ctx.Session.State.User.ID {
		return nil
	}
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "A message was sent in the channel.")
	return nil
}

// Catch catches handler errors
func (*OnMsg) Catch(ctx sayori.Context) {

}
