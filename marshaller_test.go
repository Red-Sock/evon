package evon

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	testPrefix = "COOLOBJ"
	noPrefix   = ""
)

func TestMarshalling(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          TestObject
		expectedErr error
	}

	testCases := map[string]struct{ constructor func() testCase }{
		"OK": {
			constructor: func() testCase {
				return testCase{
					in: NewTestObject(),
				}
			},
		},

		"OK_WITH_NIL_FIELD": {
			constructor: func() testCase {
				nilTo := NewTestObject()
				nilTo.PointerValue = nil

				return testCase{
					in: nilTo,
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc := tc.constructor()
			actualOut, actualErr := MarshalEnv(tc.in)

			require.ErrorIs(t, actualErr, tc.expectedErr)
			require.Equal(t, tc.in.ExpectedObjNodes(noPrefix), actualOut)
		})
	}

}

func TestMarshallingWithPrefix(t *testing.T) {
	t.Parallel()

	type testCase struct {
		prefix      string
		in          any
		expectedOut *Node
		expectedErr error
	}

	to := NewTestObject()

	testCases := map[string]testCase{
		"OK": {
			prefix:      testPrefix,
			in:          to,
			expectedOut: to.ExpectedObjNodes(testPrefix),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualOut, actualErr := MarshalEnvWithPrefix(tc.prefix, tc.in)

			require.ErrorIs(t, actualErr, tc.expectedErr)
			require.Equal(t, tc.expectedOut, actualOut)
		})
	}

}

func TestMarshallingToFile(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          []*Node
		expectedOut []byte
	}

	to := NewTestObject()

	testCases := map[string]testCase{
		"OK": {
			in:          to.ExpectedObjNodes(noPrefix).InnerNodes,
			expectedOut: fullObjectDotEnv,
		},
		"OK_PREFIXED": {
			in:          to.ExpectedObjNodes(testPrefix).InnerNodes,
			expectedOut: prefixedExpectedDotEnv,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualOut := Marshal(tc.in)
			require.Equal(t, string(tc.expectedOut), string(actualOut))
		})
	}

}

func TestMarshalFromYaml(t *testing.T) {
	expectedYamlMap := map[string]any{}

	err := yaml.Unmarshal(complexYamlConfig, expectedYamlMap)
	require.NoError(t, err)

	n, err := MarshalEnv(expectedYamlMap)
	require.NoError(t, err)

	//TODO check n is correct

	ns := NodeStorage{}
	ns.AddNode(n)

	evonMap := map[string]any{}
	err = UnmarshalWithNodes(ns, evonMap,
		WithSnakeUnmarshal())
	require.NoError(t, err)

	actualYamlMap, err := yaml.Marshal(evonMap)
	require.NoError(t, err)

	require.YAMLEq(t, string(complexYamlConfig), string(actualYamlMap))

}

//
