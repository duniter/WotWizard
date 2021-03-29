/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

//This package implements balanced and threaded trees (D. E. Knuth, The Art of Computer Programming, vol. 3, Sorting and Searching).
// Trees may be sorted or not; they are guaranteed to be sorted if insertion of values are made through the method 'SearchIns' (with a correct total 'Comparer'); if methods 'Insert', 'Prepend', 'Append' or 'Cat' are used, sorting may not be maintained.
// With sorted trees, methods 'SearchIns', 'Search', 'SearchNext' and 'Delete' can be used, but not with unsorted trees; all other methods may be used with all trees.
package avl

type (
	
	// Result of a comparison.
	Comp int8
	
	// Type of function used by 'WalkThrough'; 'v' is the value of an element of tree, 'p' are parameters.
	DoFunc func (v interface{}, p ...interface{})

)

//Results of comparison.
const (
	
	Lt = Comp(-1) //less than
	Eq = Comp(0) //equal
	Gt = Comp(+1) //greater than

)

const
	maxint int = 0x7fffffff

type (
	
	// Values in elements of trees must verify the interface 'Comparer' if these trees implement the methods 'SearchIns', 'Search', 'SearchNext' or 'Delete'; in this case, if the methods 'Insert', 'Prepend', 'Append' and 'Cat' are not used, trees remain sorted.
	Comparer interface {
		// Compare two values; the comparison must be reflexive, transitive, antisymmetric and total.
		Compare(Comparer) Comp
	}
	
	// Values in elements of trees must verify the interface 'Copier' if these trees should be copied.
	Copier interface {
		// Copy a value.
		Copy() Copier
	}
	
	// Element of a tree.
	Elem struct {
		left, right *Elem
		lTag, rTag  bool
		bal  Comp
		rank int
		cop *Elem
		val interface{}
	}
	
	// A tree.
	Tree struct {
		root *Elem
	}

)

// Is 't' empty?
func (t *Tree) IsEmpty () bool {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {panic("Invalid Tree")}
	return !t.root.lTag
} //IsEmpty

// 'Empty' empties 't'
func (t *Tree) Empty () {
	if t == nil {panic("Nil Tree")}
	e := Elem{lTag: false, rTag: true}
	e.left = &e
	e.right = &e
	t.root = &e
} //Empty

// New creates a new tree.
func New () *Tree {
	t := &Tree{}
	t.Empty()
	return t
} //New

func newElem (data interface{}) *Elem {
	return &Elem{val: data}
} //newElem

// 'Val' returns the value contained in the element 'e'.
func (e *Elem) Val () interface{} {
	if e == nil {panic("Invalid Elem")}
	return e.val
} //Val

// 'SetVal' sets the value 'v' into the element 'e'.
func (e *Elem) SetVal (v interface{}) {
	if e == nil {panic("Invalid Elem")}
	e.val = v
} //SetVal

func copy1 (e *Elem, t bool) {
	if t {
		f := new(Elem)
		f.lTag = e.lTag
		f.rTag = e.rTag
		f.bal = e.bal
		f.rank = e.rank
		f.val = e.val.(Copier).Copy()
		e.cop = f
		copy1(e.left, e.lTag)
		copy1(e.right, e.rTag)
	}
} //copy1

func copy2 (e *Elem, t bool) *Elem {
	f := e.cop
	if t {
		f.left = copy2(e.left, e.lTag)
		f.right = copy2(e.right, e.rTag)
	}
	return f
} //copy2

// 'Copy' returns a copy of 't'; values owned by elements of 't' must verify the interface 'Copier'
func (t *Tree) Copy () *Tree {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {panic("Invalid Tree")}
	u := New()
	t.root.cop = u.root
	copy1(t.root.left, t.root.lTag)
	u.root.left = copy2(t.root.left, t.root.lTag)
	u.root.lTag = t.root.lTag
	return u
} //Copy

