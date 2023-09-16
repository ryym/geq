package geq

import "fmt"

type Expr interface {
	Selection

	Eq(v any) AnonExpr
	Neq(v any) AnonExpr
	Gt(v any) AnonExpr
	Gte(v any) AnonExpr
	Lt(v any) AnonExpr
	Lte(v any) AnonExpr
	Add(v any) AnonExpr
	Sbt(v any) AnonExpr
	Mlt(v any) AnonExpr
	Dvd(v any) AnonExpr
	LikePrefix(v any) AnonExpr
	LikeSuffix(v any) AnonExpr
	LikePartial(v any) AnonExpr
	InAny(vals ...any) AnonExpr
	IsNull() AnonExpr
	IsNotNull() AnonExpr

	appendExpr(w *queryWriter, c *QueryConfig)
}

type AnonExpr interface {
	Expr
	As(name string) Aliased
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

func (c *Column[F]) appendExpr(w *queryWriter, cfg *QueryConfig) {
	w.Printf("%s.%s", cfg.dialect.Ident(c.tableName), cfg.dialect.Ident(c.columnName))
}

func (c *Column[F]) In(values []F) Expr {
	anyVals := make([]any, 0, len(values))
	for _, v := range values {
		anyVals = append(anyVals, v)
	}
	return implOps(&inExpr{operand: c, values: anyVals})
}

func (c *Column[F]) Set(value F) ValuePair {
	return ValuePair{column: c, value: toExpr(value)}
}

func toExpr(v any) Expr {
	if v == nil {
		return implOps(&nullExpr{})
	}
	switch val := v.(type) {
	case Expr:
		return val
	case *aliased:
		return val.expr
	default:
		return implOps(&litExpr{val: val})
	}
}

type nullExpr struct {
	ops
}

func (e *nullExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	w.Write("NULL")
}

type litExpr struct {
	ops
	val any
}

func (e *litExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	w.Write(cfg.dialect.Placeholder("", w.Args), e.val)
}

type infixExpr struct {
	ops
	op    string
	left  Expr
	right Expr
}

func (e *infixExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	e.left.appendExpr(w, cfg)
	w.Printf(" %s ", e.op)
	e.right.appendExpr(w, cfg)
}

type suffixExpr struct {
	ops
	op  string
	val Expr
}

func (e *suffixExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	e.val.appendExpr(w, cfg)
	w.Write(" ")
	w.Write(e.op)
}

type concatExpr struct {
	ops
	vals []Expr
}

func newConcatExpr(vals ...any) *concatExpr {
	exprs := make([]Expr, 0, len(vals))
	for _, v := range vals {
		exprs = append(exprs, toExpr(v))
	}
	return implOps(&concatExpr{vals: exprs})
}

func (e *concatExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	stype := cfg.dialect.StrConcatType()
	switch stype {
	case StrConcatStandard:
		for i, v := range e.vals {
			if i > 0 {
				w.Write(" || ")
			}
			v.appendExpr(w, cfg)
		}
	case StrConcatFunc:
		fe := &funcExpr{name: "CONCAT", args: e.vals}
		fe.appendExpr(w, cfg)
	default:
		panic(fmt.Sprintf("unknown string concat type: %v", stype))
	}
}

type inExpr struct {
	ops
	operand Expr
	values  []any
}

func (e *inExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	e.operand.appendExpr(w, cfg)
	w.Write(" IN (")
	for i, v := range e.values {
		if i > 0 {
			w.Write(", ")
		}
		toExpr(v).appendExpr(w, cfg)
	}
	w.Write(")")
}

type funcExpr struct {
	ops
	name string
	args []Expr
}

func (e *funcExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	w.Write(e.name)
	w.Write("(")
	for i, arg := range e.args {
		if i > 0 {
			w.Write(", ")
		}
		arg.appendExpr(w, cfg)
	}
	w.Write(")")
}

type RawExpr struct {
	ops
	sql string
}

func newRawExpr(sql string) *RawExpr {
	return implOps(&RawExpr{sql: sql})
}

func (e *RawExpr) appendExpr(w *queryWriter, cfg *QueryConfig) {
	w.Write(e.sql)
}

func (e *RawExpr) appendTable(w *queryWriter, cfg *QueryConfig) {
	w.Write(e.sql)
}
