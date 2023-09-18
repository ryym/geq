🚧 WIP

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

Geq generates a query builder package based on a file named geqbld.go.
With just an empty file, You can try building queries even without an actual database.

```bash
# Install CLI.
go install github.com/ryym/geq/cmd/geq@latest

# Set up the sample package.
mkdir geqsample && cd geqsample
go mod init example.com/geqsample

# Generate the query builder (in ./b by default).
echo 'package main' > geqbld.go
geq .

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

This may be sufficient for using it as a query builder, but defining a data model will make it more convenient and safe for writing queries.

### Define data model

Define a struct that corresponds to records in your database tables, and write it in `geqbld.go` .

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
# Re-generate your query builder.
geq .
```

Now you can rewrite the query in `main.go` like this:

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

Once you have a database corresponding to the definitions in `geqbld.go` , you can actually execute queries and retrieve records.

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
	q := b.SelectFrom(b.Users).Where(b.Users.Name.Eq("foo")).OrderBy(b.Users.ID)

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

You can define and utilize table relationships.

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

The relationships are used to build join queries or relevant data loading.

```go
// Use in table join.
_, _ = b.SelectFrom(b.Users).Joins(b.Users.Posts).Where(b.Posts.Title.Eq("")).Build()

// Use in data loading.
posts, err := b.SelectFrom(b.Posts).Where(b.Posts.Author.In(users)).Load(ctx, db)
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

#### Load by one query using joins

```go
var users []mdl.User
var postsMap map[uint64][]mdl.Post
q := b.SelectFrom(b.Users).Joins(b.Users.Posts).OrderBy(b.Users.ID, b.Posts.ID)
err := q.WillScan(
    b.ToSlice(b.Users, &users),
    b.ToSliceMap(b.Posts, b.Posts.AuthorID, &postsMap),
).Load(ctx, db)
```

- It requires only one round-trip to the database.
- It may load duplicate records if the relationship is not 1:1.

#### Load individually

```go
users, err := b.SelectFrom(b.Users).OrderBy(b.Users.ID).Limit(50).Load(ctx, db)
postsMap, err := b.AsSliceMap(
    b.Posts.ID,
    b.SelectFrom(b.Posts).Where(b.Posts.Author.In(users)).OrderBy(b.Posts.ID).Load(ctx, db),
).Load(ctx, db)
```

- It requires multiple round-trips to the database.
- It retrieves records without duplicate due to table joining.

## Non-table result mapping

When you want to load rows not corresponding to database tables, you generate custom mappers.

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
# Re-generate your query builder with custom row mappers.
geq .
```

Then you can load results into `mdl.PostStat` .

```go
q := b.SelectAs(&b.PostStats{
	AuthorID: b.Posts.AuthorID,
	PostCount: b.Count(b.Posts.ID),
}).From(b.Posts).GroupBy(b.Posts.AuthorID)
stats, err := q.Load(ctx, db) // stats: []mdl.PostStat
```

Or you can select single values:

```go
// userIDs: []uint64
userIDs, err := b.SelectOnly(b.Users.ID).OrderBy(b.Users.ID).Load(ctx, db)

// Currently non-column expressions are not supported.
// _, _ = b.SelectOnly(b.Count(b.Users.ID)).From(b.Users)
```