func balLI (pp **Elem, h *bool) {
	switch (*pp).bal {
		case Gt: {
			(*pp).bal = Eq
			*h = false
		}
		case Eq: {
			(*pp).bal = Lt
		}
		case Lt: {
			p1 := (*pp).left
			if p1.bal == Lt {
				if p1.rTag {
					(*pp).left = p1.right
				} else {
					(*pp).left = p1
				}
				(*pp).lTag = p1.rTag
				p1.right = *pp
				p1.rTag = true
				(*pp).bal = Eq
				p1.bal = Eq
				(*pp).rank -= p1.rank
				*pp = p1
			} else {
				p2 := p1.right
				if p2.lTag {
					p1.right = p2.left
				} else {
					p1.right = p2
				}
				p1.rTag = p2.lTag
				p2.left = p1
				p2.lTag = true
				if p2.rTag {
					(*pp).left = p2.right
				} else {
					(*pp).left = p2
				}
				(*pp).lTag = p2.rTag
				p2.right = *pp
				p2.rTag = true
				if p2.bal == Lt {
					(*pp).bal = Gt
				} else {
					(*pp).bal = Eq
				}
				if p2.bal == Gt {
					p1.bal = Lt
				} else {
					p1.bal = Eq
				}
				p2.bal = Eq
				p2.rank += p1.rank
				(*pp).rank -= p2.rank
				*pp = p2
			}
			*h = false
		}
	}
} //balLI

func balRI (pp **Elem, h *bool) {
	switch (*pp).bal {
		case Lt: {
			(*pp).bal = Eq
			*h = false
		}
		case Eq: {
			(*pp).bal = Gt
		}
		case Gt: {
			p1 := (*pp).right
			if p1.bal == Gt {
				if p1.lTag {
					(*pp).right = p1.left
				} else {
					(*pp).right = p1
				}
				(*pp).rTag = p1.lTag
				p1.left = *pp
				p1.lTag = true
				(*pp).bal = Eq
				p1.bal = Eq
				p1.rank += (*pp).rank
				*pp = p1
			} else {
				p2 := p1.left
				if p2.rTag {
					p1.left = p2.right
				} else {
					p1.left = p2
				}
				p1.lTag = p2.rTag
				p2.right = p1
				p2.rTag = true
				if p2.lTag {
					(*pp).right = p2.left
				} else {
					(*pp).right = p2
				}
				(*pp).rTag = p2.lTag
				p2.left = *pp
				p2.lTag = true
				if p2.bal == Gt {
					(*pp).bal = Lt
				} else {
					(*pp).bal = Eq
				}
				if p2.bal == Lt {
					p1.bal = Gt
				} else {
					p1.bal = Eq
				}
				p2.bal = Eq
				p1.rank -= p2.rank
				p2.rank += (*pp).rank
				*pp = p2
			}
			*h = false
		}
	}
} //balRI

func balLE (pp **Elem, h *bool) {
	switch (*pp).bal {
		case Lt: {
			(*pp).bal = Eq
		}
		case Eq: {
			(*pp).bal = Gt
			*h = false
		}
		case Gt: {
			p1 := (*pp).right
			if p1.bal == Lt {
				p2 := p1.left
				if p2.rTag {
					p1.left = p2.right
				} else {
					p1.left = p2
				}
				p1.lTag = p2.rTag
				p2.right = p1
				p2.rTag = true
				if p2.lTag {
					(*pp).right = p2.left
				} else {
					(*pp).right = p2
				}
				(*pp).rTag = p2.lTag
				p2.left = *pp
				p2.lTag = true
				if p2.bal == Gt {
					(*pp).bal = Lt
				} else {
					(*pp).bal = Eq
				}
				if p2.bal == Lt {
					p1.bal = Gt
				} else {
					p1.bal = Eq
				}
				p2.bal = Eq
				p1.rank -= p2.rank
				p2.rank += (*pp).rank
				*pp = p2
			} else {
				if p1.lTag {
					(*pp).right = p1.left
				} else {
					(*pp).right = p1
				}
				(*pp).rTag = p1.lTag
				p1.left = *pp
				p1.lTag = true
				if p1.bal == Eq {
					(*pp).bal = Gt
					p1.bal = Lt
					*h = false
				} else {
					(*pp).bal = Eq
					p1.bal = Eq
				}
				p1.rank += (*pp).rank
				*pp = p1
			}
		}
	}
} //balLE

