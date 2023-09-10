package geq

import (
	"context"
	"database/sql"
	"errors"
)

type ValueMap map[AnyColumn]any

type ValuePair struct {
	column AnyColumn
	value  Expr
}

type InsertQuery struct {
	table     AnyTable
	valueMaps []map[AnyColumn]Expr
}

func newInsertQuery(table AnyTable) *InsertQuery {
	return &InsertQuery{table: table}
}

func (q *InsertQuery) Values(pairs ...ValuePair) *InsertQuery {
	m := make(map[AnyColumn]Expr, len(pairs))
	for _, p := range pairs {
		m[p.column] = p.value
	}
	q.valueMaps = append(q.valueMaps, m)
	return q
}

func (q *InsertQuery) ValueMaps(vms ...ValueMap) *InsertQuery {
	for _, vm := range vms {
		em := make(map[AnyColumn]Expr, len(vm))
		for k, v := range vm {
			em[k] = toExpr(v)
		}
		q.valueMaps = append(q.valueMaps, em)
	}
	return q
}

func (q *InsertQuery) Finalize() (fq *FinalQuery, err error) {
	cfg := &QueryConfig{dialect: defaultDialect}
	return q.FinalizeWith(cfg)
}

func (q *InsertQuery) FinalizeWith(cfg *QueryConfig) (fq *FinalQuery, err error) {
	w := newQueryWriter()
	w.Printf("INSERT INTO %s ", cfg.dialect.Ident(q.table.TableName()))

	if len(q.valueMaps) == 0 {
		return nil, errors.New("[geq.InsertInto] no values provided")
	}

	valsLen := len(q.valueMaps[0])
	if valsLen == 0 {
		return nil, errors.New("[geq.InsertInto] values empty")
	}

	columns := make([]AnyColumn, 0, valsLen)
	for _, c := range q.table.Columns() {
		_, ok := q.valueMaps[0][c]
		if ok {
			columns = append(columns, c)
		}
	}
	if len(columns) < valsLen {
		return nil, errors.New("[geq.InsertInto] other table columns exist")
	}

	w.Write("(")
	for i, col := range columns {
		if i > 0 {
			w.Write(", ")
		}
		w.Write(col.ColumnName())
	}
	w.Write(") VALUES ")

	for i, m := range q.valueMaps {
		if len(m) != valsLen {
			return nil, errors.New("[geq.InsertInto] values length not match")
		}
		if i > 0 {
			w.Write(", ")
		}
		w.Write("(")
		for i, c := range columns {
			if i > 0 {
				w.Write(", ")
			}
			v, ok := m[c]
			if !ok {
				return nil, errors.New("[geq.InsertInto] values columns not match")
			}
			v.appendExpr(w, cfg)
		}
		w.Write(")")
	}

	return &FinalQuery{Query: w.String(), Args: w.Args}, nil
}

func (q *InsertQuery) Exec(ctx context.Context, db QueryExecutor) (result sql.Result, err error) {
	fq, err := q.Finalize()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, fq.Query, fq.Args...)
}
