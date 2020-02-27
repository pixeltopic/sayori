package sayori

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type testPrefixer struct{}

func (*testPrefixer) Load(_ string) (string, bool) {
	return "", true
}

func (*testPrefixer) Default() string {
	return ""
}

type testOnMsg struct {
	HandleCallback func(ctx Context) error
	CatchCallback  func(ctx Context)
	ParseCallback  func(toks Toks) (Args, error)
}

func (m *testOnMsg) Parse(toks Toks) (Args, error) {
	return m.ParseCallback(toks)
}

// Handle will run on a MessageCreate event.
func (m *testOnMsg) Handle(ctx Context) error {
	return m.HandleCallback(ctx)
}

// Catch catches handler errors
func (m *testOnMsg) Catch(ctx Context) {
	m.CatchCallback(ctx)
}

type testCmd struct {
	testOnMsg
	MatchCallback func(toks Toks) (string, bool)
}

// Match identifies the alias of a command. It can support multiple aliases per command.
func (m *testCmd) Match(toks Toks) (string, bool) {
	return m.MatchCallback(toks)
}

// testMockSession returns a fake discordgo Session with the ID of the session user populated
func testMockSession(selfUserID string) *discordgo.Session {
	state := discordgo.NewState()
	state.User = &discordgo.User{
		ID: selfUserID,
	}
	session := &discordgo.Session{
		State: state,
	}

	return session
}

func testMockMessageCreate(authorID string, authorBot bool, msgGuildID, msgContent string) *discordgo.MessageCreate {
	message := &discordgo.Message{
		Author: &discordgo.User{
			ID:  authorID,
			Bot: authorBot,
		},
		Content: msgContent,
		GuildID: msgGuildID,
	}

	return &discordgo.MessageCreate{
		Message: message,
	}
}

// testEvent accepts callbacks that contain tests within and creates an testOnMsg
// instance and creates the wrapper function that will be run on the targeted event.
// mock data will be fed into this wrapper function to ensure it is processed in a valid manner.
func testEvent(
	t *testing.T,
	parseCallback func(toks Toks) (Args, error),
	handleCallback func(ctx Context) error,
	catchCallback func(ctx Context),
	filter Filter,
	mockSession *discordgo.Session,
	incomingMockMessage *discordgo.MessageCreate,
) {
	r := &Router{
		session: mockSession,
		p:       nil,
	}

	event := &testOnMsg{
		ParseCallback:  parseCallback,
		HandleCallback: handleCallback,
		CatchCallback:  catchCallback,
	}

	switch {
	case event.ParseCallback == nil:
		t.Fatal("ParseCallback cannot be nil")
	case event.HandleCallback == nil:
		t.Fatal("HandleCallback cannot be nil")
	case event.CatchCallback == nil:
		t.Fatal("CatchCallback cannot be nil")
	}

	r.makeEvent(event, filter)(mockSession, incomingMockMessage)
}

func testCommand(
	t *testing.T,
	prefixer Prefixer,
	parseCallback func(toks Toks) (Args, error),
	handleCallback func(ctx Context) error,
	catchCallback func(ctx Context),
	matchCallback func(toks Toks) (string, bool),
	filter Filter,
	mockSession *discordgo.Session,
	incomingMockMessage *discordgo.MessageCreate,
) {
	r := &Router{
		session: mockSession,
		p:       prefixer,
	}

	cmd := &testCmd{
		testOnMsg: testOnMsg{
			ParseCallback:  parseCallback,
			HandleCallback: handleCallback,
			CatchCallback:  catchCallback,
		},
		MatchCallback: matchCallback,
	}

	switch {
	case cmd.ParseCallback == nil:
		t.Fatal("ParseCallback cannot be nil")
	case cmd.HandleCallback == nil:
		t.Fatal("HandleCallback cannot be nil")
	case cmd.CatchCallback == nil:
		t.Fatal("CatchCallback cannot be nil")
	case cmd.MatchCallback == nil:
		t.Fatal("MatchCallback cannot be nil")
	}

	r.makeCommand(cmd, filter)(mockSession, incomingMockMessage)

}

