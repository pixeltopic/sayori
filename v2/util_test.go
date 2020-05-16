package v2

import (
	"errors"
	"reflect"
	"sort"
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

func (c *testCmd) Handle(ctx *context.Context) error {
	if c.HandleCallback != nil {
		return c.HandleCallback(ctx)
	}
	return nil
}

func (c *testCmd) Resolve(ctx *context.Context) {
	if c.ResolveCallback != nil {
		c.ResolveCallback(ctx)
	}
}

//Parse parses a command. If ParseCallback is nil, will default to strings.Fields
func (c *testCmd) Parse(cmd string) ([]string, error) {
	if c.ParseCallback != nil {
		return c.ParseCallback(cmd)
	}
	return cmdParserDefault(cmd), nil
}

func (p *testIOParams) createCmd(t *testing.T) *testCmd {
	testFunc := func(ctx *context.Context) {

		if ctx.Prefix != p.expectedPrefix {
			t.Errorf("expected prefix to be equal, got %s, want %s", ctx.Prefix, p.expectedPrefix)
		}

		if p.expectedDepth != len(ctx.Alias) {
			t.Error("expected depth to equal length of context alias")
		}

		if !strSliceEqual(p.expectedAlias, ctx.Alias, false) {
			t.Error("expected alias to be equal")
		}
		if !strSliceEqual(p.expectedArgs, ctx.Args, false) {
			t.Error("expected args to be equal")
		}
	}
	handleCB := func(ctx *context.Context) error {
		testFunc(ctx)
		return p.expectedErr
	}

	resolveCB := func(ctx *context.Context) {
		if ctx.Err != p.expectedErr {
			t.Errorf("expected err to be equal, got %v, want %v", ctx.Err, p.expectedErr)
		}
		if ctx.Err == nil && ctx.Err == p.expectedErr {
			// fields here will only be valid if err was nil (this will not be run, if say - a parser or middleware err'd
			testFunc(ctx)
		}
	}

	return &testCmd{
		HandleCallback:  handleCB,
		ResolveCallback: resolveCB,
		ParseCallback:   func(cmd string) ([]string, error) { return cmdParserDefault(cmd), nil },
	}
}

// strSliceEqual is a helper function to ensure that 2 string slices are equal.
// if sorted true, will copy the slice and sort before comparing
func strSliceEqual(a, b []string, sorted bool) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	if sorted {
		sort.Strings(aCopy)
		sort.Strings(bCopy)
	}

	return reflect.DeepEqual(aCopy, bCopy)
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

func testCreateRouteHelper(root *testRouteDefns, cmd *testCmd) *Route {
	r := NewRoute(root.p)

	r.On(root.aliases...)
	r.Use(root.middlewares...)

	if cmd == nil {
		r.Do(root.c)
	} else {
		r.Do(cmd)
	}

	for _, sr := range root.subroutes {
		r.Has(testCreateRouteHelper(sr, cmd))
	}

	return r

}

// createRoute creates a route and ALL its subroutes recursively
func (p *testParams) createRoute() *Route {
	return testCreateRouteHelper(p.routeParams, nil)
}

// createRoute creates a route and ALL its subroutes recursively
// but generates a cmd from testIOParams to internally check the expected values.
// despite all routes/subroutes having the same handler it should be fine because
// the algorithm should find the deepest subcommand and run that handler.
func (p *testParams) createRouteWithTestIOParams(testIO *testIOParams, t *testing.T) *Route {
	return testCreateRouteHelper(p.routeParams, testIO.createCmd(t))
}

func testGetAllAliasRecursively(route *Route) []string {
	var alias []string
	alias = append(alias, route.aliases...)
	for _, sr := range route.subroutes {
		alias = append(alias, testGetAllAliasRecursively(sr)...)
	}
	return alias
}
