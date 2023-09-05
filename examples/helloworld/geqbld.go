package main

import "github.om/ryym/geq/examples/helloworld/mdl"

//geq:package ./gen/b

type GeqTables struct {
	Users     mdl.User
	Posts     mdl.Post
	Countries mdl.Country
	Cities    mdl.City
}

type GeqRelationships struct {
	Posts struct {
		Author mdl.User `geq:"Posts.AuthorID = Users.ID"`
	}
}