func balRE (pp **Elem, h *bool) {
	switch (*pp).bal {
		case Gt: {
			(*pp).bal = Eq
		}
		case Eq: {
			(*pp).bal = Lt
			*h = false
		}
		case Lt: {
			p1 := (*pp).left
			if p1.bal == Gt {
				p2 := p1.right
				if p2.lTag {
					p1.right = p2.left
				} else {
					p1.right = p2
				}
				p1.rTag = p2.lTag
				p2.left = p1
				p2.lTag = true
				if p2.rTag {
					(*pp).left = p2.right
				} else {
					(*pp).left = p2
				}
				(*pp).lTag = p2.rTag
				p2.right = *pp
				p2.rTag = true
				if p2.bal == Lt {
					(*pp).bal = Gt
				} else {
					(*pp).bal = Eq
				}
				if p2.bal == Gt {
					p1.bal = Lt
				} else {
					p1.bal = Eq
				}
				p2.bal = Eq
				p2.rank += p1.rank
				(*pp).rank -= p2.rank
				*pp = p2
			} else {
				if p1.rTag {
					(*pp).left = p1.right
				} else {
					(*pp).left = p1
				}
				(*pp).lTag = p1.rTag
				p1.right = *pp
				p1.rTag = true
				if p1.bal == Eq {
					(*pp).bal = Lt
					p1.bal = Gt
					*h = false
				} else {
					(*pp).bal = Eq
					p1.bal = Eq
				}
				(*pp).rank -= p1.rank
				*pp = p1
			}
		}
	}
} //balRE

func delL (first bool, pr **Elem, t *bool) (s *Elem, h bool) {
	if !(*pr).rTag {
		s = *pr
		*t = (*pr).lTag
		if *t || first {
			*pr = (*pr).left
		}
		h = true
	} else {
		s, h = delL(false, &(*pr).right, &(*pr).rTag)
		if h {
			balRE(pr, &h)
		}
	}
	return
} //delL

func sIns (q *Elem, l bool, pkey, pp **Elem, rank *int, t, h, found *bool) {
	key := *pkey
	if !*t {
		*found = false
		*h = true
		*t = true
		if l {
			key.left = *pp
			key.right = q
		} else {
			key.right = *pp
			key.left = q
		}
		*pp = key
		(*pp).lTag = false
		(*pp).rTag = false
		(*pp).bal = Eq
		(*pp).rank = 1
		*rank = 1
	} else {
		switch key.val.(Comparer).Compare((*pp).val.(Comparer)) {
			case Lt: {
				sIns(*pp, true, pkey, &(*pp).left, rank, &(*pp).lTag, h, found)
				if !*found {
					(*pp).rank++
				}
				if *h {
					balLI(pp, h)
				}
			}
			case Gt: {
				sIns(*pp, false, pkey, &(*pp).right, rank, &(*pp).rTag, h, found)
				*rank += (*pp).rank
				if *h {
					balRI(pp, h)
				}
			}
			case Eq: {
				*h = false
				*found = true
				*pkey = *pp
				*rank = (*pp).rank
			}
		}
	}
} //sIns

// 'SearchIns' searches the value 'key' in the sorted tree 't'; if 'key' is found, 'found' is true, the element containing this value is returned in 'res' and its rank in the tree in 'rank'; if 'key' is not found, it is inserted into a new element returned in 'res', 'found' is false and 'rank' returns the rank of the new element.
// The rank of the first element in a tree is 1.
func (t *Tree) SearchIns (key Comparer) (res *Elem, found bool, rank int) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if key == nil {
		panic("Nil Key")
	}
	k := newElem(key)
	var h bool
	kp := &k
	sIns(t.root, true, kp, &t.root.left, &rank, &t.root.lTag, &h, &found)
	res = *kp
	return
} //SearchIns

// 'Search' searches the value 'key' in the sorted tree 't'; if 'key' is found, 'found' is true, the element containing this value is returned in 'res' and its rank in the tree in 'rank'; if 'key' is not found, 'res' is nil, 'found' is false and 'rank' is 0.
// The rank of the first element in a tree is 1.
func (t *Tree) Search  (key Comparer) (res *Elem, found bool, rank int) {
	if t == nil {
		panic("Nil Tree")
	}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if key == nil {
		panic("Nil Key")
	}
	tag := t.root.lTag
	val := t.root.left
	rank = 0
	for {
		if !tag {
			res = nil
			found = false
			rank = 0
			return
		} else {
			switch key.Compare(val.val.(Comparer)) {
				case Lt: {
					tag = val.lTag
					val = val.left
				}
				case Eq: {
					rank += val.rank
					res = val
					found = true
					return
				}
				case Gt: {
					rank += val.rank
					tag = val.rTag
					val = val.right
				}
			}
		}
	}
	return
} //Search

