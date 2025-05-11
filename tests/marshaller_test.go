package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.redsock.ru/evon"
)

func TestMarshalling(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          any
		expectedOut *evon.Node
		expectedErr error
	}

	testCases := map[string]testCase{
		"OK": {
			in:          newTestObject(),
			expectedOut: expectedObjNodes(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualOut, actualErr := evon.MarshalEnv(tc.in)

			require.ErrorIs(t, actualErr, tc.expectedErr)
			require.Equal(t, tc.expectedOut, actualOut)
		})
	}

}
