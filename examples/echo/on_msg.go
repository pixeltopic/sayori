package main

import (
	"log"
	"sync"

	"github.com/pixeltopic/sayori"
)

// OnMsg will run on every message event and count
// the number of messages sent and seen by the session.
type OnMsg struct {
	sync.Mutex
	totalSent int
}

// Handle will run on a MessageCreate event.
func (m *OnMsg) Handle(ctx sayori.Context) error {
	m.Lock()
	defer m.Unlock()

	m.totalSent++
	log.Printf("Message count: %d\n", m.totalSent)
	return nil
}

// Resolve catches handler errors
func (*OnMsg) Resolve(ctx sayori.Context) {
	if ctx.Err == nil {
		return
	}
	switch e := ctx.Err.(type) {
	case *sayori.FilterError:
		if e.Filter().Contains(sayori.MessagesSelf) {
			log.Printf("Filter failed for OnMsg; id:'%d'", e.Filter())
		} else {
			ctx.Session.ChannelMessageSend(
				ctx.Message.ChannelID, e.Error())
		}

	default:
		ctx.Session.ChannelMessageSend(
			ctx.Message.ChannelID, e.Error())
	}

}
