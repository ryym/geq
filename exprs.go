package geq

import (
	"fmt"
	"strings"
)

type Expr interface {
	Selection
	As(name string) *Aliased
	appendQuery(p *queryPart)
}

type Aliased struct {
	expr  Expr
	alias string
}

func (a *Aliased) Expr() Expr    { return a.expr }
func (a *Aliased) Alias() string { return a.alias }

type Column[F any] struct {
	TableName  string
	ColumnName string
}

func (c *Column[F]) Expr() Expr    { return c }
func (c *Column[F]) Alias() string { return "" }

func (c *Column[F]) As(alias string) *Aliased {
	return &Aliased{expr: c, alias: alias}
}

func (c *Column[F]) appendQuery(p *queryPart) {
	fmt.Fprintf(p.sb, "%s.%s", c.TableName, c.ColumnName)
}

func buildExprPart(expr Expr) *queryPart {
	p := &queryPart{sb: new(strings.Builder)}
	expr.appendQuery(p)
	return p
}
