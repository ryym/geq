package tests

import "time"

type GeqRows struct {
	PostStats        PostStat
	TransactionStats TransactionStat
}

type PostStat struct {
	AuthorID  int64
	PostCount int64
	LastTitle string
}

type TransactionStat struct {
	UserID        int64
	TotalAmount   int64
	LastCreatedAt time.Time
}
