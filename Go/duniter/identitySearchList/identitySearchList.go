/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identitySearchList

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	C	"duniter/centralities"
	G	"duniter/gqlReceiver"
	H	"duniter/history"
	J	"util/json"
	M	"util/misc"
	S	"duniter/sandbox"
		"strings"
		"util/sort"

)

const (
	
	findName = "IdSearchFind";
	fixName = "IdSearchFix";
	
	findHintName = "Hint"
	findOldName = "OldMembers"
	findMemName = "Members"
	findFutName = "FutureMembers"
	
	fixHashName = "Hash"
	fixDistName = "Distance"
	fixQualName = "Quality"
	fixCentrName = "Centrality"
	
	findA = iota
	fixA

)

type (
	
	action struct {
		what int
		output,
		hint string
		hash B.Hash
		b1, b2, b3 bool
	}
	
	idET struct {
		uid string
		hash B.Hash
		limit int64
		future,
		active bool
	}
	
	certifiersT []B.Pubkey
	
	expSort struct {
		exp []int64
	}

)

func (e *expSort) Less (i, j int) bool {
	return e.exp[i] < e.exp[j]
}

func (e *expSort) Swap (i, j int) {
	exp := e.exp[i]; e.exp[i] = e.exp[j]; e.exp[j] = exp
}

func (i1 *idET) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*idET)
	b := BA.CompP(i1.uid, ii2.uid)
	if b != A.Eq {
		return b
	}
	if i1.future && !ii2.future {
		return A.Lt
	}
	if !i1.future && ii2.future {
		return A.Gt
	}
	if i1.hash < ii2.hash {
		return A.Lt
	}
	if i1.hash > ii2.hash {
		return A.Gt
	}
	return A.Eq
}

func doFind (hint string, old, mem, fut bool) J.Json {
	t := A.New()
	set := A.New()
	hintD := BA.ToDown(hint)
	ir := B.IdPosUid(hintD)
	nO := 0; nA := 0
	uid, ok := B.IdNextUid(false, &ir)
	for ok {
		ok = BA.Prefix(hintD, BA.ToDown(uid))
		var (active bool; hash B.Hash)
		if ok {
			_, active, hash, _, _, _, ok = B.IdUidComplete(uid)
		}
		if ok {
			_, b, _ := set.SearchIns(&idET{uid: uid, hash: hash}); M.Assert(!b, 100)
			if active {
				nA++
			} else {
				nO++
			}
			if (old && !active) || (mem && active) {
				_, b, _ := t.SearchIns(&idET{uid: uid, hash: hash, future: false, active: active}); M.Assert(!b, 101)
			}
			uid, ok = B.IdNextUid(false, &ir)
		}
	}
	if len(hint) <= B.PubkeyLen {
		ir := B.IdPosPubkey(B.Pubkey(hint))
		p, ok := B.IdNextPubkey(false, &ir)
		for ok {
			ok = strings.Index(string(p), hint) == 0
			var (uid string; active bool; hash B.Hash)
			if ok {
				uid, active, hash, _, _, _, ok = B.IdPubComplete(p)
			}
			if ok {
				idE := &idET{uid: uid, hash: hash, future: false, active: active}
				if _, b, _ := set.Search(idE); !b {
					if active {
						nA++
					} else {
						nO++
					}
				}
				if old && !active || mem && active {
					t.SearchIns(idE)
				}
				p, ok = B.IdNextPubkey(false, &ir)
			}
		}
	}
	nF := 0
	el := S.IdPosUid(hintD)
	uid, hash, ok := S.IdNextUid(false, &el)
	for ok {
		ok = BA.Prefix(hintD, BA.ToDown(uid))
		if ok {
			set.SearchIns(&idET{uid: uid, hash: hash})
			nF++
			if fut {
				_, b, _ := t.SearchIns(&idET{uid: uid, hash: hash, future: true, active: false}); M.Assert(!b, 103)
			}
			uid, hash, ok = S.IdNextUid(false, &el)
		}
	}
	if len(hint) <= B.PubkeyLen {
		el := S.IdPosPubkey(B.Pubkey(hint))
		p, hash, ok := S.IdNextPubkey(false, &el)
		for ok {
			ok = strings.Index(string(p), hint) == 0
			var uid string
			if ok {
				_, _, uid, _, ok = S.IdHash(hash)
			}
			if ok  {
				idE := &idET{uid: uid, hash: hash, future: true, active: false}
				if _, b, _ := set.Search(idE); !b {
					nF++
				}
				if fut {
					t.SearchIns(idE)
				}
				p, hash, ok = S.IdNextPubkey(false, &el)
			}
		}
	}
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushInteger(int64(nO))
	mk.BuildField("nb_old")
	mk.PushInteger(int64(nA))
	mk.BuildField("nb_member")
	mk.PushInteger(int64(nF))
	mk.BuildField("nb_future")
	mk.StartArray()
	e := t.Next(nil)
	for e != nil {
		idE := e.Val().(*idET)
		mk.StartObject()
		mk.PushString(idE.uid)
		mk.BuildField("uid")
		mk.PushString(string(idE.hash))
		mk.BuildField("hash")
		mk.PushBoolean(idE.future)
		mk.BuildField("future")
		mk.PushBoolean(idE.active)
		mk.BuildField("active")
		mk.BuildObject()
		e = t.Next(e)
	}
	mk.BuildArray()
	mk.BuildField("ids")
	mk.BuildObject()
	return mk.GetJson()
}

