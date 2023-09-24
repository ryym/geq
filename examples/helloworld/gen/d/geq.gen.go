// Code generated by geq. DO NOT EDIT.
// https://github.com/ryym/geq

package d

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/examples/helloworld/mdl"
)

var Users = NewUsers("users")
var Posts = NewPosts("posts")
var Countries = NewCountries("countries")
var Cities = NewCities("cities")

func init() {
	Users.InitRelships()
	Posts.InitRelships()
	Countries.InitRelships()
	Cities.InitRelships()
}

type TableUsers struct {
	*geq.TableBase
	relshipsSet bool
	ID          *geq.Column[uint64]
	Name        *geq.Column[string]
	Posts       *geq.Relship[*TablePosts, mdl.Post, uint64]
}

func NewUsers(alias string) *TableUsers {
	t := &TableUsers{
		ID:   geq.NewColumn[uint64](alias, "id"),
		Name: geq.NewColumn[string](alias, "name"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name}
	sels := []geq.Selection{t.ID, t.Name}
	t.TableBase = geq.NewTableBase("users", alias, columns, sels)
	return t
}

func (t *TableUsers) InitRelships() {
	if t.relshipsSet {
		return
	}
	func() {
		r := NewPosts("posts")
		t.Posts = geq.NewRelship(r, t.ID, r.AuthorID)
	}()
	t.relshipsSet = true
}
func (t *TableUsers) FieldPtrs(r *mdl.User) []any {
	return []any{&r.ID, &r.Name}
}
func (t *TableUsers) As(alias string) *TableUsers {
	return NewUsers(alias)
}

type TablePosts struct {
	*geq.TableBase
	relshipsSet bool
	ID          *geq.Column[uint64]
	Title       *geq.Column[string]
	AuthorID    *geq.Column[uint64]
	Published   *geq.Column[bool]
	Author      *geq.Relship[*TableUsers, mdl.User, uint64]
}

func NewPosts(alias string) *TablePosts {
	t := &TablePosts{
		ID:        geq.NewColumn[uint64](alias, "id"),
		Title:     geq.NewColumn[string](alias, "title"),
		AuthorID:  geq.NewColumn[uint64](alias, "author_id"),
		Published: geq.NewColumn[bool](alias, "published"),
	}
	columns := []geq.AnyColumn{t.ID, t.Title, t.AuthorID, t.Published}
	sels := []geq.Selection{t.ID, t.Title, t.AuthorID, t.Published}
	t.TableBase = geq.NewTableBase("posts", alias, columns, sels)
	return t
}

func (t *TablePosts) InitRelships() {
	if t.relshipsSet {
		return
	}
	func() {
		r := NewUsers("users")
		t.Author = geq.NewRelship(r, t.AuthorID, r.ID)
	}()
	t.relshipsSet = true
}
func (t *TablePosts) FieldPtrs(r *mdl.Post) []any {
	return []any{&r.ID, &r.Title, &r.AuthorID, &r.Published}
}
func (t *TablePosts) As(alias string) *TablePosts {
	return NewPosts(alias)
}

type TableCountries struct {
	*geq.TableBase
	relshipsSet bool
	ID          *geq.Column[uint32]
	Name        *geq.Column[string]
}

func NewCountries(alias string) *TableCountries {
	t := &TableCountries{
		ID:   geq.NewColumn[uint32](alias, "id"),
		Name: geq.NewColumn[string](alias, "name"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name}
	sels := []geq.Selection{t.ID, t.Name}
	t.TableBase = geq.NewTableBase("countries", alias, columns, sels)
	return t
}

func (t *TableCountries) InitRelships() {
	if t.relshipsSet {
		return
	}
	t.relshipsSet = true
}
func (t *TableCountries) FieldPtrs(r *mdl.Country) []any {
	return []any{&r.ID, &r.Name}
}
func (t *TableCountries) As(alias string) *TableCountries {
	return NewCountries(alias)
}

type TableCities struct {
	*geq.TableBase
	relshipsSet bool
	ID          *geq.Column[uint64]
	Name        *geq.Column[string]
	CountryID   *geq.Column[uint32]
}

func NewCities(alias string) *TableCities {
	t := &TableCities{
		ID:        geq.NewColumn[uint64](alias, "id"),
		Name:      geq.NewColumn[string](alias, "name"),
		CountryID: geq.NewColumn[uint32](alias, "country_id"),
	}
	columns := []geq.AnyColumn{t.ID, t.Name, t.CountryID}
	sels := []geq.Selection{t.ID, t.Name, t.CountryID}
	t.TableBase = geq.NewTableBase("cities", alias, columns, sels)
	return t
}

func (t *TableCities) InitRelships() {
	if t.relshipsSet {
		return
	}
	t.relshipsSet = true
}
func (t *TableCities) FieldPtrs(r *mdl.City) []any {
	return []any{&r.ID, &r.Name, &r.CountryID}
}
func (t *TableCities) As(alias string) *TableCities {
	return NewCities(alias)
}
