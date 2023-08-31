package geq

func AsSlice[R any](q *Query[R]) *SliceLoader[R, R] {
	return &SliceLoader[R, R]{query: q, mapper: q.from}
}

func AsMap[R any, K comparable](q *Query[R], key *Column[K]) *MapLoader[R, R, K] {
	return &MapLoader[R, R, K]{query: q, mapper: q.from, key: key}
}

func AsThese[Q any](q *Query[Q], scanners ...RowsScanner) *MultiScanLoader[Q] {
	sels := make([]Selection, 0)
	for _, s := range scanners {
		sels = append(sels, s.Selections()...)
	}
	q.selections = sels
	return &MultiScanLoader[Q]{query: q, scanners: scanners}
}

func ToSlice[R any](table Table[R], dest *[]R) *SliceScanner[R] {
	return &SliceScanner[R]{mapper: table, dest: dest}
}

func ToMap[R any, K comparable](table Table[R], key TypedSelection[K], dest *map[K]R) *MapScanner[R, K] {
	return &MapScanner[R, K]{mapper: table, dest: dest, key: key}
}

func Via[S, T, C any](srcs []S, from Table[T], relship *Relship[S, C]) *Query[T] {
	sels := relship.tableR.Selections()
	colIdx := selectionIndex(sels[0], sels, relship.colR)
	if colIdx < 0 {
		panic("right table column not in selections")
	}

	keys := make([]C, len(srcs))
	for i, s := range srcs {
		ptrs := relship.tableR.FieldPtrs(&s)
		ptr := ptrs[colIdx]
		keys[i] = *ptr.(*C)
	}

	q := newQuery(from)
	q.Where(relship.colL.In(keys))

	return q
}

func From[R any](table Table[R]) *Query[R] {
	return newQuery(table)
}

func FromNothing() *Query[struct{}] {
	return newQuery[struct{}](nil)
}
