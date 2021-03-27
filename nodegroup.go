package main

import (
	"errors"
	"log"
)

type NodeGroup struct {
	NodeImpl

	list map[Node]struct{}
}

func NewNodeGroup(name string) *NodeGroup {
	var ng NodeGroup
	ng.letter = 'g'
	ng.name = name

	ng.list = make(map[Node]struct{})

	return &ng
}

func (ng *NodeGroup) GetNode(name string) (Node, error) {
	for value, _ := range ng.list {
		if _, ok := ng.list[value]; ok {
			if value.Name() == name {
				return value, nil
			}
		}
	}
	return nil, errors.New("Node not found")
}

func (ng *NodeGroup) GetNodeDeep(name string) (Node, error) {
	nodes := ng.SelectDeep(func(n Node) (bool, bool) {
		if n.Name() == name {
			return true, true
		}
		return false, false
	})

	if len(nodes) == 1 {
		return nodes[0], nil
	} else if len(nodes) > 1 {
		log.Fatal("NodeGroup.GetNodeDeep: conflicting node names")
	}

	return nil, errors.New("Node not found")
}

func (ng *NodeGroup) AddNode(node Node) {
	ng.list[node] = struct{}{}
}

func (ng *NodeGroup) RemoveNode(node Node) {
	delete(ng.list, node)
}

func (ng *NodeGroup) RemoveNodeDeep(node Node) {
	for _, parent := range ng.ParentsOf(node) {
		parent.RemoveNode(node)
	}
}

func (ng *NodeGroup) HasNode(node Node) bool {
	_, ok := ng.list[node]
	return ok
}

func (ng *NodeGroup) DeepHasNode(node Node) bool {
	return len(ng.SelectDeep(func(n Node) (bool, bool) {
		if n == node {
			return true, true
		}
		return false, false
	})) > 0
}

func (ng *NodeGroup) ParentsOf(node Node) []*NodeGroup {
	nodes := ng.SelectDeep(func(n Node) (bool, bool) {
		if group, ok := n.(*NodeGroup); ok {
			if group.HasNode(node) {
				return true, false
			}
		}
		return false, false
	})

	var groups []*NodeGroup
	if ng.HasNode(node) {
		groups = append(groups, ng)
	}
	for _, node := range nodes {
		groups = append(groups, node.(*NodeGroup))
	}
	return groups
}

func (ng *NodeGroup) AncestorsOf(node Node) []*NodeGroup {
	var ancestors []*NodeGroup
	parents := ng.ParentsOf(node)
	ancestors = append(ancestors, parents...)
	for _, parent := range parents {
		ancestors = append(ancestors, ng.AncestorsOf(parent)...)
	}

	return ancestors
}

func (ng *NodeGroup) Children() []Node {
	var ret []Node
	for value, _ := range ng.list {
		if _, ok := ng.list[value]; ok {
			ret = append(ret, value)
		}
	}
	return ret
}

func (ng *NodeGroup) Endpoints() []*Endpoint {
	var endpoints []*Endpoint
	for node, _ := range ng.list {
		n, ok := node.(*Endpoint)
		if ok {
			endpoints = append(endpoints, n)
		}
	}
	return endpoints
}

func (ng *NodeGroup) NodeGroups() []*NodeGroup {
	var nodegroups []*NodeGroup
	for node, _ := range ng.list {
		n, ok := node.(*NodeGroup)
		if ok {
			nodegroups = append(nodegroups, n)
		}
	}
	return nodegroups
}

// filter func checks whether Node should be selected
// returns (bool,bool) - whether to select Node, and whether to end selecting early
func (ng *NodeGroup) SelectDeep(filter func(Node) (bool, bool)) []Node {
	var set map[Node]struct{}

	for _, node := range ng.Children() {
		if _, exists := set[node]; !exists {
			if want, done := filter(node); want {
				set[node] = struct{}{}
				if done {
					break
				}
			}
			if group, ok := node.(*NodeGroup); ok {
				for _, child := range group.Children() {
					if _, exists := set[child]; !exists {
						if want, done := filter(node); want {
							set[child] = struct{}{}
							if done {
								break
							}
						}
					}
				}
			}
		}
	}

	var ret []Node
	for value, _ := range set {
		if _, ok := set[value]; ok {
			ret = append(ret, value)
		}
	}
	return ret
}

func (ng *NodeGroup) DeepChildren() []Node {
	return ng.SelectDeep(func(n Node) (bool, bool) {
		return true, false
	})
}

func (ng *NodeGroup) DeepEndpoints() []*Endpoint {
	nodes := ng.SelectDeep(func(n Node) (bool, bool) {
		if _, ok := n.(*Endpoint); ok {
			return true, false
		}
		return false, false
	})

	var endpoints []*Endpoint
	for _, node := range nodes {
		endpoints = append(endpoints, node.(*Endpoint))
	}
	return endpoints
}

func (ng *NodeGroup) DeepNodeGroups() []*NodeGroup {
	nodes := ng.SelectDeep(func(n Node) (bool, bool) {
		if _, ok := n.(*NodeGroup); ok {
			return true, false
		}
		return false, false
	})

	var groups []*NodeGroup
	for _, node := range nodes {
		groups = append(groups, node.(*NodeGroup))
	}
	return groups
}
