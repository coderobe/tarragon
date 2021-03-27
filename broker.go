package main

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type Broker struct {
	listenAddr string

	rootNode *RootNode
}

func NewBroker(addr string) *Broker {
	var b Broker
	b.listenAddr = addr

	b.rootNode = NewRootNode()

	return &b
}

func (b *Broker) Root() *RootNode {
	return b.rootNode
}

func (b *Broker) ListenAndServe() (err error) {
	http.Handle("/broker", websocket.Handler(func(ws *websocket.Conn) {
		log.Println("Broker.ListenAndServe: New connection")
		// TODO: logic
		io.Copy(ws, ws)
	}))

	log.Fatal(http.ListenAndServe(b.listenAddr, nil))
	return
}
