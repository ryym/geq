package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b"
	"github.com/ryym/geq/tests/mdl"
)

func TestGeq(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(:3990)/geq")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping to DB: %v", err)
	}

	ctx := context.Background()
	runTestCases(t, db, []testCase{
		{
			name: "load as single slice",
			run: func(db *sql.Tx) (err error) {
				users, err := b.SelectFrom(b.Users).OrderBy(b.Users.ID).Load(ctx, db)
				if err != nil {
					return err
				}
				want := []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(want, users); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as single map",
			run: func(db *sql.Tx) (err error) {
				userMap, err := b.AsMap(b.Users.Name, b.SelectFrom(b.Users).OrderBy(b.Users.ID)).Load(ctx, db)
				if err != nil {
					return err
				}
				want := map[string]mdl.User{
					"user1": {ID: 1, Name: "user1"},
					"user2": {ID: 2, Name: "user2"},
					"user3": {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(want, userMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as single slice map",
			run: func(db *sql.Tx) (err error) {
				q := b.SelectFrom(b.Posts).OrderBy(b.Posts.ID)
				postsMap, err := b.AsSliceMap(b.Posts.AuthorID, q).Load(ctx, db)
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
				if diff := cmp.Diff(want, postsMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as non-table row slice",
			run: func(db *sql.Tx) (err error) {
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
					return err
				}
				want := []PostStat{
					{AuthorID: 1, PostCount: 2, LastTitle: "user1-post2"},
					{AuthorID: 2, PostCount: 1, LastTitle: "user2-post1"},
					{AuthorID: 3, PostCount: 3, LastTitle: "user3-post3"},
				}
				if diff := cmp.Diff(want, stats); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as values",
			run: func(db *sql.Tx) (err error) {
				ids, err := b.SelectOnly(b.Users.ID).From(b.Users).OrderBy(b.Users.ID).Load(ctx, db)
				if err != nil {
					return err
				}
				want := []int64{1, 2, 3}
				if diff := cmp.Diff(want, ids); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "load as sql.Rows",
			run: func(db *sql.Tx) (err error) {
				q := b.Select(b.Raw("2-9"), b.Posts.AuthorID, b.Max(b.Posts.ID)).From(b.Posts).GroupBy(b.Posts.AuthorID)
				rows, err := q.LoadRows(ctx, db)
				if err != nil {
					return err
				}
				results := make([][]int, 0, 3)
				for rows.Next() {
					var v1, v2, v3 int
					rows.Scan(&v1, &v2, &v3)
					results = append(results, []int{v1, v2, v3})
				}
				want := [][]int{{-7, 1, 2}, {-7, 2, 3}, {-7, 3, 6}}
				if diff := cmp.Diff(want, results); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into single slice",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(want, users); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into single map",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(want, userMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into single slice map",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(want, postsMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "scan into multiple results",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(wantPostSlice, posts); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				wantUserMap := map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				}
				if diff := cmp.Diff(wantUserMap, userMap); diff != "" {
					return fmt.Errorf("wrong result:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select from table",
			run: func(db *sql.Tx) (err error) {
				got := b.SelectFrom(b.Users).Finalize()
				want := newFinalQuery("SELECT users.id, users.name FROM users")
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select via table relationship",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select with join using relationship",
			run: func(db *sql.Tx) (err error) {
				q := b.SelectFrom(b.Posts).Joins(b.Posts.Author).OrderBy(b.Posts.AuthorID)
				got := q.Finalize()
				want := newFinalQuery(
					strings.Join([]string{
						"SELECT posts.id, posts.author_id, posts.title FROM posts",
						"INNER JOIN users ON posts.author_id = users.id",
						"ORDER BY posts.author_id",
					}, " "),
				)
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select expressions",
			run: func(db *sql.Tx) (err error) {
				q := b.Select(
					b.Users.ID,
					b.Users.Name.As("foo"),
					b.Posts.ID.Eq(3),
					b.Posts.Title.Eq("title").As("bar"),
					b.Null(),
				)
				got := q.Finalize()
				want := newFinalQuery(
					"SELECT users.id, users.name AS foo, posts.id = ?, posts.title = ? AS bar, NULL",
					3, "title",
				)
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select function calls",
			run: func(db *sql.Tx) (err error) {
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
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select with grouping",
			run: func(db *sql.Tx) (err error) {
				q := b.SelectFrom(b.Users, b.Users.Name, b.Max(b.Users.ID)).GroupBy(b.Users.Name)
				got := q.Finalize()
				want := newFinalQuery("SELECT users.name, MAX(users.id) FROM users GROUP BY users.name")
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select with limit",
			run: func(db *sql.Tx) (err error) {
				q := b.SelectFrom(b.Users, b.Users.ID).Limit(2)
				got := q.Finalize()
				want := newFinalQuery("SELECT users.id FROM users LIMIT 2")
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "select from sub query",
			run: func(db *sql.Tx) (err error) {
				q := b.Select(b.Raw("t.id"), b.Raw("t.title")).From(b.SelectFrom(b.Posts).As("t"))
				got := q.Finalize()
				want := newFinalQuery("SELECT t.id, t.title FROM (SELECT posts.id, posts.author_id, posts.title FROM posts) AS t")
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "build insert query by values and value maps",
			run: func(db *sql.Tx) (err error) {
				q := b.InsertInto(b.Users).
					Values(
						b.Users.ID.Set(1),
						b.Users.Name.Set("name"),
					).
					ValueMaps(
						geq.ValueMap{
							b.Users.ID:   2,
							b.Users.Name: nil,
						},
						geq.ValueMap{
							b.Users.ID:   "invalid-id",
							b.Users.Name: b.Func("NOW"),
						},
					)

				got, err := q.Finalize()
				if err != nil {
					return err
				}
				want := newFinalQuery(
					"INSERT INTO users (id, name) VALUES (?, ?), (?, NULL), (?, NOW())",
					int64(1), "name",
					2,
					"invalid-id",
				)
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}
				return nil
			},
		},
		{
			name: "insert records",
			run: func(db *sql.Tx) (err error) {
				q := b.InsertInto(b.Users).
					Values(
						b.Users.ID.Set(100),
						b.Users.Name.Set("name100"),
					).
					ValueMaps(
						geq.ValueMap{
							b.Users.ID:   200,
							b.Users.Name: "name200",
						},
					)
				fq, err := q.Finalize()
				if err != nil {
					return err
				}
				ret, err := db.Exec(fq.Query, fq.Args...)
				if err != nil {
					return err
				}
				nAffected, err := ret.RowsAffected()
				if err != nil {
					return err
				}
				if diff := cmp.Diff(nAffected, int64(2)); diff != "" {
					return fmt.Errorf("wrong affected:%s", diff)
				}
				return nil
			},
		},
		{
			name: "update records",
			run: func(db *sql.Tx) (err error) {
				q := b.Update(b.Posts).Set(
					b.Posts.AuthorID.Set(1),
					b.Posts.Title.Set("title"),
				).Where(b.Posts.ID.In([]int64{3, 4}))
				got, err := q.Finalize()
				if err != nil {
					return err
				}
				want := newFinalQuery(
					"UPDATE posts SET author_id = ?, title = ? WHERE posts.id IN (?,?)",
					int64(1), "title", int64(3), int64(4),
				)
				if diff := cmp.Diff(want, got); diff != "" {
					return fmt.Errorf("wrong final query:%s", diff)
				}

				_, err = db.Exec(got.Query, got.Args...)
				if err != nil {
					return err
				}
				posts, err := b.SelectFrom(b.Posts).
					Where(b.Posts.ID.In([]int64{3, 4})).
					OrderBy(b.Posts.ID).
					Load(ctx, db)
				if err != nil {
					return err
				}
				wantPosts := []mdl.Post{
					{ID: 3, AuthorID: 1, Title: "title"},
					{ID: 4, AuthorID: 1, Title: "title"},
				}
				if diff := cmp.Diff(wantPosts, posts); diff != "" {
					return fmt.Errorf("wrong affected:%s", diff)
				}
				return nil
			},
		},
	})
}

func newFinalQuery(query string, args ...any) *geq.FinalQuery {
	return &geq.FinalQuery{Query: query, Args: args}
}
