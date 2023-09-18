package main

import (
	"fmt"

	"github.com/ryym/geq"
)

func main() {
	q := geq.Select(geq.Raw("1"), geq.Raw("u.id, u.name as foo")).
		From(geq.Raw("users as u")).
		Where(geq.Raw("u.id").Eq(1))
	sql, err := q.Build()
	fmt.Println(sql.Query, sql.Args, err)
}
