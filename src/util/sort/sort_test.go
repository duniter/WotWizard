package sort

import (
	
	"testing"
	"util/alea"
	"fmt"

)

type (
	
	list []int

)

func (l list) Less (p1, p2 int) bool {
	return l[p1] < l[p2]
}

func (l list) Swap (p1, p2 int) {
	l[p1], l[p2] = l[p2], l[p1]
}

func TestSort (t *testing.T) {
	
	const (
		n = 10000000
		m = 1000000
	)
	
	l := make(list, n, n)
	for i := 0; i < n; i++ {
		l[i] = int(alea.Random() * float64(m))
	}
	var u TS
	u.Sorter = l
	u.QuickSort(0, len(l) - 1)
	fmt.Println(l[0])
	for i := 1; i < len(l); i++ {
		if l[i] < l[i - 1] {
			t.Fail()
		}
		//fmt.Println(l[i])
	}
}
