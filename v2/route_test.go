package v2

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/pixeltopic/sayori/v2/utils"
)

func makeMockMsg(content string) *discordgo.MessageCreate {
	message := &discordgo.Message{
		Author: &discordgo.User{
			ID:  "author_id_1",
			Bot: false,
		},
		Content: content,
		GuildID: "guild_id_1",
	}

	return &discordgo.MessageCreate{
		Message: message,
	}
}

func makeMockSes() *discordgo.Session {
	state := discordgo.NewState()
	state.User = &discordgo.User{ID: "self_id_1"}
	session := &discordgo.Session{
		State: state,
	}

	return session
}

// Note: If the message content arg does not match root, it will exit the test case without running any tests!
func TestRoute_createHandlerFunc(t *testing.T) {
	type subCase struct {
		name           string // descriptive string describing what the test case is doing
		content        string
		expectedDepth  int // depth of a subcommand token (zero indexed)
		expectedPrefix string
		expectedErr    error // error the ctx should contain
		expectedCmdID  int
	}
	type testCase struct {
		route     func(c subCase, t *testing.T) *Route
		aliasTree []string // registered aliases in the route and all subroutes. Must include duplicates and order does not matter
		subCases  []subCase
	}

	createCmd := func(id int, c subCase, t *testing.T) *testCmd {
		testFunc := func(ctx context.Context) {
			cmd := CmdFromContext(ctx)

			if c.expectedCmdID != id {
				t.Errorf("expected cmd ID to be equal, want %d, got %d", id, c.expectedCmdID)
			}

			if cmd.Prefix != c.expectedPrefix {
				t.Errorf("expected prefix to be equal, want %s, got %s", cmd.Prefix, c.expectedPrefix)
			}

			if c.expectedDepth != len(cmd.Alias) {
				t.Errorf("expected depth (%d) to equal length of context alias (%d)", c.expectedDepth, len(cmd.Alias))
			}

			content := c.content
			if c.expectedPrefix != "" {
				content = strings.TrimPrefix(c.content, c.expectedPrefix)
			}
			toks := cmdParserDefault(content)
			if !strSliceEqual(toks[:c.expectedDepth], cmd.Alias, false) {
				t.Errorf("expected alias %v to be equal to %v", toks[:c.expectedDepth], cmd.Alias)
			}
			if !strSliceEqual(toks[c.expectedDepth:], cmd.Args, false) {
				t.Errorf("expected args %v to be equal %v", toks[c.expectedDepth:], cmd.Args)
			}
		}
		handleCB := func(ctx context.Context) error {
			testFunc(ctx)
			return c.expectedErr
		}

		resolveCB := func(ctx context.Context) {
			cmd := CmdFromContext(ctx)

			if cmd.Err != nil && cmd.Err.Error() != c.expectedErr.Error() {
				t.Errorf("expected err to be equal, want %v, got %v", cmd.Err, c.expectedErr)
			}
			if cmd.Err == nil && c.expectedErr == nil {
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

	testCases := []testCase{
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewRoute(nil).On("root").Do(createCmd(0, c, t)).Has(
					NewSubroute().On("sub1").Do(createCmd(1, c, t)).Has(
						NewSubroute().On("subsub1").Do(createCmd(3, c, t)),
					),
					NewSubroute().On("sub1").Do(createCmd(2, c, t)).Has(
						NewSubroute().On("subsub2").Do(createCmd(4, c, t)),
					),
				)
				return r
			},
			aliasTree: []string{"root", "sub1", "sub1", "subsub1", "subsub2"},
			subCases: []subCase{
				{
					name:           "alias tiebreak will result in most recently added route being selected",
					content:        "root sub1",
					expectedDepth:  2,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  2,
				},
				{
					name:           "common aliases shared between routes in the same depth should properly route to the correct subroute",
					content:        "root sub1 subsub2 sub2 arg1 arg2",
					expectedDepth:  3,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  4,
				},
				{
					name:           "common aliases shared between routes in the same depth should properly route to the correct subroute",
					content:        "root sub1 subsub1 sub2 arg1 arg2",
					expectedDepth:  3,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  3,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewRoute(nil).On("root").Do(createCmd(0, c, t)).Has(
					NewSubroute().On("sub1").Do(createCmd(1, c, t)).Has(
						NewSubroute().On("subsub1").Do(createCmd(3, c, t)).Has(
							NewSubroute().On("sub1").Do(createCmd(5, c, t)),
						),
					),
					NewSubroute().On("sub1").Do(createCmd(2, c, t)).Has(
						NewSubroute().On("subsub1").Do(createCmd(4, c, t)),
					),
				)

				return r
			},
			aliasTree: []string{"root", "sub1", "sub1", "subsub1", "subsub1", "sub1"},
			subCases: []subCase{
				{
					name:           "deepest sub1 shall be selected",
					content:        "root sub1 subsub1 sub1",
					expectedDepth:  4,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  5,
				},
				{
					name:           "newest subsub1 shall be selected",
					content:        "root sub1 subsub1",
					expectedDepth:  3,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  4,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewRoute(nil).On("root").Do(createCmd(0, c, t))
				sub1A := NewSubroute().On("sub1").Do(createCmd(1, c, t)).Has(
					NewSubroute().On("subsub1").Do(createCmd(3, c, t)),
				)
				sub1B := NewSubroute().On("sub2").Do(createCmd(2, c, t))
				r.Has(sub1A, sub1B)
				return r
			},
			aliasTree: []string{"root", "sub1", "sub2", "subsub1"},
			subCases: []subCase{
				{
					name:           "should route to root",
					content:        "root",
					expectedDepth:  1,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  0,
				},
				{
					name:           "should route to sub2",
					content:        "root sub2",
					expectedDepth:  2,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  2,
				},
				{
					name:           "should route to subsub1",
					content:        "root sub1 subsub1 sub2 arg1 arg2",
					expectedDepth:  3,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  3,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewRoute(&testPref{}).On("root").Do(createCmd(0, c, t))
				r1 := NewSubroute().On("sub", "sub1", "s").Do(createCmd(1, c, t)).Has(
					NewSubroute().On("subsub1", "ss1").Do(createCmd(2, c, t)),
					NewSubroute().On("subsub2").Do(createCmd(3, c, t)),
					NewSubroute().On("subsub3", "sub").Do(createCmd(4, c, t)),
				)
				r.Has(r1)
				return r
			},
			aliasTree: []string{"root", "sub", "sub1", "s", "subsub1", "ss1", "subsub2", "subsub3", "sub"},
			subCases: []subCase{
				{
					name:           "should route to subsub3 with a prefix",
					content:        testDefaultPrefix + "root s subsub3 sub arg1 arg2",
					expectedDepth:  3,
					expectedPrefix: testDefaultPrefix,
					expectedErr:    nil,
					expectedCmdID:  4,
				},
				{
					name:           "should route to sub (2) with a prefix",
					content:        testDefaultPrefix + "root sub sub sub arg1 arg2",
					expectedDepth:  3,
					expectedPrefix: testDefaultPrefix,
					expectedErr:    nil,
					expectedCmdID:  4,
				},
				{
					name:           "should route to sub (1) with a prefix",
					content:        testDefaultPrefix + "root s sub",
					expectedDepth:  3,
					expectedPrefix: testDefaultPrefix,
					expectedErr:    nil,
					expectedCmdID:  4,
				},
				{
					name:           "should route to nowhere",
					content:        testDefaultPrefix + "s",
					expectedDepth:  0,
					expectedPrefix: testDefaultPrefix,
					expectedErr:    nil,
					expectedCmdID:  -1,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewRoute(&testPref{}).On().Do(createCmd(1, c, t))
				return r
			},
			aliasTree: []string{},
			subCases: []subCase{
				{
					name:           "valid prefix and a default route should return all tokens as args, without the prefix",
					content:        testDefaultPrefix + "root s hi there",
					expectedDepth:  0,
					expectedPrefix: testDefaultPrefix,
					expectedErr:    nil,
					expectedCmdID:  1,
				},
				{
					name:           "invalid prefix will not run despite being default route",
					content:        "root s hi there",
					expectedDepth:  0,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  -1,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewSubroute().On().Do(createCmd(1, c, t))
				r1 := NewSubroute().On("sub").Do(createCmd(2, c, t))
				r.Has(r1)
				return r
			},
			aliasTree: []string{"sub"},
			subCases: []subCase{
				{
					name:           "should run the first handler as there is no aliases, hence gets auto-selected",
					content:        testDefaultPrefix + "sub sub hi there",
					expectedDepth:  0,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  1,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewSubroute().On("root").Do(createCmd(1, c, t)).Has(
					NewSubroute().On("sub").Do(createCmd(2, c, t)).Has(
						NewSubroute().Do(createCmd(3, c, t)),
					),
				)

				return r
			},
			aliasTree: []string{"root", "sub"},
			subCases: []subCase{
				{
					name:           "should ignore the route with no aliases and use the last valid parent instead",
					content:        "root sub subsub",
					expectedDepth:  2,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  2,
				},
				{
					name:           "should fail to route to any handler",
					content:        testDefaultPrefix + "root sub subsub",
					expectedDepth:  0,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  -1,
				},
			},
		},
		{
			route: func(c subCase, t *testing.T) *Route {
				r := NewSubroute().On("root").Do(createCmd(0, c, t)).Has(
					NewSubroute().On("sub", "s").Do(createCmd(1, c, t)).Has(
						NewSubroute().Do(createCmd(2, c, t)),
					),
					NewSubroute().On().Do(createCmd(3, c, t)).Has(
						NewSubroute().On("subsub1").Do(createCmd(4, c, t)),
					),
				)

				return r
			},
			aliasTree: []string{"root", "sub", "s", "subsub1"},
			subCases: []subCase{
				{
					name:           "default route should be ignored",
					content:        "root s hi there",
					expectedDepth:  2,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  1,
				},
				{
					name:           "default route and its subroutes should be ignored",
					content:        "root arg subsub1 arg",
					expectedDepth:  1,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  0,
				},
				{
					name:           "default route and its subroutes should be ignored",
					content:        "root subsub1 arg",
					expectedDepth:  1,
					expectedPrefix: "",
					expectedErr:    nil,
					expectedCmdID:  0,
				},
			},
		},
	}

	for _, tc := range testCases {
		for _, c := range tc.subCases {
			// create test route
			route := tc.route(c, t)

			content := c.content
			if c.expectedPrefix != "" {
				content = strings.TrimPrefix(c.content, c.expectedPrefix)
			}

			// first, test the route finding algorithm itself.
			rr, depth := findRouteRecursive(route, cmdParserDefault(content), 1)
			if depth != c.expectedDepth {
				t.Errorf("findRouteRecursive returned unexpected depth, got %d, want %d", depth, c.expectedDepth)
			}
			if rr == nil && c.expectedDepth != 0 {
				t.Errorf("expected non-nil route because expectedDepth was %v, but route was nil", c.expectedDepth)
			}

			ctx := context.Background()
			msgCreate := makeMockMsg(c.content)
			ses := makeMockSes()

			// next, test that context is passed down to handlers properly
			createHandlerFunc(route)(utils.WithSes(utils.WithMsg(ctx, msgCreate.Message), ses))

			// lastly ensure the route contains all expected aliases.
			found := testGetAllAliasRecursively(route)
			if !strSliceEqual(tc.aliasTree, found, true) {
				t.Errorf("expected alias tree to be equal; got %v, want %v", found, tc.aliasTree)
			}

		}
	}

}
