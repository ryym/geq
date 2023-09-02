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

type AnyColumn interface {
	Expr
	ColumnName() string
}

type Column[F any] struct {
	ops
	tableName  string
	columnName string
}

func NewColumn[F any](tableName, columnName string) *Column[F] {
	return implOps(&Column[F]{tableName: tableName, columnName: columnName})
}

func (c *Column[F]) ColumnName() string {
	return c.columnName
}

func (c *Column[F]) appendQuery(p *queryPart) {
	fmt.Fprintf(p.sb, "%s.%s", c.tableName, c.columnName)
}

func (c *Column[F]) In(values []F) Expr {
	anyVals := make([]any, 0, len(values))
	for _, v := range values {
		anyVals = append(anyVals, v)
	}
	return implOps(&inExpr{operand: c, values: anyVals})
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

type inExpr struct {
	ops
	operand Expr
	values  []any
}

func (e *inExpr) appendQuery(p *queryPart) {
	placeholders := strings.Repeat(",?", len(e.values))[1:]
	e.operand.appendQuery(p)
	p.sb.WriteString(" IN (")
	p.sb.WriteString(placeholders)
	p.sb.WriteString(")")
	p.args = append(p.args, e.values...)
}

type funcExpr struct {
	ops
	name string
	args []Expr
}

func (e *funcExpr) appendQuery(p *queryPart) {
	p.sb.WriteString(e.name)
	p.sb.WriteRune('(')
	for i, arg := range e.args {
		if i > 0 {
			p.sb.WriteString(", ")
		}
		arg.appendQuery(p)
	}
	p.sb.WriteRune(')')
}
