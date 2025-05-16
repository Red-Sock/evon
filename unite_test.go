package evon

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_UniteNodes(t *testing.T) {
	src, trg := nodes()

	expected := &Node{
		Name: "",
		InnerNodes: []*Node{
			{
				Name:  "plain-text",
				Value: src.InnerNodes[0].Value,
			},
			{
				Name: "sub-module",
				InnerNodes: []*Node{
					{
						Name:  "sub-module_exact-number",
						Value: 1,
					},
				},
			},
			{
				Name: "sub-module-2",
				InnerNodes: []*Node{
					{
						Name:  "sub-module-2_exact-number",
						Value: 2,
					},
				},
			},
		},
	}

	Unite(src, trg)

	sort.Slice(trg.InnerNodes, func(i, j int) bool {
		return trg.InnerNodes[i].Name < trg.InnerNodes[j].Name
	})

	require.Equal(t, expected, trg)
}

func Test_UniteStorages(t *testing.T) {
	src, trg := nodes()
	srcStorage := NodesToStorage(src)
	trgStorage := NodesToStorage(trg)

	UniteStorages(srcStorage, trgStorage)
	expected := NodesToStorage(&Node{
		Name: "",
		InnerNodes: []*Node{
			{
				Name:  "plain-text",
				Value: src.InnerNodes[0].Value,
			},
			{
				Name: "sub-module",
				InnerNodes: []*Node{
					{
						Name:  "sub-module_exact-number",
						Value: 1,
					},
				},
			},
			{
				Name: "sub-module-2",
				InnerNodes: []*Node{
					{
						Name:  "sub-module-2_exact-number",
						Value: 2,
					},
				},
			},
		},
	})

	sort.Slice(trgStorage[""].InnerNodes, func(i, j int) bool {
		return trg.InnerNodes[i].Name < trg.InnerNodes[j].Name
	})

	require.Equal(t, expected, trgStorage)
}

func nodes() (*Node, *Node) {
	return &Node{
			Name: "",
			InnerNodes: []*Node{
				{
					Name:  "plain-text",
					Value: "123",
				},
				{
					Name: "sub-module",
					InnerNodes: []*Node{
						{
							Name:  "exact-number",
							Value: 1,
						},
					},
				},
			},
		}, &Node{
			Name: "",
			InnerNodes: []*Node{
				{
					Name:  "plain-text",
					Value: "321",
				},
				{
					Name: "sub-module-2",
					InnerNodes: []*Node{
						{
							Name:  "exact-number",
							Value: 2,
						},
					},
				},
			},
		}
}
