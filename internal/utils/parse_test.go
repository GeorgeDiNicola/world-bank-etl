package utils

import "testing"

func TestParseCSVStringIntoSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single value",
			input:    "USA",
			expected: []string{"USA"},
		},
		{
			name:     "multiple values with whitespace",
			input:    "USA, CAN, MEX",
			expected: []string{"USA", "CAN", "MEX"},
		},
		{
			name:     "ignores blank values",
			input:    "USA, , CAN,  ,MEX",
			expected: []string{"USA", "CAN", "MEX"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseCSVStringIntoSlice(tt.input)

			if len(actual) != len(tt.expected) {
				t.Fatalf("ParseCSVStringIntoSlice() returned %d items, expected %d", len(actual), len(tt.expected))
			}

			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Fatalf("ParseCSVStringIntoSlice() item %d = %q, expected %q", i, actual[i], tt.expected[i])
				}
			}
		})
	}
}
