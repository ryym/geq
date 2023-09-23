package tests

import "github.com/ryym/geq/internal/tests/mdl"

type GeqTables struct {
	Users        mdl.User
	Posts        mdl.Post
	Transactions mdl.Transaction
}

type GeqRelationships struct {
	Posts struct {
		Author mdl.User `geq:"Posts.AuthorID = Users.ID"`
	}
}

type GeqMappers struct {
	PostStats        mdl.PostStat
	TransactionStats mdl.TransactionStat
}
