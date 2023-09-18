package codegen

import (
	"errors"
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type rowsFileDef struct {
	PkgName    string
	Imports    []string
	RowMappers []rowMapperDef
}

type rowMapperDef struct {
	Name    string
	RowName string
	Fields  []rowFieldDef
}

type rowFieldDef struct {
	Name string
}

func genRowsFile(rootPath string, pkg *packages.Package) (err error) {
	def, err := buildRowsFileDef(pkg)
	if err != nil {
		return err
	}

	src, err := buildGoCode("rowsFile", rowsFileTmpl, def)
	if err != nil {
		return err
	}

	err = writeFile(rootPath, "mappers.gen.go", src)
	if err != nil {
		return err
	}

	return nil
}

func buildRowsFileDef(pkg *packages.Package) (def *rowsFileDef, err error) {
	imports := make(map[string]struct{}, 0)
	imports[geqPkgPath] = struct{}{}

	mappers, err := parseRowMappers(pkg, imports)
	if err != nil {
		return nil, err
	}
	return &rowsFileDef{
		PkgName:    pkg.Name,
		Imports:    mapKeys(imports),
		RowMappers: mappers,
	}, nil
}

func parseRowMappers(pkg *packages.Package, imports map[string]struct{}) (mappers []rowMapperDef, err error) {
	rowsStruct, err := lookupStruct(pkg, "GeqMappers")
	if errors.Is(err, errNoStruct) {
		return mappers, nil
	}
	if err != nil {
		return nil, err
	}

	nFields := rowsStruct.NumFields()
	mappers = make([]rowMapperDef, 0, nFields)
	for i := 0; i < nFields; i++ {
		field := rowsStruct.Field(i)
		fieldType, ok := field.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("type of GeqMappers field %s must be named", field.Name())
		}

		ftPkg := fieldType.Obj().Pkg()
		rowName := fmt.Sprintf("%s.%s", ftPkg.Name(), fieldType.Obj().Name())
		imports[ftPkg.Path()] = struct{}{}

		fieldStruct, ok := field.Type().Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("type of GeqMappers field %s must be struct", field.Name())
		}

		nRowFields := fieldStruct.NumFields()
		if nRowFields == 0 {
			return nil, fmt.Errorf("type of GeqMappers field %s must have one or more fields", field.Name())
		}

		rowFields := make([]rowFieldDef, 0, nRowFields)
		for j := 0; j < nRowFields; j++ {
			name := fieldStruct.Field(j).Name()
			rowFields = append(rowFields, rowFieldDef{Name: name})
		}

		mappers = append(mappers, rowMapperDef{
			Name:    field.Name(),
			RowName: rowName,
			Fields:  rowFields,
		})
	}

	return mappers, nil
}

const rowsFileTmpl = `
package {{.PkgName}}

import (
{{range .Imports -}}
	"{{.}}"
{{end}}
)

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
