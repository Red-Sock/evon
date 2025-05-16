package evon

type NodeDiff struct {
	NewNodes     []*Node
	RemovedNodes []*Node
}

// Diff returns difference between new and old nodes
// e.g.
//
//	if [new] contains sub-node that is not presented in [old] - it goes to NewNodes
//	if [old] contains bun-node that is not presented in [new] - it goes to RemovedNodes
func Diff(old, new *Node) NodeDiff {
	nd := NodeDiff{}

	oldStorage := NodesToStorage(old)
	newStorage := NodesToStorage(new)
	for _, newNode := range newStorage {
		oldNode, exists := oldStorage[newNode.Name]
		if !exists {
			nd.NewNodes = append(nd.NewNodes, newNode)
		} else if oldNode.Value != newNode.Value {
			nd.NewNodes = append(nd.NewNodes, newNode)
		}
	}

	for _, oldNode := range oldStorage {
		_, exists := newStorage[oldNode.Name]
		if !exists {
			nd.RemovedNodes = append(nd.RemovedNodes, oldNode)
		}
	}

	return nd
}
