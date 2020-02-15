package sayori

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

const sessionID = "myid"

type testOnMsg struct {
	HandleCallback func(ctx Context) error
	CatchCallback  func(ctx Context)
	ParseCallback  func(toks Toks) (Args, error)
	t              *testing.T
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

func testMockMessageCreate(content string) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: content,
		},
	}
}

func testEvent(
	t *testing.T,
	parseCallback func(toks Toks) (Args, error),
	handleCallback func(ctx Context) error,
	catchCallback func(ctx Context),
	rule *Rule,
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

	r.makeMsgEvent(event, rule)(mockSession, incomingMockMessage)
}

func TestRouter(t *testing.T) {
	t.Run("test ctx values", func(t *testing.T) {
		const (
			testMessage = `The placeholder text, beginning with the line 
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit", 
	looks like Latin because in its youth, centuries ago, it was Latin.`
			testMessageLen = len(testMessage)
		)

		mockSession := testMockSession(sessionID)
		mockMessageCreate := testMockMessageCreate(testMessage)

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
			t, parseCallback, handleCallback, catchCallback, nil, mockSession, mockMessageCreate)
	})
}
