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
			nodesMap.addNode(&Node{
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
		ns.addNode(node)
	}

	return ns
}

func (s NodeStorage) addNode(node *Node) {
	nameParts := strings.Split(node.Name, "_")
	parentNodePath := nameParts[0]

	var lastNode = &Node{
		Name: parentNodePath,
	}

	for _, namePart := range nameParts[1 : len(nameParts)-1] {
		parentNode := s[parentNodePath]
		if parentNode == nil {
			parentNode = &Node{
				Name: parentNodePath,
			}
			s[parentNodePath] = parentNode
		}

		if parentNodePath != "" {
			parentNodePath += "_"
		}
		currentNodePath := parentNodePath + namePart

		var ok bool
		lastNode, ok = s[currentNodePath]
		if !ok {
			lastNode = &Node{
				Name: currentNodePath,
			}
			s[lastNode.Name] = lastNode

			parentNode.InnerNodes = append(parentNode.InnerNodes, lastNode)
		}

		parentNodePath = currentNodePath
	}
	lastNode.InnerNodes = append(lastNode.InnerNodes, node)
	s[node.Name] = node

	for _, n := range node.InnerNodes {
		s.addNode(n)
	}
}
