package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	"github.com/ryym/geq"
	"github.com/ryym/geq/internal/tests/d"
	"github.com/ryym/geq/internal/tests/mdl"
)

func TestPostgreSQL(t *testing.T) {
	db, err := openDB("postgres", "port=3991 user=geq password=geq sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = initDB(db, initPostgreSQL, fixtureSQL)
	if err != nil {
		t.Fatal(err)
	}

	geq.SetDefaultDialect(&geq.DialectPostgres{})
	runIntegrationTest(t, db)
}

func TestMySQL(t *testing.T) {
	db, err := openDB("mysql", "root:root@tcp(:3990)/geq?multiStatements=true&parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = initDB(db, initMySQL, fixtureSQL)
	if err != nil {
		t.Fatal(err)
	}

	geq.SetDefaultDialect(&geq.DialectMySQL{})
	runIntegrationTest(t, db)
}

func openDB(driver, dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping to DB: %w", err)
	}
	return db, nil
}

func initDB(db *sql.DB, initSQL, fixtureSQL string) (err error) {
	_, err = db.Exec(initSQL)
	if err != nil {
		return fmt.Errorf("failed to run initSQL: %w", err)
	}
	_, err = db.Exec(fixtureSQL)
	if err != nil {
		return fmt.Errorf("failed to run fixtureSQL: %w", err)
	}
	return nil
}

func runIntegrationTest(t *testing.T, db *sql.DB) {
	ctx := context.Background()
	runTestCases(t, db, []testCase{
		{
			name: "load as single slice",
			run: func(db *sql.Tx) (err error) {
				users, err := geq.SelectFrom(d.Users).OrderBy(d.Users.ID).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(users, []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load as single map",
			run: func(db *sql.Tx) (err error) {
				userMap, err := geq.AsMap(d.Users.Name, geq.SelectFrom(d.Users).OrderBy(d.Users.ID)).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(userMap, map[string]mdl.User{
					"user1": {ID: 1, Name: "user1"},
					"user2": {ID: 2, Name: "user2"},
					"user3": {ID: 3, Name: "user3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load as single slice map",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Posts).OrderBy(d.Posts.ID)
				postsMap, err := geq.AsSliceMap(d.Posts.AuthorID, q).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(postsMap, map[int64][]mdl.Post{
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
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load as non-table row slice",
			run: func(db *sql.Tx) (err error) {
				stats, err := geq.SelectAs(&d.PostStats{
					AuthorID:  d.Posts.AuthorID,
					PostCount: geq.Count(d.Posts.ID),
					LastTitle: geq.Max(d.Posts.Title),
				}).From(d.Posts).
					GroupBy(d.Posts.AuthorID).
					OrderBy(d.Posts.AuthorID).
					Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(stats, []mdl.PostStat{
					{AuthorID: 1, PostCount: 2, LastTitle: "user1-post2"},
					{AuthorID: 2, PostCount: 1, LastTitle: "user2-post1"},
					{AuthorID: 3, PostCount: 3, LastTitle: "user3-post3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load as values",
			run: func(db *sql.Tx) (err error) {
				ids, err := geq.SelectOnly(d.Users.ID).From(d.Users).OrderBy(d.Users.ID).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(ids, []int64{1, 2, 3})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load as sql.Rows",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(geq.Raw("2-9"), d.Posts.AuthorID, geq.Max(d.Posts.ID).As("maxpostid")).
					From(d.Posts).
					GroupBy(d.Posts.AuthorID).
					OrderBy(d.Posts.AuthorID, geq.Raw("maxpostid"))
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
				err = assertEqual(results, [][]int{{-7, 1, 2}, {-7, 2, 3}, {-7, 3, 6}})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "scan into single slice",
			run: func(db *sql.Tx) (err error) {
				var users []mdl.User
				q := geq.SelectFrom(d.Users).OrderBy(d.Users.ID)
				err = q.WillScan(geq.ToSlice(d.Users, &users)).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(users, []mdl.User{
					{ID: 1, Name: "user1"},
					{ID: 2, Name: "user2"},
					{ID: 3, Name: "user3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "scan into single map",
			run: func(db *sql.Tx) (err error) {
				var userMap map[int64]mdl.User
				q := geq.SelectFrom(d.Users).OrderBy(d.Users.ID)
				err = q.WillScan(geq.ToMap(d.Users, d.Users.ID, &userMap)).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(userMap, map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "scan into single slice map",
			run: func(db *sql.Tx) (err error) {
				var postsMap map[int64][]mdl.Post
				q := geq.SelectFrom(d.Posts).OrderBy(d.Posts.ID)
				err = q.WillScan(geq.ToSliceMap(d.Posts, d.Posts.AuthorID, &postsMap)).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(postsMap, map[int64][]mdl.Post{
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
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "scan into multiple results",
			run: func(db *sql.Tx) (err error) {
				var posts []mdl.Post
				var userMap map[int64]mdl.User
				q := geq.SelectFrom(d.Posts).JoinRels(d.Posts.Author).OrderBy(d.Posts.ID)
				err = q.WillScan(
					geq.ToSlice(d.Posts, &posts),
					geq.ToMap(d.Posts.Author, d.Posts.Author.T().ID, &userMap),
				).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(posts, []mdl.Post{
					{ID: 1, AuthorID: 1, Title: "user1-post1"},
					{ID: 2, AuthorID: 1, Title: "user1-post2"},
					{ID: 3, AuthorID: 2, Title: "user2-post1"},
					{ID: 4, AuthorID: 3, Title: "user3-post1"},
					{ID: 5, AuthorID: 3, Title: "user3-post2"},
					{ID: 6, AuthorID: 3, Title: "user3-post3"},
				})
				if err != nil {
					return err
				}
				err = assertEqual(userMap, map[int64]mdl.User{
					1: {ID: 1, Name: "user1"},
					2: {ID: 2, Name: "user2"},
					3: {ID: 3, Name: "user3"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load non-builtin type fields",
			data: `
				INSERT INTO transactions (id, user_id, amount, created_at) VALUES
					(1, 1, 93, '2023-09-24 08:45:01'),
					(2, 1, 78, '2023-09-24 08:45:02')
			`,
			run: func(db *sql.Tx) (err error) {
				ts, err := geq.SelectFrom(d.Transactions).OrderBy(d.Transactions.CreatedAt).Load(ctx, db)
				if err != nil {
					return err
				}
				t1 := time.Date(2023, time.September, 24, 8, 45, 1, 0, time.UTC)
				err = assertEqual(ts, []mdl.Transaction{
					{ID: 1, UserID: 1, Amount: 93, CreatedAt: t1},
					{ID: 2, UserID: 1, Amount: 78, CreatedAt: t1.Add(time.Second)},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "load non-table rows with non-builtin type fields",
			data: `
				INSERT INTO transactions (id, user_id, amount, created_at) VALUES
					(1, 1, 120, '2023-07-12 08:45:01'),
					(2, 1, 31, '2023-07-12 08:45:02')
			`,
			run: func(db *sql.Tx) (err error) {
				stats, err := geq.SelectAs(&d.TransactionStats{
					UserID:        d.Transactions.UserID,
					TotalAmount:   geq.Sum(d.Transactions.Amount),
					LastCreatedAt: geq.Max(d.Transactions.CreatedAt),
				}).From(d.Transactions).GroupBy(d.Transactions.UserID).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(stats, []mdl.TransactionStat{
					{
						UserID:        1,
						TotalAmount:   151,
						LastCreatedAt: time.Date(2023, time.July, 12, 8, 45, 2, 0, time.UTC),
					},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select from table",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Users)
				err = assertQuery(q, "SELECT users.id, users.name FROM users")
				if err != nil {
					return err
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
				q := geq.SelectVia(users, d.Posts, d.Posts.Author)
				err = assertQuery(q, sjoin(
					"SELECT posts.id, posts.author_id, posts.title FROM posts",
					"WHERE posts.author_id IN (?, ?)",
				), int64(2), int64(3))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select with join using relationship",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Posts).JoinRels(d.Posts.Author).OrderBy(d.Posts.AuthorID)
				err = assertQuery(q, sjoin(
					"SELECT posts.id, posts.author_id, posts.title FROM posts",
					"INNER JOIN users AS posts_users ON posts.author_id = posts_users.id",
					"ORDER BY posts.author_id",
				))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select expressions",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					d.Users.ID,
					d.Users.Name.As("foo"),
					d.Posts.ID.Eq(3),
					d.Posts.Title.Eq("title").As("bar"),
					geq.Null(),
				)
				err = assertQuery(q, sjoin(
					"SELECT users.id, users.name AS foo, posts.id = ?, posts.title = ? AS bar, NULL",
				), 3, "title")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select logical expressions",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					d.Users.ID.IsNull().And(d.Users.Name.IsNotNull()),
					d.Users.ID.Eq(5).Or(d.Users.ID.Eq(7)),
					d.Users.ID.IsNotNull().And(
						d.Users.ID.Eq(2).Or(d.Users.ID.Gt(4).And(d.Users.ID.Lt(8))),
					),
				)
				err = assertQuery(q, sjoin(
					"SELECT users.id IS NULL AND users.name IS NOT NULL,",
					"users.id = ? OR users.id = ?,",
					"users.id IS NOT NULL AND (users.id = ? OR users.id > ? AND users.id < ?)",
				), 5, 7, 2, 4, 8)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select basic operations",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					d.Users.ID.Eq(3),
					d.Users.ID.Neq(3),
					d.Users.ID.Gt(4),
					d.Users.ID.Gte(4),
					d.Users.ID.Lt(4),
					d.Users.ID.Lte(4),
					d.Users.ID.Add(5),
					d.Users.ID.Sbt(5),
					d.Users.ID.Mlt(5),
					d.Users.ID.Dvd(5),
					d.Users.ID.IsNull(),
					d.Users.ID.IsNotNull(),
				)
				err = assertQuery(q, sjoin(
					"SELECT users.id = ?, users.id <> ?,",
					"users.id > ?, users.id >= ?, users.id < ?, users.id <= ?,",
					"users.id + ?, users.id - ?, users.id * ?, users.id / ?,",
					"users.id IS NULL, users.id IS NOT NULL",
				), 3, 3, 4, 4, 4, 4, 5, 5, 5, 5)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select LIKE expressions",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					d.Users.Name.LikePrefix("user"),
					d.Users.Name.LikeSuffix(d.Users.ID),
					d.Users.Name.LikePartial("s"),
				)
				err = assertQuery(q, sjoin(
					"SELECT users.name LIKE ? || '%', users.name LIKE '%' || users.id,",
					"users.name LIKE '%' || ? || '%'",
				), "user", "s")
				if err != nil {
					return err
				}

				users, err := geq.SelectFrom(d.Users).Where(d.Users.Name.LikePrefix("user")).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(len(users), 3)
				if err != nil {
					return err
				}

				users, err = geq.SelectFrom(d.Users).Where(d.Users.Name.LikeSuffix(d.Users.ID)).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(len(users), 3)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			name: "select IN expressions",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					d.Users.ID.InAny(1, "2", "str"),
					d.Users.Name.InAny(geq.Select(geq.Raw("'user'"))),
				)
				err = assertQuery(q,
					"SELECT users.id IN (?, ?, ?), users.name IN ((SELECT 'user'))",
					1, "2", "str",
				)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select function calls",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					geq.Count(d.Users.ID),
					geq.Max(d.Users.Name),
					geq.Func("MYFUNC", 1, d.Users.ID),
				)
				err = assertQuery(q, "SELECT COUNT(users.id), MAX(users.name), MYFUNC(?, users.id)", 1)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "handle expression precedences",
			run: func(db *sql.Tx) (err error) {
				a, b := geq.Raw("3"), geq.Raw("5")
				q := geq.Select(
					a.Mlt(b).Add(a),
					a.Mlt(b.Add(a)),
					a.Add(b).Mlt(b),
					geq.Parens(a.Add(b)).Mlt(b),
					a.Eq(b).And(b.Eq(a)),
					a.Add(b).Gt(b.Eq(a)),
				)
				err = assertQuery(q, sjoin(
					"SELECT 3 * 5 + 3,",
					"3 * (5 + 3),",
					"3 + 5 * 5,",
					"(3 + 5) * 5,",
					"3 = 5 AND 5 = 3,",
					"3 + 5 > (5 = 3)",
				))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select distinct",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectOnly(d.Posts.AuthorID).Distinct().From(d.Posts).OrderBy(d.Posts.AuthorID)
				err = assertQuery(q, "SELECT DISTINCT posts.author_id FROM posts ORDER BY posts.author_id")
				if err != nil {
					return err
				}
				authorIDs, err := q.Load(ctx, db)
				err = assertEqual(authorIDs, []int64{1, 2, 3})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "distinct in aggregate functions",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(
					geq.Count(d.Users.ID).Distinct(),
					geq.Max(d.Users.Name).Distinct(),
				)
				err = assertQuery(q, "SELECT COUNT(DISTINCT users.id), MAX(DISTINCT users.name)")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select with grouping",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Users, d.Users.Name, geq.Max(d.Users.ID)).GroupBy(d.Users.Name)
				err = assertQuery(q, "SELECT users.name, MAX(users.id) FROM users GROUP BY users.name")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select with grouping and having",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Users, d.Users.Name, geq.Max(d.Users.ID)).
					GroupBy(d.Users.Name).
					Having(geq.Count(d.Users.ID).Gt(1))
				err = assertQuery(q, sjoin(
					"SELECT users.name, MAX(users.id) FROM users",
					"GROUP BY users.name",
					"HAVING COUNT(users.id) > ?",
				), 1)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select with limit",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Users, d.Users.ID).Limit(2)
				err = assertQuery(q, "SELECT users.id FROM users LIMIT 2")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select with offset",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Users, d.Users.ID).Offset(2)
				err = assertQuery(q, "SELECT users.id FROM users OFFSET 2")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "select from sub query",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(geq.Raw("t.id"), geq.Raw("t.title")).From(geq.SelectFrom(d.Posts).As("t"))
				err = assertQuery(q, sjoin(
					"SELECT t.id, t.title FROM (SELECT posts.id, posts.author_id, posts.title FROM posts) AS t",
				))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "use queries as expressions",
			run: func(db *sql.Tx) (err error) {
				one, two := geq.Raw("1"), geq.Raw("2")
				q := geq.Select(geq.Select(one), geq.Select(two).As("two")).Where(two.Gte(geq.Select(one).Add(one)))
				err = assertQuery(q, "SELECT (SELECT 1), (SELECT 2) AS two WHERE 2 >= (SELECT 1) + 1")
				if err != nil {
					return err
				}
				rows, err := q.LoadRows(ctx, db)
				if err != nil {
					return err
				}
				var ret [2]int
				rows.Next()
				rows.Scan(&ret[0], &ret[1])
				err = assertEqual(ret, [2]int{1, 2})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "use raw expression in anywhere",
			run: func(db *sql.Tx) (err error) {
				q := geq.Select(geq.Raw("1"), geq.Raw("u.id, u.name as foo")).
					From(geq.Raw("users aS u")).
					Where(geq.Raw("1=1 and u.name is not null"), geq.Raw("u.id").InAny(1, 2)).
					OrderBy(geq.Raw("u.id Desc"))
				err = assertQuery(q, sjoin(
					"SELECT 1, u.id, u.name as foo FROM users aS u",
					"WHERE 1=1 and u.name is not null AND u.id IN (?, ?) ORDER BY u.id Desc",
				), 1, 2)
				if err != nil {
					return err
				}
				rows, err := q.LoadRows(ctx, db)
				if err != nil {
					return err
				}
				type row struct {
					N    int
					ID   int
					Name string
				}
				ret := make([]row, 2)
				for i := 0; rows.Next(); i++ {
					rows.Scan(&ret[i].N, &ret[i].ID, &ret[i].Name)
				}
				err = assertEqual(ret, []row{
					{N: 1, ID: 2, Name: "user2"},
					{N: 1, ID: 1, Name: "user1"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "inner join",
			data: `
                INSERT INTO users VALUES (4, 'no-posts-user');
            `,
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectOnly(d.Users.ID).From(d.Users).Distinct().InnerJoin(
					d.Posts,
					d.Posts.Title.LikePrefix(geq.Concat("user", d.Users.ID)),
				).OrderBy(d.Users.ID)
				err = assertQuery(q, sjoin(
					"SELECT DISTINCT users.id FROM users",
					"INNER JOIN posts ON posts.title LIKE ? || users.id || '%'",
					"ORDER BY users.id",
				), "user")
				if err != nil {
					return err
				}
				userIDs, err := q.Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(userIDs, []int64{1, 2, 3})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "left join",
			data: `
                INSERT INTO users VALUES (4, 'no-posts-user');
            `,
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectOnly(d.Users.ID).From(d.Users).LeftJoin(
					d.Posts,
					d.Posts.AuthorID.Eq(d.Users.ID),
				).Where(d.Posts.ID.IsNull())
				err = assertQuery(q, sjoin(
					"SELECT users.id FROM users",
					"LEFT JOIN posts ON posts.author_id = users.id",
					"WHERE posts.id IS NULL",
				))
				if err != nil {
					return err
				}
				userIDs, err := q.Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(userIDs, []int64{4})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "right join",
			data: `
                INSERT INTO users VALUES (4, 'no-posts-user');
            `,
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectOnly(d.Users.ID).Distinct().From(d.Posts).RightJoin(
					d.Users,
					d.Users.ID.Eq(d.Posts.AuthorID),
				).Where(d.Posts.ID.IsNull())
				err = assertQuery(q, sjoin(
					"SELECT DISTINCT users.id FROM posts",
					"RIGHT JOIN users ON users.id = posts.author_id",
					"WHERE posts.id IS NULL",
				))
				if err != nil {
					return err
				}
				userIDs, err := q.Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(userIDs, []int64{4})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "cross join",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectOnly(d.Users.ID).From(d.Users).CrossJoin(d.Posts)
				err = assertQuery(q, "SELECT users.id FROM users CROSS JOIN posts")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "order by desc",
			run: func(db *sql.Tx) (err error) {
				q := geq.SelectFrom(d.Posts).OrderBy(d.Posts.AuthorID.Desc(), d.Posts.ID.Asc())
				err = assertQuery(q, sjoin(
					"SELECT posts.id, posts.author_id, posts.title FROM posts",
					"ORDER BY posts.author_id DESC, posts.id",
				))
				if err != nil {
					return err
				}
				posts, err := q.Load(ctx, db)
				if err != nil {
					return err
				}
				ids := make([][2]int64, 0, len(posts))
				for _, p := range posts {
					ids = append(ids, [2]int64{p.AuthorID, p.ID})
				}
				err = assertEqual(ids, [][2]int64{{3, 4}, {3, 5}, {3, 6}, {2, 3}, {1, 1}, {1, 2}})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "build insert query by values and value maps",
			run: func(db *sql.Tx) (err error) {
				q := geq.InsertInto(d.Users).
					Values(
						d.Users.ID.Set(1),
						d.Users.Name.Set("name"),
					).
					ValueMaps(
						geq.ValueMap{
							d.Users.ID:   2,
							d.Users.Name: nil,
						},
						geq.ValueMap{
							d.Users.ID:   "invalid-id",
							d.Users.Name: geq.Func("NOW"),
						},
					)
				err = assertQuery(q, sjoin(
					"INSERT INTO users (id, name) VALUES (?, ?), (?, NULL), (?, NOW())",
				), int64(1), "name", 2, "invalid-id")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "insert records",
			run: func(db *sql.Tx) (err error) {
				q := geq.InsertInto(d.Users).
					Values(
						d.Users.ID.Set(100),
						d.Users.Name.Set("name100"),
					).
					ValueMaps(
						geq.ValueMap{
							d.Users.ID:   200,
							d.Users.Name: "name200",
						},
					)
				ret, err := q.Exec(ctx, db)
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
				q := geq.Update(d.Posts).Set(
					d.Posts.AuthorID.Set(1),
					d.Posts.Title.Set("title"),
				).Where(d.Posts.ID.In([]int64{3, 4}))
				err = assertQuery(q, sjoin(
					"UPDATE posts SET author_id = ?, title = ? WHERE posts.id IN (?, ?)",
				), int64(1), "title", int64(3), int64(4))
				if err != nil {
					return err
				}
				_, err = q.Exec(ctx, db)
				if err != nil {
					return err
				}
				posts, err := geq.SelectFrom(d.Posts).
					Where(d.Posts.ID.In([]int64{3, 4})).
					OrderBy(d.Posts.ID).
					Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(posts, []mdl.Post{
					{ID: 3, AuthorID: 1, Title: "title"},
					{ID: 4, AuthorID: 1, Title: "title"},
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "delete records",
			run: func(db *sql.Tx) (err error) {
				q := geq.DeleteFrom(d.Users).Where(d.Users.ID.In([]int64{1, 3}))
				err = assertQuery(q, "DELETE FROM users WHERE users.id IN (?, ?)", int64(1), int64(3))
				if err != nil {
					return err
				}
				_, err = q.Exec(ctx, db)
				if err != nil {
					return err
				}
				ids, err := geq.SelectOnly(d.Users.ID).From(d.Users).OrderBy(d.Users.ID).Load(ctx, db)
				if err != nil {
					return err
				}
				err = assertEqual(ids, []int64{2})
				if err != nil {
					return err
				}
				return nil
			},
		},
	})
}
