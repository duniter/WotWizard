/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package members

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BT	"util/gbTree"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	O	"util/operations"

)
	
const (
	
	countName = "MembersCount"
	countNameAll = "MembersCountAll"
	countFName = "MembersCountFlux"
	countFNameAll = "MembersCountFluxAll"
	countFPMName = "MembersCountFluxPM"
	countFPMNameAll = "MembersCountFluxPMAll"
	firstEName = "MembersFirstEntry"
	firstENameAll = "MembersFirstEntryAll"
	fEFluxName = "MembersFEFlux"
	fEFluxNameAll = "MembersFEFluxAll"
	fEFluxPMName = "MembersFEFluxPM"
	fEFluxPMNameAll = "MembersFEFluxPMAll"
	lossName = "MembersLoss"
	lossNameAll = "MembersLossAll"
	lossFluxName = "MembersLossFlux"
	lossFluxNameAll = "MembersLossFluxAll"
	lossFluxPMName = "MembersLossFluxPM"
	lossFluxPMNameAll = "MembersLossFluxPMAll"
	
	fromName = "from"
	toName = "to"
	timeUnitName = "timeUnit"
	
	countA = iota
	countFA
	countFPMA
	firstEntriesA
	fEFluxA
	fEFluxPerMemberA
	lossA
	lossFluxA
	lossFluxPMA

)

type (
	
	action struct {
		what int
		from,
		to,
		timeUnit int64
		output string
	}
	
	uidList struct {
		next *uidList
		in	bool
		uid string
	}
	
	event struct {
		date,
		mTime int64
		uidL *uidList
		number int
	}
	
	events []event
	
	eventsR []struct {
		date int64
		number float64
	}

)

func (ev1 *event) Compare (ev2 A.Comparer) A.Comp {
	eev2 := ev2.(*event)
	if ev1.date < eev2.date {
		return A.Lt
	}
	if ev1.date > eev2.date {
		return A.Gt
	}
	return A.Eq
}

func listBlock (mk *J.Maker) {
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
}

func listEvents (evts events, mk *J.Maker) {
	mk.StartArray()
	for _, ev := range evts {
		mk.StartObject()
		mk.PushInteger(ev.date)
		mk.BuildField("date")
		mk.PushInteger(ev.mTime)
		mk.BuildField("mTime")
		mk.StartArray()
		uidL := ev.uidL
		for uidL != nil {
			mk.StartObject()
			mk.PushBoolean(uidL.in)
			mk.BuildField("in")
			mk.PushString(uidL.uid)
			mk.BuildField("uid")
			mk.BuildObject()
			uidL = uidL.next
		} 
		mk.BuildArray()
		mk.BuildField("uidList")
		mk.PushInteger(int64(ev.number))
		mk.BuildField("number")
		mk.BuildObject()
	}
	mk.BuildArray()
}

func listEventsR (evts eventsR, mk *J.Maker) {
	mk.StartArray()
	for _, ev := range evts {
		mk.StartObject()
		mk.PushInteger(ev.date)
		mk.BuildField("date")
		mk.PushFloat(ev.number)
		mk.BuildField("number")
		mk.BuildObject()
	}
	mk.BuildArray()
}

