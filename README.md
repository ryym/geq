🚧 WIP

# Geq

Yet another SQL query builder for Go, with moderate type safety powered by generics and code generation.

## Features

- SQL friendly (query builder rather than ORM)
- Performative (no runtime reflections)

Unsupported:

- Schema migration
- Fixture file loading

## Quick start

### Hello world

As a first step, you can try using Geq without any database or code generation.

```bash
mkdir geqsample && cd geqsample
go mod init example.com/geqsample
```

`main.go`:

```go
package main

import (
	"fmt"

	"github.com/ryym/geq"
)

func main() {
	q := geq.Select(geq.Raw("*")).From(geq.Raw("users")).Where(geq.Raw("name").Eq("foo"))
	sql, err := q.Build()
	fmt.Println(sql.Query, sql.Args, err)
}
```

```bash
% go mod tidy
% go run .
SELECT * FROM users WHERE name = ? [foo] <nil>
```

This may be somewat useful already, but defining a data model makes it more convenient and safe for writing queries.

### Define data model

Define a struct that corresponds to records in your database tables, and write it in `geqbld.go` .
This is a configuration file for Geq.

`geqbld.go`:

```go
package main

import "example.com/geqsample/mdl"

type GeqTables struct {
	Users mdl.User
}
```

`mdl/models.go`:

```go
package mdl

type User struct {
    ID   uint64
    Name string
}
```

```bash
go install github.com/ryym/geq/cmd/geq@latest
geq .
```

The above command generates a query helper package in `./d` by default.
Now you can rewrite the query in `main.go` like this:

```diff
  import (
  	"fmt"
  
+ 	"example.com/geqsample/d"
  	"github.com/ryym/geq"
  )
 func main() {
-	q := geq.Select(geq.Raw("*")).From(geq.Raw("users")).Where(geq.Raw("name").Eq("foo"))
+	q := geq.SelectFrom(d.Users).Where(d.Users.Name.Eq("foo"))
 	sql, err := q.Build()
 	fmt.Println(sql.Query, sql.Args, err)
 }
```

```bash
% go run .
SELECT users.id, users.name FROM users WHERE users.name = ? ["foo"] <nil>
```

### Run query

Finally you can actually execute the query and retrieve records once you have a database corresponding to the definitions in `geqbld.go` .
Let's try it with PostgreSQL.

`docker-compose.yml`:

```yml
version: '3'
services:
  pg:
    image: postgres:15.4
    ports:
      - '5499:5432'
    environment:
      - POSTGRES_USER=geqsample
      - POSTGRES_PASSWORD=geqsample
```

`main.go`:

```go
package main

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/geqsample/d"
	_ "github.com/lib/pq"
	"github.com/ryym/geq"
)

const initSQL = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (id serial NOT NULL, name varchar(30) NOT NULL);
INSERT INTO users VALUES (1, 'foo'), (2, 'bar'), (3, 'foo');
`

func main() {
	// Connect to DB.
	db, err := sql.Open("postgres", "port=5499 user=geqsample password=geqsample sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Prepare the data.
	_, err = db.Exec(initSQL)
	if err != nil {
		panic(err)
	}

	// Specify the database type for query building.
	geq.SetDefaultDialect(&geq.DialectPostgres{})

	// Write a query.
	q := geq.SelectFrom(d.Users).Where(d.Users.Name.Eq("foo")).OrderBy(d.Users.ID)

	// Load the data.
	users, err := q.Load(context.Background(), db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#+v\n", users) // users: []mdl.User
}
```

```bash
% docker-compose up --build -d
% go mod tidy
% go run .
[]mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x3, Name:"foo"}}
```

# Guides

## Table relationships management

Optionally you can define and utilize table relationships.

`mdl/models.go`:

```go
package mdl

type User struct {
	ID   uint64
	Name string
}

// User has-many Posts (users.id = posts.author_id)
type Post struct {
	ID       uint64
	AuthorID uint64
	Title    string
}
```

`geqbld.go`:

```go
package main

import "example.com/geqsample/mdl"

type GeqTables struct {
	Users mdl.User
	Posts mdl.Post
}

// Define table relationships here.
type GeqRelationships struct {
	Users struct {
		Posts mdl.Post `geq:"Users.ID = Posts.AuthorID"`
	}
	Posts struct {
		Author mdl.User `geq:"Posts.AuthorID = Users.ID"`
	}
}
```

The relationship definitions make it easier to build join queries or load relevant data.

```go
// Use in data loading.
geq.SelectFrom(d.Posts).Where(d.Posts.Author.In(users))

// Use in table join.
geq.SelectFrom(d.Users).Joins(d.Users.Posts).Where(d.Users.Posts.T().Title.Eq(""))
```

### Relevant models retrieval

Unlike typical ORM libraries, Geq does not support nested relation loading such as below:

```go
type User struct {
    ID    uint64
    Name  string
    Posts []mdl.Post // NOT supported
}
```

Nested relation loading is powerful but sometimes make things so complicated.
Instead, you can load relevant records in two ways:

#### Load by one query

```go
var posts []mdl.Post
var userMap map[int64]mdl.User
q := geq.SelectFrom(d.Posts).Joins(d.Posts.Author).OrderBy(d.Posts.ID)
err = q.WillScan(
	geq.ToSlice(d.Posts, &posts),
	geq.ToMap(d.Posts.Author, d.Posts.Author.T().ID, &userMap),
).Load(ctx, db)

