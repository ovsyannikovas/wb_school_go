package main

import (
	"math"
	"testing"
)

func TestNewPoint(t *testing.T) {
	tests := []struct {
		name     string
		x        float64
		y        float64
		expected Point
	}{
		{"Positive coordinates", 3.0, 4.0, Point{3.0, 4.0}},
		{"Negative coordinates", -2.5, -1.5, Point{-2.5, -1.5}},
		{"Zero coordinates", 0.0, 0.0, Point{0.0, 0.0}},
		{"Mixed coordinates", 5.5, -3.2, Point{5.5, -3.2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := NewPoint(tt.x, tt.y)
			if point.x != tt.expected.x || point.y != tt.expected.y {
				t.Errorf("NewPoint(%f, %f) = (%f, %f), expected (%f, %f)",
					tt.x, tt.y, point.x, point.y, tt.expected.x, tt.expected.y)
			}
		})
	}
}

func TestPoint_Distance(t *testing.T) {
	tests := []struct {
		name     string
		p1       *Point
		p2       Point
		expected float64
	}{
		{
			name:     "Same point",
			p1:       NewPoint(3.0, 4.0),
			p2:       Point{3.0, 4.0},
			expected: 0.0,
		},
		{
			name:     "Horizontal distance",
			p1:       NewPoint(0.0, 0.0),
			p2:       Point{5.0, 0.0},
			expected: 5.0,
		},
		{
			name:     "Vertical distance",
			p1:       NewPoint(0.0, 0.0),
			p2:       Point{0.0, 3.0},
			expected: 3.0,
		},
		{
			name:     "Diagonal distance",
			p1:       NewPoint(0.0, 0.0),
			p2:       Point{3.0, 4.0},
			expected: 5.0,
		},
		{
			name:     "Negative coordinates",
			p1:       NewPoint(-1.0, -1.0),
			p2:       Point{2.0, 3.0},
			expected: 5.0,
		},
		{
			name:     "Your example from main",
			p1:       NewPoint(6.0, 4.0),
			p2:       Point{3.0, 5.0},
			expected: math.Sqrt(10), // √((6-3)² + (4-5)²) = √(9 + 1) = √10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := tt.p1.Distance(tt.p2)

			// Используем допуск для сравнения чисел с плавающей точкой
			if math.Abs(distance-tt.expected) > 1e-9 {
				t.Errorf("Distance() = %f, expected %f", distance, tt.expected)
			}
		})
	}
}

func TestPoint_Distance_Commutative(t *testing.T) {
	p1 := NewPoint(1.0, 2.0)
	p2 := Point{4.0, 6.0}

	distance1 := p1.Distance(p2)
	distance2 := p2.Distance(*p1) // Обратное расстояние

	if math.Abs(distance1-distance2) > 1e-9 {
		t.Errorf("Distance should be commutative: %f != %f", distance1, distance2)
	}
}
