package tests

type GeqRows struct {
	PostStats PostStat
}

type PostStat struct {
	AuthorID  int64
	PostCount int64
	LastTitle string
}
