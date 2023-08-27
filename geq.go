package geq

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
