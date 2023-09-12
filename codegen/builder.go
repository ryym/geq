package codegen

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

type builderFileDef struct {
	PkgName  string
	Imports  []string
	Tables   []tableDef
	Relships []*relshipDef
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
	MapperL   string
	MapperR   string
	RowNameR  string
	RelName   string
	FieldL    string
	FieldR    string
	FieldType string
}

type builderGeqFileDef struct {
	PkgName string
	Import  string
}

func genBuilderFile(rootPath string, pkg *packages.Package) (err error) {
	def, err := buildBuilderFileDef(pkg)
	if err != nil {
		return err
	}
	def.PkgName = "b"

	destDir := filepath.Join(rootPath, def.PkgName)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	src, err := buildGoCode("builderFile", builderFileTmpl, def)
	if err != nil {
		return err
	}
	err = writeFile(destDir, "schema.gen.go", src)
	if err != nil {
		return err
	}

	geqDef := &builderGeqFileDef{
		PkgName: def.PkgName,
		Import:  geqPkgPath,
	}
	src, err = buildGoCode("builderGeqFile", builderGeqFileTmpl, geqDef)
	if err != nil {
		return err
	}
	err = writeFile(destDir, "geq.gen.go", src)
	if err != nil {
		return err
	}

	return nil
}

func buildBuilderFileDef(pkg *packages.Package) (def *builderFileDef, err error) {
	imports := map[string]struct{}{geqPkgPath: {}}

	tables, err := parseTables(pkg, imports)
	if err != nil {
		return nil, err
	}

	relsMap, err := parseRelationships(pkg, tables)
	if err != nil {
		return nil, err
	}

	rels := make([]*relshipDef, 0)
	for i := range tables {
		rs := relsMap[tables[i].Name]
		tables[i].Relships = rs
		rels = append(rels, rs...)
	}

	def = &builderFileDef{
		Imports:  mapKeys(imports),
		Tables:   tables,
		Relships: rels,
	}
	return def, nil
}

func parseTables(pkg *packages.Package, imports map[string]struct{}) (tables []tableDef, err error) {
	tablesStruct, err := lookupStruct(pkg, "GeqTables")
	if err != nil {
		return nil, err
	}

	nTables := tablesStruct.NumFields()
	tables = make([]tableDef, 0, nTables)
	for i := 0; i < nTables; i++ {
		field := tablesStruct.Field(i)
		mapperName := field.Name()
		fieldType, ok := field.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("type of GeqTables field %s must be named", mapperName)
		}
		fieldStruct, ok := fieldType.Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("type of GeqTables field %s must be struct", mapperName)
		}

		nTableFields := fieldStruct.NumFields()
		if nTableFields == 0 {
			return nil, fmt.Errorf("type of GeqTables field %s must have one or more fields", mapperName)
		}

		fPkg := fieldType.Obj().Pkg()
		imports[fPkg.Path()] = struct{}{}
		rowName := fmt.Sprintf("%s.%s", fPkg.Name(), fieldType.Obj().Name())

		tableFields := make([]tableFieldDef, 0, nTableFields)
		for j := 0; j < nTableFields; j++ {
			f := fieldStruct.Field(j)
			var fTypeName string
			switch ft := f.Type().(type) {
			case *types.Basic:
				fTypeName = ft.Name()
			// case *types.Named:
			default:
				return nil, fmt.Errorf("type of field of table row %s.%s is invalid", rowName, f.Name())
			}

			tableFields = append(tableFields, tableFieldDef{
				Name:   f.Name(),
				DbName: toSnake(f.Name()),
				Type:   fTypeName,
			})
		}

		td := tableDef{
			Name:    mapperName,
			DbName:  toSnake(mapperName),
			RowName: rowName,
			Fields:  tableFields,
		}
		tables = append(tables, td)
	}

	return tables, nil
}

func parseRelationships(pkg *packages.Package, tables []tableDef) (relsMap map[string][]*relshipDef, err error) {
	relStruct, err := lookupStruct(pkg, "GeqRelationships")
	if err != nil {
		return nil, err
	}

	mapperMap := make(map[string]*tableDef, len(tables))
	for i := range tables {
		mapperMap[tables[i].RowName] = &tables[i]
	}

	relsMap = make(map[string][]*relshipDef, 0)
	for i := 0; i < relStruct.NumFields(); i++ {
		field := relStruct.Field(i)
		mapperL := field.Name()
		fieldStruct, ok := field.Type().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("type of field of GeqRelationships %s must be unnamed struct", mapperL)
		}
		for j := 0; j < fieldStruct.NumFields(); j++ {
			f := fieldStruct.Field(j)
			relName := f.Name()

			ft, ok := f.Type().(*types.Named)
			if !ok {
				return nil, fmt.Errorf("type of field of GeqRelationships %s.%s must be named struct", mapperL, f.Name())
			}

			rowType := fmt.Sprintf("%s.%s", ft.Obj().Pkg().Name(), ft.Obj().Name())
			mapperR, ok := mapperMap[rowType]
			if !ok {
				return nil, fmt.Errorf("invalid relationship row type: %s", rowType)
			}

			tag := reflect.StructTag(fieldStruct.Tag(j))
			relDef := tag.Get("geq")
			if relDef == "" {
				return nil, fmt.Errorf("relationship of %s must be defined in tag", relName)
			}
			relParts := strings.SplitN(relDef, " = ", 2)
			fParts1 := strings.SplitN(relParts[0], ".", 2)
			fParts2 := strings.SplitN(relParts[1], ".", 2)

			rs := &relshipDef{
				MapperL:  mapperL,
				MapperR:  mapperR.Name,
				RowNameR: mapperR.RowName,
				RelName:  relName,
				FieldL:   fParts1[1],
				FieldR:   fParts2[1],
			}
			relsMap[mapperL] = append(relsMap[mapperL], rs)

			var ftL, ftR string
			for _, t := range tables {
				switch t.Name {
				case rs.MapperL:
					for _, f := range t.Fields {
						if f.Name == rs.FieldL {
							ftL = f.Type
						}
					}
				case rs.MapperR:
					for _, f := range t.Fields {
						if f.Name == rs.FieldR {
							ftR = f.Type
						}
					}
				}
			}
			if ftL != ftR {
				return nil, fmt.Errorf("relationship field types invalid (%s, %s)", ftL, ftR)
			}
			rs.FieldType = ftL
		}
	}

	return relsMap, nil
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

