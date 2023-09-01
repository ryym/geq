package tests

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b"
	"github.com/ryym/geq/tests/mdl"
)

// type Value

func TestResultMappings(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3990)/geq")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping to DB: %v", err)
	}

	ctx := context.Background()

	runTestCases(t, []testCase{
		{
			name: "select into single slice",
			run: func() bool {
				q := b.From(b.Users).OrderBy(b.Users.ID)
				var users []mdl.User
				err := geq.AsThese(q, geq.ToSlice(b.Users, &users)).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(users, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "select into single map",
			run: func() bool {
				q := b.From(b.Users).OrderBy(b.Users.ID)
				var userMap map[int64]mdl.User
				err := geq.AsThese(q, geq.ToMap(b.Users, b.Users.ID, &userMap)).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(userMap, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "load as single slice",
			run: func() bool {
				q := b.From(b.Users).OrderBy(b.Users.ID)
				users, err := geq.AsSlice(q).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(users, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "load as single map",
			run: func() bool {
				q := b.From(b.Users).OrderBy(b.Users.ID)
				userMap, err := geq.AsMap(q, b.Users.Name).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[string]mdl.User{
					"user1": {ID: 1, Name: "user1"},
					"user2": {ID: 2, Name: "user2"},
					"user3": {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(userMap, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "load as custom slice",
			run: func() bool {
				q := b.From(b.Posts).
					Joins(b.Posts.Author).
					GroupBy(b.Posts.AuthorID).
					OrderBy(b.Posts.AuthorID)

				stats, err := geq.AsSliceOf(q, &PostStats{
					AuthorID:  b.Users.ID,
					PostCount: b.Count(b.Posts.ID),
					LastTitle: b.Max(b.Posts.Title),
				}).Load(ctx, db)

				if err != nil {
					t.Error(err)
				}
				want := []PostStat{
					{AuthorID: 1, PostCount: 2, LastTitle: "user1-post2"},
					{AuthorID: 2, PostCount: 1, LastTitle: "user2-post1"},
					{AuthorID: 3, PostCount: 3, LastTitle: "user3-post3"},
				}
				if diff := cmp.Diff(stats, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
		{
			name: "load as values",
			run: func() bool {
				q := b.From(b.Users).OrderBy(b.Users.ID)
				ids, err := geq.AsValues(q, b.Users.ID).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []int64{1, 2, 3}
				if diff := cmp.Diff(ids, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
	})
}
