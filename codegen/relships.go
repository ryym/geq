package codegen

import (
	"errors"
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

func parseRelationships(pkg *packages.Package, cfg *builderConfig, tables []tableDef) (relsMap map[string][]*relshipDef, err error) {
	relStruct, err := lookupStruct(pkg, "GeqRelationships")
	if errors.Is(err, errNoStruct) {
		return relsMap, nil
	}
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

			var rowType string
			ftPkg := ft.Obj().Pkg()
			if cfg.outPkgPath == ftPkg.Path() {
				rowType = ft.Obj().Name()
			} else {
				rowType = fmt.Sprintf("%s.%s", ft.Obj().Pkg().Name(), ft.Obj().Name())
			}
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
