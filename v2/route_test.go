package v2

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/pixeltopic/sayori/v2/utils"

	"context"
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
				t.Errorf("expected cmd ID to be equal, got %d, want %d", id, c.expectedCmdID)
			}

			if cmd.Prefix != c.expectedPrefix {
				t.Errorf("expected prefix to be equal, got %s, want %s", cmd.Prefix, c.expectedPrefix)
			}

			if c.expectedDepth != len(cmd.Alias) {
				t.Errorf("expected depth (%d) to equal length of context alias (%d)", c.expectedDepth, len(cmd.Alias))
			}

			toks := cmdParserDefault(c.content)
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
				t.Errorf("expected err to be equal, got %v, want %v", cmd.Err, c.expectedErr)
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
				r := NewRoute(nil).On("root").Do(createCmd(0, c, t))
				sub1A := NewRoute(nil).On("sub1").Do(createCmd(1, c, t)).Has(
					NewRoute(nil).On("subsub1").Do(createCmd(3, c, t)),
				)
				sub1B := NewRoute(nil).On("sub1").Do(createCmd(2, c, t)).Has(
					NewRoute(nil).On("subsub2").Do(createCmd(4, c, t)),
				)
				r.Has(sub1A, sub1B)
				return r
			},
			aliasTree: []string{"root", "sub1", "sub1", "subsub1", "subsub2"},
			subCases: []subCase{
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
	}

	for _, tc := range testCases {
		for _, c := range tc.subCases {
			// create test route
			route := tc.route(c, t)

			// first, test the route finding algorithm itself.
			rr, depth := findRouteRecursive(route, cmdParserDefault(c.content), 1)
			if depth != c.expectedDepth {
				t.Errorf("findRouteRecursive returned unexpected depth, got %d, want %d", depth, c.expectedDepth)
			}
			if rr == nil && c.expectedDepth != 0 {
				t.Errorf("expected non-nil route because expectedDepth was %v, but route was nil", c.expectedDepth)
				continue
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
						msgContent: "root sub1 subsub2 sub2 arg1 arg2",
					},
					msgContentTokenized: []string{"root", "sub1", "subsub2", "sub2", "arg1", "arg2"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub1", "sub1", "subsub1", "subsub2"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "sub1", "subsub2"},
					expectedArgs:        []string{"sub2", "arg1", "arg2"},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root sub1 subsub1 sub2 arg1 arg2",
					},
					msgContentTokenized: []string{"root", "sub1", "subsub1", "sub2", "arg1", "arg2"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub1", "sub1", "subsub1", "subsub2"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "sub1", "subsub1"},
					expectedArgs:        []string{"sub2", "arg1", "arg2"},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				aliases:     []string{"root"},
				middlewares: nil,
				subroutes: []*testRouteDefns{
					{
						aliases:     []string{"sub1"},
						middlewares: nil,
						subroutes: []*testRouteDefns{
							{
								aliases:     []string{"subsub1"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
						},
					},
					{
						aliases:     []string{"sub1"},
						middlewares: nil,
						subroutes: []*testRouteDefns{
							{
								aliases:     []string{"subsub2"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
						},
					},
				},
			},
		},
		{
			testIOParams: []*testIOParams{
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root sub1 subsub1 sub2 arg1 arg2",
					},
					msgContentTokenized: []string{"root", "sub1", "subsub1", "sub2", "arg1", "arg2"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub1", "sub2", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "sub1", "subsub1"},
					expectedArgs:        []string{"sub2", "arg1", "arg2"},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root",
					},
					msgContentTokenized: []string{"root"},
					expectedDepth:       1,
					expectedAliasTree:   []string{"root", "sub1", "sub2", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root"},
					expectedArgs:        []string{},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root sub2",
					},
					msgContentTokenized: []string{"root", "sub2"},
					expectedDepth:       2,
					expectedAliasTree:   []string{"root", "sub1", "sub2", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "sub2"},
					expectedArgs:        []string{},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				aliases:     []string{"root"},
				middlewares: nil,
				subroutes: []*testRouteDefns{
					{
						aliases:     []string{"sub1"},
						middlewares: nil,
						subroutes: []*testRouteDefns{
							{
								aliases:     []string{"subsub1"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
						},
					},
					{
						aliases:     []string{"sub2"},
						middlewares: nil,
						subroutes:   []*testRouteDefns{},
					},
				},
			},
		},
		{
			testIOParams: []*testIOParams{
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "root s subsub3 sub arg1 arg2",
					},
					msgContentTokenized: []string{"root", "s", "subsub3", "sub", "arg1", "arg2"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub", "sub1", "s", "subsub1", "ss1", "subsub2", "subsub3", "sub"},
					expectedPrefix:      testDefaultPrefix,
					expectedAlias:       []string{"root", "s", "subsub3"},
					expectedArgs:        []string{"sub", "arg1", "arg2"},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "root sub sub sub arg1 arg2",
					},
					msgContentTokenized: []string{"root", "sub", "sub", "sub", "arg1", "arg2"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub", "sub1", "s", "subsub1", "ss1", "subsub2", "subsub3", "sub"},
					expectedPrefix:      testDefaultPrefix,
					expectedAlias:       []string{"root", "sub", "sub"},
					expectedArgs:        []string{"sub", "arg1", "arg2"},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "root s sub",
					},
					msgContentTokenized: []string{"root", "s", "sub"},
					expectedDepth:       3,
					expectedAliasTree:   []string{"root", "sub", "sub1", "s", "subsub1", "ss1", "subsub2", "subsub3", "sub"},
					expectedPrefix:      testDefaultPrefix,
					expectedAlias:       []string{"root", "s", "sub"},
					expectedArgs:        []string{},
					expectedErr:         nil,
				},
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "s",
					},
					msgContentTokenized: []string{"s"},
					expectedDepth:       0,
					expectedAliasTree:   []string{"root", "sub", "sub1", "s", "subsub1", "ss1", "subsub2", "subsub3", "sub"},
					expectedPrefix:      "",
					expectedAlias:       []string{},
					expectedArgs:        []string{},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				p:           &testPref{},
				aliases:     []string{"root"},
				middlewares: nil,
				subroutes: []*testRouteDefns{
					{
						p:           &testPref{},
						aliases:     []string{"sub", "sub1", "s"},
						middlewares: nil,
						subroutes: []*testRouteDefns{
							{
								aliases:     []string{"subsub1", "ss1"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
							{
								aliases:     []string{"subsub2"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
							{
								aliases:     []string{"subsub3", "sub"},
								middlewares: nil,
								subroutes:   []*testRouteDefns{},
							},
						},
					},
				},
			},
		},
		{
			testIOParams: []*testIOParams{
				// tests a default route with a subroute which should not be considered.
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "sub sub hi there",
					},
					msgContentTokenized: []string{testDefaultPrefix + "sub", "sub", "hi", "there"},
					expectedDepth:       0,
					expectedAliasTree:   []string{"sub"},
					expectedPrefix:      "",
					expectedAlias:       []string{},
					expectedArgs:        []string{testDefaultPrefix + "sub", "sub", "hi", "there"},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				aliases: []string{},
				subroutes: []*testRouteDefns{
					{
						aliases:   []string{"sub"},
						subroutes: []*testRouteDefns{},
					},
				},
			},
		},
		{
			testIOParams: []*testIOParams{
				// Default route should be ignored
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root s hi there",
					},
					msgContentTokenized: []string{"root", "s", "hi", "there"},
					expectedDepth:       2,
					expectedAliasTree:   []string{"root", "sub", "s", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "s"},
					expectedArgs:        []string{"hi", "there"},
					expectedErr:         nil,
				},
				// Default route and its subroutes should be ignored
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root arg subsub1 arg",
					},
					msgContentTokenized: []string{"root", "arg", "subsub1", "arg"},
					expectedDepth:       1,
					expectedAliasTree:   []string{"root", "sub", "s", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root"},
					expectedArgs:        []string{"arg", "subsub1", "arg"},
					expectedErr:         nil,
				},
				// Default route and its subroutes should be ignored
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root subsub1 arg",
					},
					msgContentTokenized: []string{"root", "subsub1", "arg"},
					expectedDepth:       1,
					expectedAliasTree:   []string{"root", "sub", "s", "subsub1"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root"},
					expectedArgs:        []string{"subsub1", "arg"},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				aliases: []string{"root"},
				subroutes: []*testRouteDefns{
					{
						aliases: []string{"sub", "s"},
						subroutes: []*testRouteDefns{
							{
								aliases:   []string{},
								subroutes: []*testRouteDefns{},
							},
						},
					},
					{
						aliases: []string{},
						subroutes: []*testRouteDefns{
							{
								aliases:   []string{"subsub1"},
								subroutes: []*testRouteDefns{},
							},
						},
					},
				},
			},
		},
		{
			testIOParams: []*testIOParams{
				// Default route with a correct prefixer will return all tokens as args
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: testDefaultPrefix + "root s hi there",
					},
					msgContentTokenized: []string{"root", "s", "hi", "there"},
					expectedDepth:       0,
					expectedAliasTree:   []string{},
					expectedPrefix:      testDefaultPrefix,
					expectedAlias:       []string{},
					expectedArgs:        []string{"root", "s", "hi", "there"},
					expectedErr:         nil,
				},
				// Default route with invalid prefix will not run
				{
					sesParams: &mockSesParams{selfUserID: "self_id_1"},
					msgParams: &mockMsgParams{
						authorBot:  false,
						authorID:   "author_id_1",
						msgGuildID: "guild_id_1",
						msgContent: "root s hi there",
					},
					msgContentTokenized: []string{"root", "s", "hi", "there"},
					expectedDepth:       0,
					expectedAliasTree:   []string{},
					expectedPrefix:      "",
					expectedAlias:       []string{},
					expectedArgs:        []string{},
					expectedErr:         nil,
				},
			},
			routeParams: &testRouteDefns{
				p:         &testPref{},
				aliases:   []string{},
				subroutes: []*testRouteDefns{},
			},
		},
	}

	for _, tt := range testTrees {
		for _, io := range tt.testIOParams {
			t.Run("test handler func", func(t *testing.T) {
				rr := tt.createRouteWithTestIOParams(io, t)
				ctx := context.Background()
				msgCreate, err := io.createMockMsg()
				if err != nil {
					t.FailNow()
				}

				ses, err := io.createMockSes()
				if err != nil {
					t.FailNow()
				}

				createHandlerFunc(rr)(utils.WithSes(utils.WithMsg(ctx, msgCreate.Message), ses))
			})

			t.Run("test that all aliases in the route tree are present", func(t *testing.T) {
				rr := tt.createRoute()
				found := testGetAllAliasRecursively(rr)
				if !strSliceEqual(io.expectedAliasTree, found, true) {
					t.Errorf("expected alias tree to be equal; got %v, want %v", found, io.expectedAliasTree)
				}
			})
			t.Run("test findRoute algorithm", func(t *testing.T) {
				rr := tt.createRoute()
				found, depth := findRouteRecursive(rr, io.msgContentTokenized, 1)
				if depth != io.expectedDepth {
					t.Errorf("got %d, want %d", depth, io.expectedDepth)
				}
				if found == nil && io.expectedDepth != 0 {
					t.Error("expected non-nil route")
				}
			})
		}
	}
}
