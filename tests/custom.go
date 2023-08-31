package tests

import "github.com/ryym/geq"

type PostStat struct {
	AuthorID int64
	Foo      bool
}

type PostStats struct {
	AuthorID geq.Expr
	Foo      geq.Expr
}

func (r *PostStats) FieldPtrs(p *PostStat) []any {
	return []any{&p.AuthorID, &p.Foo}
}

func (r *PostStats) Selections() []geq.Selection {
	return []geq.Selection{r.AuthorID, r.Foo}
}
