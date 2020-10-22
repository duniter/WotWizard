/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package centralities
	
// Calculate the stress centrality with Ulrik Brandes' algorithm, slightly modified to deal with the fact that only paths between members have to be considered, and limited to B.pars.stepMax distance.

import (
	
	B	"duniter/blockchain"
	BA	"duniter/basic"
	M	"util/misc"
	N	"util/netStressD"
		"math"
		"util/sort"

)

const (
	
	updateName = "Centralities"
	
	oneUidName = "Uid"
	
	allAction = iota
	oneAction

)

type (
	
	netT struct {
		ir *B.Position
	}
	
	nodeT struct {
		p B.Pubkey
		pos B.CertPos
	}
	
	one struct {
		p B.Pubkey
		c float64
	}
	
	ones []one
	
	onesSort struct {
		os ones
	}

	central struct {
		id string
		c float64
	}
	
	centrals []central
	
	centralSort struct {
		c centrals
	}
	
	centralSortId struct {
		*centralSort
	}

)

var (
	
	mustUpdate,
	askAllOnes chan<- bool
	getAllOnes <-chan *onesSort

)

func (s *onesSort) Swap (p1, p2 int) {
	s.os[p1], s.os[p2] = s.os[p2], s.os[p1]
}

func (s *onesSort) Less (p1, p2 int) bool {
	return s.os[p1].p < s.os[p2].p
}

func (s *centralSort) Swap (p1, p2 int) {
	s.c[p1], s.c[p2] = s.c[p2], s.c[p1]
}

func (s *centralSort) Less (p1, p2 int) bool {
	return s.c[p1].c > s.c[p2].c || s.c[p1].c == s.c[p2].c && BA.CompP(s.c[p1].id, s.c[p2].id) == BA.Lt
}

func (s *centralSortId) Less (p1, p2 int) bool {
	return BA.CompP(s.c[p1].id, s.c[p2].id) == BA.Lt
}

func newNode (p B.Pubkey) *nodeT {
	return &nodeT{p: p}
}

func (*netT) Number () int {
	return B.IdLen()
}

func (net *netT) Enumerate (first bool) (node N.Node, member bool, ok bool) {
	var (id string; p B.Pubkey)
	if id, ok = B.IdNextUid(first, &net.ir); ok {
		var b bool
		p, member, _, _, _, _, b = B.IdUidComplete(id); M.Assert(b, 100)
		node = newNode(p)
	}
	return
}

func (n1 *nodeT) Compare (n2 N.Node) N.Comp {
	nn2 := n2.(*nodeT)
	if n1.p < nn2.p {
		return N.Lt
	}
	if n1.p > nn2.p {
		return N.Gt
	}
	return N.Eq
}

func (n *nodeT) FromTo (first bool) (follow N.Node, ok bool) {
	// Counterintuitive : the result doesn't depend on the direction of arrows
	ok = true
	if first {
		ok = B.CertFrom(n.p, &n.pos)
	}
	var to B.Pubkey
	if ok {
		_, to, ok = n.pos.CertNextPos()
	}
	if ok {
		follow = newNode(to)
	}
	return
}

func allOnesP () *onesSort {
	askAllOnes <- true
	return <-getAllOnes
}

func doCount () (centers, centersId centrals) {
	allOnes := allOnesP()
	if allOnes == nil {
		return nil, nil
	}
	var l int
	l = len(allOnes.os) - 1;
	centers = make(centrals, l)
	centersId = make(centrals, l)
	for i := 0; i < l; i++ {
		var (c central; b bool)
		c.id, b = B.IdPub(allOnes.os[i].p); M.Assert(b, 100)
		c.c = allOnes.os[i].c
		centers[i] = c
		centersId[i] = c
	}
	var (
		s = &centralSort{c: centers}
		sId = &centralSortId{centralSort: &centralSort{c: centersId}}
		ts sort.TS
	)
	s.c = centers
	ts.Sorter = s
	ts.QuickSort(0, l - 1)
	sId.c = centersId
	ts.Sorter = sId
	ts.QuickSort(0, l - 1)
	return
}

func doCountOne (p B.Pubkey) float64 {
	allOnes := allOnesP()
	if allOnes == nil {
		return 0.
	}
	l := len(allOnes.os) - 1
	if l == 0 {
		return 0.
	}
	allOnes.os[l].p = p
	var tf sort.TF
	tf.Finder = allOnes
	tf.BinSearch(0, l - 1, &l);
	M.Assert(l < len(allOnes.os) - 1, 100)
	return allOnes.os[l].c
}

func CountOne (p B.Pubkey) float64 {
	return doCountOne(p)
}

func Count () (centers, centersId centrals) {
	centers, centersId = doCount()
	return 
}

func countAllOnes (net *N.Net) *onesSort {
	cT := net.Centralities(int(B.Pars().StepMax))
	l := net.NbNodes()
	max := 0.
	allOnes := new(onesSort)
	allOnes.os = make(ones, l + 1)
	i := 0
	n, cV, ok := cT.Walk(true)
	for ok {
		node := n.(*nodeT)
		allOnes.os[i].p = node.p
		allOnes.os[i].c = math.Log(float64(1 + cV))
		max = M.MaxF64(max, allOnes.os[i].c)
		i++
		n, cV, ok = cT.Walk(false)
	}
	M.Assert(i == l, 60)
	for i := 0; i < l; i++ {
		allOnes.os[i] .c = allOnes.os[i].c / max
	}
	var ts = sort.TS{Sorter: allOnes}
	ts.QuickSort(0, l - 1)
	return allOnes
}

func update () *onesSort {
	net := N.NewNet(new(netT))
	net.Update()
	return countAllOnes(net)
}

func updateManager (mustUpdt, askAllOnes <-chan bool, getAllOnes chan<- *onesSort) {
	var (mustUpdate = true; allOnes *onesSort = nil)
	for {
		select {
		case <-mustUpdt:
			mustUpdate = true
		case <-askAllOnes:
			if mustUpdate {
				allOnes = update()
				mustUpdate = false
			}
			getAllOnes <- allOnes
		}
	}
}

func recordUpdate (... interface{}) {
	mustUpdate <- true
}

func init () {
	mustU := make(chan bool)
	askAll := make(chan bool)
	getAll := make(chan *onesSort)
	mustUpdate = mustU
	askAllOnes = askAll
	getAllOnes = getAll
	go updateManager(mustU, askAll, getAll)
	B.AddUpdateProc(updateName, recordUpdate)
}