// Number of members, event by event
func doCount (from, to int64, json bool) (counts events, j J.Json) {
	M.Assert(from <= to, 20)
	var ir *BT.IndexReader
	t := A.New()
	var jb, lb int32
	pk, ok := B.JLNextPubkey(true, &ir)
	for ok {
		uid, b := B.IdPub(pk); M.Assert(b, 100)
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, lb, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
		}
		for ok2 {
			mTime, time, b := B.TimeOf(jb); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, mTime: mTime, uidL: nil, number: 0})
			ev := e.Val().(*event)
			uidL := &uidList{next: ev.uidL, in: true, uid: uid}
			ev.uidL = uidL
			ev.number++
			if lb != B.HasNotLeaved {
				mTime, time, b = B.TimeOf(lb); M.Assert(b, 102)
				e, _, _ = t.SearchIns(&event{date: time, mTime: mTime, uidL: nil, number: 0})
				ev := e.Val().(*event)
				uidL := &uidList{next: ev.uidL, in: false, uid: uid}
				ev.uidL = uidL
				ev.number--
			}
			jb, lb, ok2 = B.JLPubLNext(&list)
		}
		pk, ok = B.JLNextPubkey(false, &ir)
	}
	counts = make(events, t.NumberOfElems())
	n := 0; i := -1
	e := t.Next(nil)
	for e != nil {
		ev := e.Val().(*event)
		if i < 0 && ev.date >= from {
			i = n
		}
		if ev.date >= to {break}
		counts[n] = *ev
		n++
		e = t.Next(e)
	}
	for k := 1; k < n; k++ {
		counts[k].number += counts[k - 1].number
	}
	counts = counts[i:n]
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEvents(counts, mk)
	mk.BuildField("events")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Flux of members, event by event
