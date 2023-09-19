package codegen

import (
	"errors"
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func parseTables(pkg *packages.Package, cfg *builderConfig, imports map[string]struct{}) (tables []tableDef, err error) {
	tablesStruct, err := lookupStruct(pkg, "GeqTables")
	if errors.Is(err, errNoStruct) {
		return tables, nil
	}
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

		var rowName string
		fPkg := fieldType.Obj().Pkg()
		if cfg.outPkgPath == fPkg.Path() {
			rowName = fieldType.Obj().Name()
		} else {
			imports[fPkg.Path()] = struct{}{}
			rowName = fmt.Sprintf("%s.%s", fPkg.Name(), fieldType.Obj().Name())
		}

		tableFields := make([]tableFieldDef, 0, nTableFields)
		for j := 0; j < nTableFields; j++ {
			f := fieldStruct.Field(j)
			tfd, err := parseTableField(f, cfg, imports)
			if err != nil {
				return nil, fmt.Errorf("table row %s invalid: %w", rowName, err)
			}
			tableFields = append(tableFields, *tfd)
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

func parseTableField(f *types.Var, cfg *builderConfig, imports map[string]struct{}) (tfd *tableFieldDef, err error) {
	var typeName string
	switch ft := f.Type().(type) {
	case *types.Basic:
		typeName = ft.Name()
	case *types.Named:
		ftPkg := ft.Obj().Pkg()
		if cfg.outPkgPath == ftPkg.Path() {
			typeName = ft.Obj().Name()
		} else {
			imports[ftPkg.Path()] = struct{}{}
			typeName = fmt.Sprintf("%s.%s", ftPkg.Name(), ft.Obj().Name())
		}
	default:
		return nil, fmt.Errorf("type of field %s invalid", f.Name())
	}
	return &tableFieldDef{
		Name:   f.Name(),
		DbName: toSnake(f.Name()),
		Type:   typeName,
	}, nil
}
