package geq

import (
	"strings"
)

type Expr interface {
	Selection
	As(name string) Aliased
	Eq(v any) Expr
	appendExpr(w *queryWriter)
}

type Aliased interface {
	Expr() Expr
	Alias() string
}

type aliased struct {
	expr  Expr
	alias string
}

func (a *aliased) Expr() Expr    { return a.expr }
func (a *aliased) Alias() string { return a.alias }

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

func (c *Column[F]) appendExpr(w *queryWriter) {
	w.Printf("%s.%s", c.tableName, c.columnName)
}

func (c *Column[F]) In(values []F) Expr {
	anyVals := make([]any, 0, len(values))
	for _, v := range values {
		anyVals = append(anyVals, v)
	}
	return implOps(&inExpr{operand: c, values: anyVals})
}

func lift(v any) Expr {
	switch val := v.(type) {
	case Expr:
		return val
	case *aliased:
		return val.expr
	default:
		return implOps(&litExpr{val: val})
	}
}

type litExpr struct {
	ops
	val any
}

func (e *litExpr) appendExpr(w *queryWriter) {
	w.Write("?", e.val)
}

type infixExpr struct {
	ops
	op    string
	left  Expr
	right Expr
}

func (e *infixExpr) appendExpr(w *queryWriter) {
	e.left.appendExpr(w)
	w.Printf(" %s ", e.op)
	e.right.appendExpr(w)
}

type inExpr struct {
	ops
	operand Expr
	values  []any
}

func (e *inExpr) appendExpr(w *queryWriter) {
	placeholders := strings.Repeat(",?", len(e.values))[1:]
	e.operand.appendExpr(w)
	w.Printf(" IN (%s)", placeholders)
	w.Args = append(w.Args, e.values...)
}

type funcExpr struct {
	ops
	name string
	args []Expr
}

func (e *funcExpr) appendExpr(w *queryWriter) {
	w.Write(e.name)
	w.Write("(")
	for i, arg := range e.args {
		if i > 0 {
			w.Write(", ")
		}
		arg.appendExpr(w)
	}
	w.Write(")")
}
