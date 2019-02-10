// SPDX-License-Identifier: GPL-2.0
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

func PrintReflectValues(s reflect.Value) {
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("-- %s %s = %v\n",
			typeOfT.Field(i).Name, f.Type(), f.Interface())

		if f.Kind().String() == "struct" {
			x1 := reflect.ValueOf(f.Interface())
			PrintReflectValues(x1)
			fmt.Printf("\n")
		}
	}
}
