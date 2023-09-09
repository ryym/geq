package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"go/types"
	"os"
	"path/filepath"
	"text/template"
	"unicode"

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
	rowsPaths := make([]string, 0)
	err = filepath.Walk(cfg.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch info.Name() {
		case "geqbld.go":
			bldPaths = append(bldPaths, absPath(cwd, filepath.Dir(path)))
		case "geqrows.go":
			rowsPaths = append(rowsPaths, absPath(cwd, filepath.Dir(path)))
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

	err = genRowsFiles(rowsPaths)
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

func genRowsFiles(rowsPaths []string) (err error) {
	pkgCfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax}
	for _, rowsPath := range rowsPaths {
		pkgs, err := loadPkgs(pkgCfg, rowsPath)
		if err != nil {
			return err
		}
		pkg := pkgs[0]
		err = genRowsFile(rowsPath, pkg)
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
	if len(pkgs) == 0 {
		return nil, errors.New("no Go package found")
	}
	return pkgs, nil
}

func absPath(wd string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(wd, path)
}

func lookupStruct(pkg *packages.Package, name string) (strct *types.Struct, err error) {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("no %s struct found", name)
	}
	strct, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%s must be struct", name)
	}
	return strct, nil
}

func buildGoCode(name string, codeTmpl string, data any) (src []byte, err error) {
	tmpl := template.Must(template.New(name).Parse(codeTmpl))
	buf := new(bytes.Buffer)
	_, err = buf.WriteString(autoGenWarning)
	if err != nil {
		return nil, fmt.Errorf("failed to write warning header: %w", err)
	}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Go code: %w", err)
	}

	src, err = format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}
	return src, nil
}

func writeFile(dir, fileName string, content []byte) (err error) {
	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", fileName, err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", fileName, err)
	}

	return nil
}

func toSnake(s string) string {
	if s == "" {
		return s
	}

	src := []rune(s)
	out := []rune{unicode.ToLower(src[0])}
	lastIdx := len(src) - 1
	for i := 1; i <= lastIdx; i++ {
		if unicode.IsUpper(src[i]) {
			if !unicode.IsUpper(src[i-1]) || (i < lastIdx && !unicode.IsUpper(src[i+1])) {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(src[i]))
		} else {
			out = append(out, src[i])
		}
	}

	return string(out)
}

const autoGenWarning = `
// Code generated by geq. DO NOT EDIT.
// https://github.com/ryym/geq
`
