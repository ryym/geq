package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryym/geq/internal/codegen"
)

func main() {
	flag.Parse()

	cfg := &codegen.Config{
		RootPath: flag.Args()[0],
	}

	err := codegen.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
