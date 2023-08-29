package geq

import (
	"fmt"
	"strings"
)

type Expr interface {
	Selection
	As(name string) *Aliased
	Eq(v any) Expr
	appendQuery(p *queryPart)
}

type Aliased struct {
	expr  Expr
	alias string
}

func (a *Aliased) Expr() Expr    { return a.expr }
func (a *Aliased) Alias() string { return a.alias }

type Column[F any] struct {
	ops
	TableName  string
	ColumnName string
}

func NewColumn[F any](tableName, columnName string) *Column[F] {
	return implOps(&Column[F]{TableName: tableName, ColumnName: columnName})
}

func (c *Column[F]) appendQuery(p *queryPart) {
	fmt.Fprintf(p.sb, "%s.%s", c.TableName, c.ColumnName)
}

func buildExprPart(expr Expr) *queryPart {
	p := &queryPart{sb: new(strings.Builder)}
	expr.appendQuery(p)
	return p
}

func lift(v any) Expr {
	switch val := v.(type) {
	case Expr:
		return val
	case *Aliased:
		return val.expr
	default:
		return implOps(&litExpr{val: val})
	}
}

type litExpr struct {
	ops
	val any
}

func (e *litExpr) appendQuery(p *queryPart) {
	p.sb.WriteString("?")
	p.args = append(p.args, e.val)
}

type infixExpr struct {
	ops
	op    string
	left  Expr
	right Expr
}

func (e *infixExpr) appendQuery(p *queryPart) {
	e.left.appendQuery(p)
	fmt.Fprintf(p.sb, " %s ", e.op)
	e.right.appendQuery(p)
}
