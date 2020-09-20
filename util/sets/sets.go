/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package sets

// The module UtilSets implements the type Set, a set of non-negative integers, represented by intervals.

const (
	SMin = 0          // The smallest integer in sets.
	SMax = 1 << 31 - 2 // The largest integer in sets.

	taSet = SMax - SMin + 1
)

type (
	doublet [2]int

	segment struct {
		suivant,
		precedent *segment
		largeurs doublet
	}

	// Set of non-negative integers.
	Set struct {
		nbElems int // Number of elements in the set.
		elems   *segment
	}

	tableVer [2][2]int

	SetIterator struct {
		s         *Set
		cour      *segment
		pos, posE int
	}
)

func insereSegment(apresSeg *segment, plein, vide int) {
	s := new(segment)
	s.suivant = apresSeg.suivant
	s.precedent = apresSeg
	s.largeurs[1] = plein
	s.largeurs[0] = vide
	apresSeg.suivant.precedent = s
	apresSeg.suivant = s
}

func (s *Set) trans(src *Set) {
	s.nbElems = src.nbElems
	s.elems = src.elems
}

// Creates and returns a new empty set.
func NewSet() *Set {
	s := new(Set)
	s.elems = new(segment)
	s.nbElems = 0
	s.elems.largeurs[0] = taSet
	s.elems.suivant = s.elems
	s.elems.precedent = s.elems
	return s
}

// Returns the largest set available: SMin..SMax.
func Full() *Set {
	s := NewSet()
	s.nbElems = taSet
	s.elems.largeurs[0] = 0
	insereSegment(s.elems, taSet, 0)
	return s
}

// Returns the set corresponding to the interval min..max.
func Interval(min, max int) *Set {
	if !((min >= SMin) && (min <= SMax) && (max >= SMin) && (max <= SMax)) {
		panic(20)
	}
	s := NewSet()
	if min <= max {
		s.nbElems = max - min + 1
		s.elems.largeurs[0] = min
		insereSegment(s.elems, s.nbElems, taSet-max-1)
	}
	return s
}

// Tests whether s is empty.
func (s *Set) IsEmpty() bool {
	return s.elems.suivant == s.elems
}

// Tests whether s is empty.
func (s *Set) NbElems() int {
	return int(s.nbElems)
}

// Creates and returns a copy of s.
func (s *Set) Copy() *Set {
	f := NewSet()
	f.nbElems = s.nbElems
	f.elems.largeurs[0] = s.elems.largeurs[0]
	p := s.elems.suivant
	for p != s.elems {
		insereSegment(f.elems.precedent, p.largeurs[1], p.largeurs[0])
		p = p.suivant
	}
	return f
}

func (s1 *Set) melange(s2 *Set, combi tableVer) *Set {
	s := NewSet()
	combine := func(c1, c2, c *segment) {
		s1 := c1
		plein1 := 0
		s2 := c2
		plein2 := 0
		plein3 := 0
		fin := false
		h1 := 0
		h2 := 0
		h3 := 0
		larg := 0
		for {
			if (combi[plein1][plein2] == plein3) && !fin {
				long1 := h1 + s1.largeurs[plein1]
				long2 := h2 + s2.largeurs[plein2]
				if long1 <= long2 {
					h1 = long1
					plein1 = 1 - plein1
					if plein1 != 0 {
						s1 = s1.suivant
						fin = s1 == c1
					}
					larg += long1 - h3
					h3 = long1
				}
				if long2 <= long1 {
					h2 = long2
					plein2 = 1 - plein2
					if plein2 != 0 {
						s2 = s2.suivant
					}
					larg += long2 - h3
					h3 = long2
				}
			} else {
				if plein3 == 0 {
					c.precedent.largeurs[plein3] = larg
				} else {
					insereSegment(c.precedent, larg, 0)
					s.nbElems += larg
				}
				if fin {
					break
				}
				larg = 0
				plein3 = 1 - plein3
			}
		}
	}
	combine(s1.elems, s2.elems, s.elems)
	return s
}

// Returns the union of s1 and s2.
func (s1 *Set) Union(s2 *Set) *Set {
	var s *Set
	if s1.IsEmpty() {
		s = s2.Copy()
	} else if s2.IsEmpty() {
		s = s1.Copy()
	} else {
		s = s1.melange(s2, tableVer{[2]int{0, 1}, [2]int{1, 1}})
	}
	return s
}

// Returns the intersection of s1 and s2.
func (s1 *Set) Inter(s2 *Set) *Set {
	var s *Set
	if s1.IsEmpty() || s2.IsEmpty() {
		s = NewSet()
	} else {
		s = s1.melange(s2, tableVer{[2]int{0, 0}, [2]int{0, 1}})
	}
	return s
}

// Returns the difference between s1 and s2.
func (s1 *Set) Diff(s2 *Set) *Set {
	var s *Set
	if s2.IsEmpty() {
		s = s1.Copy()
	} else if s1.IsEmpty() {
		s = NewSet()
	} else {
		s = s1.melange(s2, tableVer{[2]int{0, 0}, [2]int{1, 0}})
	}
	return s
}

