/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package members

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"
	O	"util/operations"

)

const (
	
	defaultPointsNb = 80
	defaultdegree = 2

)

type (
	
	idList struct {
		next *idList
		in	bool
		id B.Hash
	}
	
	event struct {
		date int64
		block int32
		idL *idList
		number int
	}
	
	events []event
	
	eventR struct {
		block int32
		value float64
	}
	
	eventsR []eventR

)

var (
	
	pointsNb = defaultPointsNb
	degree = defaultdegree

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

func countLimits () (min, max int32) {
	return 0, B.LastBlock()
}

// Number of members, event by event
func doCount (from, to int64) (counts events) {
	M.Assert(from <= to, 20)
	var pst *B.Position
	t := A.New()
	var jb, lb int32
	pk, ok := B.JLNextPubkey(true, &pst)
	for ok {
		_, _, hash, _, _, _, b := B.IdPubComplete(pk); M.Assert(b, 100)
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, lb, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
		}
		for ok2 {
			_, time, b := B.TimeOf(jb); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, block: jb, idL: nil, number: 0})
			ev := e.Val().(*event)
			ev.idL = &idList{next: ev.idL, in: true, id: hash}
			ev.number++
			if lb != B.HasNotLeaved {
				_, time, b = B.TimeOf(lb); M.Assert(b, 102)
				e, _, _ = t.SearchIns(&event{date: time, block: lb, idL: nil, number: 0})
				ev := e.Val().(*event)
				ev.idL = &idList{next: ev.idL, in: false, id: hash}
				ev.number--
			}
			jb, lb, ok2 = B.JLPubLNext(&list)
		}
		pk, ok = B.JLNextPubkey(false, &pst)
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
	if i == -1 {
		i = n
	}
	counts = counts[i:n]
	return
}