// SearchNext searches the value 'key' in the sorted tree 't', or the next value if not found; if 'key' is found, 'found' is true, the element containing this value is returned in 'res' and its rank in the tree in 'rank'; if 'key' is not found, 'found' is false, 'res' returns the element containing the next value and its rank is returned in 'rank'; if 'key' is not found and there is no next value, 'res' is nil and 'rank' is 0.
// The rank of the first element in a tree is 1.
func (t *Tree) SearchNext (key Comparer) (res *Elem, found bool, rank int) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	var valNext *Elem
	if key == nil {
		panic("Nil Key")
	}
	tag := t.root.lTag
	val := t.root.left
	rank = 0
	comp := Gt
	for {
		if !tag {
			if comp == Lt {
				val = valNext
				rank++
			}
			if val == t.root {
				res = nil
				rank = 0
			} else {
				res = val
			}
			found = false
			return
		}
		comp = key.Compare(val.val.(Comparer))
		switch comp {
			case Lt: {
				tag = val.lTag
				valNext = val
				val = val.left
			}
			case Eq: {
				rank += val.rank
				res = val
				found = true
				return
			}
			case Gt: {
				rank += val.rank
				tag = val.rTag
				val = val.right
			}
		}
	}
} //SearchNext

func fixLThread (p, q *Elem) {
	for p.lTag {
		p = p.left
	}
	p.left = q
} //fixLThread

func fixRThread (p, q *Elem) {
	for p.rTag {
		p = p.right
	}
	p.right = q
} //fixRThread

func delD (key Comparer, l bool, pp **Elem, t *bool) (h, found bool) {
	if !*t {
		found = false
		h = false
	} else {
		switch key.Compare((*pp).val.(Comparer)) {
			case Lt: {
				h, found = delD(key, true, &(*pp).left, &(*pp).lTag)
				if found {
					(*pp).rank--
				}
				if h {
					balLE(pp, &h)
				}
			}
			case Gt: {
				h, found = delD(key, false, &(*pp).right, &(*pp).rTag)
				if h {
					balRE(pp, &h)
				}
			}
			case Eq: {
				found = true
				if !(*pp).lTag {
					if (*pp).rTag {
						fixLThread((*pp).right, (*pp).left)
						*pp = (*pp).right
					} else {
						if l {
							*pp = (*pp).left
						} else {
							*pp = (*pp).right
						}
						*t = false
					}
					h = true
				} else if !(*pp).rTag {
					fixRThread((*pp).left, (*pp).right)
					*pp = (*pp).left
					h = true
				} else {
					s := *pp
					*pp, h = delL(true, &s.left, &s.lTag)
					(*pp).left = s.left
					(*pp).lTag = s.lTag
					(*pp).right = s.right
					(*pp).rTag = s.rTag
					(*pp).bal = s.bal
					(*pp).rank = s.rank - 1
					fixLThread((*pp).right, *pp)
					if h {
						balLE(pp, &h)
					}
				}
			}
		}
	}
	return
} //delD

// 'Delete' erases the value 'key', if it exists, in the sorted tree 't'; returns true if 'key' is found, else does nothing.
func (t *Tree) Delete (key Comparer) bool {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if key == nil {
		panic("Nil Key")
	}
	_, found := delD(key, true, &t.root.left, &t.root.lTag)
	return found
} //Delete

func nOE (p *Elem, tag bool) int {
	n := 0
	for tag {
		n += p.rank
		tag = p.rTag
		p = p.right
	}
	return n
} //nOE

// 'NumberOfElems' returns the number of elements in the tree 't'.
func (t *Tree) NumberOfElems () int {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	return nOE(t.root.left, t.root.lTag)
} //NumberOfElems

func ins (pos int, key, q *Elem, l bool, pp **Elem, t *bool) (h bool) {
	if !*t {
		h = true
		*t = true
		if l {
			key.left = *pp
			key.right = q
		} else {
			key.right = *pp
			key.left = q
		}
		*pp = key
		(*pp).lTag = false
		(*pp).rTag = false
		(*pp).bal = Eq
		(*pp).rank = 1
	} else if pos <= (*pp).rank {
		h = ins(pos, key, *pp, true, &(*pp).left, &(*pp).lTag)
		(*pp).rank++
		if h {
			balLI(pp, &h)
		}
	} else {
		pos -= (*pp).rank
		h = ins(pos, key, *pp, false, &(*pp).right, &(*pp).rTag)
		if h {
			balRI(pp, &h)
		}
	}
	return
} //ins

