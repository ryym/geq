// Code generated by geq. DO NOT EDIT.
// https://github.com/ryym/geq

package d

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/internal/tests/mdl"
	"time"
)

var Users = NewUsers("users")
var Posts = NewPosts("posts")
var Transactions = NewTransactions("transactions")

func init() {
	Posts.Author = geq.NewRelship(Users, Posts.AuthorID, Users.ID)
}

type TableUsers struct {
	*geq.TableBase
	ID   *geq.Column[int64]
	Name *geq.Column[string]
}

func NewUsers(alias string) *TableUsers {
	t := &TableUsers{
		ID:   geq.NewColumn[int64](alias, "id"),
		Name: geq.NewColumn[string](alias, "name"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name}
	sels := []geq.Selection{t.ID, t.Name}
	t.TableBase = geq.NewTableBase("users", alias, columns, sels)
	return t
}
func (t *TableUsers) FieldPtrs(r *mdl.User) []any {
	return []any{&r.ID, &r.Name}
}
func (t *TableUsers) As(alias string) *TableUsers {
	return NewUsers(alias)
}

type TablePosts struct {
	*geq.TableBase
	ID       *geq.Column[int64]
	AuthorID *geq.Column[int64]
	Title    *geq.Column[string]
	Author   *geq.Relship[*TableUsers, mdl.User, int64]
}

func NewPosts(alias string) *TablePosts {
	t := &TablePosts{
		ID:       geq.NewColumn[int64](alias, "id"),
		AuthorID: geq.NewColumn[int64](alias, "author_id"),
		Title:    geq.NewColumn[string](alias, "title"),
	}
	columns := []geq.AnyColumn{t.ID, t.AuthorID, t.Title}
	sels := []geq.Selection{t.ID, t.AuthorID, t.Title}
	t.TableBase = geq.NewTableBase("posts", alias, columns, sels)
	return t
}
func (t *TablePosts) FieldPtrs(r *mdl.Post) []any {
	return []any{&r.ID, &r.AuthorID, &r.Title}
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
	CreatedAt   *geq.Column[time.Time]
}

func NewTransactions(alias string) *TableTransactions {
	t := &TableTransactions{
		ID:          geq.NewColumn[uint32](alias, "id"),
		UserID:      geq.NewColumn[uint32](alias, "user_id"),
		Amount:      geq.NewColumn[int32](alias, "amount"),
		Description: geq.NewColumn[string](alias, "description"),
		CreatedAt:   geq.NewColumn[time.Time](alias, "created_at"),
	}
	columns := []geq.AnyColumn{t.ID, t.UserID, t.Amount, t.Description, t.CreatedAt}
	sels := []geq.Selection{t.ID, t.UserID, t.Amount, t.Description, t.CreatedAt}
	t.TableBase = geq.NewTableBase("transactions", alias, columns, sels)
	return t
}
func (t *TableTransactions) FieldPtrs(r *mdl.Transaction) []any {
	return []any{&r.ID, &r.UserID, &r.Amount, &r.Description, &r.CreatedAt}
}
func (t *TableTransactions) As(alias string) *TableTransactions {
	return NewTransactions(alias)
}

type PostStats struct {
	AuthorID  geq.Expr
	PostCount geq.Expr
	LastTitle geq.Expr
}

func (m *PostStats) FieldPtrs(r *mdl.PostStat) []any {
	return []any{&r.AuthorID, &r.PostCount, &r.LastTitle}
}
func (m *PostStats) Selections() []geq.Selection {
	return []geq.Selection{m.AuthorID, m.PostCount, m.LastTitle}
}

type TransactionStats struct {
	UserID        geq.Expr
	TotalAmount   geq.Expr
	LastCreatedAt geq.Expr
}

func (m *TransactionStats) FieldPtrs(r *mdl.TransactionStat) []any {
	return []any{&r.UserID, &r.TotalAmount, &r.LastCreatedAt}
}
func (m *TransactionStats) Selections() []geq.Selection {
	return []geq.Selection{m.UserID, m.TotalAmount, m.LastCreatedAt}
}
