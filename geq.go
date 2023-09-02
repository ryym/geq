package geq

func Builder_AsMap[R any, K comparable](key *Column[K], q *Query[R]) *MapLoader[R, R, K] {
	return &MapLoader[R, R, K]{query: q, mapper: q.mapper, key: key}
}

func Builder_AsSliceMap[R any, K comparable](key *Column[K], q *Query[R]) *SliceMapLoader[R, R, K] {
	return &SliceMapLoader[R, R, K]{query: q, mapper: q.mapper, key: key}
}

func Builder_ToSlice[R any](mapper RowMapper[R], dest *[]R) *SliceScanner[R] {
	return &SliceScanner[R]{mapper: mapper, dest: dest}
}

func Builder_ToMap[R any, K comparable](mapper RowMapper[R], key TypedSelection[K], dest *map[K]R) *MapScanner[R, K] {
	return &MapScanner[R, K]{mapper: mapper, dest: dest, key: key}
}

func Builder_ToSliceMap[R any, K comparable](mapper RowMapper[R], key TypedSelection[K], dest *map[K][]R) *SliceMapScanner[R, K] {
	return &SliceMapScanner[R, K]{mapper: mapper, dest: dest, key: key}
}

func Builder_SelectFrom[R any](table Table[R], sels ...Selection) *Query[R] {
	q := newQuery(table).From(table)
	if len(sels) > 0 {
		q.selections = sels
	}
	return q
}

func Builder_SelectAs[R any](mapper RowMapper[R]) *Query[R] {
	return newQuery(mapper)
}

func Builder_SelectOnly[V any](col *Column[V]) *Query[V] {
	mapper := &ValueMapper[V]{sels: []Selection{col}}
	return newQuery(mapper)
}

func Builder_Select(sels ...Selection) *Query[struct{}] {
	mapper := &EmptyMapper{}
	q := newQuery[struct{}](mapper)
	q.selections = sels
	return q
}

func Builder_SelectVia[S, T, C any](srcs []S, table Table[T], relship *Relship[S, C]) *Query[T] {
	return newQuery(table).From(table).Where(relship.In(srcs))
}

func Builder_InsertInto(table AnyTable) *InsertQuery {
	return newInsertQuery(table)
}

func Builder_Func(name string, args ...any) Expr {
	exprs := make([]Expr, 0, len(args))
	for _, arg := range args {
		exprs = append(exprs, lift(arg))
	}
	return implOps(&funcExpr{name: name, args: exprs})
}

func Builder_Raw(expr string, args ...any) Expr {
	return implOps(&rawExpr{expr: expr, args: args})
}
