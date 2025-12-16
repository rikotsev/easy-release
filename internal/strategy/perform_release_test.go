package strategy

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestExtractSemVerFromTitle(t *testing.T) {
	tcs := []struct {
		input    string
		expected *semver.Version
	}{
		{
			input:    "1.0.0",
			expected: semver.New(1, 0, 0, "", ""),
		},
		{
			input:    "8.12.123 asd",
			expected: semver.New(8, 12, 123, "", ""),
		},
		{
			input:    "20.30.500 #8",
			expected: semver.New(20, 30, 500, "", ""),
		},
		{
			input:    "1000.20.123435 (#8)",
			expected: semver.New(1000, 20, 123435, "", ""),
		},
		{
			input:    "     1000.20.123435       (#8)",
			expected: semver.New(1000, 20, 123435, "", ""),
		},
	}

	for idx, tc := range tcs {
		t.Run(fmt.Sprintf("tc: #%d", idx), func(t *testing.T) {
			actual, err := extractSemVerFromTitle(tc.input)

			if err != nil {
				t.Fatalf("got err: %v", err)
			}

			if tc.expected.Compare(actual) != 0 {
				t.Errorf("expected: %s, got: %s", tc.expected.String(), tc.input)
			}
		})
	}

}