func certs (mk *J.Maker, h S.Hash, pubkey B.Pubkey, inBC bool) (certifiers certifiersT) {
	mk.StartObject()
	mk.StartObject()
	t := A.New()
	sentNb := 0
	sentFNb := 0
	var posB B.CertPos
	if inBC {
		okB := B.CertFrom(pubkey, &posB)
		if okB {
			sentNb = posB.CertPosLen()
		}
		var posS S.CertPos
		okS := S.CertFrom(pubkey, &posS)
		if okS {
			sentFNb = posS.CertPosLen()
		}
		var (from, to B.Pubkey)
		if okB {
			from, to, okB = posB.CertNextPos()
		}
		for okB {
			idE := new(idET)
			uid, _, hash, _, _, _, b := B.IdPubComplete(to); M.Assert(b, 100)
			idE.uid = uid
			idE.hash = hash
			idE.future = false
			_, idE.limit, b = B.Cert(from, to); M.Assert(b, 101)
			_, b, _ = t.SearchIns(idE); M.Assert(!b, 102)
			from, to, okB = posB.CertNextPos()
		}
		var toH B.Hash
		if okS {
			from, toH, okS = posS.CertNextPos()
		}
		for okS {
			idE := new(idET)
			idE.hash = toH
			var b bool
			to, idE.limit, b = S.Cert(from, toH)
			if b {
				idE.uid, b = B.IdPub(to)
				if !b {
					_, _, idE.uid, _, b = S.IdHash(toH)
				}
			}
			M.Assert(b, 103)
			idE.future = true
			_, b, _ = t.SearchIns(idE); M.Assert(!b, 104)
			from, toH, okS = posS.CertNextPos()
		}
	}
	mk.PushInteger(int64(sentNb))
	mk.BuildField("nb_member")
	mk.PushInteger(int64(sentFNb))
	mk.BuildField("nb_future")
	mk.StartArray()
	e := t.Next(nil)
	for e != nil {
		idE := e.Val().(*idET)
		mk.StartObject()
		mk.PushString(idE.uid);
		mk.BuildField("uid");
		mk.PushInteger(idE.limit);
		mk.BuildField("limit");
		mk.PushBoolean(idE.future);
		mk.BuildField("future");
		mk.BuildObject()
		e = t.Next(e)
	}
	mk.BuildArray()
	mk.BuildField("partners")
	mk.StartArray()
	if inBC {
		uid, b := B.IdPub(pubkey); M.Assert(b, 105)
		cert := B.AllCertified(uid)
		if cert != nil {
			for _, c := range cert {
				mk.StartObject()
				mk.PushString(c)
				mk.BuildField("cert")
				mk.BuildObject()
			}
		}
	}
	mk.BuildArray()
	mk.BuildField("all_partners")
	mk.BuildObject()
	mk.BuildField("sent_certs")
	
	mk.StartObject()
	t = A.New()
	okB := inBC
	if okB {
		okB = B.CertTo(pubkey, &posB)
	}
	recNb := 0
	if okB {
		recNb = posB.CertPosLen()
	}
	var (okS bool; toH B.Hash; posS S.CertPos)
	if inBC {
		_, _, toH, _, _, _, okS = B.IdPubComplete(pubkey)
		if okS {
			okS = S.CertTo(toH, &posS)
		}
	} else {
		okS = S.CertTo(h, &posS)
	}
	recFNb := 0
	if okS {
		recFNb = posS.CertPosLen()
	}
	certifiers = make(certifiersT, recNb + recFNb)
	var (es expSort; ts = sort.TS{Sorter: &es})
	if recNb >= int(B.Pars().SigQty) {
		es.exp = make([]int64, recNb);
	} else {
		es.exp = nil
	}
	i := 0
	var from, to B.Pubkey
	if okB {
		from, to, okB = posB.CertNextPos()
	}
	for okB {
		certifiers[i] = from
		idE := new(idET)
		uid, _, hash, _, _, _, b := B.IdPubComplete(from); M.Assert(b, 106)
		idE.uid = uid
		idE.hash = hash
		idE.future = false
		_, idE.limit, b = B.Cert(from, to); M.Assert(b, 107)
		if es.exp != nil {
			es.exp[i] = idE.limit
		}
		_, b, _ = t.SearchIns(idE); M.Assert(!b, 108)
		i++
		from, to, okB = posB.CertNextPos()
	}
	if okS {
		from, toH, okS = posS.CertNextPos()
	}
	for okS {
		certifiers[i] = from
		i++
		idE := new(idET)
		var b bool
		idE.uid, b = B.IdPub(from); M.Assert(b, 109)
		idE.hash = toH
		idE.future = true
		_, idE.limit, b = S.Cert(from, toH, ); M.Assert(b, 110)
		_, b, _ = t.SearchIns(idE); M.Assert(!b, 111)
		from, toH, okS = posS.CertNextPos()
	}
	mk.PushInteger(int64(recNb))
	mk.BuildField("nb_member")
	mk.PushInteger(int64(recFNb))
	mk.BuildField("nb_future")
	mk.StartArray()
	e = t.Next(nil)
	for e != nil {
		idE := e.Val().(*idET)
		mk.StartObject()
		mk.PushString(idE.uid);
		mk.BuildField("uid");
		mk.PushInteger(idE.limit);
		mk.BuildField("limit");
		mk.PushBoolean(idE.future);
		mk.BuildField("future");
		mk.BuildObject()
		e = t.Next(e)
	}
	mk.BuildArray()
	mk.BuildField("partners")
	mk.StartArray()
	if inBC {
		uid, b := B.IdPub(pubkey); M.Assert(b, 112)
		cert := B.AllCertifiers(uid)
		if cert != nil {
			for _, c := range cert {
				mk.StartObject()
				mk.PushString(c)
				mk.BuildField("cert")
				mk.BuildObject()
			}
		}
	}
	mk.BuildArray()
	mk.BuildField("all_partners")
	mk.BuildObject()
	mk.BuildField("received_certs")
	rCertsLimit := int64(- 1)
	if es.exp != nil {
		ts.QuickSort(0, recNb - 1)
		rCertsLimit = es.exp[recNb - int(B.Pars().SigQty)]
	}
	mk.PushInteger(rCertsLimit)
	mk.BuildField("received_certs_limit")
	mk.BuildObject()
	return
}

