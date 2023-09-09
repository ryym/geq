// Code generated by geq. DO NOT EDIT.
// https://github.com/ryym/geq

package tests

import (
	"github.com/ryym/geq"
)

type PostStats struct {
	AuthorID  geq.Expr
	PostCount geq.Expr
	LastTitle geq.Expr
}

func (m *PostStats) FieldPtrs(r *PostStat) []any {
	return []any{&r.AuthorID, &r.PostCount, &r.LastTitle}
}
func (m *PostStats) Selections() []geq.Selection {
	return []geq.Selection{m.AuthorID, m.PostCount, m.LastTitle}
}
