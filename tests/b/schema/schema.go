package schema

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/mdl"
)

type UsersTable struct {
	*geq.TableBase
	ID   *geq.Column[int64]
	Name *geq.Column[string]
}

func NewUsersTable(alias string) *UsersTable {
	t := &UsersTable{
		ID:   geq.NewColumn[int64]("users", "id"),
		Name: geq.NewColumn[string]("users", "name"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name}
	sels := []geq.Selection{t.ID, t.Name}
	t.TableBase = geq.NewTableBase("users", alias, columns, sels)
	return t
}

func (t *UsersTable) FieldPtrs(u *mdl.User) []any {
	return []any{&u.ID, &u.Name}
}

func (t *UsersTable) As(alias string) *UsersTable {
	return NewUsersTable(alias)
}

type PostsTable struct {
	*geq.TableBase
	ID       *geq.Column[int64]
	AuthorID *geq.Column[int64]
	Title    *geq.Column[string]

	Author *geq.Relship[mdl.User, int64]
}

func NewPostsTable(alias string) *PostsTable {
	t := &PostsTable{
		ID:       geq.NewColumn[int64]("posts", "id"),
		AuthorID: geq.NewColumn[int64]("posts", "author_id"),
		Title:    geq.NewColumn[string]("posts", "title"),
	}
	columns := []geq.AnyColumn{t.ID, t.AuthorID, t.Title}
	sels := []geq.Selection{t.ID, t.AuthorID, t.Title}
	t.TableBase = geq.NewTableBase("posts", alias, columns, sels)
	return t
}

func (t *PostsTable) FieldPtrs(p *mdl.Post) []any {
	return []any{&p.ID, &p.AuthorID, &p.Title}
}

func (t *PostsTable) As(alias string) *PostsTable {
	return NewPostsTable(alias)
}
