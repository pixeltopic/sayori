package v2

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/pixeltopic/sayori/v2/context"
)

const testDefaultPrefix = "t!"

type (
	// testPrefEmpty defaults the prefix to an empty string.
	testPrefEmpty struct{}

	// testPref defaults the prefix to an empty string.
	testPref struct{}

	// testCmd is a test type. The callbacks are able to access the stdlib test scope so we can easily test
	testCmd struct {
		HandleCallback  func(*context.Context) error
		ResolveCallback func(*context.Context)
		ParseCallback   func(string) ([]string, error)
	}

	mockSesParams struct {
		selfUserID string
	}

	mockMsgParams struct {
		authorBot                        bool
		authorID, msgGuildID, msgContent string
	}

	// testRouteParams helps bootstrap a Route for testing.
	testRouteParams struct {
		c *testCmd
		p Prefixer

		aliases     []string
		subroutes   []*testRouteParams
		middlewares []Middlewarer
	}

	// testParams helps bootstrap table driven tests
	testParams struct {
		sesParams *mockSesParams
		msgParams *mockMsgParams

		routeParams *testRouteParams
	}
)

func (p *testPrefEmpty) Load(_ string) (string, bool) { return p.Default(), true }

func (*testPrefEmpty) Default() string { return "" }

func (p *testPref) Load(_ string) (string, bool) { return p.Default(), false }

func (*testPref) Default() string { return testDefaultPrefix }

func (c *testCmd) Handle(ctx *context.Context) error { return c.HandleCallback(ctx) }

func (c *testCmd) Resolve(ctx *context.Context) { c.ResolveCallback(ctx) }

// Parse parses a command. If ParseCallback is nil, will default to strings.Fields
func (c *testCmd) Parse(cmd string) ([]string, error) {
	if c.ParseCallback == nil {
		return cmdParserDefault(cmd), nil
	}
	return c.ParseCallback(cmd)
}

// testMockSes returns a fake discordgo Session with the ID of the session user populated
func (p *testParams) createMockSes() (*discordgo.Session, error) {
	if p.sesParams == nil {
		return nil, errors.New("p.sesParams is nil")
	}

	state := discordgo.NewState()
	state.User = &discordgo.User{
		ID: p.sesParams.selfUserID,
	}
	session := &discordgo.Session{
		State: state,
	}

	return session, nil
}

func (p *testParams) createMockMsg() (*discordgo.MessageCreate, error) {
	if p.msgParams == nil {
		return nil, errors.New("p.msgParams is nil")
	}
	message := &discordgo.Message{
		Author: &discordgo.User{
			ID:  p.msgParams.authorID,
			Bot: p.msgParams.authorBot,
		},
		Content: p.msgParams.msgContent,
		GuildID: p.msgParams.msgGuildID,
	}

	return &discordgo.MessageCreate{
		Message: message,
	}, nil
}

func testCreateRouteHelper(root *testRouteParams) *Route {
	r := NewRoute(root.p)

	r.On(root.aliases...)
	r.Use(root.middlewares...)
	r.Do(root.c)

	for _, sr := range root.subroutes {
		r.Has(testCreateRouteHelper(sr))
	}

	return r

}

// createRoute creates a route and ALL its subroutes recursively
func (p *testParams) createRoute() *Route {
	return testCreateRouteHelper(p.routeParams)
}

func testGetAllAliasRecursively(route *Route) []string {
	var alias []string
	alias = append(alias, route.aliases...)
	for _, sr := range route.subroutes {
		alias = append(alias, testGetAllAliasRecursively(sr)...)
	}
	return alias
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestRoute(t *testing.T) {

	t.Run("generate a route from params", func(t *testing.T) {
		tree := &testParams{
			sesParams: nil,
			msgParams: nil,
			routeParams: &testRouteParams{
				c: nil, p: nil, aliases: []string{"root"},
				middlewares: nil,
				subroutes: []*testRouteParams{
					{
						c: nil, p: nil, aliases: []string{"sub1"},
						middlewares: nil,
						subroutes: []*testRouteParams{
							{
								c: nil, p: nil, aliases: []string{"subsub1"},
								middlewares: nil,
								subroutes:   []*testRouteParams{},
							},
						},
					},
					{
						c: nil, p: nil, aliases: []string{"sub2"},
						middlewares: nil,
						subroutes:   []*testRouteParams{},
					},
				},
			},
		}

		// do tests here after generating a route from testParams.RouteParams
		r := tree.createRoute()

		expected := testGetAllAliasRecursively(r)

		if len(expected) != 4 {
			t.FailNow()
		}
		//if !r.HasAlias("root") {
		//	t.FailNow()
		//}
		//if subr := r.Find("sub1"); subr != nil {
		//	if subr.Find("subsub1") == nil {
		//		t.FailNow()
		//	}
		//} else {
		//	t.FailNow()
		//}
		//if r.Find("sub2") == nil {
		//	t.FailNow()
		//}
	})
}
