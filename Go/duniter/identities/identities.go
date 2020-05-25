/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identities

import (
	
	A	"util/avl"
	BA	"duniter/basic"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	IS	"duniter/identitySearchList"
	M	"util/misc"
	S	"duniter/sandbox"
		/*
		"fmt"
		*/

)

type (
		
	filter func (member bool, expires_on int64) bool

)

var (
	
	filters = [...]filter{
		func (member bool, exp int64) bool { //IS.Revoked
			return exp == BA.Revoked
		},
		func (member bool, exp int64) bool { //IS.Missing
			return !member && (exp != BA.Revoked)
		},
		func (member bool, exp int64) bool { //IS.Member
			return member
		},
	}

)

func insert (l *G.ListValue, hash B.Hash) {
	l.Append(GQ.Wrap(hash))
} //insert

func listBC (f filter, sortedByPubkey bool, from, to string) *G.ListValue {
	l := G.NewListValue()
	if sortedByPubkey {
		fromP := B.Pubkey(from)
		toP := B.Pubkey(to)
		ir := B.IdPosPubkey(fromP)
		pubkey, ok := B.IdNextPubkey(false, &ir)
		for ok && (toP == "" || pubkey < toP)  {
			_, member, hash, _, _, exp, b := B.IdPubComplete(pubkey); M.Assert(b, 100)
			if f(member, exp) {
				insert(l, hash)
			}
			pubkey, ok = B.IdNextPubkey(false, &ir)
		}
	} else {
		ir := B.IdPosUid(from)
		uid, ok := B.IdNextUid(false, &ir)
		for ok && (to == "" || BA.CompP(uid, to) == BA.Lt) {
			_, member, hash, _, _, exp, b := B.IdUidComplete(uid); M.Assert(b, 101)
			if f(member, exp) {
				insert(l, hash)
			}
			uid, ok = B.IdNextUid(false, &ir)
		}
	}
	return l
} //listBC

func listSB (sortedByPubkey bool, from, to string) *G.ListValue {
	l := G.NewListValue()
	if sortedByPubkey {
		fromP := B.Pubkey(from)
		toP := B.Pubkey(to)
		el := S.IdPosPubkey(fromP)
		p, hash, ok := S.IdNextPubkey(false, &el)
		for ok && (toP == "" || p < toP)  {
			if _, ok := B.IdHash(hash); !ok {
				insert(l, hash)
			}
			p, hash, ok = S.IdNextPubkey(false, &el)
		}
	} else {
		el := S.IdPosUid(from)
		uid, hash, ok := S.IdNextUid(false, &el)
		for ok && (to == "" || BA.CompP(uid, to) == BA.Lt) {
			if _, ok := B.IdHash(hash); !ok {
				insert(l, hash)
			}
			uid, hash, ok = S.IdNextUid(false, &el)
		}
	}
	return l
} //listSB

func takeStatus (enum *G.EnumValue) (status int) {
	switch enum.Enum.S {
	case "REVOKED":
		status = IS.Revoked
	case "MISSING":
		status = IS.Missing
	case "MEMBER":
		status = IS.Member
	case "NEWCOMER":
		status = IS.Newcomer
	}
	return
} //takeStatus

func getStatus (argumentValues *A.Tree) (status int) {
	var v G.Value
	ok := G.GetValue(argumentValues, "status", &v)
	M.Assert(ok, 100)
	switch v := v.(type) {
	case *G.EnumValue:
		status = takeStatus(v)
	default:
		M.Halt(v, 101)
	}
	return
} //getStatus

func getOrder (argumentValues *A.Tree) bool {
	var v G.Value
	ok := G.GetValue(argumentValues, "sortedBy", &v)
	M.Assert(ok, 100)
	switch v := v.(type) {
	case *G.EnumValue:
		return v.Enum.S == "PUBKEY"
	default:
		M.Halt(v, 101)
		return false
	}
} //getOrder

func getLimits (argumentValues *A.Tree) (from, to string) {
	
	get := func (name string) string {
		var v G.Value
		ok := G.GetValue(argumentValues, name, &v)
		M.Assert(ok, 100)
		switch v := v.(type) {
		case *G.StringValue:
			return v.String.S
		default:
			M.Halt(v, 100)
			return ""
		}
	} //get
	
	//getLimits
	from = get("start")
	to = get("end")
	return
} //getLimits

