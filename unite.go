package evon

// Unite adds missing variables from src to trg
// If value exist in trg changes it onto the one from src
func Unite(src, trg *Node) {
	UniteStorages(NodesToStorage(src), NodesToStorage(trg))
}

// UniteStorages adds missing variables from src to trg
// If value exist in trg changes it onto the one from src
func UniteStorages(src, trg NodeStorage) {
	for _, sNode := range src {
		if sNode.Name != "" {
			trg.AddNode(sNode)
		}
	}
}
