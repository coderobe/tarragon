package main

import (
	//"errors"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

type Instance struct {
	brokerAddr string
	secure     bool
	socket     *websocket.Conn
	send       chan string
	recv       chan string

	token string
	user  *User

	rootNode *RootNode
}

/**
func (i *Instance) Emitter() (*Emitter, error) {
	if i.emitter == nil {
		return nil, errors.New("Instance not authenticated")
	}

	return i.emitter, nil
}

func (i *Instance) Remote() (*Remote, error) {
	if i.remote == nil {
		return nil, errors.New("Instance not logged in")
	}

	return i.remote, nil
}
**/

/**
func (i *Instance) Auth(token []byte) (*Emitter, error) {
	// TODO
	log.Fatal("Instance.Auth: unimplemented")

	// i.emitter = ...

	return i.emitter, nil
}

func (i *Instance) Login(username string, password string) (*Remote, error) {
	// TODO
	log.Fatal("Instance.Login: unimplemented")

	//i.Send()

	// i.remote = ...

	return i.remote, nil
}
**/

func NewInstance(addr string, secure bool) *Instance {
	var i Instance
	i.brokerAddr = addr
	i.secure = secure

	i.rootNode = NewRootNode()

	return &i
}

func (i *Instance) Root() *RootNode {
	return i.rootNode
}

func (i *Instance) ConnectAndRecv() (err error) {
	proto := "wss"
	if !i.secure {
		proto = "ws"
		log.Println("Warning: Instance.secure = false - connecting to plaintext websocket")
	}

	i.socket, err = websocket.Dial(fmt.Sprintf("%s://%s/broker", proto, i.brokerAddr), "", i.brokerAddr)

	if err != nil {
		return err
	}

	go func() {
		for {
			i.SendRaw(<-i.send)
		}
	}()

	var recvStr string
	for {
		websocket.Message.Receive(i.socket, &recvStr)
		i.recv <- recvStr
		log.Printf("Instance.HandleRecv: received '%s'\n", recvStr)
		// TODO: socket event handler logic
	}

	return
}

func (i *Instance) Disconnect() {
	i.socket.Close()
}

func (i *Instance) SendRaw(command string) error {
	return websocket.Message.Send(i.socket, command)
}

func (i *Instance) Send(command string) {
	i.send <- command
}

func (i *Instance) Receive() string {
	return <-i.recv
}

func (i *Instance) Execute(command string) string {
	i.Send(command)
	return i.Receive()
}
