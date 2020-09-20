/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identitySearchList

// À corriger dans certs

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	C	"duniter/centralities"
	M	"util/misc"
	S	"duniter/sandbox"
		"strings"
		"util/sort"

)

const (
	
	Revoked = iota
	Missing
	Member
	Newcomer

)

type (
	
	StatusList []int
	
	IdET struct {
		uid string
		Hash B.Hash
		Status int
		limit int64
	}
	
	expSort struct {
		exp []int64
	}

)

func (e *expSort) Less (i, j int) bool {
	return e.exp[i] < e.exp[j]
} //Less

func (e *expSort) Swap (i, j int) {
	exp := e.exp[i]; e.exp[i] = e.exp[j]; e.exp[j] = exp
} //Swap

func (i1 *IdET) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*IdET)
	b := BA.CompP(i1.uid, ii2.uid)
	if b != A.Eq {
		return b
	}
	if i1.Status == Newcomer && ii2.Status != Newcomer {
		return A.Lt
	}
	if i1.Status != Newcomer && ii2.Status == Newcomer {
		return A.Gt
	}
	if i1.Hash < ii2.Hash {
		return A.Lt
	}
	if i1.Hash > ii2.Hash {
		return A.Gt
	}
	return A.Eq
} //Compare

func getStatus (inBC, active bool, exp int64) (status int) {
	if active {
		status = Member
	} else if exp == BA.Revoked {
		status = Revoked
	} else if inBC {
		status = Missing
	} else {
		status = Newcomer
	}
	return
} //getStatus

func Find (hint string, sl StatusList) (nR, nM, nA, nF int, ids *A.Tree) { //IdET
	
	getFlags := func (sl StatusList) M.Set {
		set := M.MakeSet()
		for _, s := range sl {
			set = M.Add(set, s)
		}
		return set
	} //getFlags
	
	incNbs :=func (status int, nR, nM, nA, nF *int) {
		switch status {
		case Revoked:
			*nR++
		case Missing:
			*nM++
		case Member:
			*nA++
		case Newcomer:
			*nF++
		}
	} //incNbs
	
	//Find
	asked := getFlags(sl)
	ids = A.New() //IdET
	set := A.New() //IdET
	hintD := BA.ToDown(hint)
	ir := B.IdPosUid(hintD)
	nR = 0; nM = 0; nA = 0; nF = 0
	uid, okL := B.IdNextUid(false, &ir)
	for okL {
		ok := BA.Prefix(hintD, BA.ToDown(uid))
		var (active bool; hash B.Hash; exp int64)
		if ok {
			_, active, hash, _, _, exp, ok = B.IdUidComplete(uid); M.Assert(ok, 100)
			_, b, _ := set.SearchIns(&IdET{uid: uid, Hash: hash}); M.Assert(!b, 101)
			status := getStatus(true, active, exp)
			incNbs(status, &nR, &nM, &nA, &nF)
			if M.In(status, asked) {
				_, b, _ := ids.SearchIns(&IdET{uid: uid, Hash: hash, Status: status}); M.Assert(!b, 102)
			}
		}
		uid, okL = B.IdNextUid(false, &ir)
	}
	if len(hint) <= B.PubkeyLen {
		ir := B.IdPosPubkey(B.Pubkey(hint))
		p, okL := B.IdNextPubkey(false, &ir)
		for okL {
			ok := strings.Index(string(p), hint) == 0
			var (uid string; active bool; hash B.Hash; exp int64)
			if ok {
				uid, active, hash, _, _, exp, ok = B.IdPubComplete(p); M.Assert(ok, 103)
				status := getStatus(true, active, exp)
				if _, b, _ := set.Search(&IdET{uid: uid, Hash: hash}); !b {
					incNbs(status, &nR, &nM, &nA, &nF)
				}
				if M.In(status, asked) {
					ids.SearchIns(&IdET{uid: uid, Hash: hash, Status: status})
				}
			}
			p, okL = B.IdNextPubkey(false, &ir)
		}
	}
	el := S.IdPosUid(hintD)
	uid, hash, okL := S.IdNextUid(false, &el)
	for okL {
		ok := BA.Prefix(hintD, BA.ToDown(uid))
		if ok {
			set.SearchIns(&IdET{uid: uid, Hash: hash})
			_, inBC := B.IdHash(hash)
			active := inBC
			if active {
				_, active, _, _, _, _, ok = B.IdUidComplete(uid); M.Assert(ok, 104)
			}
			status := getStatus(inBC, active, 0)
			if !inBC {
				incNbs(status, &nR, &nM, &nA, &nF)
			}
			if M.In(status, asked) {
				ids.SearchIns(&IdET{uid: uid, Hash: hash, Status: status})
			}
		}
		uid, hash, okL = S.IdNextUid(false, &el)
	}
	if len(hint) <= B.PubkeyLen {
		el := S.IdPosPubkey(B.Pubkey(hint))
		p, hash, okL := S.IdNextPubkey(false, &el)
		for okL {
			ok := strings.Index(string(p), hint) == 0
			if ok {
				_, inBC := B.IdHash(hash)
				active := inBC
				if active {
					_, active, _, _, _, _, ok = B.IdPubComplete(p); M.Assert(ok, 105)
				}
				status := getStatus(inBC, active, 0)
				var uid string
				_, _, uid, _, _, ok = S.IdHash(hash); M.Assert(ok, 106)
				if _, b, _ := set.Search(&IdET{uid: uid, Hash: hash}); !b {
					incNbs(status, &nR, &nM, &nA, &nF)
				}
				if M.In(status, asked) {
					ids.SearchIns(&IdET{uid: uid, Hash: hash, Status: status})
				}
			}
			p, hash, okL = S.IdNextPubkey(false, &el)
		}
	}
	return
} //Find

