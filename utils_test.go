package main

import (
	"reflect"
	"testing"
)

func TestDeleteElementFromStrSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		element  string
		expected []string
	}{
		{
			name:     "Delete existing element (case sensitive)",
			input:    []string{"apple", "banana", "cherry"},
			element:  "banana",
			expected: []string{"apple", "cherry"},
		},
		{
			name:     "Delete existing element (case insensitive)",
			input:    []string{"apple", "Banana", "cherry"},
			element:  "banana",
			expected: []string{"apple", "cherry"},
		},
		{
			name:     "Element not found",
			input:    []string{"apple", "banana", "cherry"},
			element:  "orange",
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Delete first element",
			input:    []string{"Apple", "banana", "cherry"},
			element:  "apple",
			expected: []string{"banana", "cherry"},
		},
		{
			name:     "Delete last element",
			input:    []string{"apple", "banana", "Cherry"},
			element:  "cherry",
			expected: []string{"apple", "banana"},
		},
		{
			name:     "Delete only element",
			input:    []string{"apple"},
			element:  "apple",
			expected: []string{},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			element:  "apple",
			expected: []string{},
		},
		{
			name:     "Multiple matches, only first deleted",
			input:    []string{"apple", "banana", "apple", "cherry"},
			element:  "apple",
			expected: []string{"banana", "apple", "cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deleteElementFromStrSlice(tt.input, tt.element)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("deleteElementFromStrSlice(%v, %q) = %v; want %v", tt.input, tt.element, result, tt.expected)
			}
		})
	}
}
