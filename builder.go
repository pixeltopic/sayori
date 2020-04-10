package sayori

// Builder generates the framework of a command.
type Builder struct {
	handler     interface{}
	filter      Filter
	middlewares []Middleware
}

// NewBuilder returns a new Command or Event builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Command binds a `Command` implementation to the builder.
// Only one Command can be built at once.
func (b *Builder) Command(c Command) *Builder {
	b.handler = c
	return b
}

// Event binds an `Event` implementation to the builder.
// Only one Event can be built at once.
func (b *Builder) Event(e Event) *Builder {
	b.handler = e
	return b
}

// Filter prevents sent messages from firing the Command or Event by filtering out whatever meets the criteria
func (b *Builder) Filter(f Filter) *Builder {
	b.filter = NewFilter(b.filter, f)
	return b
}

// Use a custom middleware. Middlewares are executed in the order they are chained via Builder.
// Middlewares run AFTER all Filters are run.
func (b *Builder) Use(m Middleware) *Builder {
	b.middlewares = append(b.middlewares, m)
	return b
}
