package mdl

import "time"

type User struct {
	ID   int64
	Name string
}

type Post struct {
	ID       int64
	AuthorID int64
	Title    string
}

type Transaction struct {
	ID          uint32
	UserID      uint32
	Amount      int32
	Description string
	CreatedAt   time.Time
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
