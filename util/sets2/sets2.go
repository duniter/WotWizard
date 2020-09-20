/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package sets2

// The module UtilSets2 implements the type Set, a set of non-negative integers, represented by map(s).

const (
	
	SMin = 0          // The smallest integer in sets.
	SMax = 1 << 31 - 2 // The largest integer in sets.

	taSet = SMax - SMin + 1

)

type (
	
	void struct {}

	// Set of non-negative integers.
	Set map [int] void

	SetIterator struct {
		s Set
		c <-chan int
	}

)

var v void

// Creates and returns a new empty set.
func NewSet () Set {
	return make(Set)
}

// Returns the set corresponding to the interval min..max.
func Interval (min, max int) Set {
	if !((min >= SMin) && (min <= SMax) && (max >= SMin) && (max <= SMax)) {
		panic(20)
	}
	s := NewSet()
	for e := min; e <= max; e++ {
		s[e] = v
	}
	return s
}

// Tests whether s is empty.
func (s Set) IsEmpty () bool {
	return len(s) == 0
}

// Return the number of elements in s
func (s Set) NbElems () int {
	return len(s)
}

// Creates and returns a copy of s.
func (s Set) Copy () Set {
	f := NewSet()
	for e := range s {
		f[e] = v
	}
	return f
}

func (s1 Set) Add (s2 Set) {
	for e := range s2 {
		s1[e] = v
	}
}

// Returns the union of s1 and s2.
func (s1 Set) Union (s2 Set) Set {
	s := s1.Copy()
	s.Add(s2)
	return s
}

// Returns the intersection of s1 and s2.
func (s1 Set) Inter (s2 Set) Set {
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	s := NewSet()
	for e := range s1 {
		if _, b := s2[e]; b {
			s[e] = v
		}
	}
	return s
}

// Returns the difference between s1 and s2.
func (s1 Set) Diff (s2 Set) Set {
	s := NewSet()
	for e := range s1 {
		if _, b := s2[e]; !b {
			s[e] = v
		}
	}
	return s
}

// Returns the exclusive union of s1 and s2.
func (s1 Set) XOR (s2 Set) Set {
	s := NewSet()
	for e := range s1 {
		if _, b := s2[e]; !b {
			s[e] = v
		}
	}
	for e := range s2 {
		if _, b := s1[e]; !b {
			s[e] = v
		}
	}
	return s
}

// Includes the integer e in the set s.
func (s Set) Incl (e int) {
	s[e] = v
}

// Excludes the integer e from the set s.
func (s Set) Excl (e int) {
	delete(s, e)
}

// Adds the interval min..max to the set s.
func (s Set) Fill (min, max int) {
	for e := min; e <= max; e++ {
		s[e] = v
	}
}

// Removes the interval min..max from the set s.
func (s Set) Clear (min, max int) {
	for e := min; e <= max; e++ {
		delete(s, e)
	}
}

// Returns the set corresponding to the bitset se
func Small (se uint64) Set {
	s := NewSet()
	var m uint64 = 1
	for i := 0; i < 64; i++ {
		if se & m == 1 {
			s[i] = v
		}
		se >>= 1
	}
	return s
}

// Tests whether the integer e is in the set s.
func (s Set) In (e int) bool {
	_, b := s[e]
	return b
}

// Tests whether s1 = s2.
func (s1 Set) Equal (s2 Set) bool {
	return s1.XOR(s2).IsEmpty()
}

// Tests whether s1 is a subset of s2.
func (s1 Set) Subset (s2 Set) bool {
	return s1.Diff(s2).IsEmpty()
}

// Attach the set s to a SetIterator and return the later.
func (s Set) Attach () *SetIterator {
	i := new(SetIterator)
	i.s = s
	return i
}

func yields (s Set, c chan<- int) {
	for e := range s {
		c <- e
	}
	close(c)
}

// On return, e contains the first element of the set attached to i. Returns true in ok if such an element exists. Usage: e, ok := i.FirstE(); for ok { ... e, ok = i.NextE()}
func (i *SetIterator) FirstE () (e int, ok bool) {
	c := make(chan int, 0)
	i.c = c
	go yields (i.s, c)
	e, ok = <-c
	return
}

// On return, e contains the next element of the set attached to i. Returns true in ok if such an element exists. i.FirstE must have been called once before i.NextE. Usage: e, ok := i.FirstE(); for ok { ... e, ok = i.NextE()}
func (i *SetIterator) NextE () (e int, ok bool) {
	e, ok = <-i.c
	return
}
