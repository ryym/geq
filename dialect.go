package geq

import "fmt"

type StrConcatType uint

const (
	StrConcatStandard StrConcatType = iota
	StrConcatFunc
)

type Dialect interface {
	Placeholder(typeName string, prevArgs []any) string
	Ident(v string) string
	StrConcatType() StrConcatType
}

func DialectByName(driverName string) (d Dialect, err error) {
	switch driverName {
	case "postgres":
		return &DialectPostgres{}, nil
	case "mysql":
		return &DialectMySQL{}, nil
	default:
		return nil, fmt.Errorf("unsupported DB dialect: %s", driverName)
	}
}

type DialectGeneric struct{}

func (d *DialectGeneric) Placeholder(typeName string, prevArgs []any) string {
	return "?"
}

func (d *DialectGeneric) Ident(v string) string {
	return v
}

func (d *DialectGeneric) StrConcatType() StrConcatType {
	return StrConcatStandard
}

type DialectPostgres struct{}

func (d *DialectPostgres) Placeholder(typeName string, prevArgs []any) string {
	phNum := len(prevArgs) + 1
	if typeName == "" {
		return fmt.Sprintf("$%d", phNum)
	}
	return fmt.Sprintf("$%d::%s", phNum, typeName)
}

func (d *DialectPostgres) Ident(v string) string {
	return fmt.Sprintf(`"%s"`, v)
}

func (d *DialectPostgres) StrConcatType() StrConcatType {
	return StrConcatStandard
}

type DialectMySQL struct{}

func (d *DialectMySQL) Placeholder(typeName string, prevArgs []any) string {
	return "?"
}

func (d *DialectMySQL) Ident(v string) string {
	return fmt.Sprintf("`%s`", v)
}

func (d *DialectMySQL) StrConcatType() StrConcatType {
	return StrConcatFunc
}
