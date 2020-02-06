package main

import (
	"log"

	"github.com/pixeltopic/sayori"
)

// OnMsg will run on every message event.
type OnMsg struct {
}

// Handle will run on a MessageCreate event.
func (*OnMsg) Handle(ctx sayori.Context) error {
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "A message was sent in the channel.")
	return nil
}

// Catch catches handler errors
func (*OnMsg) Catch(ctx sayori.Context) {
	if ctx.Err == nil {
		return
	}
	switch e := ctx.Err.(type) {
	case *sayori.RuleError:
		if e.Rule() != sayori.RuleHandleSelf {
			ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, e.Error())
		} else {
			log.Println("Not sending a message because I'd loop forever!")
		}
	default:
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, e.Error())
	}

}
