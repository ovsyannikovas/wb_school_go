package main

import (
	"testing"
)

func TestUnpackString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic unpacking",
			input:       "a4bc2d5e",
			expected:    "aaaabccddddde",
			expectError: false,
		},
		{
			name:        "no digits - unchanged",
			input:       "abcd",
			expected:    "abcd",
			expectError: false,
		},
		{
			name:        "single character with repeat",
			input:       "a5",
			expected:    "aaaaa",
			expectError: false,
		},
		{
			name:        "multiple same characters",
			input:       "a2b3c4",
			expected:    "aabbbcccc",
			expectError: false,
		},

		// Errors
		{
			name:        "only digits",
			input:       "45",
			expected:    "",
			expectError: true,
			errorMsg:    "некорректная строка: цифра без предшествующего символа",
		},
		{
			name:        "starts with digit",
			input:       "5abc",
			expected:    "",
			expectError: true,
			errorMsg:    "некорректная строка: цифра без предшествующего символа",
		},
		{
			name:        "starts with zero",
			input:       "0abc",
			expected:    "",
			expectError: true,
			errorMsg:    "некорректное число: 0",
		},
		{
			name:        "multiple digits at start",
			input:       "123abc",
			expected:    "",
			expectError: true,
		},
		{
			name:        "zero repeat",
			input:       "a0b",
			expected:    "",
			expectError: true,
			errorMsg:    "некорректное число: 0",
		},

		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: false,
		},

		// Escape
		{
			name:        "escaped digits",
			input:       "qwe\\4\\5",
			expected:    "qwe45",
			expectError: false,
		},
		{
			name:        "mixed escape sequence",
			input:       "qwe\\45",
			expected:    "qwe44444",
			expectError: false,
		},
		{
			name:        "escape with letter",
			input:       "qwe\\a",
			expected:    "qwea",
			expectError: false,
		},
		{
			name:        "escape backslash",
			input:       "qwe\\\\5",
			expected:    "qwe\\\\\\\\\\",
			expectError: false,
		},
		{
			name:        "incomplete escape sequence",
			input:       "qwe\\",
			expected:    "",
			expectError: true,
			errorMsg:    "некорректная escape-последовательность",
		},

		// Многозначные числа
		{
			name:        "multi-digit repeat",
			input:       "a10b",
			expected:    "aaaaaaaaaab",
			expectError: false,
		},
		{
			name:        "large number repeat",
			input:       "a12",
			expected:    "aaaaaaaaaaaa",
			expectError: false,
		},
		{
			name:        "multiple multi-digit numbers",
			input:       "a10b20",
			expected:    "aaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
			expectError: false,
		},

		// Специальные символы и Unicode
		{
			name:        "unicode characters",
			input:       "ф2ц3",
			expected:    "ффццц",
			expectError: false,
		},
		{
			name:        "special characters",
			input:       "!2@3#4",
			expected:    "!!@@@####",
			expectError: false,
		},
		{
			name:        "spaces in string",
			input:       "a 2b 3",
			expected:    "a  b   ",
			expectError: false,
		},

		{
			name:        "single character",
			input:       "a",
			expected:    "a",
			expectError: false,
		},
		{
			name:        "single digit after valid char",
			input:       "a1",
			expected:    "a",
			expectError: false,
		},
		{
			name:        "only one escaped digit",
			input:       "\\4",
			expected:    "4",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UnpackString(tt.input)

			if (err != nil) != tt.expectError {
				t.Errorf("UnpackString(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
				return
			}

			if tt.expectError && tt.errorMsg != "" && err != nil {
				if err.Error() != tt.errorMsg {
					t.Errorf("UnpackString(%q) error message = %q, expected %q", tt.input, err.Error(), tt.errorMsg)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("UnpackString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
