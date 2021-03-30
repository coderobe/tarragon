package main

const (
	MessageLogin = iota
	MessageAuth
	MessageLogoff
	MessageDeauth
	MessageIdentify
	MessageNewAuthToken
	MessageDeleteAuthToken
	MessageEventNewGroup
	MessageEventNewUser
	MessageEventNewEndpoint
	MessageEventRemoveEndpoint
	MessageEventRemoveGroup
	MessageEventRemoveUser
	MessageEventEndpointOnline
	MessageEventEndpointOffline
	MessageEventGroupGroupJoin
	MessageEventGroupGroupLeave
	MessageEventGroupEndpointJoin
	MessageEventGroupEndpointLeave
)

type Message struct {
	Type    int
	Reply   bool
	Success bool
	Data    map[string]string
}

func NewMessage(typ int) Message {
	var msg Message
	msg.Type = typ
	msg.Data = make(map[string]string)
	return msg
}
