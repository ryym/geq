package reports

//geq:package .

type GeqMappers struct {
	PostStats     PostStat
	SameNameUsers SameNameUser
}

type PostStat struct {
	AuthorID  uint64
	PostCount int
	LastTitle string
}

type SameNameUser struct {
	Name  string
	Count int
}