const builderFileTmpl = `
package {{.PkgName}}

import (
{{range .Imports -}}
	"{{.}}"
{{end}}
)

{{range .Tables -}}
var {{.Name}} = New{{.Name}}("")
{{end}}

func init() {
	{{range .Relships -}}
	{{.MapperL}}.{{.RelName}} = geq.NewRelship({{.MapperR}}, {{.MapperL}}.{{.FieldL}}, {{.MapperR}}.{{.FieldR}})
	{{end -}}
}

{{range .Tables}}
type Table{{.Name}} struct {
	*geq.TableBase
	{{range .Fields -}}
	{{.Name}} *geq.Column[{{.Type}}]
	{{end -}}
	{{range .Relships -}}
	{{.RelName}} *geq.Relship[{{.RowNameR}}, {{.FieldType}}]
	{{end -}}
}

func New{{.Name}}(alias string) *Table{{.Name}} {
	t := &Table{{.Name}}{
		{{$dbTable := .DbName}}
		{{- range .Fields -}}
		{{.Name}}: geq.NewColumn[{{.Type}}]("{{$dbTable}}", "{{.DbName}}"),
		{{end -}}
	}
	columns := []geq.AnyColumn{ {{- range .Fields}} t.{{.Name}}, {{end -}} }
	sels := []geq.Selection{ {{- range .Fields}} t.{{.Name}}, {{end -}} }
	t.TableBase = geq.NewTableBase("{{.DbName}}", alias, columns, sels)
	return t
}
func (t *Table{{.Name}}) FieldPtrs(r *{{.RowName}}) []any {
	return []any{ {{- range .Fields}} &r.{{.Name}}, {{end -}} }
}
func (t *Table{{.Name}}) As(alias string) *Table{{.Name}} {
	return New{{.Name}}(alias)
}

{{end}}

`

const builderGeqFileTmpl = `
package {{.PkgName}}

import "{{.Import}}"

func AsMap[R any, K comparable](key *geq.Column[K], q *geq.Query[R]) *geq.MapLoader[R, R, K] {
	return geq.Builder_AsMap(key, q)
}

func AsSliceMap[R any, K comparable](key *geq.Column[K], q *geq.Query[R]) *geq.SliceMapLoader[R, R, K] {
	return geq.Builder_AsSliceMap(key, q)
}

func ToSlice[R any](mapper geq.RowMapper[R], dest *[]R) *geq.SliceScanner[R] {
	return geq.Builder_ToSlice(mapper, dest)
}

func ToMap[R any, K comparable](mapper geq.RowMapper[R], key geq.TypedSelection[K], dest *map[K]R) *geq.MapScanner[R, K] {
	return geq.Builder_ToMap(mapper, key, dest)
}

func ToSliceMap[R any, K comparable](mapper geq.RowMapper[R], key geq.TypedSelection[K], dest *map[K][]R) *geq.SliceMapScanner[R, K] {
	return geq.Builder_ToSliceMap(mapper, key, dest)
}

func SelectFrom[R any](table geq.Table[R], sels ...geq.Selection) *geq.Query[R] {
	return geq.Builder_SelectFrom(table, sels...)
}

func SelectAs[R any](mapper geq.RowMapper[R]) *geq.Query[R] {
	return geq.Builder_SelectAs(mapper)
}

func SelectOnly[V any](col *geq.Column[V]) *geq.Query[V] {
	return geq.Builder_SelectOnly(col)
}

func Select(sels ...geq.Selection) *geq.Query[struct{}] {
	return geq.Builder_Select(sels...)
}

func SelectVia[S, T, C any](srcs []S, from geq.Table[T], relship *geq.Relship[S, C]) *geq.Query[T] {
	return geq.Builder_SelectVia(srcs, from, relship)
}

func InsertInto(table geq.AnyTable) *geq.InsertQuery {
	return geq.Builder_InsertInto(table)
}

func Update(table geq.AnyTable) *geq.UpdateQuery {
	return geq.Builder_Update(table)
}

func Null() geq.AnonExpr {
	return geq.Builder_Null()
}

func Concat(vals ...any) geq.AnonExpr {
	return geq.Builder_Concat(vals...)
}

func Func(name string, args ...any) geq.AnonExpr {
	return geq.Builder_Func(name, args...)
}

func Count(expr geq.Expr) geq.AnonExpr {
	return Func("COUNT", expr)
}

func Max(expr geq.Expr) geq.AnonExpr {
	return Func("MAX", expr)
}

func Raw(expr string, args ...any) geq.AnonExpr {
	return geq.Builder_Raw(expr, args...)
}
`
