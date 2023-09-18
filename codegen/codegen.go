package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

const geqPkgPath = "github.com/ryym/geq"

type Config struct {
	RootPath string
}

func Run(cfg *Config) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	bldPaths := make([]string, 0)
	err = filepath.Walk(cfg.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "geqbld.go" {
			bldPaths = append(bldPaths, absPath(cwd, filepath.Dir(path)))
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = genBuildersFiles(bldPaths)
	if err != nil {
		return err
	}

	return nil
}

func genBuildersFiles(bldPaths []string) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax}
	for _, bldPath := range bldPaths {
		pkgs, err := loadPkgs(pkgCfg, bldPath)
		if err != nil {
			return err
		}
		pkg := pkgs[0]
		err = genBuilderFile(bldPath, pkg)
		if err != nil {
			return fmt.Errorf("builder generation failed at: %s: %w", pkg.ID, err)
		}
	}
	return nil
}
