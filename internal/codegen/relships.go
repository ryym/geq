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
		mapperLName := field.Name()
		fieldStruct, ok := field.Type().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("type of field of GeqRelationships %s must be unnamed struct", mapperLName)
		}
		for j := 0; j < fieldStruct.NumFields(); j++ {
			f := fieldStruct.Field(j)
			relName := f.Name()
			fullRelName := fmt.Sprintf("%s.%s", mapperLName, relName)

			ft, ok := f.Type().(*types.Named)
			if !ok {
				return nil, fmt.Errorf("type of field of GeqRelationships %s must be named struct", fullRelName)
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
				return nil, fmt.Errorf("row type of relationship of %s is invalid: %s", fullRelName, rowType)
			}

			fieldL, fieldR, err := parseRelshipStructTag(fullRelName, mapperLName, mapperR.Name, fieldStruct.Tag(j))
			if err != nil {
				return nil, err
			}

			rs := &relshipDef{
				MapperR:  mapperR,
				RowNameR: mapperR.RowName,
				RelName:  relName,
				FieldL:   fieldL,
				FieldR:   fieldR,
			}
			relsMap[mapperLName] = append(relsMap[mapperLName], rs)

			var ftL, ftR string
			for _, t := range tables {
				switch t.Name {
				case mapperLName:
					for _, f := range t.Fields {
						if f.Name == rs.FieldL {
							ftL = f.Type
						}
					}
				case rs.MapperR.Name:
					for _, f := range t.Fields {
						if f.Name == rs.FieldR {
							ftR = f.Type
						}
					}
				}
			}
			if ftL != ftR {
				return nil, fmt.Errorf("relationship field types of %s is invalid (%s, %s)", fullRelName, ftL, ftR)
			}
			rs.FieldType = ftL
		}
	}

	return relsMap, nil
}

func parseRelshipStructTag(fullRelName, mapperLName, mapperRName, tagStr string) (fieldL string, fieldR string, err error) {
	tag := reflect.StructTag(tagStr)
	relDef := tag.Get("geq")
	if relDef == "" {
		err = fmt.Errorf("relationship of %s must be defined in tag", fullRelName)
		return
	}

	relParts := strings.SplitN(relDef, "=", 2)
	fParts1 := strings.SplitN(strings.TrimSpace(relParts[0]), ".", 2)
	fParts2 := strings.SplitN(strings.TrimSpace(relParts[1]), ".", 2)

	var fPartsL, fPartsR []string
	switch mapperLName {
	case fParts1[0]:
		fPartsL, fPartsR = fParts1, fParts2
	case fParts2[0]:
		fPartsL, fPartsR = fParts2, fParts1
	default:
		err = fmt.Errorf("relationship definition of %s is invalid: no %s", fullRelName, mapperLName)
		return
	}
	if fPartsR[0] != mapperRName {
		err = fmt.Errorf("relationship definition of %s is invalid: no %s", fullRelName, mapperRName)
		return
	}

	return fPartsL[1], fPartsR[1], nil
}
