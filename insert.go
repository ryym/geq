package geq

import (
	"errors"
)

type ValueMap map[AnyColumn]any

type ValuePair struct {
	column AnyColumn
	value  any
}

type InsertQuery struct {
	table     AnyTable
	rowValues []map[AnyColumn]Expr
}

func newInsertQuery(table AnyTable) *InsertQuery {
	return &InsertQuery{table: table}
}

func (q *InsertQuery) Values(sets ...ValuePair) *InsertQuery {
	m := make(map[AnyColumn]Expr, len(sets))
	for _, p := range sets {
		m[p.column] = toExpr(p.value)
	}
	q.rowValues = append(q.rowValues, m)
	return q
}

func (q *InsertQuery) ValueMaps(vms ...ValueMap) *InsertQuery {
	ems := make([]map[AnyColumn]Expr, 0, len(vms))
	for _, vm := range vms {
		em := make(map[AnyColumn]Expr, len(vm))
		for k, v := range vm {
			em[k] = toExpr(v)
		}
		ems = append(ems, em)
	}
	q.rowValues = append(q.rowValues, ems...)
	return q
}

func (q *InsertQuery) Finalize() (fq *FinalQuery, err error) {
	w := newQueryWriter()

	w.Printf("INSERT INTO %s ", q.table.TableName())

	if len(q.rowValues) == 0 {
		return nil, errors.New("[geq.InsertInto] values must be provided")
	}

	valsLen := len(q.rowValues[0])
	for _, m := range q.rowValues {
		if valsLen != len(m) {
			return nil, errors.New("[geq.InsertInto] all values length must be same")
		}
	}
	if valsLen == 0 {
		return nil, errors.New("[geq.InsertInto] values must not be empty")
	}

	columns := make([]AnyColumn, 0, valsLen)
	for _, c := range q.table.Columns() {
		_, ok := q.rowValues[0][c]
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
	w.Write(") ")

	w.Write("VALUES ")

	for i, m := range q.rowValues {
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
				return nil, errors.New("[geq.InsertInto] all values columns must be same")
			}
			v.appendExpr(w)
		}
		w.Write(")")
	}

	return &FinalQuery{Query: w.String(), Args: w.Args}, nil
}
