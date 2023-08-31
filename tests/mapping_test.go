package tests

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
)

func TestResultMappings(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3990)/geq")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping to DB: %v", err)
	}

	b := NewQueryBuilder()
	ctx := context.Background()

	runTestCases(t, []testCase{
		{
			name: "select into single slice",
			run: func() bool {
				q := geq.From(b.Users).OrderBy(b.Users.ID)
				var users []User
				err := geq.AsThese(q, geq.ToSlice(b.Users, &users)).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []User{
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
				q := geq.From(b.Users).OrderBy(b.Users.ID)
				var userMap map[int64]User
				err := geq.AsThese(q, geq.ToMap(b.Users, b.Users.ID, &userMap)).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[int64]User{
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
				q := geq.From(b.Users).OrderBy(b.Users.ID)
				users, err := geq.AsSlice(q).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []User{
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
				q := geq.From(b.Users).OrderBy(b.Users.ID)
				userMap, err := geq.AsMap(q, b.Users.Name).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[string]User{
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
				q := geq.From(b.Posts).Joins(b.Posts.Author).OrderBy(b.Posts.AuthorID).Limit(3)
				stats, err := geq.AsSliceOf(q, &PostStats{
					AuthorID: b.Users.ID,
					Foo:      b.Posts.ID.Eq(b.Users.ID),
				}).Load(ctx, db)

				if err != nil {
					t.Error(err)
				}
				want := []PostStat{
					{AuthorID: 1, Foo: true},
					{AuthorID: 1, Foo: false},
					{AuthorID: 2, Foo: false},
				}
				if diff := cmp.Diff(stats, want); diff != "" {
					t.Errorf("wrong result:%s", diff)
					return false
				}
				return true
			},
		},
	})
}
