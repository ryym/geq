package reports

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b"
)

func SomeInnerFunc() string {
	return "this does nothing"
}

func PostStatsQuery() (*geq.BuiltQuery, error) {
	q := b.SelectAs(&PostStats{
		AuthorID:  b.Posts.AuthorID,
		PostCount: b.Count(b.Posts.ID),
		LastTitle: b.Max(b.Posts.Title),
	}).GroupBy(b.Posts.AuthorID)
	// postStats, err := q.Load(ctx, db)
	return q.Build()
}

func SameNameUsersQuery() (*geq.BuiltQuery, error) {
	q := b.SelectAs(&SameNameUsers{
		Name:  b.Users.Name,
		Count: b.Count(b.Users.ID),
	}).From(b.Users).GroupBy(b.Users.Name)
	// sameNameUsers, err := q.Load(ctx, db)
	return q.Build()
}