// Returns the exclusive union of s1 and s2.
func (s1 *Set) XOR(s2 *Set) *Set {
	var s *Set
	if s1.IsEmpty() {
		s = s2.Copy()
	} else if s2.IsEmpty() {
		s = s1.Copy()
	} else {
		s = s1.melange(s2, tableVer{[2]int{0, 1}, [2]int{1, 0}})
	}
	return s
}

// Includes the integer e in the set s.
func (s *Set) Incl(e int) {
	ss := Interval(e, e)
	ss = s.Union(ss)
	s.trans(ss)
}

// Excludes the integer e from the set s.
func (s *Set) Excl(e int) {
	ss := Interval(e, e)
	ss = s.Diff(ss)
	s.trans(ss)
}

// Adds the interval min..max to the set s.
func (s *Set) Fill(min, max int) {
	ss := Interval(min, max)
	ss = s.Union(ss)
	s.trans(ss)
}

// Removes the interval min..max from the set s.
func (s *Set) Clear(min, max int) {
	ss := Interval(min, max)
	ss = s.Diff(ss)
	s.trans(ss)
}

// Returns the set corresponding to the bitset se
func Small(se uint64) *Set {
	s := NewSet()
	var m uint64 = 1
	for i := 0; i < 64; i++ {
		if se & m == 1 {
			s.Incl(i)
		}
		se >>= 1
	}
	return s
}

// Tests whether the integer e is in the set s.
func (s *Set) In(e int) bool {
	if e < SMin || e > SMax {
		return false
	}
	se := s.elems
	plein := 0
	for {
		if e < se.largeurs[plein] {
			return plein != 0
		} else {
			e -= se.largeurs[plein]
		}
		plein = 1 - plein
		if plein != 0 {
			se = se.suivant
		}
	}
}

// Tests whether s1 = s2.
func (s1 *Set) Equal(s2 *Set) bool {
	return s1.XOR(s2).IsEmpty()
}

// Tests whether s1 is a subset of s2.
func (s1 *Set) Subset(s2 *Set) bool {
	return s1.Diff(s2).IsEmpty()
}

// Attach the set s to a SetIterator and return the later.
func (s *Set) Attach() *SetIterator {
	i := new(SetIterator)
	i.s = s
	i.cour = nil
	return i
}

// On return, min..max is the first interval of the set attached to i. Returns true in ok if such an interval exists. Usage: min, max, ok := i.First(); for ok { ... min, max, ok = i.Next()}
func (i *SetIterator) First() (min, max int, ok bool) {
	if i.s == nil {
		panic(20)
	}
	if i.s.IsEmpty() {
		ok = false
		return
	}
	i.cour = i.s.elems.suivant
	i.pos = SMin + i.s.elems.largeurs[0]
	min = i.pos
	max = min + i.cour.largeurs[1] - 1
	ok = true
	return
}

// On return, min..max is the next interval of the set attached to i. Returns true in ok if such an interval exists. i.First must have been called once before i.Next. Usage: min, max, ok := i.First(); for ok { ... min, max, ok = i.Next()}
func (i *SetIterator) Next() (min, max int, ok bool) {
	if i.s == nil {
		panic(20)
	}
	if i.cour == nil {
		panic(21)
	}
	i.pos += i.cour.largeurs[1] + i.cour.largeurs[0]
	i.cour = i.cour.suivant
	if i.cour == i.s.elems {
		i.cour = nil
		ok = false
		return
	}
	min = i.pos
	max = min + i.cour.largeurs[1] - 1
	ok = true
	return
}

// On return, e contains the first element of the set attached to i. Returns true in ok if such an element exists. Usage: e, ok := i.FirstE(); for ok { ... e, ok = i.NextE()}
func (i *SetIterator) FirstE() (e int, ok bool) {
	if i.s == nil {
		panic(20)
	}
	if i.s.IsEmpty() {
		ok = false
		return
	}
	i.cour = i.s.elems.suivant
	i.pos = SMin + i.s.elems.largeurs[0]
	i.posE = i.pos
	e = i.posE
	ok = true
	return
}

// On return, e contains the next element of the set attached to i. Returns true in ok if such an element exists. i.FirstE must have been called once before i.NextE. Usage: e, ok := i.FirstE(); for ok { ... e, ok = i.NextE()}
func (i *SetIterator) NextE() (e int, ok bool) {
	if i.s == nil {
		panic(20)
	}
	if i.cour == nil {
		panic(21)
	}
	max := i.pos + i.cour.largeurs[1] - 1
	if i.posE < max {
		i.posE++
		e = i.posE
		ok = true
		return
	}
	i.pos += i.cour.largeurs[1] + i.cour.largeurs[0]
	i.cour = i.cour.suivant
	if i.cour == i.s.elems {
		i.cour = nil
		ok = false
		return
	}
	i.posE = i.pos
	e = i.posE
	ok = true
	return
}
