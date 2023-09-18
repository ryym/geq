package codegen

import (
	"errors"
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type rowMapperDef struct {
	Name    string
	RowName string
	Fields  []rowFieldDef
}

type rowFieldDef struct {
	Name string
}

func parseRowMappers(pkg *packages.Package, cfg *builderConfig, imports map[string]struct{}) (mappers []rowMapperDef, err error) {
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

		var rowName string
		ftPkg := fieldType.Obj().Pkg()
		if cfg.outPkgPath == ftPkg.Path() {
			rowName = fieldType.Obj().Name()
		} else {
			rowName = fmt.Sprintf("%s.%s", ftPkg.Name(), fieldType.Obj().Name())
			imports[ftPkg.Path()] = struct{}{}
		}

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
