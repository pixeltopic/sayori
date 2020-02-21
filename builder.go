package sayori

// Builder generates the framework of a command.
type Builder struct {
	handler interface{}
	rule    *Rule
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

// WithRule adds a new Rule to the builder.
func (b *Builder) WithRule(r Rule) *Builder {
	if b.rule != nil {
		b.rule = NewRule(*b.rule, r)
	} else {
		b.rule = NewRule(r)
	}
	return b
}
