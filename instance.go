package main

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

type Instance struct {
	brokerAddr string
	secure     bool
	socket     *websocket.Conn
	send       chan Message
	recv       chan Message

	self *Endpoint

	state *State
}

func (i *Instance) Self() *Endpoint {
	return i.self
}

func (i *Instance) SetSelf(e *Endpoint) *Instance {
	i.self = e
	return i
}

func (i *Instance) Auth(token string) error {
	var msg Message
	msg.Type = MessageAuth
	msg.Data = make(map[string]string)
	msg.Data["token"] = token

	msg = i.Execute(msg)

	if msg.Success {
		return nil
	}

	return errors.New(msg.Data["message"])
}

func (i *Instance) NewAuthToken() (string, error) {
	var msg Message
	msg.Type = MessageNewAuthToken

	msg = i.Execute(msg)

	if msg.Success {
		return msg.Data["token"], nil
	}

	return "", errors.New(msg.Data["message"])
}

func (i *Instance) DeleteAuthToken(token string) error {
	var msg Message
	msg.Type = MessageDeleteAuthToken
	msg.Data = make(map[string]string)
	msg.Data["token"] = token

	msg = i.Execute(msg)

	if msg.Success {
		return nil
	}

	return errors.New(msg.Data["message"])
}

func (i *Instance) Login(username string, password string) error {
	var msg Message
	msg.Type = MessageLogin
	msg.Data = make(map[string]string)
	msg.Data["username"] = username
	msg.Data["password"] = password

	msg = i.Execute(msg)

	if msg.Success {
		return nil
	}

	return errors.New(msg.Data["message"])
}

func (i *Instance) Logoff() error {
	var msg Message
	msg.Type = MessageLogoff

	msg = i.Execute(msg)

	if msg.Success {
		log.Println("Instance: Logged out")
		return nil
	}

	return errors.New(msg.Data["message"])
}

func (i *Instance) Deauth() error {
	var msg Message
	msg.Type = MessageDeauth

	msg = i.Execute(msg)

	if msg.Success {
		log.Println("Instance: Deauthenticated")
		return nil
	}

	return errors.New(msg.Data["message"])
}

func (i *Instance) Identify(hostname string) error {
	var msg Message
	msg.Type = MessageIdentify
	msg.Data = make(map[string]string)
	msg.Data["hostname"] = hostname

	msg = i.Execute(msg)

	if msg.Success {
		if e, err := i.State().GetEndpoint(msg.Data["hostname"]); err == nil {
			i.SetSelf(e)
		}
		return nil
	}

	return errors.New(msg.Data["message"])
}

func NewInstance(addr string, secure bool) *Instance {
	var i Instance
	i.brokerAddr = addr
	i.secure = secure

	i.send = make(chan Message)
	i.recv = make(chan Message)

	i.state = NewState()

	return &i
}

func (i *Instance) State() *State {
	return i.state
}

func (i *Instance) ConnectAndRecv() (err error) {
	log.Printf("Instance: Connecting to %v\n", i.brokerAddr)

	proto := "wss"
	if !i.secure {
		proto = "ws"
		log.Println("Instance: [Warning] Instance.secure = false - connecting to plaintext websocket")
	}

	i.socket, err = websocket.Dial(fmt.Sprintf("%s://%s/broker", proto, i.brokerAddr), "", i.brokerAddr)

	if err != nil {
		log.Fatal(err)
		return err
	}

	go func() {
		for {
			if err := websocket.JSON.Send(i.socket, <-i.send); err != nil {
				log.Fatal(err)
			}
		}
	}()

	for {
		var msg Message
		err := websocket.JSON.Receive(i.socket, &msg)
		if err != nil {
			log.Println(err)
			break
		}
		if msg.Reply {
			i.recv <- msg
		} else {
			switch msg.Type {
			case MessageEventNewGroup:
				if _, err := i.State().GetGroup(msg.Data["name"]); err == nil {
					continue // group exists
				}
				user, err := i.State().GetUser(msg.Data["owner"])
				if err != nil {
					user, err = i.State().NewUser(msg.Data["owner"])
					if err != nil {
						log.Fatal("Instance: unresolvable state inconsistency", err)
					}
				}
				if _, err := i.State().NewGroup(msg.Data["name"], user); err != nil {
					log.Fatal("Instance: unresolvable state inconsistency", err)
				}
			case MessageEventNewUser:
				if _, err := i.State().GetUser(msg.Data["name"]); err != nil {
					if _, err := i.State().NewUser(msg.Data["name"]); err != nil {
						log.Fatal("Instance: unresolvable state inconsistency", err)
					}
				}
			case MessageEventNewEndpoint:
				if _, err := i.State().GetEndpoint(msg.Data["name"]); err == nil {
					continue // endpoint exists
				}
				user, err := i.State().GetUser(msg.Data["owner"])
				if err != nil {
					user, err = i.State().NewUser(msg.Data["owner"])
					if err != nil {
						log.Fatal("Instance: unresolvable state inconsistency", err)
					}
				}
				if _, err := i.State().NewEndpoint(msg.Data["name"], user); err != nil {
					log.Fatal("Instance: unresolvable state inconsistency", err)
				}
			case MessageEventRemoveEndpoint:
				if endpoint, err := i.State().GetEndpoint(msg.Data["name"]); err == nil {
					i.State().RemoveEndpoint(endpoint)
				}
			case MessageEventRemoveGroup:
				if group, err := i.State().GetGroup(msg.Data["name"]); err == nil {
					i.State().RemoveGroup(group)
				}
			case MessageEventRemoveUser:
				if user, err := i.State().GetUser(msg.Data["name"]); err == nil {
					i.State().RemoveUser(user)
				}
			case MessageEventEndpointOnline:
				if endpoint, err := i.State().GetEndpoint(msg.Data["name"]); err == nil {
					endpoint.SetStaticOnline(true)
				}
			case MessageEventEndpointOffline:
				if endpoint, err := i.State().GetEndpoint(msg.Data["name"]); err == nil {
					endpoint.SetStaticOnline(false)
				}
			case MessageEventGroupGroupJoin:
				if group, err := i.State().GetGroup(msg.Data["group"]); err == nil {
					if target, err := i.State().GetGroup(msg.Data["target"]); err == nil {
						group.AddGroup(target)
					}
				}
			case MessageEventGroupGroupLeave:
				if group, err := i.State().GetGroup(msg.Data["group"]); err == nil {
					if target, err := i.State().GetGroup(msg.Data["target"]); err == nil {
						group.RemoveGroup(target)
					}
				}
			case MessageEventGroupEndpointJoin:
				if group, err := i.State().GetGroup(msg.Data["group"]); err == nil {
					if target, err := i.State().GetEndpoint(msg.Data["target"]); err == nil {
						group.AddEndpoint(target)
					}
				}
			case MessageEventGroupEndpointLeave:
				if group, err := i.State().GetGroup(msg.Data["group"]); err == nil {
					if target, err := i.State().GetEndpoint(msg.Data["target"]); err == nil {
						group.RemoveEndpoint(target)
					}
				}
			default:
				log.Fatal("Instance: unhandled event message %v\n", msg)
			}
		}
	}

	return
}

func (i *Instance) Send(command Message) {
	i.send <- command
}

func (i *Instance) Receive() Message {
	return <-i.recv
}

func (i *Instance) Execute(command Message) Message {
	i.Send(command)
	return i.Receive()
}
