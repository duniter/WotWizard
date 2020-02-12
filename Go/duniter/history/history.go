/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package history

import (
	
	B	"duniter/blockchain"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"

)

const (
	
	historyName = "History"
	
	uidName = "Uid"

)

type (
	
	action struct {
		id,
		output string
	}
	
	event struct {
		next *event
		block int
		date int64
	}
	
	Event struct {
		Block int
		Date int64
	}
	
	History []Event

)

func BuildHistoryP (pubkey B.Pubkey) History {
	var ev *event = nil
	var jb, lb int32
	n := 0
	list, ok := B.JLPub(pubkey)
	if ok {
		jb, lb, ok = B.JLPubLNext(&list); M.Assert(ok, 100)
	}
	for ok {
		var b bool
		if lb != B.HasNotLeaved {
			ev = &event{next: ev}
			ev.block = int(lb)
			ev.date, _, b = B.TimeOf(lb); M.Assert(b, 102)
			n++
		}
		ev = &event{next: ev}
		ev.block = int(jb)
		ev.date, _, b = B.TimeOf(jb); M.Assert(b, 101)
		n++
		jb, lb, ok = B.JLPubLNext(&list)
	}
	evts := make(History, n)
	n = 0
	for ev != nil {
		evts[n] = Event{Block: ev.block, Date: ev.date}
		n++
		ev = ev.next
	}
	return evts
}

func BuildHistoryI (id string) History {
	if pk, ok := B.IdUid(id); ok {
		return BuildHistoryP(pk)
	}
	return nil
}

func BuildHistoryH (hash B.Hash) History {
	if pk, ok := B.IdHash(hash); ok {
		return BuildHistoryP(pk)
	}
	return nil
}

func listBlock (mk *J.Maker) {
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
}

func listHistory (id string) J.Json {
	evts := BuildHistoryI(id)
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushString(id)
	mk.BuildField("uid")
	in := true
	mk.StartArray()
	for _, ev := range evts {
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
	listBlock(mk)
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	return historyName
}

func (a *action) Activate () {
	j := listHistory(a.id)
	G.Json(j, a.output)
}

func history (id, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{id: id, output: output}
}

func init () {
	G.AddAction(historyName, history, G.Arguments{uidName})
}
