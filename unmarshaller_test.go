package evon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalToMap(t *testing.T) {
	targetMap := map[string]any{}
	err := Unmarshal(simpleObjectDotEnv, targetMap)
	require.NoError(t, err)

	expected := map[string]any{
		"POINTER-VALUE":  "31",
		"ROOT-INT-VALUE": "3",
		"CHILD-OBJ-VALUE": map[string]any{
			"BOOL-VALUE":   "true",
			"STRING-VALUE": "12",
		},
	}
	require.Equal(t, expected, targetMap)
}
