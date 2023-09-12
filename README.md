(WIP)

# Geq

Yet another SQL query builder for Go, with moderate type safety powered by generics and code generation.

- SQL friendly (Not ORM)
- Performative (No runtime reflections)

```go
package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ryym/geq/examples/helloworld/b"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost)")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Load records as a slice.
	users, err := b.SelectFrom(b.Users).OrderBy(b.Users.ID).Load(ctx, db)
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		fmt.Println(u.ID, u.Name)
	}

	// Or Load as other forms, like a map of slices.
	q := b.SelectFrom(b.Posts).Where(b.Posts.Author.In(users))
	postsMap, err := b.AsSliceMap(b.Posts.AuthorID, q).Load(ctx, db)
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		posts := postsMap[u.ID]
		if len(posts) > 0 {
			fmt.Println(posts, posts[0].Title)
		}
	}

	// Or Scan into multiple results by one query.
	err = b.SelectFrom(b.Users).Joins(b.Users.Posts).OrderBy(b.Users.ID).WillScan(
		b.ToSlice(b.Users, &users),
		b.ToSliceMap(b.Posts, b.Posts.AuthorID, &postsMap),
	).Load(ctx, db)
	if err != nil {
		panic(err)
	}
}
```


