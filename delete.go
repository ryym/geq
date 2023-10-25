package geq

import (
	"context"
	"database/sql"
)

type DeleteQuery struct {
	table  AnyTable
	wheres []Expr
}

func newDeleteQuery(table AnyTable) *DeleteQuery {
	return &DeleteQuery{table: table, wheres: nil}
}

func (q *DeleteQuery) Where(exprs ...Expr) *DeleteQuery {
	q.wheres = append(q.wheres, exprs...)
	return q
}

func (q *DeleteQuery) Build() (bq *BuiltQuery, err error) {
	cfg := &QueryConfig{dialect: defaultDialect}
	return q.BuildWith(cfg)
}

func (q *DeleteQuery) BuildWith(cfg *QueryConfig) (bq *BuiltQuery, err error) {
	w := newQueryWriter()
	w.Write("DELETE FROM ")
	w.Write(cfg.dialect.Ident(q.table.getTableName()))

	if len(q.wheres) > 0 {
		w.Write(" WHERE ")
		andAll(q.wheres...).appendExpr(w, cfg)
	}

	return &BuiltQuery{Query: w.String(), Args: w.Args}, nil
}

func (q *DeleteQuery) Exec(ctx context.Context, db QueryExecutor) (result sql.Result, err error) {
	bq, err := q.Build()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, bq.Query, bq.Args...)
}
