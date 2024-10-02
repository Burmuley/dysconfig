package main

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"alreadySnakeCase", "already_snake_case"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"aA", "a_a"},
		{"snake_case", "snake_case"},
		{"snake_Case", "snake_case"},
		{"APIResponse", "api_response"},
		{"UserID", "user_id"},
		{"HTML", "html"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toSnakeCase(test.input)
			if result != test.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}
