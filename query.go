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
	Expr() Expr
	Alias() string
}

type TypedSelection[F any] interface {
	Selection
}

type AnyRelship interface {
	RightTableName() string
	JoinColumns() (left Expr, right Expr)
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

func (r *Relship[R, C]) JoinColumns() (left Expr, right Expr) {
	return r.colL, r.colR
}

type FinalQuery struct {
	Query string
	Args  []any
}

type Query[R any] struct {
	selections []Selection
	from       Table[R]
	innerJoins []AnyRelship // For now.
	wheres     []Expr
	orders     []Expr // For now ASC only.
	limit      uint
	args       []any
}

func newQuery[R any](from Table[R]) *Query[R] {
	var sels []Selection
	if from != nil {
		sels = from.Selections()
	}
	return &Query[R]{selections: sels, from: from}
}

func (q *Query[R]) Select(sels ...Selection) *Query[R] {
	q.selections = sels
	return q
}

func (q *Query[R]) Joins(relships ...AnyRelship) *Query[R] {
	q.innerJoins = append(q.innerJoins, relships...)
	return q
}

func (q *Query[R]) Where(exprs ...Expr) *Query[R] {
	q.wheres = append(q.wheres, exprs...)
	return q
}

func (q *Query[R]) OrderBy(orders ...Expr) *Query[R] {
	q.orders = orders
	return q
}

func (q *Query[R]) Limit(n uint) *Query[R] {
	q.limit = n
	return q
}

func (q *Query[R]) Finalize() *FinalQuery {
	var sb strings.Builder

	appendQuery := func(s string, args []any) {
		sb.WriteString(s)
		q.args = append(q.args, args...)
	}

	sb.WriteString("SELECT ")
	for i, sel := range q.selections {
		if i > 0 {
			sb.WriteString(", ")
		}
		p := buildExprPart(sel.Expr())
		appendQuery(p.String(), p.args)
		alias := sel.Alias()
		if alias != "" {
			fmt.Fprintf(&sb, " AS %s", alias)
		}
	}

	if q.from != nil {
		sb.WriteString(" FROM ")
		sb.WriteString(q.from.TableName())
	}

	if len(q.innerJoins) > 0 {
		for _, r := range q.innerJoins {
			fmt.Fprintf(&sb, " INNER JOIN %s ON ", r.RightTableName())
			colL, colR := r.JoinColumns()
			partL := buildExprPart(colL)
			partR := buildExprPart(colR)
			appendQuery(partL.String(), partL.args)
			sb.WriteString(" = ")
			appendQuery(partR.String(), partR.args)
		}
	}

	if len(q.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range q.wheres {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			part := buildExprPart(w)
			appendQuery(part.String(), part.args)
		}
	}

	if len(q.orders) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, expr := range q.orders {
			if i > 0 {
				sb.WriteRune(',')
			}
			p := buildExprPart(expr)
			appendQuery(p.String(), p.args)
		}
	}

	if q.limit > 0 {
		fmt.Fprintf(&sb, " LIMIT %d", q.limit)
	}

	return &FinalQuery{
		Query: sb.String(),
		Args:  q.args,
	}
}

type queryPart struct {
	sb   *strings.Builder
	args []any
}

func (p *queryPart) String() string {
	return p.sb.String()
}
