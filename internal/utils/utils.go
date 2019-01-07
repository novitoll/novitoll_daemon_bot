package utils

import (
	"fmt"
)

func SumSliceInt(n []int) int {
	s := 0
	for i := range n {
		s += n[i]
	}
	return s
}

func CountDiffInPercent(a, b int) string {
	d := (b * 100) / a
	return fmt.Sprintf("%d%%", d-100)
}
