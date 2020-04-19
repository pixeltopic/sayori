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
	HandleCallback  func(ctx Context) error
	ResolveCallback func(ctx Context)
	ParseCallback   func(toks Toks) (Args, error)
}

func (m *testOnMsg) Parse(toks Toks) (Args, error) {
	return m.ParseCallback(toks)
}

// Handle will run on a MessageCreate event.
func (m *testOnMsg) Handle(ctx Context) error {
	return m.HandleCallback(ctx)
}

// Resolve catches handler errors
func (m *testOnMsg) Resolve(ctx Context) {
	m.ResolveCallback(ctx)
}

type testMiddleware struct{}

func (m *testMiddleware) Do(_ Context) error {
	return nil
}

type testMiddleware2 struct{}

func (m *testMiddleware2) Do(_ Context) error {
	return errors.New("failing middleware")
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
	resolveCallback func(ctx Context),
	filter Filter,
	mockSession *discordgo.Session,
	incomingMockMessage *discordgo.MessageCreate,
	middlewares []Middleware,
) {
	r := &Router{
		Session: mockSession,
		p:       nil,
	}

	event := &testOnMsg{
		ParseCallback:   parseCallback,
		HandleCallback:  handleCallback,
		ResolveCallback: resolveCallback,
	}

	switch {
	case event.ParseCallback == nil:
		t.Fatal("ParseCallback cannot be nil")
	case event.HandleCallback == nil:
		t.Fatal("HandleCallback cannot be nil")
	case event.ResolveCallback == nil:
		t.Fatal("ResolveCallback cannot be nil")
	}

	r.makeEvent(event, filter, middlewares)(mockSession, incomingMockMessage)
}

func testCommand(
	t *testing.T,
	prefixer Prefixer,
	parseCallback func(toks Toks) (Args, error),
	handleCallback func(ctx Context) error,
	resolveCallback func(ctx Context),
	matchCallback func(toks Toks) (string, bool),
	filter Filter,
	mockSession *discordgo.Session,
	incomingMockMessage *discordgo.MessageCreate,
	middlewares []Middleware,
) {
	r := &Router{
		Session: mockSession,
		p:       prefixer,
	}

	cmd := &testCmd{
		testOnMsg: testOnMsg{
			ParseCallback:   parseCallback,
			HandleCallback:  handleCallback,
			ResolveCallback: resolveCallback,
		},
		MatchCallback: matchCallback,
	}

	switch {
	case cmd.ParseCallback == nil:
		t.Fatal("ParseCallback cannot be nil")
	case cmd.HandleCallback == nil:
		t.Fatal("HandleCallback cannot be nil")
	case cmd.ResolveCallback == nil:
		t.Fatal("ResolveCallback cannot be nil")
	case cmd.MatchCallback == nil:
		t.Fatal("MatchCallback cannot be nil")
	}

	r.makeCommand(cmd, filter, middlewares)(mockSession, incomingMockMessage)

}

