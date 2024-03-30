package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	testCases := map[string]struct {
		input    int64
		expected string
	}{
		"encode 2468135791013": {
			input:    2468135791013,
			expected: "27qMi57J",
		},
		"encode 7489135791013": {
			input:    7489135791013,
			expected: "4PjAHW6Y",
		},
		"encode 5638910482": {
			input:    5638910482,
			expected: "9bHtdX",
		},
		"encode 00000000000": {
			input:    87452840931,
			expected: "3JEufoG",
		},
	}
	a := assert.New(t)
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			a.Equal(testCase.expected, encode(testCase.input))
		})
	}
}
