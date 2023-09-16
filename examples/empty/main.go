package main

import (
	"fmt"

	"github.com/ryym/geq/examples/empty/b"
)

func main() {
	q := b.Select(b.Raw("1"))
	fmt.Println(q.Build())
}
