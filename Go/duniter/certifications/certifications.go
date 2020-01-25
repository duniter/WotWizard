/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package certifications

import (
	
	B	"duniter/blockchain"
	BA	"duniter/basic"
	BT	"util/gbTree"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	S	"util/sort"
		"math"

)

const (
	
	fromName = "CertificationsFrom"
	toName = "CertificationsTo"
	
	fromAction = iota
	toAction

)

type (
	
	action struct {
		what int
		output string
	}
	
	dist = []int
	
	certif struct {
		id string
		created,
		exp int64
	}
	
	certifs = []certif;
	
	certSort struct {
		c certifs
	}

)

func (cs *certSort) Less (c1, c2 int) bool {
	return M.Abs64(cs.c[c1].exp) < M.Abs64(cs.c[c2].exp) || M.Abs64(cs.c[c1].exp) == M.Abs64(cs.c[c2].exp) && BA.CompP(cs.c[c1].id, cs.c[c2].id) == BA.Lt;
}

func (cs *certSort) Swap (c1, c2 int) {
	cs.c[c1], cs.c[c2] = cs.c[c2], cs.c[c1]
}

func moments (d dist) (mean, sDev float64, nb, median int) {
	m := len(d) - 1
	n := 0; nb = 0; nb2 := 0
	for i := 0; i <= m; i++ {
		n += d[i]
		nb += i * d[i]
		nb2 += i * i * d[i]
	}
	if n == 0 {
		mean = 0
		sDev = 0
		nb = 0
		median = 0
	} else {
		mean = float64(nb) / float64(n)
		sDev = math.Sqrt(float64(nb2) / float64(n) - mean * mean)
		median = -1; q := 0;
		for {
			median++
			q += d[median]
			if 2 * q >= n {break}
		}
	}
	return
}

func listMoments (d dist, m *J.Maker) {
	mean, sDev, nb, median := moments(d)
	m.StartObject()
	m.PushInteger(int64(nb))
	m.BuildField("number")
	m.PushFloat(mean)
	m.BuildField("mean")
	m.PushInteger(int64(median))
	m.BuildField("median")
	m.PushFloat(sDev)
	m.BuildField("standard_deviation")
	m.StartArray()
	for i := 0; i < len(d); i++ {
		m.PushInteger(int64(d[i]))
	}
	m.BuildArray()
	m.BuildField("distribution")
	m.BuildObject()
	m.BuildField("statistics")
}

func listFrom () J.Json {
	var (r *BT.IndexReader; pos B.CertPos)
	m := 0
	ok := B.CertNextFrom(true, &pos, &r)
	for ok {
		m = M.Max(m, pos.CertPosLen())
		ok = B.CertNextFrom(false, &pos, &r)
	}
	d := make(dist, m + 1)
	var (cs certSort; ts = S.TS{Sorter: &cs}; block int32)
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	uid, ok := B.IdNextUid(true, &r)
	for ok {
		from, b := B.IdUid(uid); M.Assert(b, 100)
		q := 0
		mk.StartObject()
		mk.PushString(uid)
		mk.BuildField("from")
		if B.CertFrom(from, &pos) {
			cs.c = make(certifs, pos.CertPosLen())
			_, to, okP := pos.CertNextPos()
			for okP {
				block, cs.c[q].exp, b = B.Cert(from, to); M.Assert(b, 101)
				cs.c[q].id, b = B.IdPub(to); M.Assert(b, 102)
				cs.c[q].created, _, _ = B.TimeOf(block)
				q++
				_, to, okP = pos.CertNextPos()
			}
		}
		d[q]++
		mk.StartArray()
		if q > 0 {
			ts.QuickSort(0, q - 1)
			for q := 0; q < len(cs.c); q++ {
				mk.StartObject()
				mk.PushString(cs.c[q].id)
				mk.BuildField("uid")
				mk.PushInteger(cs.c[q].created)
				mk.BuildField("created")
				mk.PushInteger(cs.c[q].exp)
				mk.BuildField("expired")
				mk.BuildObject()
			}
		}
		mk.BuildArray()
		mk.BuildField("to")
		mk.BuildObject()
		uid, ok = B.IdNextUid(false, &r)
	}
	mk.BuildArray()
	mk.BuildField("data")
	listMoments(d, mk)
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now());
	mk.BuildField("now");
	mk.BuildObject()
	return mk.GetJson()
}

func listTo () J.Json {
	var (r *BT.IndexReader; pos B.CertPos)
	m := 0;
	ok := B.CertNextTo(true, &pos, &r)
	for ok {
		m = M.Max(m, pos.CertPosLen())
		ok = B.CertNextTo(false, &pos, &r)
	}
	d := make(dist, m + 1)
	var (cs certSort; ts = S.TS{Sorter: &cs}; block int32)
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	uid, ok := B.IdNextUid(true, &r)
	for ok {
		to, b := B.IdUid(uid); M.Assert(b, 100)
		q := 0;
		mk.StartObject()
		mk.PushString(uid)
		mk.BuildField("to")
		if B.CertTo(to, &pos) {
			cs.c = make(certifs, pos.CertPosLen())
			from, _, okP := pos.CertNextPos()
			for okP {
				block, cs.c[q].exp, b = B.Cert(from, to); M.Assert(b, 101)
				cs.c[q].id, b = B.IdPub(from); M.Assert(b, 102)
				cs.c[q].created, _, _ = B.TimeOf(block)
				q++
				from, _, okP = pos.CertNextPos()
			}
		}
		d[q]++
		mk.StartArray()
		if q > 0 {
			ts.QuickSort(0, q - 1);
			for q := 0; q < len(cs.c); q++ {
				mk.StartObject()
				mk.PushString(cs.c[q].id)
				mk.BuildField("uid")
				mk.PushInteger(cs.c[q].created)
				mk.BuildField("created")
				mk.PushInteger(cs.c[q].exp)
				mk.BuildField("expired")
				mk.BuildObject()
			}
		}
		mk.BuildArray()
		mk.BuildField("from")
		mk.BuildObject()
		uid, ok = B.IdNextUid(false, &r)
	}
	mk.BuildArray()
	mk.BuildField("data")
	listMoments(d, mk)
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now());
	mk.BuildField("now");
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case fromAction:
		s = fromName
	case toAction:
		s = toName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case fromAction:
		G.Json(listFrom(), a.output)
	case toAction:
		G.Json(listTo(), a.output)
	}
}

func from (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: fromAction, output: output}
}

func to (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: toAction, output: output}
}

func init () {
	G.AddAction(fromName, from, G.Arguments{})
	G.AddAction(toName, to, G.Arguments{})
}