func SentCerts (h S.Hash, pubkey B.Pubkey, inBC bool) (sentNb, sentFNb int, certified *A.Tree) { //*IdET
	certified = A.New()
	sentNb = 0
	sentFNb = 0
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
		var from, to B.Pubkey
		if okB {
			from, to, okB = posB.CertNextPos()
		}
		for okB {
			idE := new(IdET)
			uid, _, hash, _, _, _, b := B.IdPubComplete(to); M.Assert(b, 100)
			idE.uid = uid
			idE.Hash = hash
			idE.Status = Member // Status of the certification (Newcomer or Member)
			_, idE.limit, b = B.Cert(from, to); M.Assert(b, 101)
			_, b, _ = certified.SearchIns(idE); M.Assert(!b, 102)
			from, to, okB = posB.CertNextPos()
		}
		var toH B.Hash
		if okS {
			from, toH, okS = posS.CertNextPos()
		}
		for okS {
			idE := new(IdET)
			idE.Hash = toH
			var b bool
			to, _, idE.limit, b = S.Cert(from, toH)
			if b {
				idE.uid, b = B.IdPub(to)
				if !b {
					_, _, idE.uid, _, _, b = S.IdHash(toH)
				}
			}
			M.Assert(b, 103)
			idE.Status = Newcomer
			certified.SearchIns(idE)
			from, toH, okS = posS.CertNextPos()
		}
	}
	return
} //SentCerts

func RecCerts (h S.Hash, pubkey B.Pubkey, inBC bool) (recNb, recFNb int, rCertsLimit int64, certifiers *A.Tree, certifiersA B.PubkeysT) { //*IdET
	certifiers = A.New()
	var posB B.CertPos
	okB := inBC
	if okB {
		okB = B.CertTo(pubkey, &posB)
	}
	recNb = 0
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
	recFNb = 0
	if okS {
		recFNb = posS.CertPosLen() // À corriger : Enlever les émetteurs qui ne sont plus membres
	}
	certifiersA = make(B.PubkeysT, recNb + recFNb)
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
		certifiersA[i] = from
		idE := new(IdET)
		uid, _, hash, _, _, _, b := B.IdPubComplete(from); M.Assert(b, 106)
		idE.uid = uid
		idE.Hash = hash
		idE.Status = Member
		_, idE.limit, b = B.Cert(from, to); M.Assert(b, 107)
		if es.exp != nil {
			es.exp[i] = idE.limit
		}
		_, b, _ = certifiers.SearchIns(idE); M.Assert(!b, 108)
		i++
		from, to, okB = posB.CertNextPos()
	}
	if okS {
		from, toH, okS = posS.CertNextPos()
	}
	for okS {
		certifiersA[i] = from // À corriger : vérifier que from est toujours membre et sinon ajouter une croix à côté du rond
		i++
		idE := new(IdET)
		var (b bool; hash B.Hash)
		idE.uid, _, hash, _, _, _, b = B.IdPubComplete(from); M.Assert(b, 109)
		idE.Hash = hash
		idE.Status = Newcomer
		_, _, idE.limit, b = S.Cert(from, toH, ); M.Assert(b, 110)
		_, b, _ = certifiers.SearchIns(idE)
		from, toH, okS = posS.CertNextPos()
	}
	rCertsLimit = int64(- 1)
	if es.exp != nil {
		ts.QuickSort(0, recNb - 1)
		rCertsLimit = es.exp[recNb - int(B.Pars().SigQty)]
	}
	return
} //RecCerts

func Get (hash S.Hash) (uid string, pubkey B.Pubkey, block, application int32, limitDate int64, inBC, member, ok bool) {
	pubkey, inBC = B.IdHash(hash)
	ok = false
	if inBC {
		uid, member, _, block, application, limitDate, ok = B.IdPubComplete(pubkey); M.Assert(ok, 100);
	} else if _, pubkey, uid, _, limitDate, ok = S.IdHash(hash); ok {
		member = false
		block = - 1
		application = -1
	}
	return
} //Get

func NotTooFar (p B.Pubkey, member bool, certifiers B.PubkeysT) (proportionOfSentries float64, ok bool) {
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
	certs := make(B.PubkeysT, n)
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
} //NotTooFar

func FixCertNextDate (member bool, p B.Pubkey) (date int64, passed bool) {
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
} //FixCertNextDate

func CalcQuality (p B.Pubkey) (quality float64) {
	pubs := make(B.PubkeysT, 1)
	pubs[0] = p
	quality = B.PercentOfSentries(pubs)
	return
} //CalcQuality

func CalcCentrality (p B.Pubkey, inBC bool) (centrality float64) {
	if inBC {
		centrality = C.CountOne(p)
	} else {
		centrality = 0.
	}
	return
} //CalcCentrality
