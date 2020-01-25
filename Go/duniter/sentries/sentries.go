/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package sentries

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	U	"util/sets2"
	
)

const (
	
	sentriesName = "Sentries"

)

type (
	
	action struct {
		output string
	}
	
	uid struct {
		uid string
	}

)

func (i1 *uid) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*uid)
	return BA.CompP(i1.uid, ii2.uid)
}

func list () J.Json {
	m := J.NewMaker()
	m.StartObject()
	m.PushInteger(int64(B.SentryTreshold()))
	m.BuildField("threshold")
	ids := A.New()
	var is = new(U.SetIterator)
	pubkey, ok := B.NextSentry(true, &is)
	for ok {
		var b bool
		id := new(uid)
		id.uid, b = B.IdPub(pubkey); M.Assert(b, 100)
		_, b, _ = ids.SearchIns(id); M.Assert(!b, 101)
		pubkey, ok = B.NextSentry(false, &is)
	}
	m.StartArray()
	e := ids.Next(nil)
	for e != nil {
		m.StartObject()
		m.PushString(e.Val().(*uid).uid)
		m.BuildField("name")
		m.BuildObject()
		e = ids.Next(e)
	}
	m.BuildArray()
	m.BuildField("sentries")
	m.PushInteger(int64(B.LastBlock()))
	m.BuildField("block")
	m.PushInteger(B.Now())
	m.BuildField("now")
	m.BuildObject()
	return m.GetJson()
}

func (a *action) Name () string {
	return sentriesName
}

func (a *action) Activate () {
	G.Json(list(), a.output)
}

func do (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{output: output}
}

func init () {
	G.AddAction(sentriesName, do, G.Arguments{})
}
