package geq

type RowMapper[R any] interface {
	Selections() []Selection
	FieldPtrs(*R) []any
}

type ValueMapper[V any] struct {
	sels []Selection
}

func (m *ValueMapper[V]) Selections() []Selection {
	return m.sels
}
func (m *ValueMapper[V]) FieldPtrs(p *V) []any {
	return []any{p}
}

type EmptyMapper struct{}

func (m *EmptyMapper) Selections() []Selection {
	return []Selection{}
}
func (m *EmptyMapper) FieldPtrs(_ *struct{}) []any {
	return []any{}
}