for _, p := range posts {
	author := userMap[p.AuthorID]
	fmt.Println(p, author)
}
```

- It requires only one round-trip to the database.
- It may load duplicate records if the relationship is not 1:1.

#### Load individually

```go
users, err := geq.SelectFrom(d.Users).OrderBy(d.Users.ID).Limit(50).Load(ctx, db)
postsMap, err := geq.AsSliceMap(
	d.Posts.ID,
	geq.SelectFrom(d.Posts).Where(d.Posts.Author.In(users)).OrderBy(d.Posts.ID),
).Load(ctx, db)

for _, u := range users {
	posts := postsMap[u.ID]
	fmt.Println(u, posts)
}
```

- It requires multiple round-trips to the database.
- It retrieves records without duplicate due to table joining.

## Non-table result mapping

When you want to load rows not corresponding to database tables, you generate row mappers.

```go
package mdl

type PostStat struct {
	AuthorID  int64
	PostCount int64
}
```

Define `GeqMappers` in `geqbld.go`:

```go
package main

// ...

type GeqMappers struct {
    PostStats mdl.PostStat
}
```

```bash
# Re-generate your query helper with row mappers.
geq .
```

Then you can load results into `mdl.PostStat` using `SelectAs` .

```go
q := geq.SelectAs(&d.PostStats{
	// Specify what you want to load for each field.
	AuthorID: d.Posts.AuthorID,
	PostCount: geq.Count(d.Posts.ID),
}).From(d.Posts).GroupBy(d.Posts.AuthorID)
stats, err := q.Load(ctx, db) // stats: []mdl.PostStat
```

Or when you want to select single values, use `SelectOnly` .

```go
// userIDs: []uint64
userIDs, err := geq.SelectOnly(d.Users.ID).OrderBy(d.Users.ID).Load(ctx, db)
```

## Result mapping patterns

### `Load` - Load as slice

```go
users, err := geq.SelectFrom(d.Users).OrderBy(d.Users.ID).Load(ctx, db)
fmt.Printf("%#+v\n", users)
//=> []mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x2, Name:"bar"}, mdl.User{ID:0x3, Name:"foo"}}
```

### `AsMap(key, query).Load` - Load as map

```go
userMap, err := geq.AsMap(d.Users.ID, geq.SelectFrom(d.Users)).Load(ctx, db)
fmt.Printf("%#+v\n", userMap)
//=> map[uint64]mdl.User{0x1:mdl.User{ID:0x1, Name:"foo"}, 0x2:mdl.User{ID:0x2, Name:"bar"}, 0x3:mdl.User{ID:0x3, Name:"foo"}}
```

### `AsSliceMap(key, query).Load` - Load as map of slices

```go
usersMap, err := geq.AsSliceMap(d.Users.Name, geq.SelectFrom(d.Users)).Load(ctx, db)
fmt.Printf("%#+v\n", usersMap)
//=> map[string][]mdl.User{"bar":[]mdl.User{mdl.User{ID:0x2, Name:"bar"}}, "foo":[]mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x3, Name:"foo"}}}
```

### `LoadRows` - Load as `*sql.Rows`

```go
rows, err = geq.SelectFrom(d.Users).LoadRows(ctx, db)
```

### `WillScan(scans...).Load` - Scan into multiple results

```go
var users []mdl.User
var postsMap map[uint64][]mdl.Post
err := geq.
	SelectFrom(d.Users).
	Joins(d.Users.Posts).
	OrderBy(d.Users.ID, d.Users.Posts.T().ID).
	WillScan(
		d.ToSlice(d.Users, &users),
		d.ToSliceMap(d.Users.Posts, d.Users.Posts.T().AuthorID, &postsMap),
	).
	Load(ctx, db)
fmt.Printf("%#+v\n", users)
//=> []mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x2, Name:"bar"}, mdl.User{ID:0x3, Name:"foo"}}
fmt.Printf("%#+v\n", postsMap)
//=. map[uint64][]mdl.Post{0x1:[]mdl.Post{mdl.Post{ID:0x1, AuthorID:0x1, Title:"user1-post1"}, mdl.Post{ID:0x2, AuthorID:0x1, Title:"user1-post2"}}, 0x2:[]mdl.Post{mdl.Post{ID:0x3, AuthorID:0x2, Title:"user2-post1"}}, 0x3:[]mdl.Post{mdl.Post{ID:0x4, AuthorID:0x3, Title:"user3-post1"}}}
```

## `SELECT` patterns

### `SelectFrom` - Select records from table

```go
users, err := geq.SelectFrom(d.Users).Load(ctx, db)
```

### `SelectAs` - Select non-table record results

```go
stats, err := geq.SelectAs(&d.PostStats{
	AuthorID: d.Posts.AuthorID,
	PostCount: geq.Count(d.Posts.ID),
}).From(d.Posts).GroupBy(d.Posts.AuthorID).Load(ctx, db)
```

### `SelectOnly` - Select single values

```go
userIDs, err := geq.SelectOnly(d.Users.ID).OrderBy(d.Users.ID).Load(ctx, db)

// Currently non-column expressions are not supported.
// _, _ = geq.SelectOnly(geq.Count(d.Users.ID)).From(d.Users)
```

### `SelectVia` - Select records via table relationships

```go
users, err := geq.SelectFrom(d.Users)
posts, err := geq.SelectVia(users, d.Posts, d.Posts.Author).Load(ctx, db)

// The above query is a shorthand of this query:
posts, err := geq.SelectFrom(d.Posts).Where(d.Posts.Author.In(users)).Load(ctx, db)
```

### `Select` - Select arbitrary values

```go
// Mainly used for sub queries.
geq.SelectAs(&d.Hoge{
    AuthorName: d.Users.Name,
    IsPro: geq.Exists(geq.Select().From(d.Prof))
}).From(d.Users).Where(
    d.Users.ID.InAny(geq.Select(d.Posts.AuthorID).From(d.Posts)),
).GroupBy(d.Users.Name)
```
