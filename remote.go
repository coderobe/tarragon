package main

type Remote struct {
	instance *Instance
}

func NewRemote(instance *Instance) *Remote {
	var r Remote
	r.instance = instance

	return &r
}

func (r *Remote) MakeAuth() []byte {

	return nil
}
