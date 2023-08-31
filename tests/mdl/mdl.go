package mdl

type User struct {
	ID   int64
	Name string
}

type Post struct {
	ID       int64
	AuthorID int64
	Title    string
}
