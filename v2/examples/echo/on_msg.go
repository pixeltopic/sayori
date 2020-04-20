package main

import (
	"log"
	"sync"

	"github.com/pixeltopic/sayori/v2/context"
)

// OnMsg will run on every message event and count
// the number of messages sent and seen by the session.
type OnMsg struct {
	sync.Mutex
	totalSent int
}

// Handle will run on a MessageCreate event.
func (m *OnMsg) Handle(_ *context.Context) error {
	m.Lock()
	defer m.Unlock()

	m.totalSent++
	log.Printf("Message count: %d\n", m.totalSent)
	return nil
}

// Resolve catches handler errors
func (*OnMsg) Resolve(_ *context.Context) {
}
