🚧 WIP

# Geq

Yet another SQL query builder for Go, with moderate type safety powered by generics and code generation.

## Features

- SQL friendly query builder (not ORM)
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
go get github.com/ryym/geq
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
% go run .
SELECT * FROM users WHERE name = ? [foo] <nil>
```

This may be somewat useful already, but defining a data model makes it more convenient and safe for writing queries.

### Define data model

Define a struct that corresponds to records in your database tables, and write it in `geqbld.go` .
This is a configuration file for Geq.

`mdl/models.go`:

```go
package mdl

type User struct {
    ID   uint64
    Name string
}
```

`geqbld.go`:

```go
package main

import "example.com/geqsample/mdl"

type GeqTables struct {
	Users mdl.User
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

```go
const initSQL = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (id serial NOT NULL, name varchar(30) NOT NULL);
INSERT INTO users VALUES (1, 'foo'), (2, 'bar'), (3, 'foo');

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (id serial NOT NULL, author_id int NOT NULL, title varchar(30) NOT NULL);
INSERT INTO posts (id, author_id, title) VALUES
  (1, 1, 'post1'),
  (2, 1, ''),
  (3, 2, 'Go is nice'),
  (4, 3, 'Programming is fun');
`
```

The relationship definitions make it easier to build join queries or load relevant data.

```go
// Use in table join.
_, _ = geq.SelectFrom(b.Users).Joins(b.Users.Posts).Where(b.Posts.Title.Eq("")).Build()

// Use in data loading.
posts, err := geq.SelectFrom(b.Posts).Where(b.Posts.Author.In(users)).Load(ctx, db)
fmt.Println(posts, err)
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

#### Load individually

```go
users, err := geq.SelectFrom(b.Users).OrderBy(b.Users.ID).Limit(50).Load(ctx, db)
postsMap, err := geq.AsSliceMap(
    b.Posts.ID,
    geq.SelectFrom(b.Posts).Where(b.Posts.Author.In(users)).OrderBy(b.Posts.ID).Load(ctx, db),
).Load(ctx, db)

for _, u := range users {
    fmt.Println(u, postsMap[u.ID])
}
```

- It requires multiple round-trips to the database.
- It retrieves records without duplicate due to table joining.

#### Load by one query

```go
var users []mdl.User
var postsMap map[uint64][]mdl.Post
q := geq.SelectFrom(b.Users).Joins(b.Users.Posts).OrderBy(b.Users.ID, b.Posts.ID)
err := q.WillScan(
    geq.ToSlice(b.Users, &users),
    geq.ToSliceMap(b.Posts, b.Posts.AuthorID, &postsMap),
).Load(ctx, db)

for _, u := range users {
    fmt.Println(u, postsMap[u.ID])
}
```

- It requires only one round-trip to the database.
- It may load duplicate records if the relationship is not 1:1.

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

Then you can load results into `mdl.PostStat` .

```go
q := geq.SelectAs(&b.PostStats{
	AuthorID: b.Posts.AuthorID,
	PostCount: b.Count(b.Posts.ID),
}).From(b.Posts).GroupBy(b.Posts.AuthorID)
stats, err := q.Load(ctx, db) // stats: []mdl.PostStat
```

Or you can select single values:

```go
// userIDs: []uint64
userIDs, err := geq.SelectOnly(b.Users.ID).OrderBy(b.Users.ID).Load(ctx, db)

// Currently non-column expressions are not supported.
// _, _ = geq.SelectOnly(geq.Count(b.Users.ID)).From(b.Users)
```

----

# Usage

We use these sample models in the following guides:

```go
package mdl

type User struct {
	ID   uint64
	Name string
}

type Post struct {
	ID       uint64
	AuthorID uint64
	Title    string
}
```

## Result mapping

### Load as `*sql.Rows`

```go
q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
rows, err := q.LoadRows(context.Background(), db)
```

### Load as slice

```go
q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
users, err := q.Load(context.Background(), db)
fmt.Printf("%#+v\n", users)
//=> []mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x2, Name:"bar"}, mdl.User{ID:0x3, Name:"foo"}}
```

### Load as map

```go
q := b.SelectFrom(b.Users)
userMap, err := b.AsMap(b.Users.ID, q).Load(context.Background(), db)
fmt.Printf("%#+v\n", userMap)
//=> map[uint64]mdl.User{0x1:mdl.User{ID:0x1, Name:"foo"}, 0x2:mdl.User{ID:0x2, Name:"bar"}, 0x3:mdl.User{ID:0x3, Name:"foo"}}
```

### Load as map of slices

```go
q := b.SelectFrom(b.Users)
usersMap, err := b.AsSliceMap(b.Users.Name, q).Load(context.Background(), db)
fmt.Printf("%#+v\n", usersMap)
//=> map[string][]mdl.User{"bar":[]mdl.User{mdl.User{ID:0x2, Name:"bar"}}, "foo":[]mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x3, Name:"foo"}}}
```

### Scan into multiple results

```go
var users []mdl.User
var postsMap map[uint64][]mdl.Post
q := b.SelectFrom(b.Users).Joins(b.Users.Posts).OrderBy(b.Users.ID, b.Posts.ID)
err := q.WillScan(
    b.ToSlice(b.Users, &users),
    b.ToSliceMap(b.Posts, b.Posts.AuthorID, &postsMap),
).Load(context.Background(), db)
fmt.Printf("%#+v\n", users)
//=> []mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x2, Name:"bar"}, mdl.User{ID:0x3, Name:"foo"}}
fmt.Printf("%#+v\n", postsMap)
//=. map[uint64][]mdl.Post{0x1:[]mdl.Post{mdl.Post{ID:0x1, AuthorID:0x1, Title:"user1-post1"}, mdl.Post{ID:0x2, AuthorID:0x1, Title:"user1-post2"}}, 0x2:[]mdl.Post{mdl.Post{ID:0x3, AuthorID:0x2, Title:"user2-post1"}}, 0x3:[]mdl.Post{mdl.Post{ID:0x4, AuthorID:0x3, Title:"user3-post1"}}}
```

## Querying

### Select records from table

```go
users, err := b.SelectFrom(b.Users).Load(ctx, db)
```

### Select single values

### Select records via table relationships

### Select non-table record results

