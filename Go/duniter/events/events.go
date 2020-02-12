/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package events

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
	
	memEndsName = "MemEnds";
	missEndsName = "MissEnds";
	certEndsName = "CertEnds";
	
	memA = iota
	missA
	certA

)

type (
	
	action struct {
		what int
		output string
	}
	
	doProc func () Memberships
	
	Membershiper interface {
		Id () string
		Exp () int64
	}
	
	membership struct {
		id string
		exp int64
	}
	
	MembershipCer interface {
		Membershiper
		Member () bool
	}
	
	membershipC struct {
		membership
		member bool
	}
	
	Memberships []Membershiper
	
	memSort struct {
		m Memberships
	}
	
	certif struct {
		from string
		exp int64
	}

)

func (m membership) Id () string {
	return m.id
}

func (m membership) Exp () int64 {
	return m.exp
}

func (m membershipC) Member () bool {
	return m.member
}

func (ms *memSort) Less (m1, m2 int) bool {
	return M.Abs64(ms.m[m1].Exp()) < M.Abs64(ms.m[m2].Exp()) || M.Abs64(ms.m[m1].Exp()) == M.Abs64(ms.m[m2].Exp()) && BA.CompP(ms.m[m1].Id(), ms.m[m2].Id()) == BA.Lt
}

func (ms *memSort) Swap (m1, m2 int) {
	ms.m[m1], ms.m[m2] = ms.m[m2], ms.m[m1]
}

func DoMembershipsEnds () Memberships {
	var ms memSort
	ms.m = make(Memberships, B.IdLenM())
	var ir *BT.IndexReader
	i := 0
	id, ok := B.IdNextUidM(true, &ir)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b && mem, 100)
		ms.m[i] = membership{id: id, exp: exp}
		i++
		id, ok = B.IdNextUidM(false, &ir)
	}
	ts := sort.TS{Sorter: &ms}
	ts.QuickSort(0, len(ms.m) - 1)
	return ms.m
}

func doMissingEnds () Memberships {
	var ir *BT.IndexReader
	i := 0
	id, ok := B.IdNextUid(true, &ir)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 100)
		if !mem && exp != BA.Revoked {
			i++
		}
		id, ok = B.IdNextUid(false, &ir)
	}
	var ms memSort
	ms.m = make(Memberships, i)
	i = 0
	id, ok = B.IdNextUid(true, &ir)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 101)
		if !mem && exp != BA.Revoked {
			ms.m[i] = membership{id: id, exp: exp}
			i++
		}
		id, ok = B.IdNextUid(false, &ir)
	}
	ts := sort.TS{Sorter: &ms}
	ts.QuickSort(0, len(ms.m) - 1)
	return ms.m
}

// Time of calculation reduced.
// Time in O(n), with n = number of certifications.
// Previous time in O(n^2)
// With 25246 certifications, previous time 10s, new time 3s
func DoCertifsEnds () Memberships {
	var ir *BT.IndexReader
	i := 0
	id, ok := B.IdNextUid(true, &ir)
	for ok {
		p, _, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 100)
		bb := exp != BA.Revoked
		if bb {
			var pos B.CertPos
			bb = B.CertToByExp(p, &pos)
			bb = bb && pos.CertPosLen() >= int(B.Pars().SigQty)
		}
		if bb {
			i++
		}
		id, ok = B.IdNextUid(false, &ir)
	}
	var ms memSort
	ms.m = make(Memberships, i)
	n := int(B.Pars().SigQty)
	i = 0
	id, ok = B.IdNextUid(true, &ir)
	for ok {
		p, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 101)
		var pos B.CertPos
		bb := exp != BA.Revoked
		if bb {
			bb = B.CertToByExp(p, &pos)
			bb = bb && pos.CertPosLen() >= int(B.Pars().SigQty)
		}
		if bb {
			var from, to B.Pubkey
			for j := 1; j <= n; j++ {
				from, to, b = pos.CertNextPos(); M.Assert(b, 102)
			}
			_, exp, b = B.Cert(from, to); M.Assert(b, 103)
			ms.m[i] = membershipC{membership: membership{id: id, exp: exp}, member: mem}
			i++
		}
		id, ok = B.IdNextUid(false, &ir)
	}
	ts := sort.TS{Sorter: &ms}
	ts.QuickSort(0, len(ms.m) - 1)
	return ms.m
}

func JsonCommon (mk *J.Maker) {
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
}

func list (do doProc) J.Json {
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	ms := do()
	for _, m := range ms {
		mk.StartObject()
		mk.PushString(m.Id())
		mk.BuildField("id")
		if mc, ok := m.(MembershipCer); ok {
			mk.PushBoolean(mc.Member())
		} else {
			mk.PushBoolean(true)
		}
		mk.BuildField("member")
		mk.PushInteger(m.Exp())
		mk.BuildField("limit")
		mk.BuildObject()
	}
	mk.BuildArray()
	mk.BuildField("limits")
	JsonCommon(mk)
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case memA:
		s = memEndsName
	case missA:
		s = missEndsName
	case certA:
		s = certEndsName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case memA:
		G.Json(list(DoMembershipsEnds), a.output)
	case missA:
		G.Json(list(doMissingEnds), a.output)
	case certA:
		G.Json(list(DoCertifsEnds), a.output)
	}
}

func member (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: memA, output: output}
}

func missing (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: missA, output: output}
}

func certifs (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: certA, output: output}
}

func init () {
	G.AddAction(memEndsName, member, G.Arguments{})
	G.AddAction(missEndsName, missing, G.Arguments{})
	G.AddAction(certEndsName, certifs, G.Arguments{})
}
