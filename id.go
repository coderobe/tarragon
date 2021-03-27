package main

import (
	"fmt"
	"log"
)

type Id struct {
	letter rune
	name   string
}

func (id *Id) Name() string {
	return id.name
}

func (id *Id) Identifier() string {
	if id.letter == 0 {
		log.Fatal("Id.Identifier: letter rune unset")
	}
	if id.name == "" {
		log.Fatal("Id.Identifier: name unset")
	}

	return fmt.Sprintf("%c:%s", id.letter, id.name)
}
