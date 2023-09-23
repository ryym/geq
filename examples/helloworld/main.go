package main

import (
	"fmt"

	"github.com/ryym/geq"
	"github.com/ryym/geq/examples/helloworld/gen/d"
	"github.com/ryym/geq/examples/helloworld/reports"
)

func main() {
	fmt.Println("helloworld")

	q := geq.SelectFrom(d.Users).OrderBy(d.Users.ID)
	// users, err := q.Load(ctx, db)
	bq, err := q.Build()
	if err != nil {
		fmt.Println(err)
	}
	// db.Query(bq.Query, bq.Args...)
	fmt.Println(bq)

	fmt.Println(reports.PostStatsQuery())
	fmt.Println(reports.SameNameUsersQuery())
}
