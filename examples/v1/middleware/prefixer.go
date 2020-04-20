package main

const defaultPrefix = "."

// Prefixer loads a prefix
type Prefixer struct{}

// Load returns the Default prefix no matter what the guildID given is.
func (p *Prefixer) Load(_ string) (string, bool) {
	return p.Default(), true
}

// Default returns the default router prefix
func (*Prefixer) Default() string {
	return defaultPrefix
}
