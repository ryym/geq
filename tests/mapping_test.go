package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
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
			name: "scan into single slice",
			run: func() (err error) {
				var users []mdl.User
				q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
				err = q.WillScan(b.ToSlice(b.Users, &users)).Load(ctx, db)
				if err != nil {
					return err
				}
				want := []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(users, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into single map",
			run: func() (err error) {
				var userMap map[int64]mdl.User
				q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
				err = q.WillScan(b.ToMap(b.Users, b.Users.ID, &userMap)).Load(ctx, db)
				if err != nil {
					return err
				}
				want := map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(userMap, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into single slice map",
			run: func() (err error) {
				var postsMap map[int64][]mdl.Post
				q := b.SelectFrom(b.Posts).OrderBy(b.Posts.ID)
				err = q.WillScan(b.ToSliceMap(b.Posts, b.Posts.AuthorID, &postsMap)).Load(ctx, db)
				if err != nil {
					return err
				}
				want := map[int64][]mdl.Post{
					1: {
						{ID: 1, AuthorID: 1, Title: "user1-post1"},
						{ID: 2, AuthorID: 1, Title: "user1-post2"},
					},
					2: {
						{ID: 3, AuthorID: 2, Title: "user2-post1"},
					},
					3: {
						{ID: 4, AuthorID: 3, Title: "user3-post1"},
						{ID: 5, AuthorID: 3, Title: "user3-post2"},
						{ID: 6, AuthorID: 3, Title: "user3-post3"},
					},
				}
				if diff := cmp.Diff(postsMap, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into multiple results",
			run: func() (err error) {
				var posts []mdl.Post
				var userMap map[int64]mdl.User
				q := b.SelectFrom(b.Posts).Joins(b.Posts.Author).OrderBy(b.Posts.ID)
				err = q.WillScan(
					b.ToSlice(b.Posts, &posts),
					b.ToMap(b.Users, b.Users.ID, &userMap),
				).Load(ctx, db)
				if err != nil {
					return err
				}
				wantPostSlice := []mdl.Post{
					{ID: 1, AuthorID: 1, Title: "user1-post1"},
					{ID: 2, AuthorID: 1, Title: "user1-post2"},
					{ID: 3, AuthorID: 2, Title: "user2-post1"},
					{ID: 4, AuthorID: 3, Title: "user3-post1"},
					{ID: 5, AuthorID: 3, Title: "user3-post2"},
					{ID: 6, AuthorID: 3, Title: "user3-post3"},
				}
				if diff := cmp.Diff(posts, wantPostSlice); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				wantUserMap := map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(userMap, wantUserMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as single slice",
			run: func() (err error) {
				users, err := b.SelectFrom(b.Users).OrderBy(b.Users.ID).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(users, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as single map",
			run: func() (err error) {
				userMap, err := b.AsMap(b.Users.Name, b.SelectFrom(b.Users).OrderBy(b.Users.ID)).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[string]mdl.User{
					"user1": {ID: 1, Name: "user1"},
					"user2": {ID: 2, Name: "user2"},
					"user3": {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(userMap, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as single slice map",
			run: func() (err error) {
				q := b.SelectFrom(b.Posts).OrderBy(b.Posts.ID)
				postsMap, err := b.AsSliceMap(b.Posts.AuthorID, q).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := map[int64][]mdl.Post{
					1: {
						{ID: 1, AuthorID: 1, Title: "user1-post1"},
						{ID: 2, AuthorID: 1, Title: "user1-post2"},
					},
					2: {
						{ID: 3, AuthorID: 2, Title: "user2-post1"},
					},
					3: {
						{ID: 4, AuthorID: 3, Title: "user3-post1"},
						{ID: 5, AuthorID: 3, Title: "user3-post2"},
						{ID: 6, AuthorID: 3, Title: "user3-post3"},
					},
				}
				if diff := cmp.Diff(postsMap, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as custom slice",
			run: func() (err error) {
				stats, err := b.SelectAs(&PostStats{
					AuthorID:  b.Users.ID,
					PostCount: b.Count(b.Posts.ID),
					LastTitle: b.Max(b.Posts.Title),
				}).From(b.Posts).
					Joins(b.Posts.Author).
					GroupBy(b.Posts.AuthorID).
					OrderBy(b.Posts.AuthorID).
					Load(ctx, db)

				if err != nil {
					t.Error(err)
				}
				want := []PostStat{
					{AuthorID: 1, PostCount: 2, LastTitle: "user1-post2"},
					{AuthorID: 2, PostCount: 1, LastTitle: "user2-post1"},
					{AuthorID: 3, PostCount: 3, LastTitle: "user3-post3"},
				}
				if diff := cmp.Diff(stats, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as values",
			run: func() (err error) {
				ids, err := b.SelectOnly(b.Users.ID).From(b.Users).OrderBy(b.Users.ID).Load(ctx, db)
				if err != nil {
					t.Error(err)
				}
				want := []int64{1, 2, 3}
				if diff := cmp.Diff(ids, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as sql.Rows",
			run: func() (err error) {
				q := b.Select(b.Raw("2-9"), b.Posts.AuthorID, b.Max(b.Posts.ID)).From(b.Posts).GroupBy(b.Posts.AuthorID)
				rows, err := q.LoadRows(ctx, db)
				if err != nil {
					t.Error(err)
				}

				results := make([][]int, 0, 3)
				for rows.Next() {
					var v1, v2, v3 int
					rows.Scan(&v1, &v2, &v3)
					results = append(results, []int{v1, v2, v3})
				}

				want := [][]int{{-7, 1, 2}, {-7, 2, 3}, {-7, 3, 6}}
				if diff := cmp.Diff(results, want); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
	})
}