// 'Insert' inserts the value 'key' at the position Min(Max(rank, 1), t.NumberOfElems() + 1) in the tree 't'.
// The rank of the first element in a tree is 1.
func (t *Tree) Insert (key interface{}, rank int) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if key == nil {
		panic("Nil Key")
	}
	ins(rank, newElem(key), t.root, true, &t.root.left, &t.root.lTag)
} //Insert

// 'Prepend' inserts the value 'key' at the first place into the tree 't'.
func (t *Tree) Prepend (key interface{}) {
	if t == nil {panic("Nil Tree")}
	t.Insert(key, 0)
} //Prepend

// 'Append' inserts the value 'key' at the last place into the tree 't'.
func (t *Tree) Append (key interface{}) {
	if t == nil {panic("Nil Tree")}
	t.Insert(key, maxint)
} //Append

func delE (l bool, pp **Elem, t *bool, rank *int) (h bool) {
	if !*t {
		panic("Invalid Rank")
	}
	if *rank < (*pp).rank {
		h = delE(true, &(*pp).left, &(*pp).lTag, rank)
		(*pp).rank--
		if h {
			balLE(pp, &h)
		}
	} else if *rank > (*pp).rank {
		*rank -= (*pp).rank
		h = delE(false, &(*pp).right, &(*pp).rTag, rank)
		if h {
			balRE(pp, &h)
		}
	} else if !(*pp).lTag {
		if (*pp).rTag {
			fixLThread((*pp).right, (*pp).left)
			*pp = (*pp).right
		} else {
			if l {
				*pp = (*pp).left
			} else {
				*pp = (*pp).right
			}
			*t = false
		}
		h = true
	} else if !(*pp).rTag {
		fixRThread((*pp).left, (*pp).right)
		*pp = (*pp).left
		h = true
	} else {
		s := *pp
		*pp, h = delL(true, &s.left, &s.lTag)
		(*pp).left = s.left
		(*pp).lTag = s.lTag
		(*pp).right = s.right
		(*pp).rTag = s.rTag
		(*pp).bal = s.bal
		(*pp).rank = s.rank - 1
		fixLThread((*pp).right, *pp)
		if h {
			balLE(pp, &h)
		}
	}
	return
} //delE

// 'Erase' erases the element at position 'rank' in the tree 't'; 'rank' must verify 1 <= rank <= t.NumberOfElems().
func (t *Tree) Erase (rank int) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	delE(true, &t.root.left, &t.root.lTag, &rank)
} //Erase

// 'Find' finds and returns in 'res' the element at position 'rank' in the tree 't'; if the element does not exist, 'found' is false and 'res' returns nil.
func (t *Tree) Find (rank int) (res *Elem, found bool) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	tag := t.root.lTag
	e := t.root.left
	for {
		switch {
			case !tag: {
				return nil, false
			}
			case rank < e.rank: {
				tag = e.lTag
				e = e.left
			}
			case rank > e.rank: {
				rank -= e.rank
				tag = e.rTag
				e = e.right
			}
			default: {
				return e, true
			}
		}
	}
	return
} //Find

func fixLRoot (root *Elem) {
	fixLThread(root, root)
} //fixLRoot

func fixRRoot (root *Elem) {
	if root.lTag {
		fixRThread(root.left, root)
	} else {
		root.left = root
	}
} //fixRRoot

func height (e *Elem, t bool) int {
	h := 0
	for t {
		h++
		switch e.bal {
			case Lt, Eq: {
				t = e.lTag
				e = e.left
			}
			case Gt: {
				t = e.rTag
				e = e.right
			}
		}
	}
	return h
} //height

