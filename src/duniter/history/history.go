/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package history

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	IS	"duniter/identitySearchList"
	M	"util/misc"

)

type (
	
	event struct {
		next *event
		block int32
	}

)

func buildHistory (pubkey B.Pubkey) *event {
	var ev *event = nil
	var jb, lb int32
	list, ok := B.JLPub(pubkey)
	if ok {
		jb, lb, ok = B.JLPubLNext(&list); M.Assert(ok, 100)
	}
	for ok {
		if lb != B.HasNotLeaved {
			ev = &event{next: ev}
			ev.block = lb
		}
		ev = &event{next: ev}
		ev.block = jb
		jb, lb, ok = B.JLPubLNext(&list)
	}
	return ev
} //buildHistory

func identityHistoryR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch hash := GQ.Unwrap(rootValue, 0).(type) {
	case B.Hash:
		l := G.NewListValue()
		_, pub, _, _, _, inBC, _, ok := IS.Get(hash); M.Assert(ok, 100)
		if inBC {
			ev := buildHistory(pub)
			in := true
			for ev != nil {
				l.Append(GQ.Wrap(in, ev.block))
				in = !in
				ev = ev.next
			}
		}
		return l
	case *G.NullValue:
		return hash
	default:
		M.Halt(hash, 100)
		return nil
	}
} //identityHistoryR

func historyEvInR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch in := GQ.Unwrap(rootValue, 0).(type) {
	case bool:
		return G.MakeBooleanValue(in)
	default:
		M.Halt(in, 100)
		return nil
	}
} //historyEvInR

func historyEvBlockR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch block := GQ.Unwrap(rootValue, 1).(type) {
	case int32:
		return GQ.Wrap(block)
	default:
		M.Halt(block, 100)
		return nil
	}
} //historyEvBlockR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Identity", "history", identityHistoryR)
	ts.FixFieldResolver("HistoryEvent", "in", historyEvInR)
	ts.FixFieldResolver("HistoryEvent", "block", historyEvBlockR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
} //init
