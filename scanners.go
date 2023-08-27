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
