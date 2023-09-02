package tests

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b"
	"github.com/ryym/geq/tests/mdl"
)

func TestBuiltQueries(t *testing.T) {
	runTestCases(t, []testCase{
		{
			name: "basic select",
			run: func() bool {
				got := b.SelectFrom(b.Users).Finalize()
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
				users := []mdl.User{
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				q := b.SelectVia(users, b.Posts, b.Posts.Author)
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
				q := b.SelectFrom(b.Posts).Joins(b.Posts.Author).OrderBy(b.Posts.AuthorID)
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
				q := b.Select(
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
		{
			name: "select function calls",
			run: func() bool {
				q := b.Select(
					b.Count(b.Users.ID),
					b.Max(b.Users.Name),
					b.Func("MYFUNC", 1, b.Users.ID),
				)
				got := q.Finalize()
				want := newFinalQuery(
					"SELECT COUNT(users.id), MAX(users.name), MYFUNC(?, users.id)",
					1,
				)
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select with grouping",
			run: func() bool {
				q := b.SelectFrom(b.Users, b.Users.Name, b.Max(b.Users.ID)).GroupBy(b.Users.Name)
				got := q.Finalize()
				want := newFinalQuery("SELECT users.name, MAX(users.id) FROM users GROUP BY users.name")
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select with limit",
			run: func() bool {
				q := b.SelectFrom(b.Users, b.Users.ID).Limit(2)
				got := q.Finalize()
				want := newFinalQuery("SELECT users.id FROM users LIMIT 2")
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select from sub query",
			run: func() bool {
				q := b.Select(b.Raw("t.id"), b.Raw("t.title")).From(b.SelectFrom(b.Posts).As("t"))
				got := q.Finalize()
				want := newFinalQuery("SELECT t.id, t.title FROM (SELECT posts.id, posts.author_id, posts.title FROM posts) AS t")
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("wrong final query:%s", diff)
					return false
				}
				return false
			},
		},
	})
}

func newFinalQuery(query string, args ...any) *geq.FinalQuery {
	return &geq.FinalQuery{Query: query, Args: args}
}
