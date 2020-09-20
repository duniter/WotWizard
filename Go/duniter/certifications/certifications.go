/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package certifications

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"
	S	"duniter/sandbox"
	/*
	"fmt"
	*/

)

/*
type (
	
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
*/

func rCertsCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch list := GQ.Unwrap(rootValue, 0).(type) {
	case *G.ListValue:
		return list
	default:
		M.Halt(list, 100)
		return nil
	}
} //rCertsCertsR

func rCertsLimitR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch limit := GQ.Unwrap(rootValue, 1).(type) {
	case int64:
		if limit < 0 {
			return G.MakeNullValue()
		}
		return G.MakeInt64Value(limit)
	default:
		M.Halt(limit, 100)
		return nil
	}
} //rCertsLimitR

func certFromR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch from := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		return GQ.Wrap(from)
	default:
		M.Halt(from, 100)
		return nil
	}
} //certFromR

func certToR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch to := GQ.Unwrap(rootValue, 1).(type) {
	case B.Hash:
		return GQ.Wrap(to)
	default:
		M.Halt(to, 100)
		return nil
	}
} //certToR

func certPendingR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch pending := GQ.Unwrap(rootValue, 2).(type) {
	case bool:
		return G.MakeBooleanValue(pending)
	default:
		M.Halt(pending, 100)
		return nil
	}
} //certPendingR

func certBlockR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var (from, to B.Pubkey; toH B.Hash; idInBC bool)
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		from, idInBC = B.IdHash(hash); M.Assert(idInBC, 100)
	default:
		M.Halt(hash, 100)
		return nil
	}
	switch hash := GQ.Unwrap(rootValue, 1).(type) {
	case B.Hash:
		toH = hash
		to, idInBC = B.IdHash(hash)
	default:
		M.Halt(hash, 100)
		return nil
	}
	var block int32
	ok := false
	if idInBC {
		block, _, ok = B.Cert(from, to)
	}
	if !ok {
		_, block, _, ok = S.Cert(from, toH); M.Assert(ok, 103)
	}
	return GQ.Wrap(block)
} //certBlockR

func certExpR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var (from, to B.Pubkey; toH B.Hash; idInBC, pending bool)
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		from, idInBC = B.IdHash(hash); M.Assert(idInBC, 100)
	default:
		M.Halt(hash, 101)
		return nil
	}
	switch hash := GQ.Unwrap(rootValue, 1).(type) {
	case B.Hash:
		toH = hash
		to, idInBC = B.IdHash(hash)
	default:
		M.Halt(hash, 102)
		return nil
	}
	switch p := GQ.Unwrap(rootValue, 2).(type) {
	case bool:
		pending = p
	default:
		M.Halt(pending, 103)
		return nil
	}
	var (exp int64; ok bool)
	if pending {
		_, _, exp, ok = S.Cert(from, toH)
	}
	if !pending || !ok {
		_, exp, ok = B.Cert(from, to); M.Assert(ok, 104)
	}
	return G.MakeInt64Value(exp)
} //certExpR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Received_Certifications", "certifications", rCertsCertsR)
	ts.FixFieldResolver("Received_Certifications", "limit", rCertsLimitR)
	ts.FixFieldResolver("Certification", "from", certFromR)
	ts.FixFieldResolver("Certification", "to", certToR)
	ts.FixFieldResolver("Certification", "pending", certPendingR)
	ts.FixFieldResolver("Certification", "block", certBlockR)
	ts.FixFieldResolver("Certification", "expires_on", certExpR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
} //init
