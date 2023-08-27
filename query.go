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

type TypedSelection[F any] interface {
	Selection
}

type Column[F any] struct {
	TableName  string
	ColumnName string
}

func (c *Column[F]) SelectionName() string {
	return fmt.Sprintf("%s.%s", c.TableName, c.ColumnName)
}

type AnyRelship interface {
	RightTableName() string
	JoinColumns() (left Selection, right Selection)
}

type Relship[R, C any] struct {
	tableR Table[R]
	colL   *Column[C]
	colR   *Column[C]
}

func NewRelship[R, C any](tableR Table[R], colL, colR *Column[C]) *Relship[R, C] {
	return &Relship[R, C]{tableR: tableR, colL: colL, colR: colR}
}

func (r *Relship[R, C]) RightTableName() string {
	return r.tableR.TableName()
}

func (r *Relship[R, C]) JoinColumns() (left Selection, right Selection) {
	return r.colL, r.colR
}

type FinalQuery struct {
	Query string
	Args  []any
}

type Query[R any] struct {
	selections []Selection
	from       Table[R]
	innerJoins []string    // For now.
	wheres     []string    // For now.
	orders     []Selection // For now ASC only.
	args       []any
}

func NewQuery[R any](from Table[R]) *Query[R] {
	return &Query[R]{selections: from.Selections(), from: from}
}

func (q *Query[R]) Joins(relships ...AnyRelship) *Query[R] {
	for _, r := range relships {
		var sb strings.Builder
		colL, colR := r.JoinColumns()
		fmt.Fprintf(
			&sb, " INNER JOIN %s ON %s = %s",
			r.RightTableName(), colL.SelectionName(), colR.SelectionName(),
		)
		q.innerJoins = append(q.innerJoins, sb.String())
	}
	return q
}

func (q *Query[R]) AppendWhere(wheres ...string) *Query[R] {
	q.wheres = append(q.wheres, wheres...)
	return q
}

func (q *Query[R]) AppendArgs(args ...any) {
	q.args = append(q.args, args...)
}

func (q *Query[R]) OrderBy(orders ...Selection) *Query[R] {
	q.orders = orders
	return q
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

	if len(q.innerJoins) > 0 {
		for _, j := range q.innerJoins {
			sb.WriteString(j)
		}
	}

	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(w)
		}
	}

	if len(q.orders) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, sel := range q.orders {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(sel.SelectionName())
		}
	}

	return &FinalQuery{
		Query: sb.String(),
		Args:  q.args,
	}
}
