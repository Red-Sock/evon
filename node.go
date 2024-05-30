package env

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
			nodesMap.addNode(Node{
				Name:  name,
				Value: value,
			})
			//
			//nameParts := strings.Split(name, "_")
			//// todo подумать над пустыми именами
			//parentNodePath := nameParts[0]
			//var leafNode *Node
			//for _, namePart := range nameParts[1:] {
			//	parentNode := nodesMap[parentNodePath]
			//	if parentNode == nil {
			//		parentNode = &Node{
			//			Name: parentNodePath,
			//		}
			//		nodesMap[parentNodePath] = parentNode
			//	}
			//
			//	if parentNodePath != "" {
			//		parentNodePath += "_"
			//	}
			//	currentNodePath := parentNodePath + namePart
			//
			//	leafNode = &Node{
			//		Name: currentNodePath,
			//	}
			//	if _, ok := nodesMap[leafNode.Name]; !ok {
			//		parentNode.InnerNodes = append(parentNode.InnerNodes, leafNode)
			//		nodesMap[leafNode.Name] = leafNode
			//	}
			//	parentNodePath = currentNodePath
			//}
			//leafNode.Value = value
		}
	}

	return nodesMap
}

func (s NodeStorage) addNode(node Node) {
	nameParts := strings.Split(node.Name, "_")
	// todo подумать над пустыми именами
	parentNodePath := nameParts[0]

	var leaftNode *Node
	for _, namePart := range nameParts[1:] {
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

		leaftNode = &Node{
			Name: currentNodePath,
		}
		if _, ok := s[leaftNode.Name]; !ok {
			parentNode.InnerNodes = append(parentNode.InnerNodes, leaftNode)
			s[leaftNode.Name] = leaftNode
		}
		parentNodePath = currentNodePath
	}
	leaftNode.Value = node.Value
	for _, n := range node.InnerNodes {
		s.addNode(*n)
	}
	leaftNode.InnerNodes = node.InnerNodes
}
