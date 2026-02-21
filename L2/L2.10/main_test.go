package main

import (
	"reflect"
	"testing"
)

func TestBasicSort(t *testing.T) {
	input := []string{"c", "b", "a"}
	opts := &SortOptions{}
	expected := []string{"a", "b", "c"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestReverse(t *testing.T) {
	input := []string{"a", "b", "c"}
	opts := &SortOptions{Reverse: true}
	expected := []string{"c", "b", "a"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestUnique(t *testing.T) {
	input := []string{"a", "a", "b", "b", "c"}
	opts := &SortOptions{Unique: true}
	expected := []string{"a", "b", "c"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestColumn(t *testing.T) {
	input := []string{"b\t2", "a\t1", "c\t3"}
	opts := &SortOptions{Column: 2}
	expected := []string{"a\t1", "b\t2", "c\t3"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestNumeric(t *testing.T) {
	input := []string{"10", "2", "33", "1"}
	opts := &SortOptions{Numeric: true}
	expected := []string{"1", "2", "10", "33"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestNumericWithNonNumbers(t *testing.T) {
	input := []string{"abc", "10", "2", "xyz", "1"}
	opts := &SortOptions{Numeric: true}
	expected := []string{"abc", "xyz", "1", "2", "10"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestHumanNumeric(t *testing.T) {
	input := []string{"2M", "1K", "500", "1G"}
	opts := &SortOptions{HumanNumeric: true}
	expected := []string{"500", "1K", "2M", "1G"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMonth(t *testing.T) {
	input := []string{"23 Mar 2024", "15 Jan 2024", "01 Feb 2024"}
	opts := &SortOptions{Month: true}
	expected := []string{"15 Jan 2024", "01 Feb 2024", "23 Mar 2024"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestColumnAndNumeric(t *testing.T) {
	input := []string{"b\t10", "a\t2", "c\t33"}
	opts := &SortOptions{Column: 2, Numeric: true}
	expected := []string{"a\t2", "b\t10", "c\t33"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestColumnAndReverse(t *testing.T) {
	input := []string{"b\t2", "a\t1", "c\t3"}
	opts := &SortOptions{Column: 2, Reverse: true}
	expected := []string{"c\t3", "b\t2", "a\t1"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestNumericAndReverse(t *testing.T) {
	input := []string{"10", "2", "33", "1"}
	opts := &SortOptions{Numeric: true, Reverse: true}
	expected := []string{"33", "10", "2", "1"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestUniqueAndNumeric(t *testing.T) {
	input := []string{"10", "2", "10", "33", "2", "1"}
	opts := &SortOptions{Unique: true, Numeric: true}
	expected := []string{"1", "2", "10", "33"}

	result, err := sortLines(input, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMonthDetailed(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected bool // true если a < b
	}{
		{
			name:     "Jan vs Feb",
			a:        "15 Jan 2024",
			b:        "01 Feb 2024",
			expected: true, // Jan должен быть перед Feb
		},
		{
			name:     "Feb vs Mar",
			a:        "01 Feb 2024",
			b:        "23 Mar 2024",
			expected: true, // Feb перед Mar
		},
		{
			name:     "Mar vs Jan",
			a:        "23 Mar 2024",
			b:        "15 Jan 2024",
			expected: false, // Mar после Jan
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &SortOptions{Month: true}
			lines := []string{tt.a, tt.b}
			result, err := sortLines(lines, opts)
			if err != nil {
				t.Fatal(err)
			}

			// Проверяем порядок
			if tt.expected && result[0] != tt.a {
				t.Errorf("Expected %s before %s, got %s before %s", tt.a, tt.b, result[0], result[1])
			}
			if !tt.expected && result[0] != tt.b {
				t.Errorf("Expected %s before %s, got %s before %s", tt.b, tt.a, result[0], result[1])
			}
		})
	}
}

func TestMonthCompare(t *testing.T) {
	opts := &SortOptions{Month: true}
	lines := []string{"15 Jan 2024", "01 Feb 2024"}

	result := compare(0, 1, lines, opts)
	if !result {
		t.Errorf("Expected Jan < Feb, got false")
	}

	result = compare(1, 0, lines, opts)
	if result {
		t.Errorf("Expected Feb > Jan, got true")
	}
}

func TestParseMonth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"Jan", 1, false},
		{"15 Jan 2024", 1, false},
		{"FEB", 2, false},
		{"01 Feb 2024", 2, false},
		{"mar", 3, false},
		{"23 Mar 2024", 3, false},
		{"Apr", 4, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseMonth(tt.input, false)
			if tt.hasError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.input)
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("parseMonth(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