func TestRouter(t *testing.T) {
	t.Run("test event handling", func(t *testing.T) {
		const (
			testMessage = `The placeholder text, beginning with the line 
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit", 
	looks like Latin because in its youth, centuries ago, it was Latin.`
			testMessageLen = len(testMessage)
			sessionID      = "myID"
			authorID       = "fakeauthorID"
			guildID        = "someguildID"
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(authorID, false, guildID, testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			if ctx.Message.Content != testMessage {
				t.Fatalf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
			}
			if ctx.Args == nil {
				t.Fatal("ctx.Args is nil")
			}
			val, ok := ctx.Args.Load("len")
			if !ok {
				t.Fatal("ctx.Args[len] does not exist")
			}

			stored, _ := val.(int)
			if stored != testMessageLen {
				t.Fatalf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
			}

			return errors.New("its a failure oh no")
		}

		resolveCallback := func(ctx Context) {
			if ctx.Err == nil {
				t.Fatalf("expected ctx.Err to be non-nil")
			}
		}

		testEvent(
			t, parseCallback, handleCallback, resolveCallback, NewFilter(), mockSession, mockMessageCreate, nil)
	})

	t.Run("test command handling", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
			sessionID      = "myID"
			authorID       = "fakeauthorID"
			guildID        = "someguildID"
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(
			authorID, false, guildID, testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			if ctx.Message.Content != testMessage {
				t.Fatalf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
			}
			if ctx.Args == nil {
				t.Fatal("ctx.Args is nil")
			}
			val, ok := ctx.Args.Load("len")
			if !ok {
				t.Fatal("ctx.Args[len] does not exist")
			}

			stored, _ := val.(int)
			if stored != testMessageLen {
				t.Fatalf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
			}

			if ctx.Alias != "myaliasecho" {
				t.Fatalf("expected ctx.Alias to equal %s, got %s", "myaliasecho", ctx.Alias)
			}

			return errors.New("its a failure oh no")
		}

		resolveCallback := func(ctx Context) {
			if ctx.Err == nil {
				t.Fatalf("expected ctx.Err to be non-nil")
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
			resolveCallback, matchCallback, NewFilter(), mockSession, mockMessageCreate, nil)

	})

	t.Run("test event & command handling with middlewares", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
			sessionID      = "myID"
			authorID       = "fakeauthorID"
			guildID        = "someguildID"
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(
			authorID, false, guildID, testMessage)

		parseCallback := func(toks Toks) (Args, error) { return nil, nil }

		handleCallback := func(ctx Context) error {
			if ctx.Message.Content != testMessage {
				t.Fatalf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
			}

			if ctx.Alias != "myaliasecho" {
				t.Fatalf("expected ctx.Alias to equal %s, got %s", "myaliasecho", ctx.Alias)
			}

			return nil
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

		t.Run("test a middleware for an event which passes", func(t *testing.T) {
			handleCallback := func(ctx Context) error {
				if ctx.Message.Content != testMessage {
					t.Fatalf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
				}

				if ctx.Alias != "" {
					t.Fatalf("expected ctx.Alias to be empty, got '%s'", ctx.Alias)
				}

				return nil
			}

			resolveCallback := func(ctx Context) {
				if ctx.Err != nil {
					t.Fatalf("expected ctx.Err to be nil")
				}
			}

			testEvent(
				t, parseCallback, handleCallback,
				resolveCallback, NewFilter(), mockSession, mockMessageCreate, []Middleware{&testMiddleware{}})
		})

		t.Run("test a middleware for an event which fails midway", func(t *testing.T) {
			handleCallback := func(ctx Context) error {
				if ctx.Message.Content != testMessage {
					t.Fatalf("expected ctx.Message.Content to equal '%s', got '%s'", testMessage, ctx.Message.Content)
				}

				if ctx.Alias != "" {
					t.Fatalf("expected ctx.Alias to be empty, got '%s'", ctx.Alias)
				}

				return nil
			}

			resolveCallback := func(ctx Context) {
				if ctx.Err == nil {
					t.Fatalf("expected ctx.Err to be non-nil")
				}
			}

			testEvent(
				t, parseCallback, handleCallback,
				resolveCallback, NewFilter(), mockSession, mockMessageCreate, []Middleware{&testMiddleware2{}, &testMiddleware{}})
		})

		t.Run("test a middleware which passes", func(t *testing.T) {
			resolveCallback := func(ctx Context) {
				if ctx.Err != nil {
					t.Fatalf("expected ctx.Err to be nil")
				}
			}

			testCommand(
				t, &testPrefixer{}, parseCallback, handleCallback,
				resolveCallback, matchCallback, NewFilter(), mockSession, mockMessageCreate, []Middleware{&testMiddleware{}})
		})

		t.Run("test a middleware which fails midway", func(t *testing.T) {
			resolveCallback := func(ctx Context) {
				if ctx.Err == nil {
					t.Fatalf("expected ctx.Err to be non-nil")
				}
			}

			testCommand(
				t, &testPrefixer{}, parseCallback, handleCallback,
				resolveCallback, matchCallback, NewFilter(), mockSession, mockMessageCreate, []Middleware{&testMiddleware2{}, &testMiddleware{}})
		})

	})

	t.Run("test command handling with filters", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
			alias          = "myaliasEcho"
			sessionID      = "fakeauthorID"
			authorID       = sessionID
			guildID        = "someguildID"
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(authorID, true, guildID, testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			args := NewArgs()
			args.Store("len", testMessageLen)
			return args, nil
		}

		handleCallback := func(ctx Context) error {
			t.Fatal("expected given handleCallback to never execute due to failed filters")
			return nil
		}

		resolveCallback := func(ctx Context) {
			// test ctx.Args
			if ctx.Args == nil {
				t.Fatal("ctx.Args is nil")
			}
			val, ok := ctx.Args.Load("len")
			if !ok {
				t.Fatal("ctx.Args[len] does not exist")
			}
			stored, _ := val.(int)
			if stored != testMessageLen {
				t.Fatalf("expected ctx.Args[len] to equal '%d', got '%d'", testMessageLen, stored)
			}

			// test ctx.Alias
			if ctx.Alias != alias {
				t.Fatalf("expected ctx.Alias to equal %s, got %s", alias, ctx.Alias)
			}

			// test ctx.Err for the proper Filter error
			if ctx.Err == nil {
				t.Fatal("expected ctx.Err to be non-nil")
			}
			filterErr, ok := ctx.Err.(*FilterError)
			if !ok {
				t.Fatal("expected filterErr to implement FilterError")
			}

			if filterErr.Filter() != NewFilter(MessagesGuild, MessagesSelf) {
				t.Fatalf("incorrect filter failure code; expected %d but got %d",
					NewFilter(MessagesGuild, MessagesSelf), filterErr.Filter())
			}

		}

		matchCallback := func(toks Toks) (string, bool) {
			return alias, true
		}

		testCommand(
			t, &testPrefixer{}, parseCallback, handleCallback,
			resolveCallback, matchCallback, NewFilter(MessagesGuild, MessagesSelf), mockSession, mockMessageCreate, nil)

	})

	t.Run("test command handling with filters and middlewares", func(t *testing.T) {
		const (
			testMessage    = `myaliasEcho echo me please! This is a friendly message :)`
			testMessageLen = len(testMessage)
			alias          = "myaliasEcho"
			sessionID      = "fakeauthorID"
			authorID       = "anotherauthorID"
			guildID        = "someguildID"
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(authorID, true, guildID, testMessage)

		parseCallback := func(toks Toks) (Args, error) {
			return nil, nil
		}

		handleCallback := func(ctx Context) error {
			t.Fatal("expected given handleCallback to never execute due to failed middleware")
			return nil
		}

		resolveCallback := func(ctx Context) {
			// test ctx.Alias
			if ctx.Alias != alias {
				t.Fatalf("expected ctx.Alias to equal %s, got %s", alias, ctx.Alias)
			}

			// test ctx.Err for the proper Filter error
			if ctx.Err == nil {
				t.Fatal("expected ctx.Err to be non-nil")
			}

			if ctx.Err.Error() != "failing middleware" {
				t.Fatalf("expected ctx.Err to be 'failing middleware', got '%s'", ctx.Err.Error())
			}

		}

		matchCallback := func(toks Toks) (string, bool) {
			return alias, true
		}

		testCommand(
			t, &testPrefixer{}, parseCallback, handleCallback,
			resolveCallback, matchCallback, NewFilter(MessagesSelf), mockSession, mockMessageCreate,
			[]Middleware{&testMiddleware{}, &testMiddleware2{}})

	})
}
