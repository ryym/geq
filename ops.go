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
		right: lift(v),
		op:    "=",
	})
}
