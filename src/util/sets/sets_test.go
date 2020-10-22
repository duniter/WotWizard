package sets

import (
		"fmt"
	R	"math/rand"
		"testing"
)

func createSet(max int, prop float32, print bool) *Set {
	nb := int(prop * float32(max))
	s := NewSet()
	for i := 1; i <= nb; i++ {
		n := int(R.Int31n(int32(max)))
		if print {
			fmt.Println(n)
		}
		s.Incl(n)
	}
	if print {
		fmt.Println()
	}
	return s
}

func printSet(s *Set) {
	i := s.Attach()
	e, ok := i.FirstE()
	for ok {
		fmt.Println(e)
		e, ok = i.NextE()
	}
	fmt.Println()
}

func TestCreate(t *testing.T) {
	const (
		max  = 10
		prop = 0.9
	)
	printSet(createSet(max, prop, true))
}

func TestXorUnionInterDiff(t *testing.T) {
	const (
		max  = 100
		prop = 0.9
	)
	e1 := createSet(max, prop, false)
	e2 := createSet(max, prop, false)
	f := e1.Union(e2).Diff(e1.Inter(e2))
	g := e1.XOR(e2)
	if !f.Equal(g) {
		t.Fail()
	}
	printSet(e1)
	printSet(e2)
	printSet(g)
}
