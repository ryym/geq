package b

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b/schema"
)

var Users *schema.UsersTable
var Posts *schema.PostsTable

func init() {
	Users = schema.NewUsersTable("")
	Posts = schema.NewPostsTable("")
	Posts.Author = geq.NewRelship(Users, Posts.AuthorID, Users.ID)
}

func AsMap[R any, K comparable](key *geq.Column[K], q *geq.Query[R]) *geq.MapLoader[R, R, K] {
	return geq.Builder_AsMap(key, q)
}

func ToSlice[R any](mapper geq.RowMapper[R], dest *[]R) *geq.SliceScanner[R] {
	return geq.Builder_ToSlice(mapper, dest)
}

func ToMap[R any, K comparable](mapper geq.RowMapper[R], key geq.TypedSelection[K], dest *map[K]R) *geq.MapScanner[R, K] {
	return geq.Builder_ToMap(mapper, key, dest)
}

func SelectFrom[R any](table geq.Table[R], sels ...geq.Selection) *geq.Query[R] {
	return geq.Builder_SelectFrom(table, sels...)
}

func SelectAs[R any](mapper geq.RowMapper[R]) *geq.Query[R] {
	return geq.Builder_SelectAs(mapper)
}

func SelectOnly[V any](col *geq.Column[V]) *geq.Query[V] {
	return geq.Builder_SelectOnly(col)
}

func Select(sels ...geq.Selection) *geq.Query[struct{}] {
	return geq.Builder_Select(sels...)
}

func SelectVia[S, T, C any](srcs []S, from geq.Table[T], relship *geq.Relship[S, C]) *geq.Query[T] {
	return geq.Builder_SelectVia(srcs, from, relship)
}

func Func(name string, args ...any) geq.Expr {
	return geq.Builder_Func(name, args...)
}

func Count(expr geq.Expr) geq.Expr {
	return Func("COUNT", expr)
}

func Max(expr geq.Expr) geq.Expr {
	return Func("MAX", expr)
}

func Raw(expr string, args ...any) geq.Expr {
	return geq.Builder_Raw(expr, args...)
}
