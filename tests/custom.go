package tests

import "github.com/ryym/geq"

type PostStat struct {
	AuthorID  int64
	PostCount int64
	LastTitle string
}

type PostStats struct {
	AuthorID  geq.Expr
	PostCount geq.Expr
	LastTitle geq.Expr
}

func (r *PostStats) FieldPtrs(p *PostStat) []any {
	return []any{&p.AuthorID, &p.PostCount, &p.LastTitle}
}

func (r *PostStats) Selections() []geq.Selection {
	return []geq.Selection{r.AuthorID, r.PostCount, r.LastTitle}
}