func TestRouter(t *testing.T) {
	t.Run("test event handling", func(t *testing.T) {
		const (
			testMessage = `The placeholder text, beginning with the line 
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit", 
	looks like Latin because in its youth, centuries ago, it was Latin.`
			testMessageLen = len(testMessage)
		)

		mockSession := testMockSession("myid")
		mockMessageCreate := testMockMessageCreate(
			"fakeauthorID", false, "someguildID", testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			if ctx.Message.Content != testMessage {
				t.Errorf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
			}
			if ctx.Args == nil {
				t.Error("ctx.Args is nil")
			} else {
				val, ok := ctx.Args.Load("len")
				if !ok {
					t.Error("ctx.Args[len] does not exist")
				}

				stored, _ := val.(int)
				if stored != testMessageLen {
					t.Errorf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
				}
			}

			return errors.New("its a failure oh no")
		}

		catchCallback := func(ctx Context) {
			if ctx.Err == nil {
				t.Errorf("expected ctx.Err to equal '%s', got '%s'", "its a failure oh no", "nil")
			}
		}

		testEvent(
			t, parseCallback, handleCallback, catchCallback, NewFilter(), mockSession, mockMessageCreate)
	})

	t.Run("test command handling", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
		)

		mockSession := testMockSession("myid")
		mockMessageCreate := testMockMessageCreate(
			"fakeauthorID", false, "someguildID", testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			if ctx.Message.Content != testMessage {
				t.Errorf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
			}
			if ctx.Args == nil {
				t.Error("ctx.Args is nil")
			} else {
				val, ok := ctx.Args.Load("len")
				if !ok {
					t.Error("ctx.Args[len] does not exist")
				}

				stored, _ := val.(int)
				if stored != testMessageLen {
					t.Errorf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
				}
			}

			if ctx.Alias != "myaliasecho" {
				t.Errorf("expected ctx.Alias to equal %s, got %s", "myaliasecho", ctx.Alias)
			}

			return errors.New("its a failure oh no")
		}

		catchCallback := func(ctx Context) {
			if ctx.Err == nil {
				t.Errorf("expected ctx.Err to equal '%s', got '%s'", "its a failure oh no", "nil")
			}
		}

		matchCallback := func(toks Toks) (string, bool) {
			if alias, ok := toks.Get(0); ok {
				isMatch := strings.ToLower(alias) == "myaliasecho"
				if !isMatch {
					t.Fatalf("expected alias to be %s, got %s", "myaliasecho", strings.ToLower(alias))
				}

				// we don't actually need to return the alias as all lowercase, it's up to your implementation
				return "myaliasecho", isMatch
			}

			t.Fatal("expected alias to be found in toks")
			return "", false
		}

		testCommand(
			t, &testPrefixer{}, parseCallback, handleCallback,
			catchCallback, matchCallback, NewFilter(), mockSession, mockMessageCreate)

	})

	t.Run("test command handling with filters", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
			alias          = "myaliasEcho"
		)

		mockSession := testMockSession("fakeauthorID")
		mockMessageCreate := testMockMessageCreate(
			"fakeauthorID", true, "someguildID", testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			t.Fatal("expected given handleCallback to never execute due to failed filters")
			return errors.New("oof")
		}

		catchCallback := func(ctx Context) {
			if ctx.Args == nil {
				t.Error("ctx.Args is nil")
			} else {
				val, ok := ctx.Args.Load("len")
				if !ok {
					t.Fatal("ctx.Args[len] does not exist")
				}

				stored, _ := val.(int)
				if stored != testMessageLen {
					t.Errorf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
				}
			}
			if ctx.Alias != alias {
				t.Errorf("expected ctx.Alias to equal %s, got %s", alias, ctx.Alias)
			}
			if ctx.Err == nil {
				t.Errorf("expected ctx.Err to equal '%s', got '%s'", "its a failure oh no", "nil")
			} else {
				filterErr, ok := ctx.Err.(*FilterError)
				if !ok {
					t.Fatal("expected filterErr to implement FilterError")
				}

				if filterErr.Filter() != NewFilter(MessagesBot, MessagesGuild, MessagesSelf) {
					t.Error("incorrect filter failure code")
				}
			}
		}

		matchCallback := func(toks Toks) (string, bool) {
			return alias, true
		}

		testCommand(
			t, &testPrefixer{}, parseCallback, handleCallback,
			catchCallback, matchCallback, NewFilter(MessagesGuild, MessagesBot, MessagesSelf), mockSession, mockMessageCreate)

	})
}
