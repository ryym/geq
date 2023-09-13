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

func (o *ops) Expr() Expr    { return o.expr }
func (o *ops) Alias() string { return "" }

func (o *ops) As(alias string) Aliased {
	return &aliased{expr: o.expr, alias: alias}
}

func (o *ops) Eq(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "=",
	})
}

func (o *ops) Neq(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<>",
	})
}

func (o *ops) Gt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    ">",
	})
}

func (o *ops) Gte(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    ">=",
	})
}

func (o *ops) Lt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<",
	})
}

func (o *ops) Lte(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<=",
	})
}

func (o *ops) Add(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "+",
	})
}

func (o *ops) Sbt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "-",
	})
}

func (o *ops) Mlt(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "*",
	})
}

func (o *ops) Dvd(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "/",
	})
}

func (o *ops) IsNull() AnonExpr {
	return implOps(&suffixExpr{
		val: o.expr,
		op:  "IS NULL",
	})
}

func (o *ops) IsNotNull() AnonExpr {
	return implOps(&suffixExpr{
		val: o.expr,
		op:  "IS NOT NULL",
	})
}

func (o *ops) LikePrefix(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: newConcatExpr(v, newRawExpr("'%'")),
		op:    "LIKE",
	})
}

func (o *ops) LikeSuffix(v any) AnonExpr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: newConcatExpr(newRawExpr("'%'"), v),
		op:    "LIKE",
	})
}

func (o *ops) LikePartial(v any) AnonExpr {
	pc := newRawExpr("'%'")
	return implOps(&infixExpr{
		left:  o.expr,
		right: newConcatExpr(pc, v, pc),
		op:    "LIKE",
	})
}

func (o *ops) InAny(vals ...any) AnonExpr {
	return implOps(&inExpr{
		operand: o.expr,
		values:  vals,
	})
}
