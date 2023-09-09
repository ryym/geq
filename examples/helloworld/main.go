package main

import (
	"fmt"

	"github.com/ryym/geq/examples/helloworld/reports"
	"github.com/ryym/geq/tests/b"
)

func main() {
	fmt.Println("helloworld")

	q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
	// users, err := q.Load(ctx, db)
	fmt.Println(q.Finalize().Query)

	fmt.Println(reports.PostStatsQuery())
	fmt.Println(reports.SameNameUsersQuery())
}
