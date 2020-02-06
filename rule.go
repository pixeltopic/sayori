package sayori

// Rule defines when an event handler should be run.
type Rule int

// RuleError is an error that has a failing rule attached
type RuleError struct {
	rule   Rule
	reason string
}

// Rule returns the offending rule
func (e *RuleError) Rule() Rule {
	return e.rule
}

func (e *RuleError) Error() string {
	return e.reason
}

// Rule definitions for handler execution
const (
	RuleHandleSelf Rule = 1 << iota
	RuleHandleBot
	RuleHandleEmptyContent
	RuleHandlePrivateMsgs
	RuleHandleGuildMsgs
)

// NewRule generates an Rule bitset given Rules and performing a bitwise `or` on all of them
func NewRule(rules ...Rule) Rule {
	var rule Rule
	for _, r := range rules {
		rule = rule | r
	}
	return rule
}

// HasRule checks if a rule is enabled
func (e Rule) HasRule(rule Rule) bool {
	return e&rule == rule
}

// allow inspects context and determines if it should be processed or not.
//
// returns true if allowed with an empty Rule, or false with the offending Rule.
//
// if ctx.Message or ctx.Session is nil, will return false with Rule value of 0
func (e Rule) allow(ctx Context) (bool, Rule) {
	if ctx.Message == nil || ctx.Session == nil {
		return false, NewRule()
	}

	var (
		msgAuthorID = ctx.Message.Author.ID
		selfUserID  = ctx.Session.State.User.ID
		authorIsBot = ctx.Message.Author.Bot
		contentLen  = len(ctx.Message.Content)
		guildIDLen  = len(ctx.Message.GuildID)
	)

	if !e.HasRule(RuleHandleSelf) && msgAuthorID == selfUserID {
		return false, RuleHandleSelf
	}
	if !e.HasRule(RuleHandleBot) && authorIsBot {
		return false, RuleHandleBot
	}
	if !e.HasRule(RuleHandleEmptyContent) && contentLen == 0 {
		return false, RuleHandleEmptyContent
	}
	if !e.HasRule(RuleHandlePrivateMsgs) && guildIDLen == 0 {
		return false, RuleHandlePrivateMsgs
	}
	if !e.HasRule(RuleHandleGuildMsgs) && guildIDLen != 0 {
		return false, RuleHandleGuildMsgs
	}
	return true, NewRule()
}
