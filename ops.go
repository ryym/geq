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

func (o *ops) Eq(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "=",
	})
}

func (o *ops) Neq(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<>",
	})
}

func (o *ops) Gt(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    ">",
	})
}

func (o *ops) Gte(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    ">=",
	})
}

func (o *ops) Lt(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<",
	})
}

func (o *ops) Lte(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "<=",
	})
}

func (o *ops) Add(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "+",
	})
}

func (o *ops) Sbt(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "-",
	})
}

func (o *ops) Mlt(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "*",
	})
}

func (o *ops) Dvd(v any) Expr {
	return implOps(&infixExpr{
		left:  o.expr,
		right: toExpr(v),
		op:    "/",
	})
}

func (o *ops) IsNull() Expr {
	return implOps(&suffixExpr{
		val: o.expr,
		op:  "IS NULL",
	})
}

func (o *ops) IsNotNull() Expr {
	return implOps(&suffixExpr{
		val: o.expr,
		op:  "IS NOT NULL",
	})
}
