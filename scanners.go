package geq

type RowsScanner interface {
	Selections() []Selection
	BeforeEachScan(idx int, sels []Selection) (ptrs []any)
	AfterEachScan(ptrs []any)
}

type SliceScanner[R any] struct {
	mapper RowMapper[R]
	dest   *[]R
	row    *R
}

func (s *SliceScanner[R]) Selections() []Selection {
	return s.mapper.Selections()
}

func (s *SliceScanner[R]) BeforeEachScan(idx int, _ []Selection) []any {
	if idx == 0 {
		*s.dest = make([]R, 0)
	}
	s.row = new(R)
	return s.mapper.FieldPtrs(s.row)
}
func (s *SliceScanner[R]) AfterEachScan(ptrs []any) {
	*s.dest = append(*s.dest, *s.row)
}

type MapScanner[R any, K comparable] struct {
	mapper RowMapper[R]
	key    TypedSelection[K]
	dest   *map[K]R
	keyIdx int
	row    *R
}

func (s *MapScanner[R, K]) Selections() []Selection {
	return s.mapper.Selections()
}

func (s *MapScanner[R, K]) BeforeEachScan(idx int, sels []Selection) []any {
	if idx == 0 {
		s.keyIdx = selectionIndex(s.Selections()[0], sels, s.key)
		*s.dest = make(map[K]R)
	}
	s.row = new(R)
	return s.mapper.FieldPtrs(s.row)
}

func (s *MapScanner[R, K]) AfterEachScan(ptrs []any) {
	keyPtr := ptrs[s.keyIdx]
	key := *keyPtr.(*K)
	(*s.dest)[key] = *s.row
}
