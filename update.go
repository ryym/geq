package geq

import "errors"

type UpdateQuery struct {
	table    AnyTable
	valueMap map[AnyColumn]Expr
	wheres   []Expr
}

func newUpdateQuery(table AnyTable) *UpdateQuery {
	return &UpdateQuery{table: table}
}

func (q *UpdateQuery) Set(pairs ...ValuePair) *UpdateQuery {
	m := make(map[AnyColumn]Expr, len(pairs))
	for _, p := range pairs {
		m[p.column] = p.value
	}
	q.valueMap = m
	return q
}

func (q *UpdateQuery) SetMap(vm ValueMap) *UpdateQuery {
	em := make(map[AnyColumn]Expr, len(vm))
	for k, v := range vm {
		em[k] = toExpr(v)
	}
	q.valueMap = em
	return q
}

func (q *UpdateQuery) Where(exprs ...Expr) *UpdateQuery {
	q.wheres = append(q.wheres, exprs...)
	return q
}

func (q *UpdateQuery) Finalize() (fq *FinalQuery, err error) {
	w := newQueryWriter()
	w.Write("UPDATE ")
	w.Write(q.table.TableName())
	w.Write(" SET ")

	if len(q.valueMap) == 0 {
		return nil, errors.New("[geq.Update] values empty")
	}

	setWritten := false
	for _, c := range q.table.Columns() {
		if setWritten {
			w.Write(", ")
		}
		v, ok := q.valueMap[c]
		if ok {
			setWritten = true
			w.Write(c.ColumnName())
			w.Write(" = ")
			v.appendExpr(w)
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

	return &FinalQuery{Query: w.String(), Args: w.Args}, nil
}
