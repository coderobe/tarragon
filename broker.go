package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

//go:embed status.html
var statusTemplate string

type Broker struct {
	listenAddr string

	user  *User
	state *State
}

func NewBroker(addr string) *Broker {
	var b Broker
	b.listenAddr = addr

	b.user = NewUser(".")
	b.state = NewState()

	return &b
}

func (b *Broker) User() *User {
	return b.user
}

func (b *Broker) State() *State {
	return b.state
}

func (b *Broker) HandleStatus() string {
	t, _ := template.New("status").Parse(statusTemplate)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			t.Execute(w, b.State())
		} else {
			http.Error(w, "resource unavailable", 500)
		}
	})

	url := fmt.Sprintf("http://%v/", b.listenAddr)

	log.Printf("Broker: Enabled statuspage on %v\n", url)

	return url
}

func (b *Broker) ListenAndServe() (err error) {
	http.Handle("/broker", websocket.Handler(func(ws *websocket.Conn) {
		var user *User
		fullLogin := false
		var endpoint *Endpoint
		cRecv := make(chan Message)
		cSend := make(chan Message)
		emitter := NewEmitter(cSend, cRecv)
		go func() {
			for {
				err := websocket.JSON.Send(ws, <-cSend)
				if err != nil {
					break
				}
			}
		}()

		for {
			var msg Message
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				log.Printf("Broker: Lost connection (%v)\n", err)
				if endpoint != nil {
					log.Printf("Broker: Endpoint %v (%v) disconnected\n", endpoint.Name(), endpoint.Owner().Name())
					if endpoint.Connected() {
						endpoint.Disconnect()
					}
					brc := NewMessage(MessageEventEndpointOffline)
					brc.Data["name"] = endpoint.Name()
					b.State().Broadcast(brc)
				}
				break
			}

			if msg.Data == nil {
				msg.Data = make(map[string]string)
			}
			if msg.Reply {
				cRecv <- msg
				continue
			}
			msg.Reply = true

			if msg.Type != MessageLogin && msg.Type != MessageAuth && user == nil {
				msg.Data["message"] = "Method not allowed"
				websocket.JSON.Send(ws, msg)
				continue
			}

			switch msg.Type {
			case MessageLogin:
				log.Printf("Broker: Login attempt for %v\n", msg.Data["username"])
				if u, err := b.State().GetUser(msg.Data["username"]); err == nil {
					if u.CheckPassword(msg.Data["password"]) {
						user = u
						fullLogin = true
						log.Printf("Broker: User %v logged in\n", user.Name())
						msg.Success = true
					} else {
						msg.Data["message"] = "Invalid password"
					}
				} else {
					msg.Data["message"] = "User does not exist"
				}
				websocket.JSON.Send(ws, msg)
				b.State().PushState(emitter)
			case MessageLogoff:
				fullLogin = false
				msg.Success = true
				websocket.JSON.Send(ws, msg)
			case MessageAuth:
				for _, u := range b.State().Users() {
					if u.CheckToken(msg.Data["token"]) {
						user = u
						msg.Success = true
						break
					}
				}
				if !msg.Success {
					msg.Data["message"] = "Invalid token"
				}
				websocket.JSON.Send(ws, msg)
				b.State().PushState(emitter)
			case MessageDeauth:
				fullLogin = false
				user = nil
				if endpoint != nil {
					if endpoint.Connected() {
						endpoint.Disconnect()
					}
					brc := NewMessage(MessageEventEndpointOffline)
					brc.Data["name"] = endpoint.Name()
					b.State().Broadcast(brc)
					endpoint = nil
				}
				msg.Success = true
				websocket.JSON.Send(ws, msg)
			case MessageIdentify:
				if e, err := b.State().GetEndpoint(msg.Data["hostname"]); err == nil {
					if e.Owner() == user {
						e.Disconnect()
						if endpoint != nil && endpoint.Connected() {
							endpoint.Disconnect()
						}
						endpoint = e
						msg.Success = true
					} else {
						msg.Data["message"] = "User does not own this hostname"
					}
				} else {
					if e, err = b.State().NewEndpoint(msg.Data["hostname"], user); err == nil {
						if endpoint != nil && endpoint.Connected() {
							endpoint.Disconnect()
						}
						endpoint = e
						msg.Success = true
					} else {
						msg.Data["message"] = fmt.Sprintf("%v", err)
					}
				}
				if msg.Success {
					log.Printf("Broker: Endpoint %v just identified\n", endpoint.Name())
					endpoint.Connect(emitter)
					brc := NewMessage(MessageEventEndpointOnline)
					brc.Data["name"] = endpoint.Name()
					b.State().Broadcast(brc)
				}
				websocket.JSON.Send(ws, msg)
			case MessageNewAuthToken:
				if !fullLogin {
					msg.Data["message"] = "Method requires full login"
					websocket.JSON.Send(ws, msg)
					break
				}
				msg.Data["token"] = user.NewToken()
				msg.Success = true
				websocket.JSON.Send(ws, msg)
			case MessageDeleteAuthToken:
				if !fullLogin {
					msg.Data["message"] = "Method requires full login"
					websocket.JSON.Send(ws, msg)
					break
				}
				user.RemoveToken(msg.Data["token"])
				msg.Success = true
				websocket.JSON.Send(ws, msg)
			default:
				log.Printf("Broker: Unhandled event: %v\n", msg)
			}
		}
	}))

	log.Printf("Broker: Listening on %v\n", b.listenAddr)
	log.Fatal(http.ListenAndServe(b.listenAddr, nil))
	return
}
