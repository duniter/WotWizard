/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package netStressD
	
	// Calculate the stress centrality with Ulrik Brandes's algorithm, slightly modified to deal with the fact that only paths between members have to be considered

import (
	S	"util/sets2"
	M	"util/misc"
		"util/sort"
)

type (
	
	Comp int8

)

const (
	
	// Results of comparison.
	Lt = Comp(-1) // less than
	Eq = Comp(0) //equal
	Gt = Comp(+1) //greater than

)

type (
	
	Neter interface {
		Number () int
		Enumerate (first bool) (node Node, extremity, ok bool)
	}
	
	Net struct {
		Neter
		nbNodes int
		nodes *nodesSort
		extremities S.Set
		links linksT
	}
	
	Node interface {
		Compare (Node) Comp
		FromTo (first bool) (follow Node, ok bool)
	}
	
	nodesT []Node
	
	nodesSort struct {
		t nodesT
		ex []bool
	}
	
	linksT []S.Set
	
	setCT []int64
	
	Centrals struct {
		nbNodes int
		nodes nodesT
		cT setCT
		pos int
	}
	
	queue struct {
		end *elem
	}
	
	stack struct {
		end *elem
	}
	
	elem struct {
		next *elem
		val int
	}

)

func NewNet (n Neter) *Net {
	return &Net{Neter: n}
}

func (s *stack) init () {
	s.end = nil
}

func (s *stack) isEmpty () bool {
	return s.end == nil
}

func (s *stack) push (val int) {
	s.end = &elem{val: val, next: s.end}
}

func (s *stack) pop () (val int) {
	M.Assert(s.end != nil, 20)
	val = s.end.val
	s.end = s.end.next
	return
}

func (q *queue) init () {
	q.end = nil
}

func (q *queue) isEmpty () bool {
	return q.end == nil
}

func (q *queue) put (val int) {
	e := &elem{val: val}
	if q.end == nil {
		q.end = e
		e.next = e
	} else {
		e.next = q.end.next
		q.end.next = e
		q.end = e
	}
}

func (q *queue) get () (val int) {
	M.Assert(q.end != nil, 20)
	e := q.end.next
	q.end.next = e.next
	if q.end == e {
		q.end = nil
	}
	val = e.val
	return
}

func (net Net) ExtremitiesNb () int {
	M.Assert(net.extremities != nil, 20)
	return net.extremities.NbElems()
}

func (s *nodesSort) Less (p1, p2 int) bool {
	return s.t[p1].Compare(s.t[p2]) == Lt
}

func (s *nodesSort) Swap (p1, p2 int) {
	s.t[p1], s.t[p2] = s.t[p2], s.t[p1]
	s.ex[p1], s.ex[p2] = s.ex[p2], s.ex[p1]
}

func (ct *Centrals) Walk (first bool) (node Node, c int64, ok bool) {
	if first {
		ct.pos = 0
	} else if ct.pos < ct.nbNodes {
		ct.pos++
	}
	ok = ct.pos < ct.nbNodes
	if !ok {
		return
	}
	node = ct.nodes[ct.pos]
	c = ct.cT[ct.pos]
	return
}

func (net *Net) NbNodes () int {
	return net.nbNodes
}

func (net *Net) findNode (node Node) (n int, ok bool) {
	tf := sort.TF{Finder: net.nodes}
	n = net.nbNodes
	net.nodes.t[n] = node
	tf.BinSearch(0, n - 1, &n)
	ok = n < net.nbNodes
	return
}

func (net *Net) Update() {
	net.nbNodes = net.Number()
	net.nodes = new(nodesSort)
	net.nodes.t = make(nodesT, net.nbNodes + 1)
	net.nodes.ex = make([]bool, net.nbNodes)
	i := 0
	node, extremity, ok := net.Enumerate(true)
	for ok {
		/**/
		M.Assert(node != nil, 100)
		/**/
		net.nodes.t[i] = node
		net.nodes.ex[i] = extremity
		i++
		node, extremity, ok = net.Enumerate(false)
	}
	M.Assert(i == net.nbNodes)
	ts := sort.TS{Sorter: net.nodes}
	ts.QuickSort(0, net.nbNodes - 1)
	net.extremities = S.NewSet()
	for i, x := range net.nodes.ex {
		if x {
			net.extremities.Incl(i)
		}
	}
	net.nodes.ex = nil
	
	net.links = make(linksT, net.nbNodes)
	for i := 0; i < net.nbNodes; i++ {
		net.links[i] = S.NewSet()
		node, ok := net.nodes.t[i].FromTo(true)
		for ok {
			/**/
			M.Assert(node != nil, 101)
			/**/
			n, b := net.findNode(node); M.Assert(b)
			net.links[i].Incl(n)
			node, ok = net.nodes.t[i].FromTo(false)
		}
	}
}

func (net *Net) newC () (set setCT) {
	set = make(setCT, net.nbNodes)
	for i := 0; i < net.nbNodes; i++ {
		set[i] = 0
	}
	return
}

// Modified Ulrik Brandes's algorithm
func (net *Net) stressD (maxStep int) setCT {
	var (st stack; q queue)
	p := make([]stack, net.nbNodes)
	sig := make([]int, net.nbNodes)
	d := make([]int, net.nbNodes)
	delt := make([]int, net.nbNodes)
	cS := net.newC()
	for v := 0; v < net.nbNodes; v++ {
		cS[v] = 0
	}
	// Only extremities can be sources of paths
	it1 := net.extremities.Attach()
	s, ok1 := it1.FirstE()
	for ok1 {
		st.init()
		for v := 0; v < net.nbNodes; v++ {
			p[v].init()
			sig[v] = 0; d[v] = - 1
			delt[v] = 0
		}
		sig[s] = 1; d[s] = 0
		q.init()
		q.put(s)
		for !q.isEmpty() {
			v := q.get()
			st.push(v)
			if d[v] <= maxStep {
				it2 := net.links[v].Attach()
				w, ok2 := it2.FirstE()
				for ok2 {
					if d[w] < 0 {
						q.put(w)
						d[w] = d[v] + 1
					}
					if d[w] == d[v] + 1 {
						sig[w] += sig[v]
						p[w].push(v)
					}
					w, ok2 = it2.NextE()
				}
			}
		}
		for !st.isEmpty() {
			w := st.pop()
			ext := net.extremities.In(w)
			for !p[w].isEmpty() {
				v := p[w].pop()
				if ext {
					delt[v] = delt[v] + sig[v] * (1 + delt[w] / sig[w])
				} else {
					 // Don't increase by 1 if w is not an extremity, since no path ends at w
					delt[v] = delt[v] + sig[v] * (delt[w] / sig[w])
				}
			}
			if w != s {
				cS[w] += int64(delt[w])
			}
		}
		s, ok1 = it1.NextE()
	}
	return cS
}

func (net *Net) Centralities (maxStep int) *Centrals {
	return &Centrals{nbNodes: net.nbNodes, nodes: net.nodes.t, cT: net.stressD(maxStep)}
}