// Flux of members, event by event
func doCountFlux (from, to, timeUnit int64) (countsR eventsR, counts events) {
	counts = doCount(from, to)
	dots := make(O.Dots, len(counts))
	for i := 0; i < len(counts); i++ {
		dots[i].SetXY(float64(counts[i].date) / float64(timeUnit), float64(counts[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	countsR = make(eventsR, len(counts))
	for i := 0; i < len(counts); i++ {
		countsR[i].block = counts[i].block
		countsR[i].value = dotsD[i].Y()
	}
	return
}

	// Flux of first entries per member, event by event
func doCountFluxPerMember (from, to, timeUnit int64) (countsR eventsR) {
	countsR, counts := doCountFlux(from, to, timeUnit)
	for c := 0; c < len(countsR); c++ {
		countsR[c].value /= float64(counts[c].number)
	}
	return
}

	// Number of first entries, event by event
func doFirstEntries (from, to int64) (fE events) {
	M.Assert(from <= to, 20)
	var pst *B.Position
	t := A.New()
	var jb, jb0 int32
	pk, ok := B.JLNextPubkey(true, &pst)
	for ok {
		_, _, hash, _, _, _, b := B.IdPubComplete(pk); M.Assert(b, 100)
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, _, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
			for ok2 {
				jb0 = jb
				jb, _, ok2 = B.JLPubLNext(&list)
			}
			_, time, b := B.TimeOf(jb0); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, block: jb0, idL: nil, number: 0})
			ev := e.Val().(*event)
			ev.idL = &idList{next: ev.idL, in: true, id: hash}
			ev.number++
		}
		pk, ok = B.JLNextPubkey(false, &pst)
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
	return
}

// Flux of first entries, event by event
func doFEFlux (from, to, timeUnit int64) (fER eventsR) {
	
	fE := doFirstEntries(from, to)
	dots := make(O.Dots, len(fE))
	for i := 0; i < len(fE); i++ {
		dots[i].SetXY(float64(fE[i].date) / float64(timeUnit), float64(fE[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	fER = make(eventsR, len(fE))
	for i := 0; i < len(fE); i++ {
		fER[i].block = fE[i].block
		fER[i].value = dotsD[i].Y()
	}
	return
}

// Flux of first entries per member, event by event
func doFEFluxPerMember (from, to, timeUnit int64) (fER eventsR) {
	fER = doFEFlux(from, to, timeUnit)
	counts := doCount(from, to)
	c := 0;
	for f := 0; f < len(fER); f++ {
		for c < len(counts) - 1 && counts[c + 1].block <= fER[f].block {
			c++
		}
		fER[f].value /= float64(counts[c].number)
	}
	return
}

// Loss of members, event by event
func doLoss (from, to int64) (losses events) {
	M.Assert(from <= to, 20)
	var pst *B.Position
	t := A.New()
	var jb, lb int32
	pk, ok := B.JLNextPubkey(true, &pst)
	for ok {
		_, _, hash, _, _, _, b := B.IdPubComplete(pk); M.Assert(b, 100)
		var ev0 *event
		list, ok2 := B.JLPub(pk)
		if ok2 {
			jb, lb, ok2 = B.JLPubLNext(&list); M.Assert(ok2, 100)
		}
		for ok2 {
			_, time, b := B.TimeOf(jb); M.Assert(b, 101)
			e, _, _ := t.SearchIns(&event{date: time, block: jb, idL: nil, number: 0})
			ev0 = e.Val().(*event)
			ev0.idL = &idList{next: ev0.idL, in: true, id: hash}
			ev0.number--
			if lb != B.HasNotLeaved {
				_, time, b := B.TimeOf(lb); M.Assert(b, 102)
				e, _, _ := t.SearchIns(&event{date: time, block: lb, number: 0})
				ev := e.Val().(*event)
				ev.idL = &idList{next: ev.idL, in: false, id: hash}
				ev.number++
			}
			jb, lb, ok2 = B.JLPubLNext(&list)
		}
		ev0.idL = ev0.idL.next
		ev0.number++
		pk, ok = B.JLNextPubkey(false, &pst)
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
	losses = losses[i:n]
	return
}

// Flux of losses, event by event
func doLossFlux (from, to, timeUnit int64) (lossesR eventsR) {
	
	losses := doLoss(from, to)
	dots := make(O.Dots, len(losses))
	for i := 0; i < len(losses); i++ {
		dots[i].SetXY(float64(losses[i].date) / float64(timeUnit), float64(losses[i].number))
	}
	dotsD := O.Derive(dots, pointsNb, degree)
	lossesR = make(eventsR, len(losses))
	for i := 0; i < len(losses); i++ {
		lossesR[i].block = losses[i].block
		lossesR[i].value = dotsD[i].Y()
	}
	return
}

// Flux of losses per member, event by event
func doLossFluxPerMember (from, to, timeUnit int64) (lossesR eventsR) {
	lossesR = doLossFlux(from, to, timeUnit)
	counts := doCount(from, to)
	M.Assert(counts != nil && len(counts) == len(lossesR), 100)
	for l := 0; l < len(lossesR); l++ {
		lossesR[l].value /= float64(counts[l].number)
	}
	return
}

func getStartEnd (as *A.Tree) (start, end int64) {
	var v G.Value
	if G.GetValue(as, "start", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			start = v.Int
		case *G.NullValue:
			start = 0
		default:
			M.Halt(v, 100)
		}
	} else {
		start = 0
	}
	if G.GetValue(as, "end", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			end = v.Int
		case *G.NullValue:
			end = M.MaxInt64
		default:
			M.Halt(v, 100)
		}
	} else {
		end = M.MaxInt64
	}
	return
} //getStartEnd

func getTimeUnit (as *A.Tree) (timeUnit int64) {
	const timeUnitDefault = 2629800 // s = 1 month
	var v G.Value
	if G.GetValue(as, "timeUnit", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			timeUnit = v.Int
		case *G.NullValue:
			timeUnit = timeUnitDefault
		default:
			M.Halt(v, 100)
			return
		}
	} else {
		timeUnit = timeUnitDefault
	}
	return
} //getTimeUnit

func countMinR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	min, _ := countLimits()
	return GQ.Wrap(min)
} //countMinR

func countMaxR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	_, max := countLimits()
	return GQ.Wrap(max)
} //countMaxR

func membersCountR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	if start <= end {
		for _, e := range doCount(start, end) {
			l.Append(GQ.Wrap(e))
		}
	}
	return l
} //membersCountR

func membersFluxR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		countsR, _ := doCountFlux(start, end, timeUnit)
		for _, f := range countsR {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //membersFluxR

func membersFluxPMR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		for _, f := range doCountFluxPerMember(start, end, timeUnit) {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //membersFluxPMR

func fECountR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	if start <= end {
		for _, e := range doFirstEntries(start, end) {
			l.Append(GQ.Wrap(e))
		}
	}
	return l
} //fECountR

func fEFluxR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		for _, f := range doFEFlux(start, end, timeUnit) {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //fEFluxR

func fEFluxPMR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		for _, f := range doFEFluxPerMember(start, end, timeUnit) {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //fEFluxPMR

func lossCountR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	if start <= end {
		for _, e := range doLoss(start, end) {
			l.Append(GQ.Wrap(e))
		}
	}
	return l
} //lossCountR

func lossFluxR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		for _, f := range doLossFlux(start, end, timeUnit) {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //lossFluxR

func lossFluxPMR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	start, end := getStartEnd(argumentValues)
	timeUnit := getTimeUnit(argumentValues)
	if start <= end && timeUnit > 0 {
		for _, f := range doLossFluxPerMember(start, end, timeUnit) {
			l.Append(GQ.Wrap(f))
		}
	}
	return l
} //lossFluxPMR

func eventIdListR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	l := G.NewListValue()
	switch ev := GQ.Unwrap(rootValue, 0).(type) {
	case event:
		i := ev.idL
		for i != nil {
			l.Append(GQ.Wrap(i.in, i.id))
			i = i.next
		}
	default:
		M.Halt(ev, 100)
		return nil
	}
	return l
} //eventIdListR

func eventBlockR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch ev := GQ.Unwrap(rootValue, 0).(type) {
	case event:
		return GQ.Wrap(ev.block)
	default:
		M.Halt(ev, 100)
		return nil
	}
} //eventBlockR

func eventNumberR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch ev := GQ.Unwrap(rootValue, 0).(type) {
	case event:
		return G.MakeIntValue(ev.number)
	default:
		M.Halt(ev, 100)
		return nil
	}
} //eventNumberR

