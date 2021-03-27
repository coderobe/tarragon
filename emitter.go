package main

type Emitter struct {
	instance *Instance
}

func NewEmitter(instance *Instance) *Emitter {
	var e Emitter
	e.instance = instance

	return &e
}
