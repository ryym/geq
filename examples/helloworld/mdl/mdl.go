package mdl

type User struct {
	ID   uint64
	Name string
}

type Post struct {
	ID        uint64
	Title     string
	AuthorID  uint64
	Published bool
}

type Country struct {
	ID   uint32
	Name string
}

type City struct {
	ID        uint64
	Name      string
	CountryID uint32
}
