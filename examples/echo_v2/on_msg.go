package main

import (
	"log"
	"sync"

	v2 "github.com/pixeltopic/sayori/v2"
)

// OnMsg will run on every message event and count
// the number of messages sent and seen by the session.
type OnMsg struct {
	sync.Mutex
	totalSent int
}

// Handle will run on a MessageCreate event.
func (m *OnMsg) Handle(ctx *v2.Context) error {
	m.Lock()
	defer m.Unlock()

	m.totalSent++
	log.Printf("Message count: %d\n", m.totalSent)
	return nil
}

// Resolve catches handler errors
func (*OnMsg) Resolve(ctx *v2.Context) {
}
