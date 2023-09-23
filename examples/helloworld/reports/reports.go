package reports

import (
	"github.com/ryym/geq"
	"github.com/ryym/geq/examples/helloworld/gen/d"
)

func SomeInnerFunc() string {
	return "this does nothing"
}

func PostStatsQuery() (*geq.BuiltQuery, error) {
	q := geq.SelectAs(&PostStats{
		AuthorID:  d.Posts.AuthorID,
		PostCount: geq.Count(d.Posts.ID),
		LastTitle: geq.Max(d.Posts.Title),
	}).GroupBy(d.Posts.AuthorID)
	// postStats, err := q.Load(ctx, db)
	return q.Build()
}

func SameNameUsersQuery() (*geq.BuiltQuery, error) {
	q := geq.SelectAs(&SameNameUsers{
		Name:  d.Users.Name,
		Count: geq.Count(d.Users.ID),
	}).From(d.Users).GroupBy(d.Users.Name)
	// sameNameUsers, err := q.Load(ctx, db)
	return q.Build()
}
