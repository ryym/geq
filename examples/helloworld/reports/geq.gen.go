// Code generated by geq. DO NOT EDIT.
// https://github.com/ryym/geq

package reports

import (
	"github.com/ryym/geq"
)

func init() {
}

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

type SameNameUsers struct {
	Name  geq.Expr
	Count geq.Expr
}

func (m *SameNameUsers) FieldPtrs(r *SameNameUser) []any {
	return []any{&r.Name, &r.Count}
}
func (m *SameNameUsers) Selections() []geq.Selection {
	return []geq.Selection{m.Name, m.Count}
}
