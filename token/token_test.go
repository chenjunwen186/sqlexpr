package token

import "testing"

func TestLookupIdent(t *testing.T) {
	type TestCase struct {
		input    string
		expected Type
	}
	tests := []TestCase{
		{"case", CASE},
		{"eND", END},
		{"When", WHEN},
		{"True", TRUE},
		{"FALSE", FALSE},
	}

	for _, test := range tests {
		actual := LookupIdent(test.input)
		if actual.Type != test.expected {
			t.Errorf("LookupIdent(%q) wrong. expected=%q, got=%q", test.input, test.expected, actual)
		}
	}
}
