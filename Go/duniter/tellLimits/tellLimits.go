/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package tellLimits

import (
	
	B	"duniter/blockchain"
	E	"duniter/events"
	G	"duniter/gqlReceiver"
	J	"util/json"

)

const (
	
	day = 24 * 60 * 60
	
	limitsMName = "MemLim"
	limitsCName = "CertLim"
	
	forwardName = "forward";
	
	memA = iota
	certA

)

type (
	
	action struct {
		what,
		forward int
		output string
	}
	
	doProc func (int) E.Memberships

)

func selectM (fw int) E.Memberships {
	ms := E.DoMembershipsEnds()
	t := (B.Now() + int64(fw) * day) / day * day
	i := 0
	for i < len(ms) && ms[i].Exp() < t {
		i++
	}
	if i == len(ms) {
		return make(E.Memberships, 0)
	}
	t += day
	j := i
	for j < len(ms) && ms[j].Exp() < t {
		j++
	}
	n := j - i
	ms2 := make(E.Memberships, n)
	for j := 0; j < n; j++ {
		ms2[j] = ms[i]
		i++
	}
	return ms2
}

func selectC (fw int) E.Memberships {
	ms := E.DoCertifsEnds()
	t := (B.Now() + int64(fw) * day) / day * day
	i := 0
	for i < len(ms) && ms[i].Exp() < t {
		i++
	}
	if i == len(ms) {
		return make(E.Memberships, 0)
	}
	t += day
	j := i
	for j < len(ms) && ms[j].Exp() < t {
		j++
	}
	n := j - i
	ms2 := make(E.Memberships, n)
	for j := 0; j < n; j++ {
		ms2[j] = ms[i]
		i++
	}
	return ms2
}

func jsonCommon (fw int, mk *J.Maker) {
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mt := B.Now()
	mk.PushInteger(mt)
	mk.BuildField("now")
	mt += int64(fw) * day
	mt = mt / day * day
	mk.PushInteger(mt)
	mk.BuildField("next")
}

func listCertifiers (id string, mk *J.Maker) {
	certifiers := B.AllCertifiers(id)
	mk.StartArray()
	if certifiers != nil {
		for _, c := range certifiers {
			mk.StartObject()
			mk.PushString(c)
			mk.BuildField("certifier")
			mk.BuildObject()
		}
	}
	mk.BuildArray()
}

func list (fw int, do doProc) J.Json {
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	ms := do(fw)
	for _, m := range ms {
		mk.StartObject()
		mk.PushString(m.Id())
		mk.BuildField("id")
		mk.PushInteger(m.Exp())
		mk.BuildField("limit")
		listCertifiers(m.Id(), mk)
		mk.BuildField("certifiers")
		mk.BuildObject()
	}
	mk.BuildArray()
	mk.BuildField("limits")
	jsonCommon(fw, mk)
	mk.BuildObject()
	return mk.GetJson()
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case memA:
		s = limitsMName
	case certA:
		s = limitsCName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case memA:
		G.Json(list(a.forward, selectM), a.output)
	case certA:
		G.Json(list(a.forward, selectC), a.output)
	}
}

func member (forward int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: memA, forward: int(forward), output: output}
}

func certifs (forward int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: certA, forward: int(forward), output: output}
}

func init () {
	G.AddAction(limitsMName, member, G.Arguments{forwardName})
	G.AddAction(limitsCName, certifs, G.Arguments{forwardName})
}
