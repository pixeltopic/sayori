package v2

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/pixeltopic/sayori/v2/context"
)

const testDefaultPrefix = "t!"

type (
	// testParams helps bootstrap table driven tests.
	testParams struct {
		testIOParams []*testIOParams
		routeParams  *testRouteDefns
	}

	// testIOParams contains test input-expected outputs.
	testIOParams struct {
		sesParams           *mockSesParams
		msgParams           *mockMsgParams
		msgContentTokenized []string // msgContent in mockMsgParams, but tokenized

		expectedDepth     int // depth of a subcommand token (zero indexed)
		expectedPrefix    string
		expectedAliasTree []string // represents all aliases of the root command and sub command
		expectedAlias     []string // order sensitive. alias trace generated from the command invocation
		expectedArgs      []string // order sensitive. args generated from the command invocation
		expectedErr       error    // error the ctx should contain
	}

	// testRouteDefns contains definitions to easily bootstrap route setup
	testRouteDefns struct {
		c *testCmd
		p Prefixer

		aliases     []string
		subroutes   []*testRouteDefns
		middlewares []Middlewarer
	}

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
func (p *testIOParams) createMockSes() (*discordgo.Session, error) {
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

func (p *testIOParams) createMockMsg() (*discordgo.MessageCreate, error) {
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

func testCreateRouteHelper(root *testRouteDefns) *Route {
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

// TODO: table driven test where messages are in a slice and we have one big testRouteDefns to test?
// In addition we can test multiple testRouteDefns too
// Need to find a good way to store expected test results
func TestRoute(t *testing.T) {
	testTrees := []*testParams{
		{
			testIOParams: []*testIOParams{
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root sub1 sub2 sub3 arg1 arg2",
					},
					msgContentTokenized: []string{"root", "sub1", "sub2", "arg3", "arg1", "arg2"},
					expectedDepth:       2,
				},
			},
			routeParams: &testRouteDefns{
				c: nil, p: nil, aliases: []string{"root"},
				middlewares: nil,
				subroutes: []*testRouteDefns{
					{
						c: nil, p: nil, aliases: []string{"sub1"},
						middlewares: nil,
						subroutes: []*testRouteDefns{
							{
								c: nil, p: nil, aliases: []string{"subsub1"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
						},
					},
					{
						c: nil, p: nil, aliases: []string{"sub2"},
						middlewares: nil,
						subroutes:   []*testRouteDefns{},
					},
				},
			},
		},
	}

	for _, tt := range testTrees {
		rr := tt.createRoute()

		for _, io := range tt.testIOParams {
			t.Run("test findRoute", func(t *testing.T) {
				found, depth := findRoute(rr, io.msgContentTokenized)
				if depth != io.expectedDepth {
					t.Errorf("got %d, want %d", depth, io.expectedDepth)
				}
				if found == nil {
					t.Error("expected non-nil route")
				}
			})
		}
	}
	t.Run("generate a route from params", func(t *testing.T) {

		// do tests here after generating a route from testParams.RouteParams
		//r := tree.createRoute()
		//
		//expected := testGetAllAliasRecursively(r)
		//
		//if len(expected) != 4 {
		//	t.FailNow()
		//}
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
