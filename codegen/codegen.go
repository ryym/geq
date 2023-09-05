package codegen

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

type Config struct {
	RootPath string
}

func Run(cfg *Config) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax}
	pkgs, err := packages.Load(pkgCfg, cfg.RootPath)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return errors.New("failed to load packages")
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg.ID, "files:")
		for _, f := range pkg.GoFiles {
			fmt.Println(f)
		}
	}

	return nil
}
