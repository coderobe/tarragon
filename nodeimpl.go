package main

type NodeImpl struct {
	Id

	owner *User
}

func (n *NodeImpl) Owner() *User {
	return n.owner
}
