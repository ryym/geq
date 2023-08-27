package geq

import (
	"fmt"
	"strings"
)

type RowMapper[R any] interface {
	Selections() []Selection
	FieldPtrs(*R) []any
}

type Table[R any] interface {
	RowMapper[R]
	TableName() string
}

type Selection interface {
	SelectionName() string
}

type Column[F any] struct {
	TableName  string
	ColumnName string
}

func (c *Column[F]) SelectionName() string {
	return fmt.Sprintf("%s.%s", c.TableName, c.ColumnName)
}

type FinalQuery struct {
	Query string
	Args  []any
}

type Query[R any] struct {
	selections []Selection
	from       Table[R]
}

func NewQuery[R any](from Table[R]) *Query[R] {
	return &Query[R]{selections: from.Selections(), from: from}
}

func (q *Query[R]) Finalize() *FinalQuery {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	for i, sel := range q.selections {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(sel.SelectionName())
	}
	sb.WriteString(" FROM ")
	sb.WriteString(q.from.TableName())
	if q.raw != nil {
		sb.WriteString(" ")
		sb.WriteString(q.raw.String())
	}

	return &FinalQuery{
		Query: sb.String(),
		Args:  nil,
	}
}
