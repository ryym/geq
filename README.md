(WIP)

# Geq

Yet another SQL query builder for Go, with moderate type safety powered by generics and code generation.

## Features

- SQL friendly
- Performative (No runtime reflections)

Unsupported:

- Schema migration
- Fixture file loading

## Quick start

### Hello, world

Geq uses the code-first approach, so you need to create `geqbld.go` file first.

```bash
go install github.com/ryym/geq/cmd/geq@latest

mkdir geqsample && cd geqsample
go mod init example.com/geqsample
echo 'package main' > geqbld.go
geq . # Generate the query builder in geqsample/b.
go mod tidy
touch main.go
```

`main.go`:

```go
package main

import (
    "fmt"

    "example.com/geqsample/b"
)

func main() {
   q := b.Select(b.Raw("*")).From(b.Raw("users")).Where(b.Raw("id").Eq(1))
   sql, err := q.Build()
   fmt.Println(sql.Query, sql.Args, err)
}
```

```bash
% go run .
SELECT * FROM users WHERE id = ? [1] <nil>
```

### Define data model

Using code generation, you can write an equivalent query more type-safely.

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

`main.go`:

```diff
 func main() {
-	q := b.Select(b.Raw("*")).From(b.Raw("users")).Where(b.Raw("id").Eq(1))
+	q := b.SelectFrom(b.Users).Where(b.Users.ID.Eq(1))
 	sql, err := q.Build()
 	fmt.Println(sql.Query, sql.Args, err)
 }
```

```bash
% go run .
SELECT users.id, users.name FROM users WHERE users.id = ? [1] <nil>
```

### Run query

Now you can load actual DB records into your data models.

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

	"example.com/geqsample/b"
	_ "github.com/lib/pq"
	"github.com/ryym/geq"
)

const ddl = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (id serial NOT NULL, name varchar(30) NOT NULL);
INSERT INTO users VALUES (1, 'foo'), (2, 'bar'), (3, 'foo');
`

func main() {
	db, err := sql.Open("postgres", "port=5499 user=geqsample password=geqsample sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Prepare the data.
	_, err = db.Exec(ddl)
	if err != nil {
		panic(err)
	}

	// Specify the database type for query building.
	geq.SetDefaultDialect(&geq.DialectPostgres{})

	q := b.SelectFrom(b.Users).Where(b.Users.Name.Eq("foo")).OrderBy(b.Users.ID)

	// Load the data.
	users, err := q.Load(context.Background(), db)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#+v\n", users)
}
```

```bash
% docker-compose up --build -d
% go mod tidy
% go run .
[]mdl.User{mdl.User{ID:0x1, Name:"foo"}, mdl.User{ID:0x3, Name:"foo"}}
```

## Define table relationships

Optionally you can define and utilize table relationships.

`mdl/models.go`:

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

`geqbld.go`:

```go
package main

import "example.com/geqsample/mdl"

type GeqTables struct {
	Users mdl.User
	Posts mdl.Post
}

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
const ddl = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (id serial NOT NULL, name varchar(30) NOT NULL);
INSERT INTO users VALUES (1, 'foo'), (2, 'bar'), (3, 'foo');

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (id serial NOT NULL, author_id int NOT NULL, title varchar(30) NOT NULL);
INSERT INTO posts (id, author_id, title) VALUES
  (1, 1, 'user1-post1'),
  (2, 1, 'user1-post2'),
  (3, 2, 'user2-post1'),
  (4, 3, 'user3-post1');
`
```

```go
// Use in table join.
_, _ = b.SelectFrom(b.Users).Joins(b.Users.Posts).Where(b.Posts.Title.Eq("")).Build()

// Use in data loading.
posts, err := b.SelectVia(users, b.Posts, b.Posts.Author).Load(context.Background(), db)
fmt.Println(posts, err)
```

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

