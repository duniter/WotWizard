/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package qualities

// Calculate qualities with UtilSets; numbering of members is done by array; fast

import (
	
	B	"duniter/blockchain"
	BA	"duniter/basic"
	BT	"util/gbTree"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
		"util/sort"

)

const (
	
	distancesName = "QualitiesDist"
	qualitiesName  = "QualitiesQual"
	
	distA = iota
	qualA
	
)

type (
	
	action struct {
		what int
		output string
	}

	propOfSentries struct {
		id string
		prop float64
	}
	
	propsT []*propOfSentries
	
	propsSort struct {
		t propsT
	}
	
	propsSortId struct {
		propsSort
	}

)

func (s *propsSort) Swap (i, j int) {
	s.t[i], s.t[j] = s.t[j], s.t[i]
}

func (s *propsSort) Less (p1, p2 int) bool {
	return s.t[p1].prop > s.t[p2].prop || s.t[p1].prop == s.t[p2].prop && BA.CompP(s.t[p1].id, s.t[p2].id) == BA.Lt
}

func (s *propsSortId) Less (p1, p2 int) bool {
	return BA.CompP(s.t[p1].id, s.t[p2].id) == BA.Lt
}

func percentOfSentries (dist bool, pubkey B.Pubkey) float64 {
	if dist {
		var pos B.CertPos
		ok := B.CertTo(pubkey, &pos)
		n := 1
		if ok {
			n = pos.CertPosLen() + 1
		}
		pubs := make([]B.Pubkey, n)
		pubs[0] = pubkey
		for i := 1; i < n; i++ {
			pubs[i], _, ok = pos.CertNextPos(); M.Assert(ok)
		}
		return B.PercentOfSentries(pubs)
	} else {
		pub := make([]B.Pubkey, 1)
		pub[0] = pubkey
		return B.PercentOfSentries(pub)
	}
}

func count (dist bool) (props, propsId propsT) {
	n := B.IdLenM()
	if n == 0 {
		props = nil
		propsId = nil
	} else {
		props = make(propsT, n)
		var ir *BT.IndexReader
		i := 0
		p, ok := B.IdNextPubkeyM(true, &ir)
		for ok {
			c := new(propOfSentries)
			var b bool
			c.id, b = B.IdPub(p); M.Assert(b, 100)
			c.prop = percentOfSentries(dist, p)
			props[i] = c
			i++
			p, ok = B.IdNextPubkeyM(false, &ir)
		}
		ts := sort.TS{Sorter: &propsSort{t: props}}
		ts.QuickSort(0, n - 1)
		propsId = make(propsT, n)
		for i := 0; i < n; i++ {
			propsId[i] = props[i]
		}
		ts = sort.TS{Sorter: &propsSortId{propsSort:propsSort{t: propsId}}}
		ts.QuickSort(0, n - 1)
	}
	return
}

func list (dist bool) J.Json {
	props, propsId := count(dist)
	mk := J.NewMaker();
	mk.StartObject()
	mk.StartArray()
	if props != nil {
		for i := 0; i < len(props); i++ {
			mk.StartObject()
			mk.PushString(props[i].id)
			mk.BuildField("id")
			mk.PushFloat(props[i].prop)
			mk.BuildField("prop")
			mk.BuildObject()
		}
	}
	mk.BuildArray()
	mk.BuildField("values")
	mk.StartArray()
	if propsId != nil {
		for i := 0; i < len(propsId); i++ {
			mk.StartObject()
			mk.PushString(propsId[i].id)
			mk.BuildField("id")
			mk.PushFloat(propsId[i].prop)
			mk.BuildField("prop")
			mk.BuildObject()
		}
	}
	mk.BuildArray()
	mk.BuildField("values_byId")
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mt := B.Now()
	mk.PushInteger(mt)
	mk.BuildField("now")
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case distA:
		s = distancesName
	case qualA:
		s = qualitiesName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case distA:
		G.Json(list(true), a.output)
	case qualA:
		G.Json(list(false), a.output)
	}
}

func distances (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: distA, output: output}
}

func qualities (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: qualA, output: output}
}

func init () {
	G.AddAction(distancesName, distances, G.Arguments{})
	G.AddAction(qualitiesName, qualities, G.Arguments{})
}
