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
