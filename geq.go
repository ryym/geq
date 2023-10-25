package geq

var defaultDialect Dialect = &DialectGeneric{}

func SetDefaultDialect(d Dialect) {
	defaultDialect = d
}

func AsMap[R any, K comparable](key *Column[K], q *Query[R]) *MapLoader[R, R, K] {
	return &MapLoader[R, R, K]{query: q, mapper: q.mapper, key: key}
}

func AsSliceMap[R any, K comparable](key *Column[K], q *Query[R]) *SliceMapLoader[R, R, K] {
	return &SliceMapLoader[R, R, K]{query: q, mapper: q.mapper, key: key}
}

func ToSlice[R any](mapper RowMapper[R], dest *[]R) *SliceScanner[R] {
	return &SliceScanner[R]{mapper: mapper, dest: dest}
}

func ToMap[R any, K comparable](mapper RowMapper[R], key TypedSelection[K], dest *map[K]R) *MapScanner[R, K] {
	return &MapScanner[R, K]{mapper: mapper, dest: dest, key: key}
}

func ToSliceMap[R any, K comparable](mapper RowMapper[R], key TypedSelection[K], dest *map[K][]R) *SliceMapScanner[R, K] {
	return &SliceMapScanner[R, K]{mapper: mapper, dest: dest, key: key}
}

func SelectFrom[R any](table Table[R], sels ...Selection) *Query[R] {
	q := newQuery(table).From(table)
	if len(sels) > 0 {
		q.selections = sels
	}
	return q
}

func SelectAs[R any](mapper RowMapper[R]) *Query[R] {
	return newQuery(mapper)
}

func SelectOnly[V any](col *Column[V]) *Query[V] {
	mapper := &ValueMapper[V]{sels: []Selection{col}}
	return newQuery(mapper)
}

func Select(sels ...Selection) *Query[struct{}] {
	mapper := &EmptyMapper{}
	q := newQuery[struct{}](mapper)
	q.selections = sels
	return q
}

func SelectVia[S, T, C any, RT Table[S]](srcs []S, table Table[T], relship *Relship[RT, S, C]) *Query[T] {
	return newQuery(table).From(table).Where(relship.In(srcs))
}

func InsertInto(table AnyTable) *InsertQuery {
	return newInsertQuery(table)
}

func Update(table AnyTable) *UpdateQuery {
	return newUpdateQuery(table)
}

func DeleteFrom(table AnyTable) *DeleteQuery {
	return newDeleteQuery(table)
}

func Null() AnonExpr {
	return implOps(&nullExpr{})
}

func Parens(expr Expr) AnonExpr {
	return implOps(&parensExpr{expr: expr})
}

func Concat(vals ...any) AnonExpr {
	return newConcatExpr(vals...)
}

func Raw(sql string) *RawExpr {
	return newRawExpr(sql)
}

func Func(name string, args ...any) *FuncExpr {
	exprs := make([]Expr, 0, len(args))
	for _, arg := range args {
		exprs = append(exprs, toExpr(arg))
	}
	return implOps(&FuncExpr{name: name, args: exprs})
}

func Count(expr Expr) *FuncExpr {
	return Func("COUNT", expr)
}

func Sum(expr Expr) *FuncExpr {
	return Func("SUM", expr)
}

func Min(expr Expr) *FuncExpr {
	return Func("MIN", expr)
}

func Max(expr Expr) *FuncExpr {
	return Func("MAX", expr)
}

func Avg(expr Expr) *FuncExpr {
	return Func("AVG", expr)
}

func Coalesce(exprs ...Expr) AnonExpr {
	vals := make([]any, 0, len(exprs))
	for _, e := range exprs {
		vals = append(vals, e)
	}
	return Func("COALESCE", vals...)
}
