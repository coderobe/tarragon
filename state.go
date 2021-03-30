package main

import (
	"errors"
	"log"
)

type State struct {
	users map[*User]struct{}
	root  *Group
}

func NewState() *State {
	var s State

	s.users = make(map[*User]struct{})
	s.root = NewGroup(".")

	return &s
}

func (s *State) Broadcast(msg Message) {
	for _, endpoint := range s.AllEndpoints() {
		if endpoint.Connected() {
			endpoint.Emitter().Send(msg)
		}
	}
}

func (s *State) PushState(e *Emitter) {
	for _, user := range s.Users() {
		e.Send(s.NotifyNewUser(user.Name()))
	}
	for _, endpoint := range s.AllEndpoints() {
		e.Send(s.NotifyNewEndpoint(endpoint.Name(), endpoint.Owner().Name()))
	}
	for _, group := range s.PureGroups() {
		e.Send(s.NotifyNewGroup(group.Name(), group.Owner().Name()))
		for _, inner := range group.Groups() {
			e.Send(s.NotifyGroupGroupJoin(group.Name(), inner.Name()))
		}
		for _, inner := range group.Endpoints() {
			e.Send(s.NotifyGroupEndpointJoin(group.Name(), inner.Name()))
		}
	}
}

func (s *State) PrettyPrint() {
	log.Println("==== STATE")
	log.Println("Users:")
	for _, user := range s.Users() {
		log.Printf("\t%v\n", user.Name())
		for _, endpoint := range user.Endpoints() {
			log.Printf("\t... endpoint: %v (online? %v)\n", endpoint.Name(), endpoint.Online())
		}
	}
	log.Println("Groups:")
	for _, group := range s.PureGroups() {
		log.Printf("\t%v\n", group.Name())
		for _, inner := range group.Groups() {
			if _, err := s.GetUser(inner.Name()); err == nil {
				log.Printf("\t... member user: %v\n", inner.Name())
			} else {
				log.Printf("\t... member group: %v\n", inner.Name())
			}
		}
		for _, endpoint := range group.Endpoints() {
			log.Printf("\t... member endpoint: %v\n", endpoint.Name())
		}
	}
	log.Println("====")
}

func (s *State) Root() *Group {
	return s.root
}

func (s *State) Users() []*User {
	var ret []*User
	for user, _ := range s.users {
		ret = append(ret, user)
	}
	return ret
}

func (s *State) Groups() []*Group {
	return s.Root().Groups()
}

func (s *State) PureGroups() []*Group {
	users := s.Users()
	var ret []*Group
	for _, group := range s.Groups() {
		usergroup := false
		for _, user := range users {
			if user.Name() == group.Name() {
				usergroup = true
			}
		}
		if !usergroup {
			ret = append(ret, group)
		}
	}
	return ret
}

func (s *State) GetUser(name string) (*User, error) {
	for user, _ := range s.users {
		if user.Name() == name {
			return user, nil
		}
	}
	return nil, errors.New("User not found")
}

func (s *State) GetUserGroups(user *User) []*Group {
	var ret []*Group
	for _, group := range s.Root().Groups() {
		if _, err := group.GetGroup(user.Name()); err == nil && group != user.Group() {
			ret = append(ret, group)
		}
	}
	return ret
}

func (s *State) GetGroup(name string) (*Group, error) {
	return s.Root().GetGroup(name)
}

func (s *State) GetEndpoint(name string) (*Endpoint, error) {
	for _, endpoint := range s.AllEndpoints() {
		if endpoint.Name() == name {
			return endpoint, nil
		}
	}
	return nil, errors.New("Endpoint not found")
}

func (s *State) NameUsed(name string) bool {
	for _, endpoint := range s.AllEndpoints() {
		if endpoint.Name() == name {
			return true
		}
	}
	for _, group := range s.Root().Groups() {
		if group.Name() == name {
			return true
		}
	}
	return false
}

func (s *State) AllEndpoints() []*Endpoint {
	all := make(map[*Endpoint]struct{})

	for _, endpoint := range s.Root().Endpoints() {
		all[endpoint] = struct{}{}
	}
	for _, group := range s.Root().Groups() {
		for _, endpoint := range group.Endpoints() {
			all[endpoint] = struct{}{}
		}
	}

	var ret []*Endpoint
	for endpoint, _ := range all {
		ret = append(ret, endpoint)
	}
	return ret
}

