package sayori

// EventHandlerRule defines when an event handler should be run.
type EventHandlerRule int

// Rule definitions for handler execution
const (
	RuleHandleSelf EventHandlerRule = 1 << iota
	RuleHandleBot
	RuleHandleEmptyContent
	RuleHandlePrivateMsgs
	RuleHandleGuildMsgs
)

// NewEventHandlerRule generates an EventHandlerRule bitset given Rules
func NewEventHandlerRule(rules ...EventHandlerRule) EventHandlerRule {
	var rule EventHandlerRule
	for _, r := range rules {
		rule = rule | r
	}
	return rule
}

// IsRuleEnabled checks if a rule is enabled
func (e EventHandlerRule) IsRuleEnabled(rule EventHandlerRule) bool {
	return e&rule == rule
}

// Allow returns true if allowed or false with a failing EventHandlerRule
func (e EventHandlerRule) Allow(authorID, userID string, authorIsBot bool, contentLen, guildIDLen int) (bool, EventHandlerRule) {
	if !e.IsRuleEnabled(RuleHandleSelf) && authorID == userID {
		return false, RuleHandleSelf
	}
	if !e.IsRuleEnabled(RuleHandleBot) && authorIsBot {
		return false, RuleHandleBot
	}
	if !e.IsRuleEnabled(RuleHandleEmptyContent) && contentLen == 0 {
		return false, RuleHandleEmptyContent
	}
	if !e.IsRuleEnabled(RuleHandlePrivateMsgs) && guildIDLen == 0 {
		return false, RuleHandlePrivateMsgs
	}
	if !e.IsRuleEnabled(RuleHandleGuildMsgs) && guildIDLen != 0 {
		return false, RuleHandleGuildMsgs
	}
	return true, EventHandlerRule(0)
}
