package main

import ()

type Entity struct {
	name  string
	owner *User
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) SetName(name string) *Entity {
	e.name = name
	return e
}

func (e *Entity) Owner() *User {
	return e.owner
}

func (e *Entity) SetOwner(owner *User) *Entity {
	e.owner = owner

	return e
}