func identitiesR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	order := getOrder(argumentValues)
	from, to := getLimits(argumentValues)
	status := getStatus(argumentValues)
	if status == IS.Newcomer {
		return listSB(order, from, to)
	} else {
		return listBC(filters[status], order, from, to)
	}
} //identitiesR

func idSearchR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var (v G.Value; hint string; statusList IS.StatusList)
	ok := G.GetValue(argumentValues, "with", &v)
	M.Assert(ok, 100)
	switch v := v.(type) {
	case *G.InputObjectValue:
		var vv G.Value
		ok := G.GetObjectValueInputField(v, "hint", &vv)
		M.Assert(ok, 101)
		switch vv := vv.(type) {
		case *G.StringValue:
			hint = vv.String.S
		default:
			M.Halt(vv, 102)
		}
		ok = G.GetObjectValueInputField(v, "status_list", &vv)
		M.Assert(ok, 103)
		switch vv := vv.(type) {
		case *G.ListValue:
			n := 0
			for l := vv.First(); l != nil; l = vv.Next(l) {
				n++
			}
			statusList = make(IS.StatusList, n)
			i := 0
			for l := vv.First(); l != nil; l = vv.Next(l) {
				statusList[i] = takeStatus(l.Value.(*G.EnumValue))
				i++
			}
		default:
			M.Halt(vv, 104)
		}
	default:
		M.Halt(105)
	}
	nR, nM, nA, nF, ids := IS.Find(hint, statusList)
	return GQ.Wrap(nR, nM, nA, nF, ids)
} //idSearchR

func idFromHashR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var v G.Value
	ok := G.GetValue(argumentValues, "hash", &v)
	if ! ok {
		return G.MakeNullValue() 
	}
	switch h := v.(type) {
	case *G.StringValue:
		hash := B.Hash(h.String.S)
		_, ok := B.IdHash(hash)
		if !ok {
			_, _, _, _, _, ok = S.IdHash(hash)
		}
		if !ok {
			return G.MakeNullValue()
		}
		return GQ.Wrap(hash)
	default:
		M.Halt(h, 100)
		return nil
	}
} //idFromHashR

func idSearchOutputNRR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch nR := GQ.Unwrap(rootValue, 0).(type) {
	case int:
		return G.MakeIntValue(nR)
	default:
		M.Halt(nR, 100)
		return nil
	}
} //idSearchOutputNRR

func idSearchOutputNMR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch nM := GQ.Unwrap(rootValue, 1).(type) {
	case int:
		return G.MakeIntValue(nM)
	default:
		M.Halt(nM, 100)
		return nil
	}
} //idSearchOutputNMR

func idSearchOutputNAR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch nA := GQ.Unwrap(rootValue, 2).(type) {
	case int:
		return G.MakeIntValue(nA)
	default:
		M.Halt(nA, 100)
		return nil
	}
} //idSearchOutputNAR

func idSearchOutputNFR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch nF := GQ.Unwrap(rootValue, 3).(type) {
	case int:
		return G.MakeIntValue(nF)
	default:
		M.Halt(nF, 100)
		return nil
	}
} //idSearchOutputNFR

func idSearchOutputIdsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch ids := GQ.Unwrap(rootValue, 4).(type) {
	case *A.Tree:
		l := G.NewListValue()
		e := ids.Next(nil)
		for e != nil {
			l.Append(GQ.Wrap(e.Val().(*IS.IdET).Hash))
			e = ids.Next(e)
		}
		return l
	default:
		M.Halt(ids, 100)
		return nil
	}
} //idSearchOutputIdsR

func identityPubkeyR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		return G.MakeStringValue(string(pub))
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityPubkeyR

func identityUidR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		uid, _, _, _, _, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		return G.MakeStringValue(uid)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityUidR

func identityHashR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		return G.MakeStringValue(string(hash))
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityHashR

