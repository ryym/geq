package main

import "github.com/ryym/geq/examples/helloworld/mdl"

//geq:package ./gen/b

type GeqTables struct {
	Users     mdl.User
	Posts     mdl.Post
	Countries mdl.Country
	Cities    mdl.City
}

type GeqRelationships struct {
	Users struct {
		Posts mdl.Post `geq:"Users.ID = Posts.AuthorID"`
	}
	Posts struct {
		Author mdl.User `geq:"Posts.AuthorID = Users.ID"`
	}
}
