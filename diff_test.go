package evon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Diff(t *testing.T) {
	oldNodes := &Node{
		Name: "",
		InnerNodes: []*Node{
			{
				Name:  "string",
				Value: "13",
			},
			{
				Name: "inner-nodes",
				InnerNodes: []*Node{
					{
						Name:  "deleted-string",
						Value: "13",
					},
				},
			},
		},
	}

	newNodes := &Node{
		Name: "",
		InnerNodes: []*Node{
			{
				Name:  "string",
				Value: "13",
			},
			{
				Name: "inner-nodes",
				InnerNodes: []*Node{
					{
						Name:  "added-string",
						Value: "14",
					},
				},
			},
		},
	}

	actual := Diff(oldNodes, newNodes)

	expected := NodeDiff{
		NewNodes: []*Node{
			{
				Name:  "inner-nodes_added-string",
				Value: "14",
			},
		},
		RemovedNodes: []*Node{
			{
				Name:  "inner-nodes_deleted-string",
				Value: "13",
			},
		},
	}

	require.Equal(t, expected, actual)
}