func (s *State) NewUser(name string) (*User, error) {
	if s.NameUsed(name) {
		return nil, errors.New("Name in use")
	}

	u := NewUser(name)
	s.users[u] = struct{}{}
	s.Root().AddGroup(u.Group())

	s.Broadcast(s.NotifyNewUser(name))

	return u, nil
}

func (s *State) NotifyNewUser(name string) Message {
	msg := NewMessage(MessageEventNewUser)
	msg.Data["name"] = name
	return msg
}

func (s *State) NewGroup(name string, owner *User) (*Group, error) {
	if s.NameUsed(name) {
		return nil, errors.New("Name in use")
	}

	g := NewGroup(name)
	g.SetOwner(owner)
	s.Root().AddGroup(g)

	s.Broadcast(s.NotifyNewGroup(name, owner.Name()))

	return g, nil
}

func (s *State) NotifyNewGroup(name string, owner string) Message {
	msg := NewMessage(MessageEventNewGroup)
	msg.Data["name"] = name
	msg.Data["owner"] = owner
	return msg
}

func (s *State) NewEndpoint(name string, owner *User) (*Endpoint, error) {
	if s.NameUsed(name) {
		return nil, errors.New("Name in use")
	}

	e := NewEndpoint(name)
	e.SetOwner(owner)

	s.Broadcast(s.NotifyNewEndpoint(name, owner.Name()))

	return e, nil
}

func (s *State) NotifyNewEndpoint(name string, owner string) Message {
	msg := NewMessage(MessageEventNewEndpoint)
	msg.Data["name"] = name
	msg.Data["owner"] = owner
	return msg
}

func (s *State) RemoveEndpoint(target *Endpoint) {
	for _, group := range s.Root().Groups() {
		group.RemoveEndpoint(target)
	}

	s.Broadcast(s.NotifyRemoveEndpoint(target.Name()))
}

func (s *State) NotifyRemoveEndpoint(name string) Message {
	msg := NewMessage(MessageEventRemoveEndpoint)
	msg.Data["name"] = name
	return msg
}

func (s *State) RemoveGroup(target *Group) {
	for _, group := range s.Root().Groups() {
		group.RemoveGroup(target)
	}
	s.Root().RemoveGroup(target)

	s.Broadcast(s.NotifyRemoveGroup(target.Name()))
}

func (s *State) NotifyRemoveGroup(name string) Message {
	msg := NewMessage(MessageEventRemoveGroup)
	msg.Data["name"] = name
	return msg
}

func (s *State) RemoveUser(target *User) {
	for _, endpoint := range target.Endpoints() {
		s.RemoveEndpoint(endpoint)
	}
	for _, group := range s.Root().Groups() {
		group.RemoveGroup(target.Group())
	}
	delete(s.users, target)

	s.Broadcast(s.NotifyRemoveUser(target.Name()))
}

func (s *State) NotifyRemoveUser(name string) Message {
	msg := NewMessage(MessageEventRemoveUser)
	msg.Data["name"] = name
	return msg
}

func (s *State) NotifyGroupGroupJoin(group string, target string) Message {
	msg := NewMessage(MessageEventGroupGroupJoin)
	msg.Data["group"] = group
	msg.Data["target"] = target
	return msg
}

func (s *State) NotifyGroupGroupLeave(group string, target string) Message {
	msg := NewMessage(MessageEventGroupGroupLeave)
	msg.Data["group"] = group
	msg.Data["target"] = target
	return msg
}

func (s *State) NotifyGroupEndpointJoin(group string, endpoint string) Message {
	msg := NewMessage(MessageEventGroupEndpointJoin)
	msg.Data["group"] = group
	msg.Data["endpoint"] = endpoint
	return msg
}

func (s *State) NotifyGroupEndpointLeave(group string, endpoint string) Message {
	msg := NewMessage(MessageEventGroupEndpointLeave)
	msg.Data["group"] = group
	msg.Data["endpoint"] = endpoint
	return msg
}
