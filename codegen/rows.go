package codegen

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"

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

	file, err := os.Create(filepath.Join(rootPath, "geqrows.gen.go"))
	if err != nil {
		return fmt.Errorf("failed to create rows file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(src)
	if err != nil {
		return fmt.Errorf("failed to write rows file: %w", err)
	}

	return nil
}

func buildRowsFileDef(pkg *packages.Package) (def *rowsFileDef, err error) {
	rowsStruct, err := lookupStruct(pkg, "GeqRows")
	if err != nil {
		return nil, err
	}

	nFields := rowsStruct.NumFields()
	mappers := make([]rowMapperDef, 0, nFields)
	for i := 0; i < nFields; i++ {
		field := rowsStruct.Field(i)
		fieldType, ok := field.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("type of GeqRows field %s must be named", field.Name())
		}

		fieldStruct, ok := field.Type().Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("type of GeqRows field %s must be struct", field.Name())
		}

		nRowFields := fieldStruct.NumFields()
		if nRowFields == 0 {
			return nil, fmt.Errorf("type of GeqRows field %s must have one or more fields", field.Name())
		}

		rowFields := make([]rowFieldDef, 0, nRowFields)
		for j := 0; j < nRowFields; j++ {
			name := fieldStruct.Field(j).Name()
			rowFields = append(rowFields, rowFieldDef{Name: name})
		}

		mappers = append(mappers, rowMapperDef{
			Name:    field.Name(),
			RowName: fieldType.Obj().Name(),
			Fields:  rowFields,
		})
	}

	return &rowsFileDef{
		PkgName:    pkg.Name,
		Imports:    []string{"github.com/ryym/geq"},
		RowMappers: mappers,
	}, nil
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
