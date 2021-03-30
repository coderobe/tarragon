package main

import (
	"errors"
)

type Group struct {
	Entity

	groups    map[*Group]struct{}
	endpoints map[*Endpoint]struct{}
}

func NewGroup(name string) *Group {
	var g Group
	g.name = name

	g.groups = make(map[*Group]struct{})
	g.endpoints = make(map[*Endpoint]struct{})

	return &g
}

func (g *Group) SetOwner(owner *User) *Group {
	if g.Owner() != nil {
		g.Owner().Group().RemoveGroup(g)
	}

	g.Entity.SetOwner(owner)
	g.Owner().Group().AddGroup(g)

	return g
}

func (g *Group) AddGroup(group *Group) {
	g.groups[group] = struct{}{}
}

func (g *Group) AddEndpoint(endpoint *Endpoint) {
	g.endpoints[endpoint] = struct{}{}
}

func (g *Group) RemoveGroup(group *Group) {
	delete(g.groups, group)
}

func (g *Group) RemoveEndpoint(endpoint *Endpoint) {
	delete(g.endpoints, endpoint)
}

func (g *Group) GetGroup(name string) (*Group, error) {
	for value, _ := range g.groups {
		if _, ok := g.groups[value]; ok {
			if value.Name() == name {
				return value, nil
			}
		}
	}
	return nil, errors.New("Group not found")
}

func (g *Group) GetEndpoint(name string) (*Endpoint, error) {
	for value, _ := range g.endpoints {
		if _, ok := g.endpoints[value]; ok {
			if value.Name() == name {
				return value, nil
			}
		}
	}
	return nil, errors.New("Endpoint not found")
}

func (g *Group) Groups() []*Group {
	var ret []*Group
	for value, _ := range g.groups {
		if _, ok := g.groups[value]; ok {
			ret = append(ret, value)
		} else {
			// clean up bad groups while we're at it
			delete(g.groups, value)
		}
	}
	return ret
}

func (g *Group) Endpoints() []*Endpoint {
	var ret []*Endpoint
	for value, _ := range g.endpoints {
		if _, ok := g.endpoints[value]; ok {
			ret = append(ret, value)
		} else {
			// clean up bad endpoints while we're at it
			delete(g.endpoints, value)
		}
	}
	return ret
}