func bindLeft (q1, p1 *Elem, t1 bool, h1 int, j *Elem, p2 **Elem, t2 *bool, h2 int, q2 *Elem) (h bool) {
	if !((t1 == (h1 > 0)) && (*t2 == (h2 > 0))) {
		panic("incorrect height")
	}
	if h2 > h1+1 {
		if (*p2).bal == Gt {
			h2--
		}
		h = bindLeft(q1, p1, t1, h1, j, &(*p2).left, &(*p2).lTag, h2-1, *p2)
		(*p2).rank += j.rank
		if h {
			balLI(p2, &h)
		}
	} else {
		h = true
		if *t2 {
			j.right = *p2
		} else {
			j.right = q2
		}
		j.rTag = *t2
		if t1 {
			j.left = p1
		} else {
			j.left = q1
		}
		j.lTag = t1
		if h1 == h2 {
			j.bal = Eq
		} else {
			j.bal = Gt
		}
		j.rank = nOE(p1, t1) + 1
		if t1 {
			fixRThread(p1, j)
		}
		if *t2 {
			fixLThread(*p2, j)
		}
		*p2 = j
		*t2 = true
	}
	return
} //bindLeft

func bindRight (q1 *Elem, p1 **Elem, t1 *bool, h1 int, j *Elem, p2 *Elem, t2 bool, h2 int, q2 *Elem) (h bool) {
	if !((*t1 == (h1 > 0)) && (t2 == (h2 > 0))) {
		panic("incorrect height")
	}
	if h1 > h2+1 {
		if (*p1).bal == Lt {
			h1--
		}
		h = bindRight(*p1, &(*p1).right, &(*p1).rTag, h1-1, j, p2, t2, h2, q2)
		if h {
			balRI(p1, &h)
		}
	} else {
		h = true
		if *t1 {
			j.left = *p1
		} else {
			j.left = q1
		}
		j.lTag = *t1
		if t2 {
			j.right = p2
		} else {
			j.right = q2
		}
		j.rTag = t2
		if h1 == h2 {
			j.bal = Eq
		} else {
			j.bal = Lt
		}
		j.rank = nOE(*p1, *t1) + 1
		if *t1 {
			fixRThread(*p1, j)
		}
		if t2 {
			fixLThread(p2, j)
		}
		*p1 = j
		*t1 = true
	}
	return
} //bindRight

func eraseLeft (p **Elem, t *bool) (j *Elem, h bool) {
	if (*p).lTag {
		j, h = eraseLeft(&(*p).left, &(*p).lTag)
		(*p).rank--
		if h {
			balLE(p, &h)
		}
	} else {
		j = *p
		if (*p).rTag {
			fixLThread((*p).right, (*p).left)
			*p = (*p).right
		} else {
			*p = (*p).left
			*t = false
		}
		h = true
	}
	return
} //eraseLeft

func eraseRight (p **Elem, t *bool) (j *Elem, h bool) {
	if (*p).rTag {
		j, h = eraseRight(&(*p).right, &(*p).rTag)
		if h {
			balRE(p, &h)
		}
	} else {
		j = *p
		if (*p).lTag {
			fixRThread((*p).left, (*p).right)
			*p = (*p).left
		} else {
			*p = (*p).right
			*t = false
		}
		h = true
	}
	return
} //eraseRight

// 'Cat' concatenates the trees 't1' and 't2'; it returns the result in 't1', and 't2' is no more valid.
func (t1 *Tree) Cat (t2 *Tree) {
	if t1 == nil {panic("Nil Tree")}
	if t1.root == nil {
		panic("Invalid first Tree")
	}
	if t2 == nil {panic("Nil Tree")}
	if t2.root == nil {
		panic("Invalid second Tree")
	}
	if t1 == t2 {
		panic("Same Tree")
	}
	if t2.root.lTag {
		if !t1.root.lTag {
			t1.root = t2.root
		} else {
			h1 := height(t1.root.left, t1.root.lTag)
			h2 := height(t2.root.left, t2.root.lTag)
			if h1 < h2 {
				j, h := eraseRight(&t1.root.left, &t1.root.lTag)
				if h {
					h1--
				}
				h = bindLeft(t2.root, t1.root.left, t1.root.lTag, h1, j, &t2.root.left, &t2.root.lTag, h2, t2.root)
				t1.root = t2.root
				fixLRoot(t1.root)
			} else {
				j, h := eraseLeft(&t2.root.left, &t2.root.lTag)
				if h {
					h2--
				}
				h = bindRight(t1.root, &t1.root.left, &t1.root.lTag, h1, j, t2.root.left, t2.root.lTag, h2, t1.root)
				fixRRoot(t1.root)
			}
		}
	}
	t2.root = nil
} //Cat

