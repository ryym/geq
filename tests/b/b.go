package b

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b/schema"
)

var Users *schema.UsersTable
var Posts *schema.PostsTable

func init() {
	Users = schema.NewUsersTable()
	Posts = schema.NewPostsTable()
	Posts.Author = geq.NewRelship(Users, Posts.AuthorID, Users.ID)
}

func Via[S, T, C any](srcs []S, from geq.Table[T], relship *geq.Relship[S, C]) *geq.Query[T] {
	return geq.Via(srcs, from, relship)
}

func From[R any](table geq.Table[R]) *geq.Query[R] {
	return geq.From(table)
}

func FromNothing() *geq.Query[struct{}] {
	return geq.FromNothing()
}
