package geq

type opsImpl interface {
	Expr
	wrap(expr Expr)
}

func implOps[O opsImpl](o O) O {
	o.wrap(o)
	return o
}

type ops struct {
	expr Expr
}

func (o *ops) wrap(expr Expr) { o.expr = expr }

func (o *ops) getExpr() Expr    { return o.expr }
func (o *ops) getAlias() string { return "" }

func (o *ops) As(alias string) Aliased {
	return &aliased{expr: o.expr, alias: alias}
}

func (o *ops) Eq(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "=",
		precedence: prcdEqual,
	})
}

func (o *ops) Neq(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "<>",
		precedence: prcdEqual,
	})
}

func (o *ops) Gt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         ">",
		precedence: prcdLessGreater,
	})
}

func (o *ops) Gte(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         ">=",
		precedence: prcdLessGreater,
	})
}

func (o *ops) Lt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "<",
		precedence: prcdLessGreater,
	})
}

func (o *ops) Lte(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "<=",
		precedence: prcdLessGreater,
	})
}

func (o *ops) Add(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "+",
		precedence: prcdSum,
	})
}

func (o *ops) Sbt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "-",
		precedence: prcdSum,
	})
}

func (o *ops) Mlt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "*",
		precedence: prcdProduct,
	})
}

func (o *ops) Dvd(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      toExpr(v),
		op:         "/",
		precedence: prcdProduct,
	})
}

func (o *ops) IsNull() AnonExpr {
	return implOps(&suffixExpr{
		val:        o.expr,
		op:         "IS NULL",
		precedence: prcdLowExpr,
	})
}

func (o *ops) IsNotNull() AnonExpr {
	return implOps(&suffixExpr{
		val:        o.expr,
		op:         "IS NOT NULL",
		precedence: prcdLowExpr,
	})
}

func (o *ops) LikePrefix(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      newConcatExpr(v, newRawExpr("'%'")),
		op:         "LIKE",
		precedence: prcdLowExpr,
	})
}

func (o *ops) LikeSuffix(v any) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      newConcatExpr(newRawExpr("'%'"), v),
		op:         "LIKE",
		precedence: prcdLowExpr,
	})
}

func (o *ops) LikePartial(v any) AnonExpr {
	pc := newRawExpr("'%'")
	return implOps(&infixExpr{
		left:       o.expr,
		right:      newConcatExpr(pc, v, pc),
		op:         "LIKE",
		precedence: prcdLowExpr,
	})
}

func (o *ops) InAny(vals ...any) AnonExpr {
	return implOps(&inExpr{
		operand: o.expr,
		values:  vals,
	})
}

func (o *ops) And(e Expr) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      e,
		op:         "AND",
		precedence: prcdAnd,
	})
}

func (o *ops) Or(e Expr) AnonExpr {
	return implOps(&infixExpr{
		left:       o.expr,
		right:      e,
		op:         "OR",
		precedence: prcdOr,
	})
}

func (o *ops) order() orderItem {
	return o.Asc().order()
}

type orderer struct {
	item orderItem
}

func (o *orderer) order() orderItem {
	return o.item
}

func (o *ops) Asc() Orderer {
	return &orderer{item: orderItem{
		order: "ASC",
		expr:  o.expr,
	}}
}

func (o *ops) Desc() Orderer {
	return &orderer{item: orderItem{
		order: "DESC",
		expr:  o.expr,
	}}
}
