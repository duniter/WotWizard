/* 
Duniter1: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identities

import (
	
	BA	"duniter/basic"
	B	"duniter/blockchain"
	BT	"util/gbTree"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"

)

const (
	
	revokedName = "IdentitiesRevoked"
	missingName = "IdentitiesMissing"
	membersName = "IdentitiesMembers"
	
	revokedAction = iota
	missingAction
	membersAction

)

type (
		
	filter func (member bool, expires_on int64) bool
	
	action struct {
		what int
		output string
	}

)

func list (f filter) J.Json {
	var ir *BT.IndexReader
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	pubkey, ok := B.IdNextPubkey(true, &ir)
	for ok {
		uid, member, _, bnb, _, exp, b := B.IdPubComplete(pubkey); M.Assert(b, 100)
		if f(member, exp) {
			mk.StartObject()
			mk.PushString(string(pubkey))
			mk.BuildField("pubkey")
			mk.PushString(uid)
			mk.BuildField("uid")
			mt, _, b := B.TimeOf(bnb); M.Assert(b, 101)
			mk.PushInteger(mt)
			mk.BuildField("time")
			mk.BuildObject()
		}
		pubkey, ok = B.IdNextPubkey(false, &ir)
	}
	mk.BuildArray()
	mk.BuildField("by_pubkey")
	mk.StartArray()
	uid, ok := B.IdNextUid(true, &ir)
	for ok {
		pubkey, member, _, bnb, _, exp, b := B.IdUidComplete(uid); M.Assert(b, 102)
		if f(member, exp) {
			mk.StartObject()
			mk.PushString(uid)
			mk.BuildField("uid")
			mk.PushString(string(pubkey))
			mk.BuildField("pubkey")
			mt, _, b := B.TimeOf(bnb); M.Assert(b, 103)
			mk.PushInteger(mt)
			mk.BuildField("time")
			mk.BuildObject()
		}
		uid, ok = B.IdNextUid(false, &ir)
	}
	mk.BuildArray()
	mk.BuildField("by_uid")
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mt := B.Now()
	mk.PushInteger(mt)
	mk.BuildField("now")
	mk.BuildObject()
	return mk.GetJson()
}

func filterRevoked (member bool, exp int64) bool {
	return exp == BA.Revoked
}

func filterMissing (member bool, exp int64) bool {
	return !member && (exp != BA.Revoked)
}

func filterMembers (member bool, exp int64) bool {
	return member
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case revokedAction:
		s = revokedName
	case missingAction:
		s = missingName
	case membersAction:
		s = membersName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case revokedAction:
		G.Json(list(filterRevoked), a.output)
	case missingAction:
		G.Json(list(filterMissing), a.output)
	case membersAction:
		G.Json(list(filterMembers), a.output)
	}
}

func revoked (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: revokedAction, output: output}
}

func missing (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: missingAction, output: output}
}

func members (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: membersAction, output: output}
}

func init () {
	G.AddAction(revokedName, revoked, G.Arguments{})
	G.AddAction(missingName, missing, G.Arguments{})
	G.AddAction(membersName, members, G.Arguments{})
}
