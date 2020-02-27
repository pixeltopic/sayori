package sayori

// Args is a set of args bound to identifiers that are parsed from the command
type Args map[string]interface{}

// NewArgs makes a new instance of Args for storing key-argument mappings
func NewArgs() Args {
	return make(Args)
}

// Load loads a key from args
func (a Args) Load(key string) (interface{}, bool) {
	v, ok := a[key]
	return v, ok
}

// Store stores a key that maps to val in args
func (a Args) Store(key string, val interface{}) {
	a[key] = val
}

// Delete removes a key that maps to val in args, or if key does not exist, no-op
func (a Args) Delete(key string) {
	delete(a, key)
}
