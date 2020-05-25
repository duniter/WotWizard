/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package blocks

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"

)

func nowStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return GQ.CreateStream("now", rootValue, argumentValues)
} //nowStreamResolver

func nowStopR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	GQ.Unsubscribe("now")
	return nil
} //wwStopR

func nowR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	return GQ.Wrap(B.LastBlock())
} //nowR

func blockNumberR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch block := GQ.Unwrap(rootValue, 0).(type) {
	case int32:
		return G.MakeIntValue(int(block))
	case *G.NullValue:
		return block
	default:
		M.Halt(block, 100)
		return nil
	}
} //blockNumberR

func blockBctR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch block := GQ.Unwrap(rootValue, 0).(type) {
	case int32:
		mtime, _, ok := B.TimeOf(block); M.Assert(ok, 100)
		return G.MakeInt64Value(mtime)
	case *G.NullValue:
		return block
	default:
		M.Halt(block, 100)
		return nil
	}
} //blockBctR

func blockUtcRR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch block := GQ.Unwrap(rootValue, 0).(type) {
	case int32:
		_, time, ok := B.TimeOf(block); M.Assert(ok, 100)
		return G.MakeInt64Value(time)
	case *G.NullValue:
		return block
	default:
		M.Halt(block, 100)
		return nil
	}
} //blockUtcRR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "now", nowR)
	ts.FixFieldResolver("Block", "number", blockNumberR)
	ts.FixFieldResolver("Block", "bct", blockBctR)
	ts.FixFieldResolver("Block", "utc0", blockUtcRR)
	ts.FixFieldResolver("Subscription", "now", nowR)
	ts.FixFieldResolver("Mutation", "nowStop", nowStopR)
} //fixFieldResolvers

func fixStreamResolvers (ts G.TypeSystem) {
	ts.FixStreamResolver("now", nowStreamResolver)
}

func init () {
	ts := GQ.TS()
	fixFieldResolvers(ts)
	fixStreamResolvers(ts)
} //init
