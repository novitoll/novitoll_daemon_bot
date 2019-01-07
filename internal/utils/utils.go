package utils

import (
	"fmt"
	"reflect"
)

func SumSliceInt(n []int) int {
	s := 0
	for i := range n {
		s += n[i]
	}
	return s
}

func CountDiffInPercent(a, b int) string {
	var d int = 100
	if a > 0 {
		d = (b * 100) / a
	}
	return fmt.Sprintf("%d%%", d-100)
}

func Destruct(v interface{}) {
    p := reflect.ValueOf(v).Elem()
    p.Set(reflect.Zero(p.Type()))
}
