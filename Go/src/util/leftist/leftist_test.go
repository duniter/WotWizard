package leftist

import (
	
	A "util/alea"
	     "fmt"
	     "testing"

)

type (
	
	elem struct {
		next *Elem
		n int
	}
	
	objectV int
	
	objectE struct {
		end int
		v *Elem
	}

)

var gen *A.Generator

func (e1 *elem) Comp (e2 Comparer) Comp {
	ee2 := e2.(*elem)
	switch {
		case e1.n < ee2.n: {
			return First
		}
		case e1.n > ee2.n: {
			return Last
		}
		default: {
			return Equiv
		}
	}
}

func (e1 objectV) Comp (e2 Comparer) Comp {
	i1 := int(e1)
	i2 := int(e2.(objectV))
	switch {
		case i1 < i2: {
			return First
		}
		case i1 > i2: {
			return Last
		}
		default: {
			return Equiv
		}
	}
}

func (e1 *objectE) Comp (e2 Comparer) Comp {
	ee2 := e2.(*objectE)
	switch {
		case e1.end < ee2.end: {
			return First
		}
		case e1.end > ee2.end: {
			return Last
		}
		default: {
			return Equiv
		}
	}
}

func TestL1 (tt *testing.T) {
	const Nb = 1000
	gen.Randomize(1)
	t := New()
	for i := 1; i <= Nb; i++ {
		e := new(elem)
		e.n = int(gen.IntRand(0, Nb))
		_ = t.Insert(e)
	}
	var el *Elem
	n := 0
	e := t.First(&el)
	for e != nil {
		n++
		ee := e.(*elem)
		fmt.Println(ee.n)
		t.Erase(el)
		e = t.First(&el)
	}
	if  n != Nb {tt.Fail()}
}

func TestL2 (tt *testing.T) {
	const Nb = 1000
	gen.Randomize(1);
	t := New()
	var e1 *Elem = nil
	for i := 1; i <= Nb; i++ {
		e2 := new(elem)
		e2.n = int(gen.IntRand(0, Nb))
		e2.next = e1
		e1 = t.Insert(e2)
	}
	i := 0
	for e1 != nil {
		i++
		e2 := e1.Val().(*elem)
		fmt.Println(e2.n)
		t.Erase(e1)
		e1 = e2.next
	}
	if  i != Nb || !t.IsEmpty() {tt.Fail()}
}

func TestL3 (tt *testing.T) {
	const (
		nbElems = 10000
		timeMin = 100
		timeMax = 500
		valMin = 0
		valMax = 10000
	)
	gen.Randomize(1)
	tV := New(); tE := New()
	var elE *Elem
	n := 0
	e := tE.First(&elE)
	for n < nbElems || e != nil {
		fmt.Println("n = ", n)
		if n < nbElems {
			v := objectV(gen.IntRand(valMin, valMax + 1))
			fmt.Println("+", int(v))
			elV := tV.Insert(v)
			eE := new(objectE)
			eE.end = n + int(gen.IntRand(timeMin, timeMax + 1))
			eE.v = elV
			_ = tE.Insert(eE)
			e = tE.First(&elE)
		}
		for e != nil && e.(*objectE).end == n {
			v := e.(*objectE).v
			fmt.Println("-", int(v.Val().(objectV)))
			tV.Erase(v)
			tE.Erase(elE)
			e = tE.First(&elE)
		}
		n++
	}
	if  !tV.IsEmpty() {tt.Fail()}
}

func init () {
	gen = A.New()
}
