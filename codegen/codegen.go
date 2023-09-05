package codegen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type Config struct {
	RootPath string
}

func Run(cfg *Config) (err error) {
	bldPaths := make([]string, 0)
	rowsPaths := make([]string, 0)
	filepath.Walk(cfg.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch info.Name() {
		case "geqbld.go":
			bldPaths = append(bldPaths, fmt.Sprintf("./%s", filepath.Dir(path)))
		case "geqrows.go":
			rowsPaths = append(rowsPaths, fmt.Sprintf("./%s", filepath.Dir(path)))
		}
		return nil
	})

	err = genBuilderPkgs(bldPaths)
	if err != nil {
		return err
	}
	err = genRowsFiles(rowsPaths)
	if err != nil {
		return err
	}

	return nil
}

func genBuilderPkgs(bldPaths []string) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax}
	pkgs, err := loadPkgs(pkgCfg, bldPaths...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		fmt.Println("geqbld.go in", pkg.ID)
	}
	return nil
}

func genRowsFiles(rowsPaths []string) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax}
	pkgs, err := loadPkgs(pkgCfg, rowsPaths...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		fmt.Println("geqrows.go in", pkg.ID)
	}
	return nil
}

func loadPkgs(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("failed to load packages")
	}
	return pkgs, nil
}
