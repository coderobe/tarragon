package main

type User struct {
	NodeGroup

	password string
	tokens   [][]byte
}

func NewUser(name string) *User {
	var u User
	u.letter = 'u'
	u.name = name

	u.owner = &u

	return &u
}

func (u *User) SetPassword(pass string) *User {
	// TODO: turn to hash
	u.password = pass

	return u
}
