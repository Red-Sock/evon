package evon

import (
	"strings"
)

const (
	// ObjectSplitter must be used to separate nested objects
	// e.g. evon record "OBJECT_FIELD1_FIELD2=2"
	// is equal to json record
	// {
	//		"object": {
	//		   	"field1": {
	//			   	"field2": 2
	//			}
	//		}
	//	}
	ObjectSplitter = "_"
	// FieldSplitter must be used to for fields with more than one word in name
	// e.g. evon record "OBJECT_FIELD-ONE_FIELD-TWO=2"
	// is equal to json record
	// {
	//		"object": {
	//		   	"field-one": {
	//			   	"field-two": 2
	//			}
	//		}
	//	}
	FieldSplitter = "-"
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

func NodesToStorage(n *Node) NodeStorage {
	ns := NodeStorage{}
	if n == nil {
		n = &Node{}
	}

	ns[""] = n

	for _, node := range n.InnerNodes {
		ns.AddNode(node)
	}

	return ns
}

func (s NodeStorage) AddNode(node *Node) {
	rootName := node.Name

	nameParts := strings.Split(rootName, "_")

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

	var existingNode *Node
	for _, n := range lastNode.InnerNodes {
		if n.Name == node.Name {
			existingNode = n
			break
		}
	}
	if existingNode == nil {
		lastNode.InnerNodes = append(lastNode.InnerNodes, node)
	} else {
		existingNode.Value = node.Value
	}

	s[node.Name] = node

	for _, n := range node.InnerNodes {
		if !strings.HasPrefix(n.Name, rootName) {
			n.Name = rootName + "_" + n.Name
		}

		s.AddNode(n)
	}
}
