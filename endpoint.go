package main

import (
	"log"
)

type Endpoint struct {
	Entity

	emitter      *Emitter
	staticOnline bool
}

func NewEndpoint(name string) *Endpoint {
	var e Endpoint

	e.name = name

	return &e
}

func (e *Endpoint) Online() bool {
	return e.Connected() || e.staticOnline
}

func (e *Endpoint) SetStaticOnline(status bool) *Endpoint {
	e.staticOnline = status
	return e
}

func (e *Endpoint) Disconnect() *Endpoint {
	if e.Connected() {
		e.Emitter().Close()
	}
	e.emitter = nil

	return e
}

func (e *Endpoint) Connected() bool {
	return e.Emitter() != nil
}

func (e *Endpoint) Connect(emitter *Emitter) *Endpoint {
	if e.emitter != nil {
		log.Printf("Endpoint %v is already connected. FIXME\n", e.Name())
	} else {
		e.emitter = emitter
	}
	return e
}

func (e *Endpoint) Emitter() *Emitter {
	return e.emitter
}

func (e *Endpoint) SetOwner(owner *User) *Endpoint {
	if e.Owner() != nil {
		e.Owner().Group().RemoveEndpoint(e)
	}

	e.Entity.SetOwner(owner)
	e.owner.Group().AddEndpoint(e)

	return e
}
