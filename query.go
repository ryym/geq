package geq

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type QueryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type QueryConfig struct {
	dialect Dialect
}

func NewQueryConfig(d Dialect) *QueryConfig {
	return &QueryConfig{dialect: d}
}

type TableLike interface {
	appendTable(w *queryWriter, cfg *QueryConfig)
}

type AnyTable interface {
	TableLike
	TableName() string
	Columns() []AnyColumn
}

type Table[R any] interface {
	RowMapper[R]
	AnyTable
	InitRelships()
}

type TableBase struct {
	tableName  string
	alias      string
	columns    []AnyColumn
	selections []Selection
}

func NewTableBase(tableName string, alias string, columns []AnyColumn, sels []Selection) *TableBase {
	if alias == tableName {
		alias = ""
	}
	return &TableBase{
		tableName:  tableName,
		alias:      alias,
		columns:    columns,
		selections: sels,
	}
}

func (t *TableBase) TableName() string {
	return t.tableName
}

func (t *TableBase) Columns() []AnyColumn {
	return t.columns
}

func (t *TableBase) Selections() []Selection {
	return t.selections
}

func (t *TableBase) appendTable(w *queryWriter, cfg *QueryConfig) {
	w.Write(cfg.dialect.Ident(t.tableName))
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
	toJoinClause() joinClause
}

type Relship[T Table[R], R, C any] struct {
	tableR T
	colL   *Column[C]
	colR   *Column[C]
}

func NewRelship[T Table[R], R, C any](tableR T, colL, colR *Column[C]) *Relship[T, R, C] {
	return &Relship[T, R, C]{tableR: tableR, colL: colL, colR: colR}
}

func (r *Relship[T, R, C]) Selections() []Selection {
	return r.tableR.Selections()
}

func (r *Relship[T, R, C]) FieldPtrs(row *R) []any {
	return r.tableR.FieldPtrs(row)
}

func (r *Relship[T, R, C]) appendTable(w *queryWriter, cfg *QueryConfig) {
	r.tableR.appendTable(w, cfg)
}

func (r *Relship[T, R, C]) T() T {
	r.tableR.InitRelships()
	return r.tableR
}

func (r *Relship[T, R, C]) toJoinClause() joinClause {
	return joinClause{
		mode:      "INNER",
		table:     r.tableR,
		condition: r.colL.Eq(r.colR),
	}
}

