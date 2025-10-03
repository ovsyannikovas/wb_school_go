package main

import (
	"math"
)

type Point struct {
	x float64
	y float64
}

func NewPoint(x float64, y float64) *Point {
	return &Point{
		x, y,
	}
}

func (p *Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow((p.x-other.x), 2) + math.Pow((p.y-other.y), 2))
}
