package tests

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
)

func TestBuiltQueries(t *testing.T) {
	b := NewQueryBuilder()

	runTestCases(t, []testCase{
		{
			name: "basic select",
			run: func() bool {
				got := b.Users.Query().Finalize()
				want := newFinalQuery("SELECT users.id,users.name FROM users")
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select via table relationship",
			run: func() bool {
				users := []User{
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				q := geq.QueryVia(users, b.Posts, b.Posts.Author)
				got := q.Finalize()
				want := newFinalQuery(
					"SELECT posts.id,posts.author_id,posts.title FROM posts WHERE posts.author_id IN (?,?)",
					int64(2), int64(3),
				)
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
	})
}

func newFinalQuery(query string, args ...any) *geq.FinalQuery {
	return &geq.FinalQuery{Query: query, Args: args}
}
