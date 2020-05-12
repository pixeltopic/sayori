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
				if !strSliceEqual(io.expectedAliasTree, testGetAllAliasRecursively(rr), true) {
					t.Error("expected alias tree to be equal")
				}
			})
			t.Run("test findRoute algorithm", func(t *testing.T) {
				rr := tt.createRoute()
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
}
