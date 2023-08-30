package tests

import "github.com/ryym/geq"

type User struct {
	ID   int64
	Name string
}
type Post struct {
	ID       int64
	AuthorID int64
	Title    string
}

type UsersTable struct {
	ID      *geq.Column[int64]
	Name    *geq.Column[string]
	columns []geq.Selection
}

func NewUsersTable() *UsersTable {
	t := &UsersTable{
		ID:   geq.NewColumn[int64]("users", "id"),
		Name: geq.NewColumn[string]("users", "name"),
	}
	t.columns = []geq.Selection{t.ID, t.Name}
	return t
}

func (t *UsersTable) TableName() string {
	return "users"
}

func (t *UsersTable) FieldPtrs(u *User) []any {
	return []any{&u.ID, &u.Name}
}

func (t *UsersTable) Selections() []geq.Selection {
	return t.columns
}

type PostsTable struct {
	ID       *geq.Column[int64]
	AuthorID *geq.Column[int64]
	Title    *geq.Column[string]
	columns  []geq.Selection

	Author *geq.Relship[User, int64]
}

func NewPostsTable() *PostsTable {
	t := &PostsTable{
		ID:       geq.NewColumn[int64]("posts", "id"),
		AuthorID: geq.NewColumn[int64]("posts", "author_id"),
		Title:    geq.NewColumn[string]("posts", "title"),
	}
	t.columns = []geq.Selection{t.ID, t.AuthorID, t.Title}
	return t
}

func (t *PostsTable) TableName() string {
	return "posts"
}

func (t *PostsTable) FieldPtrs(p *Post) []any {
	return []any{&p.ID, &p.AuthorID, &p.Title}
}

func (t *PostsTable) Selections() []geq.Selection {
	return t.columns
}

type QueryBuilder struct {
	Users *UsersTable
	Posts *PostsTable
}

func NewQueryBuilder() *QueryBuilder {
	b := &QueryBuilder{
		Users: NewUsersTable(),
		Posts: NewPostsTable(),
	}
	b.Posts.Author = geq.NewRelship(b.Users, b.Posts.AuthorID, b.Users.ID)
	return b
}