func identityStatusR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, _, exp, inBC, member, ok := IS.Get(hash); M.Assert(ok, 100)
		var s string
		if !inBC {
			s = "NEWCOMER"
		} else {
			if member {
				s = "MEMBER"
			} else if exp != BA.Revoked {
				s = "MISSING"
			} else {
				s = "REVOKED"
			}
		}
		return G.MakeEnumValue(s)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityStatusR

func identityMPR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, _, _, mp := S.IdHash(hash)
		return G.MakeBooleanValue(mp)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityMPR

func identityMPBR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, bnb, _, mp := S.IdHash(hash)
		if mp {
			return GQ.Wrap(bnb)
		} else {
			return G.MakeNullValue()
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityMPBR

func identityMPLDR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, _, exp, mp := S.IdHash(hash)
		if mp {
			return G.MakeInt64Value(exp)
		} else {
			return G.MakeNullValue()
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityMPLDR

func identityIdWrittenBlockR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, block, _, _, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		if block >= 0 {
			return GQ.Wrap(block)
		} else {
			return G.MakeNullValue()
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityIdWrittenBlockR

func identityLastApplicationR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, app, _, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		if app >= 0 {
			return GQ.Wrap(app)
		} else {
			return G.MakeNullValue()
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityLastApplicationR

func identityLimitDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, _, exp, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		if exp == BA.Revoked {
			return G.MakeNullValue()
		}
		return G.MakeInt64Value(M.Abs64(exp))
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityLimitDateR

func identityLeavingR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, _, _, _, exp, inBC, _, ok := IS.Get(hash); M.Assert(ok, 100)
		if !inBC || exp == BA.Revoked {
			return G.MakeNullValue()
		}
		return G.MakeBooleanValue(exp < 0)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityLeavingR

func identitySentryR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, _, member, ok := IS.Get(hash); M.Assert(ok, 100)
		if member {
			return G.MakeBooleanValue(B.IsSentry(pub))
		} else {
			return G.MakeBooleanValue(false)
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identitySentryR

func identityRecCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, inBC, _, ok := IS.Get(hash); M.Assert(ok, 100)
		_, _, limit, certifiers, _ := IS.RecCerts(hash, pub, inBC)
		l := G.NewListValue()
		e := certifiers.Next(nil)
		for e != nil {
			idE := e.Val().(*IS.IdET)
			l.Append(GQ.Wrap(idE.Hash, hash, idE.Status == IS.Newcomer)) // Status == newcomer means certification is future
			e = certifiers.Next(e)
		}
		return GQ.Wrap(l, limit)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityRecCertsR

func identitySentCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, inBC, _, ok := IS.Get(hash); M.Assert(ok, 100)
		_, _, certified := IS.SentCerts(hash, pub, inBC)
		l := G.NewListValue()
		e := certified.Next(nil)
		for e != nil {
			idE := e.Val().(*IS.IdET)
			l.Append(GQ.Wrap(hash, idE.Hash, idE.Status == IS.Newcomer)) // Status == newcomer means certification is future
			e = certified.Next(e)
		}
		return l
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identitySentCertsR

func identityAllRecCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		l := G.NewListValue()
		pubkey, inBC := B.IdHash(hash)
		if inBC {
			uid, b := B.IdPub(pubkey); M.Assert(b, 100)
			for _, uid := range B.AllCertifiers(uid) {
				_, _, h, _, _, _, b := B.IdUidComplete(uid); M.Assert(b, 101)
				l.Append(GQ.Wrap(h)) // Status == newcomer means certification is future
			}
		}
			return l
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityAllRecCertsR

func identityAllSentCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		l := G.NewListValue()
		pubkey, inBC := B.IdHash(hash)
		if inBC {
			uid, b := B.IdPub(pubkey); M.Assert(b, 100)
			for _, uid := range B.AllCertified(uid) {
				_, _, h, _, _, _, b := B.IdUidComplete(uid); M.Assert(b, 101)
				l.Append(GQ.Wrap(h)) // Status == newcomer means certification is future
			}
		}
		return l
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityAllSentCertsR

func identityDistanceR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, inBC, member, ok := IS.Get(hash); M.Assert(ok, 100)
		_, _, _, _, certifiers := IS.RecCerts(hash, pub, inBC)
		dist, distOk := IS.NotTooFar(pub, member, certifiers)
		return GQ.Wrap(dist, distOk)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityDistanceR

func identityQualityR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, _, _, ok := IS.Get(hash); M.Assert(ok, 100)
		return G.MakeFloat64Value(IS.CalcQuality(pub) * 100)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityQualityR

func identityCentralityR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, inBC, _, ok := IS.Get(hash); M.Assert(ok, 100)
		return G.MakeFloat64Value(IS.CalcCentrality(pub, inBC) * 100)
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityCentralityR

func identityMinDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, _, member, ok := IS.Get(hash); M.Assert(ok, 100)
		date, _ := IS.FixCertNextDate(member, pub)
		if date < 0 {
			return G.MakeNullValue()
		} else {
			return G.MakeInt64Value(date)
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityMinDateR

func identityMinDatePassedR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		_, pub, _, _, _, _, member, ok := IS.Get(hash); M.Assert(ok, 100)
		date, passed := IS.FixCertNextDate(member, pub)
		if date < 0 {
			return G.MakeNullValue()
		} else {
			return G.MakeBooleanValue(passed)
		}
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityMinDatePassedR

func DistanceValR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch dist := GQ.Unwrap(rootValue, 0).(type) {
	case float64:
		return G.MakeFloat64Value(dist * 100)
	default:
		M.Halt(dist, 100)
		return nil
	}
} //DistanceValR

func DistanceOkR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch ok := GQ.Unwrap(rootValue, 1).(type) {
	case bool:
		return G.MakeBooleanValue(ok)
	default:
		M.Halt(ok, 100)
		return nil
	}
} //DistanceOkR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "identities", identitiesR)
	ts.FixFieldResolver("Query", "idSearch", idSearchR)
	ts.FixFieldResolver("Query", "idFromHash", idFromHashR)
	
	ts.FixFieldResolver("IdSearchOutput", "revokedNb", idSearchOutputNRR)
	ts.FixFieldResolver("IdSearchOutput", "missingNb", idSearchOutputNMR)
	ts.FixFieldResolver("IdSearchOutput", "memberNb", idSearchOutputNAR)
	ts.FixFieldResolver("IdSearchOutput", "newcomerNb", idSearchOutputNFR)
	ts.FixFieldResolver("IdSearchOutput", "ids", idSearchOutputIdsR)
	
	ts.FixFieldResolver("Identity", "pubkey", identityPubkeyR)
	ts.FixFieldResolver("Identity", "uid", identityUidR)
	ts.FixFieldResolver("Identity", "hash", identityHashR)
	ts.FixFieldResolver("Identity", "status", identityStatusR)
	ts.FixFieldResolver("Identity", "membership_pending", identityMPR)
	ts.FixFieldResolver("Identity", "membership_pending_block", identityMPBR)
	ts.FixFieldResolver("Identity", "membership_pending_limitDate", identityMPLDR)
	ts.FixFieldResolver("Identity", "id_written_block", identityIdWrittenBlockR)
	ts.FixFieldResolver("Identity", "lastApplication", identityLastApplicationR)
	ts.FixFieldResolver("Identity", "limitDate", identityLimitDateR)
	ts.FixFieldResolver("Identity", "isLeaving", identityLeavingR)
	ts.FixFieldResolver("Identity", "sentry", identitySentryR)
	ts.FixFieldResolver("Identity", "received_certifications", identityRecCertsR)
	ts.FixFieldResolver("Identity", "sent_certifications", identitySentCertsR)
	ts.FixFieldResolver("Identity", "all_certifiers", identityAllRecCertsR)
	ts.FixFieldResolver("Identity", "all_certified", identityAllSentCertsR)
	ts.FixFieldResolver("Identity", "distance", identityDistanceR)
	ts.FixFieldResolver("Identity", "quality", identityQualityR)
	ts.FixFieldResolver("Identity", "centrality", identityCentralityR)
	ts.FixFieldResolver("Identity", "minDate", identityMinDateR)
	ts.FixFieldResolver("Identity", "minDatePassed", identityMinDatePassedR)
	
	ts.FixFieldResolver("Distance", "value", DistanceValR)
	ts.FixFieldResolver("Distance", "dist_ok", DistanceOkR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
} //init
