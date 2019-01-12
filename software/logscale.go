package main

import (
	"math"
)

func round(f float64) int {
	return int(math.Floor(f + .5))
}

func scaleVal(start, end, steps, step int) int {
	s := float64(start) - 1 + 2
	e := float64(end) + 2
	n := float64(steps)
	i := float64(step)
	p := math.Pow(e/s, 1/n)
	s1 := math.Ceil(1.5 / (p - 1)) // steps of 1 from s up to s1
	if s+i < s1 || e-s < n {
		return int(s + i - 2)
	} else {
		intSteps := s1 - s
		n1 := n - intSteps
		i1 := i - intSteps
		p1 := math.Pow(e/s1, 1/n1)
		return round(math.Pow(p1, i1) * s1) - 2
	}
}
