package b

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/mdl"
)

var Users *TableUsers
var Posts *TablePosts
var Transactions *TableTransactions

func init() {
	Users = NewUsers("")
	Posts = NewPosts("")
	Transactions = NewTransactions("")
	Posts.Author = geq.NewRelship(Users, Posts.AuthorID, Users.ID)
}

func AsMap[R any, K comparable](key *geq.Column[K], q *geq.Query[R]) *geq.MapLoader[R, R, K] {
	return geq.Builder_AsMap(key, q)
}

func AsSliceMap[R any, K comparable](key *geq.Column[K], q *geq.Query[R]) *geq.SliceMapLoader[R, R, K] {
	return geq.Builder_AsSliceMap(key, q)
}

func ToSlice[R any](mapper geq.RowMapper[R], dest *[]R) *geq.SliceScanner[R] {
	return geq.Builder_ToSlice(mapper, dest)
}

func ToMap[R any, K comparable](mapper geq.RowMapper[R], key geq.TypedSelection[K], dest *map[K]R) *geq.MapScanner[R, K] {
	return geq.Builder_ToMap(mapper, key, dest)
}

func ToSliceMap[R any, K comparable](mapper geq.RowMapper[R], key geq.TypedSelection[K], dest *map[K][]R) *geq.SliceMapScanner[R, K] {
	return geq.Builder_ToSliceMap(mapper, key, dest)
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

func InsertInto(table geq.AnyTable) *geq.InsertQuery {
	return geq.Builder_InsertInto(table)
}

func Update(table geq.AnyTable) *geq.UpdateQuery {
	return geq.Builder_Update(table)
}

func Null() geq.Expr {
	return geq.Builder_Null()
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

// -------------

type TableUsers struct {
	*geq.TableBase
	ID   *geq.Column[int64]
	Name *geq.Column[string]
}

func NewUsers(alias string) *TableUsers {
	t := &TableUsers{
		ID:   geq.NewColumn[int64]("users", "id"),
		Name: geq.NewColumn[string]("users", "name"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name}
	sels := []geq.Selection{t.ID, t.Name}
	t.TableBase = geq.NewTableBase("users", alias, columns, sels)
	return t
}

func (t *TableUsers) FieldPtrs(u *mdl.User) []any {
	return []any{&u.ID, &u.Name}
}

func (t *TableUsers) As(alias string) *TableUsers {
	return NewUsers(alias)
}

type TablePosts struct {
	*geq.TableBase
	ID       *geq.Column[int64]
	AuthorID *geq.Column[int64]
	Title    *geq.Column[string]

	Author *geq.Relship[mdl.User, int64]
}

func NewPosts(alias string) *TablePosts {
	t := &TablePosts{
		ID:       geq.NewColumn[int64]("posts", "id"),
		AuthorID: geq.NewColumn[int64]("posts", "author_id"),
		Title:    geq.NewColumn[string]("posts", "title"),
	}
	columns := []geq.AnyColumn{t.ID, t.AuthorID, t.Title}
	sels := []geq.Selection{t.ID, t.AuthorID, t.Title}
	t.TableBase = geq.NewTableBase("posts", alias, columns, sels)
	return t
}

func (t *TablePosts) FieldPtrs(p *mdl.Post) []any {
	return []any{&p.ID, &p.AuthorID, &p.Title}
}

func (t *TablePosts) As(alias string) *TablePosts {
	return NewPosts(alias)
}

type TableTransactions struct {
	*geq.TableBase
	ID          *geq.Column[uint32]
	UserID      *geq.Column[uint32]
	Amount      *geq.Column[int32]
	Description *geq.Column[string]
}

func NewTransactions(alias string) *TableTransactions {
	t := &TableTransactions{
		ID:          geq.NewColumn[uint32]("transactions", "id"),
		UserID:      geq.NewColumn[uint32]("transactions", "user_id"),
		Amount:      geq.NewColumn[int32]("transactions", "amount"),
		Description: geq.NewColumn[string]("transactions", "description"),
	}
	columns := []geq.AnyColumn{t.ID, t.UserID, t.Amount, t.Description}
	sels := []geq.Selection{t.ID, t.UserID, t.Amount, t.Description}
	t.TableBase = geq.NewTableBase("transactions", alias, columns, sels)
	return t
}

func (t *TableTransactions) FieldPtrs(r *mdl.Transaction) []any {
	return []any{&r.ID, &r.UserID, &r.Amount, &r.Description}
}

func (t *TableTransactions) As(alias string) *TableTransactions {
	return NewTransactions(alias)
}
