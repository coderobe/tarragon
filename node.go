package main

type Node interface {
	Name() string
	Owner() *User
	Identifier() string
}
