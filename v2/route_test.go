package v2

import (
	"testing"

	"github.com/pixeltopic/sayori/v2/context"
)

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

		// tests a Command with no aliases
		{
			testIOParams: []*testIOParams{
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
		// tests a subroute with no aliases. It should not be called.
		{
			testIOParams: []*testIOParams{
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
					expectedAliasTree:   []string{"root", "sub", "s"},
					expectedPrefix:      "",
					expectedAlias:       []string{"root", "s"},
					expectedArgs:        []string{"hi", "there"},
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
				},
			},
		},
		// tests a Command with no aliases but has a prefixer
		{
			testIOParams: []*testIOParams{
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
				ctx := context.New()
				msgCreate, err := io.createMockMsg()
				if err != nil {
					t.FailNow()
				}
				ctx.Msg = msgCreate.Message

				ses, err := io.createMockSes()
				if err != nil {
					t.FailNow()
				}
				ctx.Ses = ses

				rr.handler(ctx)
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
				found, depth := findRoute(rr, io.msgContentTokenized)
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
