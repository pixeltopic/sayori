package main

import (
	"context"
	"log"
	"sync"

	"github.com/pixeltopic/sayori/v2/utils"
)

// OnMsg will run on every message event and count
// the number of messages sent and seen by the session.
type OnMsg struct {
	sync.Mutex
	totalSent int
}

// Handle will run on a MessageCreate event.
func (m *OnMsg) Handle(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()

	m.totalSent++
	log.Printf("Message count: %d, args: %v\n", m.totalSent, utils.GetArgs(ctx))
	return nil
}
