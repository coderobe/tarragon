package main

type Emitter struct {
	send    chan Message
	receive chan Message
}

func NewEmitter(send chan Message, receive chan Message) *Emitter {
	var e Emitter

	e.send = send
	e.receive = receive

	return &e
}

func (e *Emitter) Close() {
	//close(e.send)
	close(e.receive)
}

func (e *Emitter) Send(msg Message) {
	e.send <- msg
}

func (e *Emitter) Receive() Message {
	return <-e.receive
}

func (e *Emitter) Execute(msg Message) Message {
	e.Send(msg)
	return e.Receive()
}
