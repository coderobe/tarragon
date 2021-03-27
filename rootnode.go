package main

type RootNode struct {
	NodeGroup
}

func NewRootNode() *RootNode {
	var r RootNode

	r.NodeGroup = *NewNodeGroup(".")

	userRoot := NewNodeGroup("user")
	r.AddNode(userRoot)

	groupRoot := NewNodeGroup("group")
	r.AddNode(groupRoot)

	return &r
}

func (r *RootNode) Users() *NodeGroup {
	userRoot, _ := r.GetNode("user")

	return userRoot.(*NodeGroup)
}

func (r *RootNode) Groups() *NodeGroup {
	groupRoot, _ := r.GetNode("group")

	return groupRoot.(*NodeGroup)
}

func (r *RootNode) AddUser(user *User) {
	r.Users().AddNode(user)
}

func (r *RootNode) AddGroup(group *NodeGroup) {
	r.Groups().AddNode(group)
}

func (r *RootNode) GetUser(name string) (*User, error) {
	node, err := r.Users().GetNode(name)
	return node.(*User), err
}

func (r *RootNode) GetGroup(name string) (*NodeGroup, error) {
	node, err := r.Groups().GetNode(name)
	return node.(*NodeGroup), err
}
