package main

import (
	"net"
)

type Endpoint struct {
	NodeImpl

	publicAddr     net.IP
	privateAddr    net.IP
	interfaceAddrs []net.IP
	routes         []*net.IPNet
}

func NewEndpoint(name string) *Endpoint {
	var e Endpoint

	e.name = name

	return &e
}

func (e *Endpoint) PrivateAddr() net.IP {
	return e.privateAddr
}

func (e *Endpoint) InterfaceAddrs() []net.IP {
	return e.interfaceAddrs
}

func (e *Endpoint) Routes() []*net.IPNet {
	return e.routes
}

func (e *Endpoint) PublicAddr() net.IP {
	return e.publicAddr
}
