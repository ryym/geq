package reports

import (
	"github.com/ryym/geq"
)

func SomeInnerFunc() string {
	return "this does nothing"
}

func PostStatsQuery() (*geq.BuiltQuery, error) {
	panic("TODO")
	// q := geq.SelectAs(&PostStats{
	// 	AuthorID:  b.Posts.AuthorID,
	// 	PostCount: geq.Count(b.Posts.ID),
	// 	LastTitle: geq.Max(b.Posts.Title),
	// }).GroupBy(b.Posts.AuthorID)
	// // postStats, err := q.Load(ctx, db)
	// return q.Build()
}

func SameNameUsersQuery() (*geq.BuiltQuery, error) {
	panic("TODO")
	// q := geq.SelectAs(&SameNameUsers{
	// 	Name:  b.Users.Name,
	// 	Count: geq.Count(b.Users.ID),
	// }).From(b.Users).GroupBy(b.Users.Name)
	// // sameNameUsers, err := q.Load(ctx, db)
	// return q.Build()
}