func doCountFlux (from, to, timeUnit int64, json bool) (countsR eventsR, j J.Json) {

const (
	
	pointsNb = 80
	degree = 2

)
	counts, _ := doCount(from, to, false)
	dots := make(O.Dots, len(counts))
	for i := 0; i < len(counts); i++ {
		dots[i].SetXY(float64(counts[i].date) / float64(timeUnit), float64(counts[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	countsR = make(eventsR, len(counts))
	for i := 0; i < len(counts); i++ {
		countsR[i].date = counts[i].date
		countsR[i].number = dotsD[i].Y()
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(countsR, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

	// Flux of first entries per member, event by event
func doCountFluxPerMember (from, to, timeUnit int64, json bool) (countsR eventsR, j J.Json) {
	countsR, _ = doCountFlux(from, to, timeUnit, false)
	counts, _ := doCount(from, to, false)
	for c := 0; c < len(countsR); c++ {
		countsR[c].number /= float64(counts[c].number)
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(countsR, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

	// Number of first entries, event by event
func doFirstEntries (from, to int64, json bool) (fE events, j J.Json) {
	M.Assert(from <= to, 20)
	var ir *BT.IndexReader
	t := A.New()
	var jb, jb0 int32
	pk, ok := B.JLNextPubkey(true, &ir)
	for ok {
		uid, b := B.IdPub(pk); M.Assert(b, 100)
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, _, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
			for ok2 {
				jb0 = jb
				jb, _, ok2 = B.JLPubLNext(&list)
			}
			mTime, time, b := B.TimeOf(jb0); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, mTime: mTime, uidL: nil, number: 0})
			ev := e.Val().(*event)
			uidL := &uidList{next: ev.uidL, in: true, uid: uid}
			ev.uidL = uidL
			ev.number++
		}
		pk, ok = B.JLNextPubkey(false, &ir)
	}
	fE = make(events, t.NumberOfElems())
	n := 0; i := -1
	e := t.Next(nil)
	for e != nil {
		ev := e.Val().(*event)
		if i < 0 && ev.date >= from {
			i = n
		}
		if ev.date >= to {break}
		fE[n] = *ev
		n++
		e = t.Next(e)
	}
	for k := 1; k < n; k++ {
		fE[k].number += fE[k - 1].number
	}
	fE = fE[i:n]
	if !json {
		return;
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEvents(fE, mk)
	mk.BuildField("events")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Flux of first entries, event by event
func doFEFlux (from, to, timeUnit int64, json bool) (fER eventsR,  j J.Json) {
	
	const (
		
		pointsNb = 80
		degree = 2
	
	)
	
	fE, _ := doFirstEntries(from, to, false)
	dots := make(O.Dots, len(fE))
	for i := 0; i < len(fE); i++ {
		dots[i].SetXY(float64(fE[i].date) / float64(timeUnit), float64(fE[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	fER = make(eventsR, len(fE))
	for i := 0; i < len(fE); i++ {
		fER[i].date = fE[i].date
		fER[i].number = dotsD[i].Y()
	}
	if !json {
		return;
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(fER, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Flux of first entries per member, event by event
func doFEFluxPerMember (from, to, timeUnit int64, json bool) (fER eventsR, j J.Json) {
	fER, _ = doFEFlux(from, to, timeUnit, false)
	counts, _ := doCount(from, to, false)
	c := 0;
	for f := 0; f < len(fER); f++ {
		for c < len(counts) - 1 && counts[c + 1].date <= fER[f].date {
			c++
		}
		fER[f].number /= float64(counts[c].number)
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(fER, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Loss of members, event by event
func doLoss (from, to int64, json bool) (losses events, j J.Json) {
	M.Assert(from <= to, 20)
	var ir *BT.IndexReader
	t := A.New()
	var jb, lb int32
	pk, ok := B.JLNextPubkey(true, &ir)
	for ok {
		uid, b := B.IdPub(pk); M.Assert(b, 100)
		var ev0 *event
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, lb, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
		}
		for ok2 {
			mTime, time, b := B.TimeOf(jb); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, mTime: mTime, uidL: nil, number: 0})
			ev0 = e.Val().(*event)
			uidL := &uidList{next: ev0.uidL, in: true, uid: uid}
			ev0.uidL = uidL
			ev0.number--
			if lb != B.HasNotLeaved {
				_, time, b := B.TimeOf(lb); M.Assert(b, 102)
				e, _, _ := t.SearchIns(&event{date: time, number: 0})
				ev := e.Val().(*event)
				uidL := &uidList{next: ev.uidL, in: false, uid: uid}
				ev.uidL = uidL
				ev.number++
			}
			jb, lb, ok2 = B.JLPubLNext(&list)
		}
		ev0.uidL = ev0.uidL.next
		ev0.number++
		pk, ok = B.JLNextPubkey(false, &ir)
	}
	losses = make(events, t.NumberOfElems())
	n := 0; i := -1
	e := t.Next(nil)
	for e != nil {
		ev := e.Val().(*event)
		if i < 0 && ev.date >= from {
			i = n
		}
		if ev.date >= to {break}
		losses[n] = *ev
		n++
		e = t.Next(e)
	}
	for k := 1; k < n; k++ {
		losses[k].number += losses[k - 1].number
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEvents(losses, mk)
	mk.BuildField("events")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Flux of losses, event by event
func doLossFlux (from, to, timeUnit int64, json bool) (lossesR eventsR, j J.Json) {
	
	const (
		
		pointsNb = 80
		degree = 2
	
	)
	
	losses, _ := doLoss(from, to, false)
	dots := make(O.Dots, len(losses))
	for i := 0; i < len(losses); i++ {
		dots[i].SetXY(float64(losses[i].date) / float64(timeUnit), float64(losses[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	lossesR = make(eventsR, len(losses))
	for i := 0; i < len(losses); i++ {
		lossesR[i].date = losses[i].date
		lossesR[i].number = dotsD[i].Y()
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(lossesR, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

// Flux of losses per member, event by event
func doLossFluxPerMember (from, to, timeUnit int64, json bool) (lossesR eventsR, j J.Json) {
	lossesR, _ = doLossFlux(from, to, timeUnit, false)
	counts, _ := doCount(from, to, false)
	M.Assert(counts != nil && len(counts) == len(lossesR), 100)
	for l := 0; l < len(lossesR); l++ {
		lossesR[l].number /= float64(counts[l].number)
	}
	if !json {
		return
	}
	mk := J.NewMaker()
	mk.StartObject()
	listEventsR(lossesR, mk)
	mk.BuildField("eventsR")
	mk.PushInteger(timeUnit)
	mk.BuildField("time_unit")
	listBlock(mk)
	mk.BuildObject()
	j = mk.GetJson()
	return
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case countA:
		s = countName
	case countFA:
		s = countFName
	case countFPMA:
		s = countFPMName
	case firstEntriesA:
		s = firstEName
	case fEFluxA:
		s = fEFluxName
	case fEFluxPerMemberA:
		s = fEFluxPMName
	case lossA:
		s = lossName
	case lossFluxA:
		s = lossFluxName
	case lossFluxPMA:
		s = lossFluxPMName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case countA:
		_, j := doCount(a.from, a.to, true)
		G.Json(j, a.output)
	case countFA:
		_, j := doCountFlux(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	case countFPMA:
		_, j := doCountFluxPerMember(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	case firstEntriesA:
		_, j := doFirstEntries(a.from, a.to, true)
		G.Json(j, a.output)
	case fEFluxA:
		_, j := doFEFlux(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	case fEFluxPerMemberA:
		_, j := doFEFluxPerMember(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	case lossA:
		_, j := doLoss(a.from, a.to, true)
		G.Json(j, a.output)
	case lossFluxA:
		_, j := doLossFlux(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	case lossFluxPMA:
		_, j := doLossFluxPerMember(a.from, a.to, a.timeUnit, true)
		G.Json(j, a.output)
	}
}

func count (from, to int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: countA, from: from, to: to, output: output}
}

func countAll (output string, newAction chan<- B.Actioner, fields ...string) {
	count(M.MinInt64, M.MaxInt64, output, newAction, fields...)
}

func countFlux (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: countFA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func countFluxAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	countFlux(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func countFluxPerMember (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: countFPMA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func countFluxPerMemberAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	countFluxPerMember(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func firstEntries (from, to int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: firstEntriesA, from: from, to: to, output: output}
}

func firstEntriesAll (output string, newAction chan<- B.Actioner, fields ...string) {
	firstEntries(M.MinInt64, M.MaxInt64, output, newAction, fields...)
}

func fEFlux (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: fEFluxA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func fEFluxAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	fEFlux(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func fEFluxPerMember (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: fEFluxPerMemberA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func fEFluxPerMemberAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	fEFluxPerMember(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func loss (from, to int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: lossA, from: from, to: to, output: output}
}

func lossAll (output string, newAction chan<- B.Actioner, fields ...string) {
	loss(M.MinInt64, M.MaxInt64, output, newAction, fields...)
}

func lossFlux (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: lossFluxA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func lossFluxAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	lossFlux(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func lossFluxPerMember (from, to, timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: lossFluxPMA, from: from, to: to, timeUnit: timeUnit, output: output}
}

func lossFluxPerMemberAll (timeUnit int64, output string, newAction chan<- B.Actioner, fields ...string) {
	lossFluxPerMember(M.MinInt64, M.MaxInt64, timeUnit, output, newAction, fields...)
}

func init () {
	G.AddAction(countName, count, G.Arguments{fromName, toName})
	G.AddAction(countNameAll, countAll, G.Arguments{})
	G.AddAction(countFName, countFlux, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(countFNameAll, countFluxAll, G.Arguments{timeUnitName})
	G.AddAction(countFPMName, countFluxPerMember, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(countFPMNameAll, countFluxPerMemberAll, G.Arguments{timeUnitName})
	G.AddAction(firstEName, firstEntries, G.Arguments{fromName, toName})
	G.AddAction(firstENameAll, firstEntriesAll, G.Arguments{})
	G.AddAction(fEFluxName, fEFlux, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(fEFluxNameAll, fEFluxAll, G.Arguments{timeUnitName})
	G.AddAction(fEFluxPMName, fEFluxPerMember, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(fEFluxPMNameAll, fEFluxPerMemberAll, G.Arguments{timeUnitName})
	G.AddAction(lossName, loss, G.Arguments{fromName, toName})
	G.AddAction(lossNameAll, lossAll, G.Arguments{})
	G.AddAction(lossFluxName, lossFlux, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(lossFluxNameAll, lossFluxAll, G.Arguments{timeUnitName})
	G.AddAction(lossFluxPMName, lossFluxPerMember, G.Arguments{fromName, toName, timeUnitName})
	G.AddAction(lossFluxPMNameAll, lossFluxPerMemberAll, G.Arguments{timeUnitName})
}
