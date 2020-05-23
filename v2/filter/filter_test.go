package filter

import (
	sayori "github.com/pixeltopic/sayori/v2"
	"testing"

	"context"

	"github.com/bwmarrin/discordgo"
)

func testNewCtx(msgGuildID, msgContent, authorID, selfUserID string, authorBot bool) context.Context {
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

	ctx := context.Background()
	return sayori.WithSes(sayori.WithMsg(ctx, message), session)
}

func testNewBadCtx(s *discordgo.Session, m *discordgo.Message) context.Context {
	ctx := context.Background()
	return sayori.WithSes(sayori.WithMsg(ctx, m),s)
}

func TestFilter(t *testing.T) {
	var (
		i      uint
		res    bool
		filter Filter
	)

	t.Run("test new empty filter", func(t *testing.T) {
		filter = New()
		for i = 0; i < 5; i++ {
			fID := Filter(1 << i)
			res = filter.Contains(fID)

			if res {
				t.Errorf("empty rule init error, has ID: %d", fID)
			}
		}
	})

	t.Run("test Filter that filters out 1 << 0 and 1 << 2", func(t *testing.T) {
		filter = New(Filter(1), Filter(4))

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
			ctx               context.Context
			ok                bool
		)

		filter = New()
		_, errFilter = filter.Validate(testNewCtx("", "my message", "sayoriuser", "sayoribot", false))
		testAllowResult(t, Filter(0), errFilter)

		// filter out bot messages and empty messages
		filter = New(MsgFromBot, MsgNoContent)
		_, errFilter = filter.Validate(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, New(MsgFromBot, MsgNoContent), errFilter)

		// should pass the filter
		filter = New(MsgFromWebhook)
		ok, errFilter = filter.Validate(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, New(0), errFilter)
		if !ok {
			t.Fatal("incorrect Filter; expected true but got false")
		}

		// Author is nil
		state := discordgo.NewState()
		state.User = &discordgo.User{}

		ctx = testNewBadCtx(
			&discordgo.Session{State: state}, &discordgo.Message{Author: nil})

		filter = New(MsgFromSelf)
		_, errFilter = filter.Validate(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// State is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: nil}, &discordgo.Message{Author: &discordgo.User{}})

		filter = New(MsgFromSelf)
		_, errFilter = filter.Validate(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// User in State is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: discordgo.NewState()}, &discordgo.Message{Author: &discordgo.User{}})

		filter = New(MsgFromSelf)
		_, errFilter = filter.Validate(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// filter out self messages
		filter = New(MsgFromSelf)
		_, errFilter = filter.Validate(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, New(MsgFromSelf), errFilter)

		// Author in Message is nil
		ctx = testNewBadCtx(
			&discordgo.Session{State: discordgo.NewState()}, &discordgo.Message{Author: nil})

		filter = New(MsgFromBot)
		_, errFilter = filter.Validate(ctx)
		testAllowResult(t, Filter(0), errFilter)

		// filter out bot messages
		filter = New(MsgFromBot)
		_, errFilter = filter.Validate(testNewCtx("myguildid", "", "sayoribot", "sayoribot", true))
		testAllowResult(t, New(MsgFromBot), errFilter)

	})
}