func get (hash S.Hash) (uid string, pubkey B.Pubkey, block int32, blockDate, limitDate int64, h H.History, inBC, member, ok bool) {
	pubkey, inBC = B.IdHash(hash)
	ok = false
	if inBC {
		uid, member, _, block, _, limitDate, ok = B.IdPubComplete(pubkey); M.Assert(ok, 100);
		blockDate, _, ok = B.TimeOf(block); M.Assert(ok, 101)
	} else if _, pubkey, uid, limitDate, ok = S.IdHash(hash); ok {
		member = false
		block = - 1
		blockDate = - 1
	}
	h = H.BuildHistoryP(pubkey)
	return
}

func notTooFar (p B.Pubkey, member bool, certifiers []B.Pubkey) (proportionOfSentries float64, ok bool) {
	n := 0
	if certifiers != nil {
		n = len(certifiers)
	}
	i := 0
	if member {
		n++
		i = 1
	}
	if n == 0 {
		proportionOfSentries = 0.
		ok = false
		return
	}
	certs := make([]B.Pubkey, n)
	if member {
		certs[0] = p
	}
	if certifiers != nil {
		for _, c := range certifiers {
			certs[i] = c
			i++
		}
	}
	proportionOfSentries = B.PercentOfSentries(certs)
	ok = proportionOfSentries >= B.Pars().Xpercent
	return
}

