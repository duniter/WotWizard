/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package leftist
	
// Implements leftist trees, which may be used for priority queues. Cf. Knuth, The art of computer programming, vol. 3, ch. 5.2.3 and exercises 32 & 35.

import (
	
	M "util/misc"
		"sync"

)

// Results of comparison.

type
	Comp int8

const (
	
	First = Comp(-1) // Comes first.
	Equiv = Comp(0) // Same level..
	Last = Comp(+ 1) // Comes after. 

)

type (
	
	Comparer interface {
		Comp (Comparer) Comp
	}
	
	Elem struct { // Element of a leftist tree.
		// Defined recursively by e.dist= e.right.dist+ 1 and e.dist= 1 if e.right= NIL; moreover, e.right.dist<= e.left.dist for all e.
		dist int16
		 // Left son, right son and father.
		left,
		right,
		up *Elem
		// Stored value
		val Comparer
	}
	
	Tree struct { // Leftist tree.
		root *Elem
	}

)

var (
	
	mut = new(sync.RWMutex)

)

// Merge the two leftist trees p and q, and return the result
func merge (p *Elem, q *Elem) *Elem {
	var r *Elem = nil
	for p != nil && q != nil { // Forward ...
		if p.val.Comp(q.val) <= Equiv {
			p.up = r
			r = p
			p = p.right
		} else {
			q.up = r
			r = q
			q = q.right
		}
	}
	var l, d int16
	if q == nil {
		if p == nil {
			d = 0
		} else {
			d = p.dist
			p.up = r
		}
	} else {
		p = q
		d = p.dist
		p.up = r
	}
	for r != nil { // ... and backward
		q = r.left
		if q == nil {
			l = 0
		} else {
			l = q.dist
		}
		if l < d {
			d = l + 1
			r.right = q
			r.left = p
		} else {
			d++
			r.right = p
		}
		r.dist = d
		p = r
		r = p.up
	}
	return p
}

// Inserts the value v in the tree t and returns the *Elem containing it.
func (t *Tree) Insert (v Comparer) *Elem {
	e := new(Elem)
	e.dist = 1
	e.left = nil
	e.right = nil
	e.up = nil
	e.val = v
	mut.Lock()
	t.root = merge(t.root, e)
	mut.Unlock()
	return e
}

// Returns one of the first values of the tree t, and the *Elem containing it in e. Returns nil and nil if the tree is empty.
func (t *Tree) First (e **Elem) Comparer {
	mut.RLock()
	r := t.root
	mut.RUnlock()
	*e = r
	if r == nil {
		return nil
	}
	return r.val
}

// Deletes the *Elem e of the tree t.
func (t *Tree) Erase (e *Elem) {
	M.Assert(e != nil, 20)
	mut.Lock()
	e.left = merge(e.left, e.right)
	p := e.up
	if e.left != nil {
		e.left.up = p
	}
	if p == nil {
		t.root = e.left
	} else {
		M.Assert(e == p.left || e == p.right, 100)
		if e == p.left {
			p.left = e.left
		} else {
			p.right = e.left
		}
		for {
			d := p.dist
			if p.right != nil && (p.left == nil || p.left.dist < p.right.dist) {
				p.left, p.right = p.right, p.left
			}
			var l int16
			if p.right == nil {
				l = 1
			} else {
				l = p.right.dist + 1
			}
			p.dist = l
			p = p.up
			if p == nil || l == d {break}
		}
	}
	mut.Unlock()
}

// Returns true if t is empty.
func (t *Tree) IsEmpty () bool {
	return t.root == nil
}

// Empties the tree t.
func (t *Tree) Empty () {
	mut.Lock()
	t.root = nil
	mut.Unlock()
}

// Creates a new empty tree t. *)
func New () *Tree {
	t := new(Tree)
	t.root = nil
	return t
}

// Returns the value stored in el
func (el *Elem) Val () Comparer {
	return el.val
}