func (r *Relship[T, R, C]) In(recs []R) Expr {
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

type joinClause struct {
	mode      string
	table     TableLike
	condition Expr
}

type AnyQuery interface {
	Build() (*BuiltQuery, error)
	BuildWith(cfg *QueryConfig) (*BuiltQuery, error)
}

type BuiltQuery struct {
	Query string
	Args  []any
}

type Query[R any] struct {
	ops
	mapper     RowMapper[R]
	distinct   bool
	selections []Selection
	from       TableLike
	innerJoins []joinClause
	wheres     []Expr
	groups     []Expr
	orders     []Expr // For now ASC only.
	limit      uint
	args       []any
}

func newQuery[R any](mapper RowMapper[R]) *Query[R] {
	q := implOps(&Query[R]{mapper: mapper})
	q.selections = mapper.Selections()
	return q
}

func (q *Query[R]) As(alias string) *QueryTable[R] {
	return &QueryTable[R]{query: q, alias: alias}
}

func (q *Query[R]) Distinct() *Query[R] {
	q.distinct = true
	return q
}

func (q *Query[R]) From(table TableLike) *Query[R] {
	q.from = table
	return q
}

func (q *Query[R]) JoinRels(relships ...AnyRelship) *Query[R] {
	for _, rs := range relships {
		join := rs.toJoinClause()
		switch join.mode {
		case "INNER":
			q.innerJoins = append(q.innerJoins, join)
		default:
			panic(fmt.Sprintf("unknown join mode: %s", join.mode))
		}
	}
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

func (q *Query[R]) Build() (bq *BuiltQuery, err error) {
	cfg := &QueryConfig{dialect: defaultDialect}
	return q.BuildWith(cfg)
}

func (q *Query[R]) BuildWith(cfg *QueryConfig) (bq *BuiltQuery, err error) {
	w := newQueryWriter()

	w.Write("SELECT ")
	if q.distinct {
		w.Write("DISTINCT ")
	}
	for i, sel := range q.selections {
		if i > 0 {
			w.Write(", ")
		}
		sel.Expr().appendExpr(w, cfg)
		alias := sel.Alias()
		if alias != "" {
			w.Printf(" AS %s", alias)
		}
	}

	if q.from != nil {
		w.Write(" FROM ")
		q.from.appendTable(w, cfg)
	}

	if len(q.innerJoins) > 0 {
		for _, j := range q.innerJoins {
			w.Write(" ")
			w.Write(j.mode)
			w.Write(" JOIN ")
			j.table.appendTable(w, cfg)
			w.Write(" ON ")
			j.condition.appendExpr(w, cfg)
		}
	}

	if len(q.wheres) > 0 {
		w.Write(" WHERE ")
		andAll(q.wheres...).appendExpr(w, cfg)
	}

	if len(q.groups) > 0 {
		w.Write(" GROUP BY ")
		for i, e := range q.groups {
			if i > 0 {
				w.Write(", ")
			}
			e.appendExpr(w, cfg)
		}
	}

	if len(q.orders) > 0 {
		w.Write(" ORDER BY ")
		for i, e := range q.orders {
			if i > 0 {
				w.Write(", ")
			}
			e.appendExpr(w, cfg)
		}
	}

	if q.limit > 0 {
		w.Printf(" LIMIT %d", q.limit)
	}

	// Check w.errs

	return &BuiltQuery{
		Query: w.String(),
		Args:  w.Args,
	}, nil
}

func andAll(exprs ...Expr) Expr {
	e := exprs[0]
	for i := 1; i < len(exprs); i++ {
		e = e.And(exprs[i])
	}
	return e
}

func (q *Query[R]) Load(ctx context.Context, db QueryRunner) (recs []R, err error) {
	l := &SliceLoader[R, R]{query: q, mapper: q.mapper}
	return l.Load(ctx, db)
}

func (q *Query[R]) LoadRows(ctx context.Context, db QueryRunner) (rows *sql.Rows, err error) {
	bq, err := q.Build()
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, bq.Query, bq.Args...)
}

func (q *Query[R]) WillScan(scanners ...RowsScanner) *MultiScanLoader[R] {
	sels := make([]Selection, 0)
	for _, s := range scanners {
		sels = append(sels, s.Selections()...)
	}
	q.selections = sels
	return &MultiScanLoader[R]{query: q, scanners: scanners}
}

func (q *Query[R]) appendExpr(w *queryWriter, c *QueryConfig) {
	bq, err := q.BuildWith(c)
	if err != nil {
		w.AddErr(err)
	}
	w.Write("(")
	w.Write(bq.Query, bq.Args...)
	w.Write(")")
}

type QueryTable[R any] struct {
	query *Query[R]
	alias string
}

func (t *QueryTable[R]) TableName() string {
	return t.alias
}

func (t *QueryTable[R]) Expr() Expr {
	return t.query
}

func (t *QueryTable[R]) Alias() string {
	return t.alias
}

func (t *QueryTable[R]) appendTable(w *queryWriter, cfg *QueryConfig) {
	t.query.appendExpr(w, cfg)
	if t.alias != "" {
		w.Write(" AS ")
		w.Write(t.alias)
	}
}

type queryWriter struct {
	sb   *strings.Builder
	Args []any
	errs []error
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

func (w *queryWriter) AddErr(err error) {
	w.errs = append(w.errs, err)
}

func (w *queryWriter) Printf(format string, fmtargs ...any) {
	fmt.Fprintf(w.sb, format, fmtargs...)
}
