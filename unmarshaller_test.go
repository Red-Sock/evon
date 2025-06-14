package evon

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

func TestUnmarshalMatreshkaToMap(t *testing.T) {
	ns := ParseToNodes(simpleObjectDotEnv)

	targetMap := map[string]any{}
	err := UnmarshalWithNodes(ns, targetMap, WithSnakeUnmarshal())
	require.NoError(t, err)

	expected := map[string]any{
		"pointer_value":  "31",
		"root_int_value": "3",
		"child_obj_value": map[string]any{
			"bool_value":   "true",
			"string_value": "12",
		},
	}

	require.Equal(t, expected, targetMap)

	r, _ := yaml.Marshal(expected)
	print(string(r))
}
