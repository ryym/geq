package tests

import (
	"strings"
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
				want := newFinalQuery("SELECT users.id, users.name FROM users")
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
					"SELECT posts.id, posts.author_id, posts.title FROM posts WHERE posts.author_id IN (?,?)",
					int64(2), int64(3),
				)
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select with join using relationship",
			run: func() bool {
				q := b.Posts.Query().Joins(b.Posts.Author).OrderBy(b.Posts.AuthorID)
				got := q.Finalize()
				want := newFinalQuery(
					strings.Join([]string{
						"SELECT posts.id, posts.author_id, posts.title FROM posts",
						"INNER JOIN users ON posts.author_id = users.id",
						"ORDER BY posts.author_id",
					}, " "),
				)
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select expressions",
			run: func() bool {
				q := geq.Select(
					b.Users.ID,
					b.Users.Name.As("foo"),
					b.Posts.ID.Eq(3),
					b.Posts.Title.Eq("title").As("bar"),
				)
				got := q.Finalize()
				want := newFinalQuery(
					"SELECT users.id, users.name AS foo, posts.id = ?, posts.title = ? AS bar",
					3, "title",
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
