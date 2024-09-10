package evon

import (
	"strings"
)

type Node struct {
	Name       string
	Value      any
	InnerNodes []*Node
}

type NodeStorage map[string]*Node

func ParseToNodes(bytes []byte) NodeStorage {
	nodesMap := NodeStorage{}

	name := ""
	var value any

	start := 0
	for idx := range bytes {
		switch bytes[idx] {
		case '=':
			name = string(bytes[start:idx])
			start = idx + 1
		case '\n':
			value = string(bytes[start:idx])
			start = idx + 1
			nodesMap.AddNode(&Node{
				Name:  name,
				Value: value,
			})

		}
	}

	return nodesMap
}

func NodesToStorage(n []*Node) NodeStorage {
	ns := NodeStorage{}

	for _, node := range n {
		ns.AddNode(node)
	}

	return ns
}

func (s NodeStorage) AddNode(node *Node) {
	nameParts := strings.Split(node.Name, "_")

	nodePath := ""
	lastNode := s[nodePath]
	if lastNode == nil {
		lastNode = &Node{}
		s[nodePath] = lastNode
	}

	for _, namePart := range nameParts[:len(nameParts)-1] {
		if nodePath != "" {
			nodePath += "_"
		}
		nodePath = nodePath + namePart

		nextNode := s[nodePath]
		if nextNode == nil {
			nextNode = &Node{
				Name: nodePath,
			}
			if lastNode != nil {
				lastNode.InnerNodes = append(lastNode.InnerNodes, nextNode)
			}
			s[nodePath] = nextNode
		}

		lastNode = nextNode
	}

	nodeAlreadyExists := false
	for _, n := range lastNode.InnerNodes {
		if n.Name == node.Name {
			nodeAlreadyExists = true
			break
		}
	}
	if !nodeAlreadyExists {
		lastNode.InnerNodes = append(lastNode.InnerNodes, node)
	}

	s[node.Name] = node

	for _, n := range node.InnerNodes {
		s.AddNode(n)
	}
}
