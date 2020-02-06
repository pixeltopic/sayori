package main

// Prefixer loads a prefix
type Prefixer struct {
}

// Load returns a prefix based on the serverID
func (p *Prefixer) Load(serverID string) (string, bool) {
	return p.Default(), true
}

// Default returns the default router prefix
func (*Prefixer) Default() string {
	return "e!"
}