func doSplit (t1 *Tree, after int, p *Elem, t bool, t2 *Tree) (e1, e2 *Elem, tag1, tag2 bool, h1, h2, hh int) {
	if after < p.rank {
		e1, e2, tag1, tag2, h1, h2, hh = doSplit(t1, after, p.left, p.lTag, t2)
		hh++
		if p.bal == Gt {
			hh++
		}
		he := hh - 1
		if p.bal == Lt {
			he--
		}
		s := p.right
		b := p.rTag
		var h bool
		if h2 < he {
			h = bindLeft(t2.root, e2, tag2, h2, p, &s, &b, he, t2.root)
			e2 = s
			tag2 = b
			h2 = he
		} else {
			h = bindRight(t2.root, &e2, &tag2, h2, p, s, b, he, t2.root)
		}
		if h {
			h2++
		}
	} else if after > p.rank {
		after -= p.rank
		e1, e2, tag1, tag2, h1, h2, hh = doSplit(t1, after, p.right, p.rTag, t2)
		hh++
		if p.bal == Lt {
			hh++
		}
		he := hh - 1
		if p.bal == Gt {
			he--
		}
		s := p.left
		b := p.lTag
		var h bool
		if he < h1 {
			h = bindLeft(t1.root, s, b, he, p, &e1, &tag1, h1, t1.root)
		} else {
			h = bindRight(t1.root, &s, &b, he, p, e1, tag1, h1, t1.root)
			e1 = s
			tag1 = b
			h1 = he
		}
		if h {
			h1++
		}
	} else {
		hh = height(p, t)
		h1 = hh - 1
		h2 = h1
		switch p.bal {
			case Lt: {
				h2--
			}
			case Eq: {
			}
			case Gt: {
				h1--
			}
		}
		e1 = p.left
		tag1 = p.lTag
		e2 = p.right
		tag2 = p.rTag
		if ins(maxint, p, t1.root, true, &e1, &tag1) {
			h1++
		}
	}
	return
} //doSplit

// Splits the tree 't1' after the element of rank 'after'; the result is returned in 't1' and 't2'.
func (t1 *Tree) Split (after int) (t2 *Tree) {
	if t1 == nil {panic("Nil Tree")}
	if t1.root == nil {panic("Invalid Tree")}
	t2 = New()
	if after < t1.NumberOfElems() {
		if after <= 0 {
			t2.root = t1.root
			t1.Empty()
		} else {
			t1.root.left, t2.root.left, t1.root.lTag, t2.root.lTag, _, _, _ = doSplit(t1, after, t1.root.left, t1.root.lTag, t2)
			fixLRoot(t1.root)
			fixRRoot(t1.root)
			fixLRoot(t2.root)
			fixRRoot(t2.root)
		}
	}
	return
} //Split

func ahead (e *Elem, t bool, do DoFunc, p ...interface{}) {
	if t {
		ahead(e.left, e.lTag, do, p...)
		do(e.val, p...)
		ahead(e.right, e.rTag, do, p...)
	}
} //ahead

// 'WalkThrough' traverses the tree 't' in inorder and calls the function 'do', with 'p' as parameters, on the value of each element.
func (t *Tree) WalkThrough (do DoFunc, p ...interface{}) {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	ahead(t.root.left, t.root.lTag, do, p...)
} //WalkThrough

// 'Next' returns the element following 'e' in the tree 't'; if 'e' is nil, the first element is returned; if 'e' is the last element of the tree, nil is returned.
func (t *Tree) Next (e *Elem) *Elem {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if e == nil {
		e = t.root
	}
	tag := e.rTag
	e = e.right
	if tag {
		for e.lTag {
			e = e.left
		}
	}
	if e == t.root {
		return nil
	}
	return e
} //Next

// 'Previous' returns the element preceding 'e' in the tree 't'; if 'e' is nil, the last element is returned; if 'e' is the first element of the tree, nil is returned.
func (t *Tree) Previous (e *Elem) *Elem {
	if t == nil {panic("Nil Tree")}
	if t.root == nil {
		panic("Invalid Tree")
	}
	if e == nil {
		e = t.root
	}
	tag := e.lTag
	e = e.left
	if tag {
		for e.rTag {
			e = e.right
		}
	}
	if e == t.root {
		return nil
	}
	return e
} //Previous
