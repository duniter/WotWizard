/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package events

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"
		"util/sort"

)

type (
	
	Membershiper interface {
		Id () string
		Exp () int64
	}
	
	membership struct {
		id string
		exp int64
	}
	
	memberships []membership
	
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
		m memberships
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

func filter (ms memSort, start, end int64) memberships {
	n := len(ms.m) - 1
	ts := sort.TS{Sorter: &ms}
	ts.QuickSort(0, n - 1)
	tf := sort.TF{Finder: &ms}
	ms.m[n] = membership{exp: start}
	beg := n
	tf.BinSearchNext(0, n - 1, &beg)
	ms.m[n] = membership{exp: end}
	stop := n
	tf.BinSearchNext(0, n - 1, &stop)
	return ms.m[beg:stop]
}

func doMembershipsEnds (start, end int64) memberships {
	var ms memSort
	n := B.IdLenM()
	ms.m = make(memberships, n + 1)
	var pst *B.Position
	i := 0
	id, ok := B.IdNextUidM(true, &pst)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b && mem, 100)
		ms.m[i] = membership{id: id, exp: exp}
		i++
		id, ok = B.IdNextUidM(false, &pst)
	}
	return filter(ms, start, end)
}

func doMissingEnds (start, end int64) memberships {
	var pst *B.Position
	n := 0
	id, ok := B.IdNextUid(true, &pst)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 100)
		if !mem && exp != BA.Revoked {
			n++
		}
		id, ok = B.IdNextUid(false, &pst)
	}
	var ms memSort
	ms.m = make(memberships, n + 1)
	i := 0
	id, ok = B.IdNextUid(true, &pst)
	for ok {
		_, mem, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 101)
		if !mem && exp != BA.Revoked {
			ms.m[i] = membership{id: id, exp: exp}
			i++
		}
		id, ok = B.IdNextUid(false, &pst)
	}
	return filter(ms, start, end)
}

// Time of calculation reduced.
// Time in O(n), with n = number of certifications.
// Previous time in O(n^2)
// With 25246 certifications, previous time 10s, new time 3s
func DoCertifsEnds (start, end int64, missingIncluded bool) memberships {
	var pst *B.Position
	n := int(B.Pars().SigQty)
	m := 0
	id, ok := B.IdNextUid(true, &pst)
	for ok {
		p, member, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 100)
		bb := missingIncluded && exp != BA.Revoked || !missingIncluded && member
		if bb {
			var pos B.CertPos
			bb = B.CertToByExp(p, &pos)
			bb = bb && pos.CertPosLen() >= n
		}
		if bb {
			m++
		}
		id, ok = B.IdNextUid(false, &pst)
	}
	var ms memSort
	ms.m = make(memberships, m + 1)
	i := 0
	id, ok = B.IdNextUid(true, &pst)
	for ok {
		p, member, _, _, _, exp, b := B.IdUidComplete(id); M.Assert(b, 101)
		var pos B.CertPos
		bb := missingIncluded && exp != BA.Revoked || !missingIncluded && member
		if bb {
			bb = B.CertToByExp(p, &pos)
			bb = bb && pos.CertPosLen() >= n
		}
		if bb {
			var from, to B.Pubkey
			for j := 1; j <= n; j++ {
				from, to, b = pos.CertNextPos(); M.Assert(b, 102)
			}
			_, exp, b = B.Cert(from, to); M.Assert(b, 103)
			ms.m[i] = membership{id: id, exp: exp}
			i++
		}
		id, ok = B.IdNextUid(false, &pst)
	}
	return filter(ms, start, end)
}

func getStartEnd (as *A.Tree) (start, end int64) {
	var v G.Value
	if G.GetValue(as, "startFromNow", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			start = v.Int
		case *G.NullValue:
			start = 0
		default:
			M.Halt(v, 100)
		}
	} else {
		start = 0
	}
	start += B.Now()
	if G.GetValue(as, "period", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			end = v.Int
		case *G.NullValue:
			end = M.MaxInt64
		default:
			M.Halt(v, 100)
		}
	} else {
		end = M.MaxInt64
	}
	if end < M.MaxInt64 - start {
		end += start
	} else {
		end = M.MaxInt64
	}
	return
} //getStartEnd

var (
	
	memStream = GQ.CreateStream("memEnds")
	missStream = GQ.CreateStream("missEnds")
	certsStream = GQ.CreateStream("certEnds")

)

func memStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return memStream
} //memStreamResolver

func missStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return missStream
} //missStreamResolver

func certStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return certsStream
} //certStreamResolver

func memEndsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	start, end := getStartEnd(argumentValues)
	l := G.NewListValue()
	if start >= end {
		return l
	}
	ms := doMembershipsEnds(start, end)
	for _, m := range ms {
		_, _, h, _, _, _, b := B.IdUidComplete(m.id); M.Assert(b, 100)
		l.Append(GQ.Wrap(h))
	}
	return l
} //memEndsR

func missEndsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	start, end := getStartEnd(argumentValues)
	l := G.NewListValue()
	if start >= end {
		return l
	}
	ms := doMissingEnds(start, end)
	for _, m := range ms {
		_, _, h, _, _, _, b := B.IdUidComplete(m.id); M.Assert(b, 100)
		l.Append(GQ.Wrap(h))
	}
	return l
} //missEndsR

func certEndsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	start, end := getStartEnd(argumentValues)
	l := G.NewListValue()
	if start >= end {
		return l
	}
	var (v G.Value; missingIncluded bool)
	if G.GetValue(argumentValues, "missingIncluded", &v) {
		switch v := v.(type) {
		case *G.BooleanValue:
			missingIncluded = v.Boolean
		default:
			M.Halt(v, 100)
		}
	} else {
		missingIncluded = true
	}
	ms := DoCertifsEnds(start, end, missingIncluded)
	for _, m := range ms {
		_, _, h, _, _, _, b := B.IdUidComplete(m.id); M.Assert(b, 100)
		l.Append(GQ.Wrap(h))
	}
	return l
} //certEndsR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "memEnds", memEndsR)
	ts.FixFieldResolver("Query", "missEnds", missEndsR)
	ts.FixFieldResolver("Query", "certEnds", certEndsR)
	ts.FixFieldResolver("Subscription", "memEnds", memEndsR)
	ts.FixFieldResolver("Subscription", "missEnds", missEndsR)
	ts.FixFieldResolver("Subscription", "certEnds", certEndsR)
} //fixFieldResolvers

func fixStreamResolvers (ts G.TypeSystem) {
	ts.FixStreamResolver("memEnds", memStreamResolver)
	ts.FixStreamResolver("missEnds", missStreamResolver)
	ts.FixStreamResolver("certEnds", certStreamResolver)
}

func init () {
	ts := GQ.TS()
	fixFieldResolvers(ts)
	fixStreamResolvers(ts)
}
