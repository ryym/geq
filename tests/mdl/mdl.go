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
