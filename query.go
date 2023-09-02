package geq

import (
	"context"
	"fmt"
	"strings"
)

type AnyTable interface {
	TableName() string
	appendTable(w *queryWriter)
}

type Table[R any] interface {
	RowMapper[R]
	AnyTable
}

type TableBase struct {
	tableName string
	alias     string
	columns   []Selection
}

func NewTableBase(tableName string, alias string, columns []Selection) *TableBase {
	return &TableBase{
		tableName: tableName,
		alias:     alias,
		columns:   columns,
	}
}

func (t *TableBase) TableName() string {
	return t.tableName
}

func (t *TableBase) Selections() []Selection {
	return t.columns
}

func (t *TableBase) appendTable(w *queryWriter) {
	w.Write(t.tableName)
	if t.alias != "" {
		w.Write(" AS ")
		w.Write(t.alias)
	}
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

func (r *Relship[R, C]) In(recs []R) Expr {
	sels := r.tableR.Selections()
	colIdx := selectionIndex(sels[0], sels, r.colR)
	if colIdx < 0 {
		panic("right table column not in selections")
	}

	vals := make([]C, 0, len(recs))
	for _, rec := range recs {
		ptrs := r.tableR.FieldPtrs(&rec)
		ptr := ptrs[colIdx]
		vals = append(vals, *ptr.(*C))
	}

	return r.colL.In(vals)
}

type FinalQuery struct {
	Query string
	Args  []any
}

type Query[R any] struct {
	mapper     RowMapper[R]
	selections []Selection
	from       AnyTable
	innerJoins []AnyRelship // For now.
	wheres     []Expr
	groups     []Expr
	orders     []Expr // For now ASC only.
	limit      uint
	args       []any
}

func newQuery[R any](mapper RowMapper[R]) *Query[R] {
	q := &Query[R]{mapper: mapper}
	q.selections = mapper.Selections()
	return q
}

func (q *Query[R]) From(table AnyTable) *Query[R] {
	q.from = table
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

func (q *Query[R]) GroupBy(exprs ...Expr) *Query[R] {
	q.groups = append(q.groups, exprs...)
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
	w := newQueryWriter()

	w.Write("SELECT ")
	for i, sel := range q.selections {
		if i > 0 {
			w.Write(", ")
		}
		sel.Expr().appendExpr(w)
		alias := sel.Alias()
		if alias != "" {
			w.Printf(" AS %s", alias)
		}
	}

	if q.from != nil {
		w.Write(" FROM ")
		q.from.appendTable(w)
	}

	if len(q.innerJoins) > 0 {
		for _, r := range q.innerJoins {
			w.Printf(" INNER JOIN %s ON ", r.RightTableName())
			colL, colR := r.JoinColumns()
			colL.appendExpr(w)
			w.Write(" = ")
			colR.appendExpr(w)
		}
	}

	if len(q.wheres) > 0 {
		w.Write(" WHERE ")
		for i, e := range q.wheres {
			if i > 0 {
				w.Write(" AND ")
			}
			e.appendExpr(w)
		}
	}

	if len(q.groups) > 0 {
		w.Write(" GROUP BY ")
		for i, e := range q.groups {
			if i > 0 {
				w.Write(", ")
			}
			e.appendExpr(w)
		}
	}

	if len(q.orders) > 0 {
		w.Write(" ORDER BY ")
		for i, e := range q.orders {
			if i > 0 {
				w.Write(", ")
			}
			e.appendExpr(w)
		}
	}

	if q.limit > 0 {
		w.Printf(" LIMIT %d", q.limit)
	}

	return &FinalQuery{
		Query: w.String(),
		Args:  w.Args,
	}
}

func (q *Query[R]) Load(ctx context.Context, db QueryRunner) (recs []R, err error) {
	l := &SliceLoader[R, R]{query: q, mapper: q.mapper}
	return l.Load(ctx, db)
}

func (q *Query[R]) WillScan(scanners ...RowsScanner) *MultiScanLoader[R] {
	sels := make([]Selection, 0)
	for _, s := range scanners {
		sels = append(sels, s.Selections()...)
	}
	q.selections = sels
	return &MultiScanLoader[R]{query: q, scanners: scanners}
}

type queryWriter struct {
	sb   *strings.Builder
	Args []any
}

func newQueryWriter() *queryWriter {
	return &queryWriter{sb: new(strings.Builder)}
}

func (w *queryWriter) String() string {
	return w.sb.String()
}

func (w *queryWriter) Write(query string, args ...any) {
	w.sb.WriteString(query)
	w.Args = append(w.Args, args...)
}

func (w *queryWriter) Printf(format string, fmtargs ...any) {
	fmt.Fprintf(w.sb, format, fmtargs...)
}
