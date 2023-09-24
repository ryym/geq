package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

type builderConfig struct {
	outdir     string
	outPkgPath string
	outPkgName string
}

type builderFileDef struct {
	PkgName    string
	Imports    []string
	Tables     []tableDef
	RowMappers []rowMapperDef
}

type tableDef struct {
	Name     string
	DbName   string
	RowName  string
	Fields   []tableFieldDef
	Relships []*relshipDef
}

type tableFieldDef struct {
	Name   string
	DbName string
	Type   string
}

type relshipDef struct {
	MapperR   *tableDef
	RowNameR  string
	RelName   string
	FieldL    string
	FieldR    string
	FieldType string
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

func genBuilderFile(rootPath string, pkg *packages.Package) (err error) {
	cfg, err := parseBuilderConfig(pkg)
	if err != nil {
		return err
	}

	def, err := buildBuilderFileDef(pkg, cfg)
	if err != nil {
		return err
	}

	destDir := filepath.Join(rootPath, cfg.outdir)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	if len(def.Tables) > 0 || len(def.RowMappers) > 0 {
		src, err := buildGoCode("builderFile", builderFileTmpl, def)
		if err != nil {
			return err
		}
		err = writeFile(destDir, "geq.gen.go", src)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildBuilderFileDef(pkg *packages.Package, cfg *builderConfig) (def *builderFileDef, err error) {
	imports := map[string]struct{}{geqPkgPath: {}}

	tables, err := parseTables(pkg, cfg, imports)
	if err != nil {
		return nil, err
	}

	relsMap, err := parseRelationships(pkg, cfg, tables)
	if err != nil {
		return nil, err
	}
	for i := range tables {
		rs := relsMap[tables[i].Name]
		tables[i].Relships = rs
	}

	mappers, err := parseRowMappers(pkg, cfg, imports)
	if err != nil {
		return nil, err
	}

	def = &builderFileDef{
		PkgName:    cfg.outPkgName,
		Imports:    mapKeys(imports),
		Tables:     tables,
		RowMappers: mappers,
	}
	return def, nil
}

func parseBuilderConfig(pkg *packages.Package) (cfg *builderConfig, err error) {
	m, err := parseGeqConfig(pkg, []string{"geqbld.go"}, []string{"geq:outdir"})
	if err != nil {
		return nil, err
	}
	outdir, ok := m["geq:outdir"]
	if !ok {
		outdir = "./d"
	}
	if strings.Contains(outdir, "..") {
		return nil, fmt.Errorf("geq:outdir must not contain '..'")
	}

	var pkgName string
	var pkgPath string
	if filepath.Clean(outdir) == "." {
		pkgName = pkg.Name
		pkgPath = pkg.PkgPath
	} else {
		pkgName = filepath.Base(outdir)
		pkgPath = fmt.Sprintf("%s/%s", pkg.PkgPath, pkgName)
	}
	cfg = &builderConfig{
		outdir:     outdir,
		outPkgPath: pkgPath,
		outPkgName: pkgName,
	}
	return cfg, nil
}

const builderFileTmpl = `
package {{.PkgName}}

import (
{{range .Imports -}}
	"{{.}}"
{{end}}
)

{{range .Tables -}}
var {{.Name}} = New{{.Name}}("{{.DbName}}")
{{end}}

func init() {
	{{range .Tables -}}
	{{.Name}}.InitRelships()
	{{end -}}
}

{{range .Tables}}
type Table{{.Name}} struct {
	*geq.TableBase
	relshipsSet bool
	{{range .Fields -}}
	{{.Name}} *geq.Column[{{.Type}}]
	{{end -}}
	{{range .Relships -}}
	{{.RelName}} *geq.Relship[*Table{{.MapperR.Name}}, {{.RowNameR}}, {{.FieldType}}]
	{{end -}}
}

func New{{.Name}}(alias string) *Table{{.Name}} {
	t := &Table{{.Name}}{
		{{range .Fields -}}
		{{.Name}}: geq.NewColumn[{{.Type}}](alias, "{{.DbName}}"),
		{{end -}}
	}
	columns := []geq.AnyColumn{ {{- range .Fields}} t.{{.Name}}, {{end -}} }
	sels := []geq.Selection{ {{- range .Fields}} t.{{.Name}}, {{end -}} }
	t.TableBase = geq.NewTableBase("{{.DbName}}", alias, columns, sels)
	return t
}

func (t *Table{{.Name}}) InitRelships()  {
	if t.relshipsSet {
		return
	}
	{{range .Relships -}}
	func() {
		r := New{{.MapperR.Name}}("{{.MapperR.DbName}}")
		t.{{.RelName}} = geq.NewRelship(r, t.{{.FieldL}}, r.{{.FieldR}})
	}()
	{{end -}}
	t.relshipsSet = true
}
func (t *Table{{.Name}}) FieldPtrs(r *{{.RowName}}) []any {
	return []any{ {{- range .Fields}} &r.{{.Name}}, {{end -}} }
}
func (t *Table{{.Name}}) As(alias string) *Table{{.Name}} {
	return New{{.Name}}(alias)
}

{{end}}

{{range .RowMappers}}

type {{.Name}} struct {
	{{range .Fields}}
	{{.Name}} geq.Expr{{end}}
}
func (m *{{.Name}}) FieldPtrs(r *{{.RowName}}) []any {
	return []any{ {{range .Fields}} &r.{{.Name}}, {{end}} }
}
func (m *{{.Name}}) Selections() []geq.Selection {
	return []geq.Selection{ {{range .Fields}} m.{{.Name}}, {{end}} }
}

{{end}}
`
