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
		"POINTER-VALUE":  31,
		"ROOT-INT-VALUE": 3,
		"CHILD-OBJ-VALUE": map[string]any{
			"BOOL-VALUE":   true,
			"STRING-VALUE": 12,
		},
	}
	require.Equal(t, expected, targetMap)
}

func TestUnmarshallWithSlice(t *testing.T) {
	type TestObject struct {
		UnknownArray []UntaggedStruct `evon:"UNKNOWN-ARRAY"`
	}

	type testCase struct {
		input []byte
	}

	tests := map[string]testCase{
		"one_line": {
			input: []byte(`UNKNOWN-ARRAY_[0]_A=1`),
		},
		"line_with_break_end": {
			input: []byte(`UNKNOWN-ARRAY_[0]_A=1
`),
		},
		"line_with_break_start": {
			input: []byte(`
UNKNOWN-ARRAY_[0]_A=1`),
		},
		"line_with_break_start_end": {
			input: []byte(`
UNKNOWN-ARRAY_[0]_A=1`),
		},
	}

	expected := &TestObject{
		UnknownArray: []UntaggedStruct{
			{A: 1},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := &TestObject{}
			err := Unmarshal(tc.input, &actual)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}

}

func TestUnmarshallWithSliceWithinSlice(t *testing.T) {
	type LowestTierStruct struct {
		B int
	}

	type MiddleStruct struct {
		A []LowestTierStruct
	}

	type TopStruct struct {
		UnknownArray []MiddleStruct
	}

	type testCase struct {
		input    []byte
		expected TopStruct
	}

	tests := map[string]testCase{
		"one_val": {
			input: []byte(`UNKNOWN-ARRAY_[0]_A_[0]_B=1`),
			expected: TopStruct{
				UnknownArray: []MiddleStruct{
					{
						A: []LowestTierStruct{
							{B: 1},
						},
					},
				},
			},
		},
		"two_middle": {
			input: []byte(`
UNKNOWN-ARRAY_[0]_A_[0]_B=1
UNKNOWN-ARRAY_[1]_A_[0]_B=2
`),
			expected: TopStruct{
				UnknownArray: []MiddleStruct{
					{
						A: []LowestTierStruct{{B: 1}},
					},
					{
						A: []LowestTierStruct{{B: 2}},
					},
				},
			},
		},
		"two_lowest": {
			input: []byte(`
UNKNOWN-ARRAY_[0]_A_[0]_B=1
UNKNOWN-ARRAY_[0]_A_[1]_B=2
`),
			expected: TopStruct{
				UnknownArray: []MiddleStruct{
					{
						A: []LowestTierStruct{{B: 1}, {B: 2}},
					},
				},
			},
		},
		"two_lowest_two_middle": {
			input: []byte(`
UNKNOWN-ARRAY_[0]_A_[0]_B=1
UNKNOWN-ARRAY_[0]_A_[1]_B=2
UNKNOWN-ARRAY_[1]_A_[0]_B=3
UNKNOWN-ARRAY_[1]_A_[1]_B=4
`),
			expected: TopStruct{
				UnknownArray: []MiddleStruct{
					{
						A: []LowestTierStruct{{B: 1}, {B: 2}},
					},
					{
						A: []LowestTierStruct{{B: 3}, {B: 4}},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := TopStruct{}
			err := Unmarshal(tc.input, &actual)
			require.NoError(t, err)

			require.Equal(t, tc.expected, actual)
		})
	}

}

func TestUnmarshalMatreshkaToMap(t *testing.T) {
	ns := ParseToNodes(simpleObjectDotEnv)

	targetMap := map[string]any{}
	err := UnmarshalWithNodes(ns, targetMap, WithSnakeUnmarshal())
	require.NoError(t, err)

	expected := map[string]any{
		"pointer_value":  31,
		"root_int_value": 3,
		"child_obj_value": map[string]any{
			"bool_value":   true,
			"string_value": 12,
		},
	}

	require.Equal(t, expected, targetMap)

	r, _ := yaml.Marshal(expected)
	print(string(r))
}
