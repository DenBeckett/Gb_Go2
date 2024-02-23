package main

import (
	"fmt"
)

func main() {
	a1 := [4]int{6, 54, 78, 93}
	a2 := [5]int{10, 23, 44, 59, 81}

	fmt.Println(a1, " + ", a2)

	mergearray(a1[0:], a2[0:])
}

func mergearray(a, b []int) []int {
	var i, j int
	l := len(a) + len(b)
	r := make([]int, 0, l)
	for {
		if i == len(a) {
			r = append(r, b[j:]...)
			fmt.Println(r)
			return r
		}

		if j == len(b) {
			r = append(r, a[i:]...)
			fmt.Println(r)
			return r
		}

		if a[i] < b[j] {
			r = append(r, a[i])
			fmt.Println(r)
			i++
		} else if b[j] < a[i] {
			r = append(r, b[j])
			fmt.Println(r)
			j++
		} else {
			r = append(r, a[i])
			r = append(r, b[j])
			fmt.Println(r)
			i++
			j++
		}
	}
}
