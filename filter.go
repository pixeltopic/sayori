package sayori

// Filter represents a condition that prevents a `Command` or `Event` from firing.
// Only use the given const filters which range from 2^0 to 2^4.
type Filter int

// FilterError is an error that has a failing Filter attached
type FilterError struct {
	f      Filter
	reason string
}

// Filter returns the violated filter
func (e *FilterError) Filter() Filter {
	return e.f
}

func (e *FilterError) Error() string {
	return e.reason
}

// Valid Filters
const (
	SelfMessages Filter = 1 << iota
	BotMessages
	EmptyMessages
	PrivateMessages
	GuildMessages
)

// NewFilter generates a Filter bitset given filters and performing a bitwise `or` on all of them
func NewFilter(filters ...Filter) Filter {
	var filter Filter
	for _, f := range filters {
		filter = filter | f
	}
	return filter
}

// Contains returns true if the given filter is part of the current filter
func (f Filter) Contains(filter Filter) bool {
	return f.filters(filter)
}

// filters returns true if the given filter is applied to the current filter
func (f Filter) filters(filter Filter) bool {
	return f&filter == filter
}

// allow inspects context and determines if it should be processed or not.
//
// returns true if allowed with a zero value Filter, or false with all failing Filters combined with a bitwise `or`.
//
// if ctx.Message or ctx.Session is nil, will return false with a zero value Filter.
func (f Filter) allow(ctx Context) (bool, Filter) {
	var failed Filter
	if ctx.Message == nil || ctx.Session == nil {
		return false, failed
	}

	var (
		contentLen = len(ctx.Message.Content)
		guildIDLen = len(ctx.Message.GuildID)
	)

	if f.filters(SelfMessages) {
		switch {
		case ctx.Message.Author == nil:
			fallthrough
		case ctx.Session.State == nil:
			fallthrough
		case ctx.Session.State.User == nil:
			return false, Filter(0)
		case ctx.Message.Author.ID == ctx.Session.State.User.ID:
			failed = failed | SelfMessages
		}
	}
	if f.filters(BotMessages) {
		switch {
		case ctx.Message.Author == nil:
			return false, Filter(0)
		case ctx.Message.Author.Bot:
			failed = failed | BotMessages
		}
	}
	if f.filters(EmptyMessages) && contentLen == 0 {
		failed = failed | EmptyMessages
	}
	if f.filters(PrivateMessages) && guildIDLen == 0 {
		failed = failed | PrivateMessages
	}
	if f.filters(GuildMessages) && guildIDLen != 0 {
		failed = failed | GuildMessages
	}

	return failed == 0, failed
}
