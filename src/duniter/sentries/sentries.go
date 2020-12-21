/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package sentries

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"
	U	"util/sets2"
	
)

type (
	
	uid struct {
		uid string
		hash B.Hash
	}

)

func (i1 *uid) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*uid)
	return BA.CompP(i1.uid, ii2.uid)
}

func sentryTR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	return G.MakeIntValue(B.SentryThreshold())
} //sentryTR

func sentriesR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	ids := A.New()
	var is = new(U.SetIterator)
	pubkey, ok := B.NextSentry(true, &is)
	for ok {
		var b bool
		id := new(uid)
		id.uid, _, id.hash, _, _, _, b = B.IdPubComplete(pubkey); M.Assert(b, 100)
		_, b, _ = ids.SearchIns(id); M.Assert(!b, 101)
		pubkey, ok = B.NextSentry(false, &is)
	}
	e := ids.Next(nil)
	for e != nil {
		l.Append(GQ.Wrap(e.Val().(*uid).hash))
		e = ids.Next(e)
	}
	return l
} //sentriesR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "SentryThreshold", sentryTR)
	ts.FixFieldResolver("Query", "sentries", sentriesR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
} //init
