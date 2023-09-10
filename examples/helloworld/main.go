package main

import (
	"fmt"

	"github.com/ryym/geq/examples/helloworld/b"
	"github.com/ryym/geq/examples/helloworld/reports"
)

func main() {
	fmt.Println("helloworld")

	q := b.SelectFrom(b.Users).OrderBy(b.Users.ID)
	// users, err := q.Load(ctx, db)
	fq, err := q.Finalize()
	if err != nil {
		fmt.Println(err)
	}
	// db.Query(fq.Query, fq.Args...)
	fmt.Println(fq)

	fmt.Println(reports.PostStatsQuery())
	fmt.Println(reports.SameNameUsersQuery())
}