func eventIdIdR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch id := GQ.Unwrap(rootValue, 1).(type) {
	case B.Hash:
		return GQ.Wrap(id)
	default:
		M.Halt(id, 100)
		return nil
	}
} //eventIdIdR

func eventIdInOutR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch in := GQ.Unwrap(rootValue, 0).(type) {
	case bool:
		return G.MakeBooleanValue(in)
	default:
		M.Halt(in, 100)
		return nil
	}
} //eventIdInOutR

func fluxBlockR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch evR := GQ.Unwrap(rootValue, 0).(type) {
	case eventR:
		return GQ.Wrap(evR.block)
	default:
		M.Halt(evR, 100)
		return nil
	}
} //fluxBlockR

func fluxValueR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch evR := GQ.Unwrap(rootValue, 0).(type) {
	case eventR:
		return G.MakeFloat64Value(evR.value)
	default:
		M.Halt(evR, 100)
		return nil
	}
} //fluxValueR

func differParamsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	old := GQ.Wrap(pointsNb, degree)
	var v G.Value
	if G.GetValue(argumentValues, "pointsNb", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			pointsNb = int(v.Int)
		case *G.NullValue:
		default:
			M.Halt(v, 100)
		}
	}
	if G.GetValue(argumentValues, "degree", &v) {
		switch v := v.(type) {
		case *G.IntValue:
			degree = int(v.Int)
		case *G.NullValue:
		default:
			M.Halt(v, 100)
		}
	}
	return old
} //differParamsR

func differPointsNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch pN := GQ.Unwrap(rootValue, 0).(type) {
	case int:
		return G.MakeIntValue(pN)
	default:
		M.Halt(pN, 100)
		return nil
	}
} //differPointsNbR

func differDegreeR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 1).(type) {
	case int:
		return G.MakeIntValue(d)
	default:
		M.Halt(d, 100)
		return nil
	}
} //differDegreeR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "countMin", countMinR)
	ts.FixFieldResolver("Query", "countMax", countMaxR)
	ts.FixFieldResolver("Query", "membersCount", membersCountR)
	ts.FixFieldResolver("Query", "membersFlux", membersFluxR)
	ts.FixFieldResolver("Query", "membersFluxPM", membersFluxPMR)
	ts.FixFieldResolver("Query", "fECount", fECountR)
	ts.FixFieldResolver("Query", "fEFlux", fEFluxR)
	ts.FixFieldResolver("Query", "fEFluxPM", fEFluxPMR)
	ts.FixFieldResolver("Query", "lossCount", lossCountR)
	ts.FixFieldResolver("Query", "lossFlux", lossFluxR)
	ts.FixFieldResolver("Query", "lossFluxPM", lossFluxPMR)
	ts.FixFieldResolver("Event", "idList", eventIdListR)
	ts.FixFieldResolver("Event", "block", eventBlockR)
	ts.FixFieldResolver("Event", "number", eventNumberR)
	ts.FixFieldResolver("EventId", "id", eventIdIdR)
	ts.FixFieldResolver("EventId", "inOut", eventIdInOutR)
	ts.FixFieldResolver("FluxEvent", "block", fluxBlockR)
	ts.FixFieldResolver("FluxEvent", "value", fluxValueR)
	ts.FixFieldResolver("Mutation", "changeDifferParams", differParamsR)
	ts.FixFieldResolver("DifferParams", "pointsNb", differPointsNbR)
	ts.FixFieldResolver("DifferParams", "degree", differDegreeR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
}
