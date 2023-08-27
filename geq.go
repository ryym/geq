package geq

func SelectSlice[R any](q *Query[R]) *SliceLoader[R, R] {
	return &SliceLoader[R, R]{query: q, mapper: q.from}
}

func SelectMap[R any, K comparable](q *Query[R], key *Column[K]) *MapLoader[R, R, K] {
	return &MapLoader[R, R, K]{query: q, mapper: q.from, key: key}
}

func SelectInto[Q any](q *Query[Q], scanners ...RowsScanner) *MultiScanLoader[Q] {
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
