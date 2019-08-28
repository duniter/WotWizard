/* 
Duniter1: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package sandboxList

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	S	"duniter/sandbox"

)

const (
	
	sandboxName = "Sandbox"

)

type (
	
	action struct {
		output string
	}

)

func list () J.Json {
	mk := J.NewMaker()
	var e *A.Elem
	
	mk.StartObject()
	
	mk.StartObject()
	mk.StartArray()
	hash, ok := S.IdNextHash(true, &e)
	for ok {
		inBC, pubkey, uid, exp, b := S.IdHash(hash); M.Assert(b, 100)
		mk.StartObject()
		mk.PushString(string(hash))
		mk.BuildField("hash")
		mk.PushString(string(pubkey))
		mk.BuildField("pubkey")
		mk.PushString(uid)
		mk.BuildField("uid")
		mk.PushBoolean(inBC)
		mk.BuildField("inBC")
		mk.PushInteger(exp)
		mk.BuildField("expired_on")
		mk.BuildObject()
		hash, ok = S.IdNextHash(false, &e)
	}
	mk.BuildArray()
	mk.BuildField("id_byHash")
	mk.StartArray()
	pubkey, hash, ok := S.IdNextPubkey(true, &e)
	for ok {
		mk.StartObject()
		mk.PushString(string(pubkey))
		mk.BuildField("pubkey")
		mk.PushString(string(hash))
		mk.BuildField("hash")
		mk.BuildObject()
		pubkey, hash, ok = S.IdNextPubkey(false, &e)
	}
	mk.BuildArray()
	mk.BuildField("id_byPubkey")
	mk.StartArray()
	uid, hash, ok := S.IdNextUid(true, &e)
	for ok {
		mk.StartObject()
		mk.PushString(uid)
		mk.BuildField("uid")
		mk.PushString(string(hash))
		mk.BuildField("hash")
		mk.BuildObject()
		uid, hash, ok = S.IdNextUid(false, &e);
	}
	mk.BuildArray()
	mk.BuildField("id_byUid")
	mk.BuildObject()
	mk.BuildField("identities")
	
	mk.StartObject()
	mk.StartArray()
	var pos S.CertPos
	ok = S.CertNextFrom(true, &pos, &e)
	for ok {
		from, toHash, okP := pos.CertNextPos()
		for okP {
			mk.StartObject()
			mk.PushString(string(from))
			mk.BuildField("from_pubkey")
			if uid, b := B.IdPub(from); b {
				mk.PushString(uid)
			} else {
				mk.PushNull()
			}
			mk.BuildField("from_uid")
			mk.PushString(string(toHash))
			mk.BuildField("to_hash")
			_, _, uid, _, b := S.IdHash(toHash)
			if !b {
				var to B.Pubkey
				to, _, b = S.Cert(from, toHash)
				if b {
					uid, b = B.IdPub(to)
				}
			}
			if b {
				mk.PushString(uid)
			} else {
				mk.PushNull()
			}
			mk.BuildField("to_uid")
			_, exp, b := S.Cert(from, toHash); M.Assert(b, 101)
			mk.PushInteger(exp)
			mk.BuildField("expired_on");
			mk.BuildObject()
			from, toHash, okP = pos.CertNextPos()
		}
		ok = S.CertNextFrom(false, &pos, &e)
	}
	mk.BuildArray()
	mk.BuildField("certFrom")
	mk.StartArray()
	ok = S.CertNextTo(true, &pos, &e)
	for ok {
		from, toHash, okP := pos.CertNextPos()
		for okP {
			mk.StartObject()
			mk.PushString(string(from))
			mk.BuildField("from_pubkey")
			if uid, b := B.IdPub(from); b {
				mk.PushString(uid)
			} else {
				mk.PushNull()
			}
			mk.BuildField("from_uid")
			mk.PushString(string(toHash))
			mk.BuildField("to_hash")
			_, _, uid, _, b := S.IdHash(toHash)
			if !b {
				var to B.Pubkey
				to, _, b = S.Cert(from, toHash)
				if b {
					uid, b = B.IdPub(to)
				}
			}
			if b {
				mk.PushString(uid)
			} else {
				mk.PushNull()
			}
			mk.BuildField("to_uid")
			_, exp, b := S.Cert(from, toHash); M.Assert(b, 102)
			mk.PushInteger(exp)
			mk.BuildField("expired_on")
			mk.BuildObject()
			from, toHash, okP = pos.CertNextPos()
		}
		ok = S.CertNextTo(false, &pos, &e)
	}
	mk.BuildArray()
	mk.BuildField("certTo")
	mk.BuildObject()
	mk.BuildField("certifications")
	
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
	
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	return sandboxName
}

func (a *action) Activate () {
	G.Json(list(), a.output)
}

func do (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{output: output}
}

func init () {
	G.AddAction(sandboxName, do, G.Arguments{})
}
