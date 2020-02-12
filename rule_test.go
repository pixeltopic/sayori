package sayori

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func testNewDgoContext(msgGuildID, msgContent, authorID, selfUserID string, authorBot bool) Context {
	message := &discordgo.Message{
		Author: &discordgo.User{
			ID:  authorID,
			Bot: authorBot,
		},
		Content: msgContent,
		GuildID: msgGuildID,
	}
	state := discordgo.NewState()
	state.User = &discordgo.User{
		ID: selfUserID,
	}
	session := &discordgo.Session{
		State: state,
	}

	ctx := NewContext()
	ctx.Message = message
	ctx.Session = session
	return ctx
}

func TestRule(t *testing.T) {
	var (
		i    uint
		res  bool
		rule *Rule
	)

	t.Run("test new empty rule", func(t *testing.T) {
		rule = NewRule()
		for i = 0; i < 5; i++ {
			ruleID := Rule(1 << i)
			res = rule.HasRule(ruleID)

			if res {
				t.Errorf("empty rule init error, has ID: %d", ruleID)
			}
		}
	})

	t.Run("test new populated rule", func(t *testing.T) {
		rule = NewRule(Rule(1), Rule(4))

		for i = 0; i < 5; i++ {
			ruleID := Rule(1 << i)
			res = rule.HasRule(ruleID)
			switch ruleID {
			case Rule(1):
				fallthrough
			case Rule(4):
				if !res {
					t.Errorf("populated rule init error, does not have ID: %d", ruleID)
				}
			default:
				if res {
					t.Errorf("populated rule init error, has ID: %d", ruleID)
				}
			}
		}
	})

	testAllowResult := func(t *testing.T, expected, got Rule) {
		if expected != got {
			t.Fatalf("incorrect rule; expected=%d got=%d", expected, got)
		}
	}

	t.Run("test allow", func(t *testing.T) {
		var (
			rule    *Rule
			errRule Rule
		)

		rule = NewRule()
		_, errRule = rule.allow(testNewDgoContext("", "my message", "sayoriuser", "sayoribot", false))
		testAllowResult(t, RuleHandlePrivateMsgs, errRule)

		rule = NewRule(RuleHandleBot, RuleHandleEmptyContent)
		_, errRule = rule.allow(testNewDgoContext("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, RuleHandleSelf, errRule)
	})
}
