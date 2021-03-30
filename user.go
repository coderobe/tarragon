package main

import (
	"math/rand"
	"time"
	"unsafe"
)

type User struct {
	Entity

	group *Group

	password string
	tokens   map[string]struct{}
}

func NewUser(name string) *User {
	var u User
	u.SetName(name)
	u.SetOwner(&u)

	u.group = NewGroup(name)
	u.tokens = make(map[string]struct{})

	return &u
}

func (u *User) Group() *Group {
	return u.group
}

func (u *User) Endpoints() []*Endpoint {
	return u.group.Endpoints()
}

func (u *User) SetPassword(pass string) *User {
	// TODO: turn to hash
	u.password = pass
	return u
}

func (u *User) CheckPassword(pass string) bool {
	return u.password == pass
}

func (u *User) NewToken() string {
	token := randString(32)
	u.tokens[token] = struct{}{}
	return token
}

func (u *User) CheckToken(token string) bool {
	_, ok := u.tokens[token]
	return ok
}

func (u *User) RemoveToken(token string) {
	delete(u.tokens, token)
}

// https://stackoverflow.com/a/31832326
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int) string {
	src := rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
