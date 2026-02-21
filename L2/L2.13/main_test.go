package main

import (
	"reflect"
	"testing"
)

func TestCutBasic(t *testing.T) {
	opts := &CutOptions{Delimiter: "\t"}
	fields := []int{1, 3}

	result, err := cut("a\tb\tc\td", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "a\tc"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCutWithoutFields(t *testing.T) {
	opts := &CutOptions{Delimiter: "\t"}
	fields := []int{}

	result, err := cut("a\tb\tc", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "a\tb\tc"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCutWithDifferentDelimiter(t *testing.T) {
	opts := &CutOptions{Delimiter: ","}
	fields := []int{2, 4}

	result, err := cut("a,b,c,d,e", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "b,d"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCutWithOutOfRangeFields(t *testing.T) {
	opts := &CutOptions{Delimiter: "\t"}
	fields := []int{1, 5, 10} // 5 и 10 вне диапазона

	result, err := cut("a\tb\tc", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "a" // только первое поле существует
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCutWithSeparatedFlag(t *testing.T) {
	opts := &CutOptions{Delimiter: "\t", Separated: true}
	fields := []int{1, 2}

	// Строка с разделителем
	result, err := cut("a\tb", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "a\tb"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Строка без разделителя - должна быть проигнорирована
	result, err = cut("ab", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string for line without delimiter, got %q", result)
	}
}

func TestCutWithEmptyDelimiter(t *testing.T) {
	opts := &CutOptions{Delimiter: ""}
	fields := []int{1}

	_, err := cut("test", opts, fields)
	if err == nil {
		t.Error("Expected error for empty delimiter, got nil")
	}
}

func TestCutWithMultipleDelimiters(t *testing.T) {
	opts := &CutOptions{Delimiter: ","}
	fields := []int{1, 3}

	// Пустые поля между разделителями
	result, err := cut("a,,c,d", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "a,c" // a - поле 1, пустое - поле 2, c - поле 3
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
		wantErr  bool
	}{
		{
			name:     "одиночное поле",
			input:    "1",
			expected: []int{1},
		},
		{
			name:     "несколько полей",
			input:    "1,3,5",
			expected: []int{1, 3, 5},
		},
		{
			name:     "диапазон",
			input:    "3-5",
			expected: []int{3, 4, 5},
		},
		{
			name:     "смешанный формат",
			input:    "1,3-5,7",
			expected: []int{1, 3, 4, 5, 7},
		},
		{
			name:     "с пробелами",
			input:    " 1 , 3-5 , 7 ",
			expected: []int{1, 3, 4, 5, 7},
		},
		{
			name:     "перекрывающиеся диапазоны",
			input:    "1-3,2-4",
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "дубликаты",
			input:    "1,1,2-3,2",
			expected: []int{1, 2, 3},
		},
		{
			name:    "пустая строка",
			input:   "",
			wantErr: true,
		},
		{
			name:    "неверный формат диапазона",
			input:   "1-",
			wantErr: true,
		},
		{
			name:    "неверный формат диапазона 2",
			input:   "-5",
			wantErr: true,
		},
		{
			name:    "диапазон start > end",
			input:   "5-3",
			wantErr: true,
		},
		{
			name:    "отрицательное число",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "не число",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFields(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCutEdgeCases(t *testing.T) {
	opts := &CutOptions{Delimiter: "\t"}
	fields := []int{1}

	// Пустая строка
	result, err := cut("", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}

	// Строка только с разделителями
	result, err = cut("\t\t", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestCutWithUnsortedFields(t *testing.T) {
	opts := &CutOptions{Delimiter: ","}
	fields := []int{3, 1, 4} // несортированный порядок

	result, err := cut("a,b,c,d,e", opts, fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Должен выводить в порядке запроса, а не сортировать
	expected := "c,a,d"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
