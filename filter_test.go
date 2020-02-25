package sayori

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func testNewCtx(msgGuildID, msgContent, authorID, selfUserID string, authorBot bool) Context {
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

func testNewBadCtx(s *discordgo.Session, m *discordgo.Message) Context {
	ctx := NewContext()
	ctx.Message = m
	ctx.Session = s
	return ctx
}

func TestFilter(t *testing.T) {
	var (
		i      uint
		res    bool
		filter Filter
	)

	t.Run("test new empty filter", func(t *testing.T) {
		filter = NewFilter()
		for i = 0; i < 5; i++ {
			fID := Filter(1 << i)
			res = filter.Contains(fID)

			if res {
				t.Errorf("empty rule init error, has ID: %d", fID)
			}
		}
	})

	t.Run("test Filter that filters out 1 << 0 and 1 << 2", func(t *testing.T) {
		filter = NewFilter(Filter(1), Filter(4))

		for i = 0; i < 5; i++ {
			fID := Filter(1 << i)
			res = filter.Contains(fID)
			switch fID {
			case Filter(1):
				fallthrough
			case Filter(4):
				if !res {
					t.Errorf("populated Filter init error, does not have ID: %d", fID)
				}
			default:
				if res {
					t.Errorf("populated Filter init error, has ID: %d", fID)
				}
			}
		}
	})

	testAllowResult := func(t *testing.T, expected, got Filter) {
		if expected != got {
			t.Fatalf("incorrect Filter; expected=%d got=%d", expected, got)
		}
	}

	t.Run("test allow", func(t *testing.T) {
		var (
			filter, errFilter Filter
			ctx               Context
		)

		filter = NewFilter()
		_, errFilter = filter.allow(testNewCtx("", "my message", "sayoriuser", "sayoribot", false))
		testAllowResult(t, Filter(0), errFilter)

		// filter out bot messages and empty messages
		filter = NewFilter(BotMessages, EmptyMessages)
		_, errFilter = filter.allow(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, NewFilter(BotMessages, EmptyMessages), errFilter)

		// Author is nil
		state := discordgo.NewState()
		state.User = &discordgo.User{}

		ctx = testNewBadCtx(
			&discordgo.Session{State: state}, &discordgo.Message{Author: nil})

		filter = NewFilter(SelfMessages)
		_, errFilter = filter.allow(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// State is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: nil}, &discordgo.Message{Author: &discordgo.User{}})

		filter = NewFilter(SelfMessages)
		_, errFilter = filter.allow(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// User in State is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: discordgo.NewState()}, &discordgo.Message{Author: &discordgo.User{}})

		filter = NewFilter(SelfMessages)
		_, errFilter = filter.allow(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// filter out self messages
		filter = NewFilter(SelfMessages)
		_, errFilter = filter.allow(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, NewFilter(SelfMessages), errFilter)

		// Author in Message is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: discordgo.NewState()}, &discordgo.Message{Author: nil})

		filter = NewFilter(BotMessages)
		_, errFilter = filter.allow(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// filter out bot messages
		filter = NewFilter(BotMessages)
		_, errFilter = filter.allow(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, NewFilter(BotMessages), errFilter)

	})
}