func fixCertNextDate (member bool, p B.Pubkey) (date int64, passed bool) {
	passed = member
	if member {
		date = 0
		var pos B.CertPos
		if B.CertFrom(p, &pos) {
			from, to, ok := pos.CertNextPos()
			for ok {
				block_number, _, b := B.Cert(from, to); M.Assert(b, 100)
				tm, _, b := B.TimeOf(block_number); M.Assert(b, 101)
				date = M.Max64(date, tm)
				from, to, ok = pos.CertNextPos()
			}
			date += int64(B.Pars().SigPeriod)
			passed = date <= B.Now()
		}
	} else {
		date = - 1
	}
	return
}

func calcQuality (p B.Pubkey) (quality float64) {
	pubs := make(B.PubkeysT, 1)
	pubs[0] = p
	quality = B.PercentOfSentries(pubs)
	return
}

func calcCentrality (p B.Pubkey, inBC bool) (centrality float64) {
	if inBC {
		centrality = C.CountOne(p)
	} else {
		centrality = 0.
	}
	return
}

func doFix (hash S.Hash, calcDist, calcQual, calcCentr bool) J.Json {
	mk := J.NewMaker()
	mk.StartObject()
	if uid, pubkey, block, blockDate, limitDate, history, inBC, member, ok := get(hash); ok {
		mk.StartObject()
		mk.PushString(string(hash))
		mk.BuildField("hash")
		mk.PushString(uid)
		mk.BuildField("uid")
		mk.PushString(string(pubkey))
		mk.BuildField("pubkey")
		mk.PushInteger(int64(block))
		mk.BuildField("block")
		mk.PushInteger(blockDate)
		mk.BuildField("blockDate")
		mk.PushInteger(limitDate)
		mk.BuildField("limitDate")
		mk.PushBoolean(member)
		mk.BuildField("member")
		sentry := member && B.IsSentry(pubkey)
		mk.PushBoolean(sentry)
		mk.BuildField("sentry")
		availability, passed := fixCertNextDate(member, pubkey)
		mk.PushInteger(availability)
		mk.BuildField("availability")
		mk.PushBoolean(passed)
		mk.BuildField("passed")
		in := true
		mk.StartArray()
		for _, ev := range history {
			mk.StartObject()
			mk.PushBoolean(in)
			in = !in
			mk.BuildField("in")
			mk.PushInteger(int64(ev.Block))
			mk.BuildField("block")
			mk.PushInteger(ev.Date)
			mk.BuildField("date")
			mk.BuildObject()
		}
		mk.BuildArray()
		mk.BuildField("history")
		certifiers := certs(mk, hash, pubkey, inBC)
		mk.BuildField("certifications")
		if calcDist {
			distance, distanceOK := notTooFar(pubkey, member, certifiers)
			mk.PushFloat(distance)
			mk.PushBoolean(distanceOK)
		} else {
			mk.PushFloat(0)
			mk.PushBoolean(false)
		}
		mk.Swap()
		mk.BuildField("distance")
		mk.Swap()
		mk.BuildField("distance_ok")
		if calcQual {
			mk.PushFloat(calcQuality(pubkey))
		} else {
			mk.PushFloat(0.)
		}
		mk.BuildField("quality")
		if calcCentr {
			mk.PushFloat(calcCentrality(pubkey, inBC))
		} else {
			mk.PushFloat(0.)
		}
		mk.BuildField("centrality")
		mk.BuildObject()
	} else {
		mk.PushNull()
	}
	mk.BuildField("res")
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case findA:
		s = findName
	case fixA:
		s = fixName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case findA:
		G.Json(doFind(a.hint, a.b1, a.b2, a.b3), a.output)
	case fixA:
		G.Json(doFix(a.hash, a.b1, a.b2, a.b3), a.output)
	}
}

func find (hint, output string, newAction chan<- B.Actioner, fields ...string) {
	a := &action{what: findA, output: output, hint: hint, b1: false, b2: false, b3: false}
	for _, f := range fields {
		switch f {
		case findOldName:
			a.b1 = true
		case findMemName:
			a.b2 = true
		case findFutName:
			a.b3 = true
		}
	}
	newAction <- a
}

func fix (hash string, output string, newAction chan<- B.Actioner, fields ...string) {
	a := &action{what: fixA, output: output, hash: B.Hash(hash), b1: false, b2: false, b3: false}
	for _, f := range fields {
		switch f {
		case fixDistName:
			a.b1 = true
		case fixQualName:
			a.b2 = true
		case fixCentrName:
			a.b3 = true
		}
	}
	newAction <- a
}

func init () {
	G.AddAction(findName, find, G.Arguments{findHintName})
	G.AddAction(fixName, fix, G.Arguments{fixHashName})
}
