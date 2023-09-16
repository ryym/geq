package main

import (
	"fmt"

	"github.com/ryym/geq/examples/empty/b"
)

func main() {
	q := b.Select(b.Raw("1"), b.Raw("u.id, u.name as foo")).
		From(b.Raw("users as u")).
		Where(b.Raw("u.id").Eq(1))
	sql, err := q.Build()
	fmt.Println(sql.Query, sql.Args, err)
}
