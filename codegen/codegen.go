package codegen

import (
	"errors"
	"fmt"
	"go/types"
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
	pkgCfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedTypes}
	pkgs, err := loadPkgs(pkgCfg, bldPaths...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		fmt.Println("geqbld.go in", pkg.ID)
		err = genBuilderPkg(pkg)
		if err != nil {
			return fmt.Errorf("builder generation failed at: %s: %w", pkg.ID, err)
		}
	}
	return nil
}

func genRowsFiles(rowsPaths []string) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedCompiledGoFiles}
	pkgs, err := loadPkgs(pkgCfg, rowsPaths...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		fmt.Println("geqrows.go in", pkg.ID)
		err = genRowsFile(pkg)
		if err != nil {
			return fmt.Errorf("row mappers generation failed at: %s: %w", pkg.ID, err)
		}
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

func genBuilderPkg(pkg *packages.Package) (err error) {
	tablesObj := pkg.Types.Scope().Lookup("GeqTables")
	if tablesObj == nil {
		return errors.New("no GeqTables struct found")
	}
	tablesStruct, ok := tablesObj.Type().Underlying().(*types.Struct)
	if !ok {
		return errors.New("GeqTables must be struct")
	}
	for i := 0; i < tablesStruct.NumFields(); i++ {
		field := tablesStruct.Field(i)
		fieldTypeNamed, ok := field.Type().(*types.Named)
		if !ok {
			return fmt.Errorf("type of GeqTables field %s is invalid", field.Name())
		}
		fieldType := fieldTypeNamed.Obj()
		fmt.Println(field.Name(), fieldType.Name(), fieldType.Pkg().Path())
	}

	return nil
}

func genRowsFile(pkg *packages.Package) (err error) {
	if len(pkg.Syntax) != len(pkg.CompiledGoFiles) {
		return errors.New("failed to parse some Go files")
	}

	rowTypeNames := make([]string, 0)
	for i, file := range pkg.CompiledGoFiles {
		if filepath.Base(file) == "geqrows.go" {
			for name := range pkg.Syntax[i].Scope.Objects {
				rowTypeNames = append(rowTypeNames, name)
			}
		}
	}
	for _, rowName := range rowTypeNames {
		rowObj := pkg.Types.Scope().Lookup(rowName)
		rowStruct, ok := rowObj.Type().Underlying().(*types.Struct)
		if !ok {
			return fmt.Errorf("row %s must be struct", rowName)
		}
		for i := 0; i < rowStruct.NumFields(); i++ {
			field := rowStruct.Field(i)
			fieldType := field.Type()
			fmt.Println(field.Name(), fieldType.String())
		}
	}

	return nil
}
