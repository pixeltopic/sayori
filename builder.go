package sayori

// Builder generates the framework of a command.
type Builder struct {
	handler interface{}
	filter  Filter
}

// command binds a `Command` implementation to the builder.
func (b *Builder) command(c Command) *Builder {
	b.handler = c
	return b
}

// event binds an `Event` implementation to the builder.
func (b *Builder) event(e Event) *Builder {
	b.handler = e
	return b
}

// Filter prevents the specified condition from firing the Command or Event
func (b *Builder) Filter(f Filter) *Builder {
	b.filter = NewFilter(b.filter, f)
	return b
}
