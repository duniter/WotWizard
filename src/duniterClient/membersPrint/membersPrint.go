/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package membersPrint

import (
	
	BA	"duniterClient/basicPrint"
	G	"util/graphQL"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"strconv"
		"strings"
		"html/template"

)

const (
	
	defaultDelay = "7"
	
	countTName = "400membersCountT"
	countGName = "403membersCountG"
	countFName = "404membersCountFlux"
	countFPMName = "405membersCountFluxPM"
	firstETName = "401membersFirstEntryT"
	firstEGName = "406membersFirstEntryG"
	fEFluxName = "407membersFEFlux"
	fEFluxPMName = "408membersFEFluxPM"
	lossTName = "402membersLossT"
	lossGName = "409membersLossG"
	lossFluxName = "410membersLossFlux"
	lossFluxPMName = "411membersLossFluxPM"
	
	queryNow = `
		query MembersNow {
			now {
				number
				bct
			}
		}
	`
	
	queryCountMax = `
		query CountMax {
			countMax {
				utc0
			}
		}
	`
	
	queryCount = `
		query MembersCount ($start: Int64, $end: Int64, $text: Boolean! = true) {
			now {
				number
				bct
			}
			events: membersCount(start: $start, end: $end) {
				block {
					number @include(if: $text)
					utc0
					bct @include(if: $text)
				}
				idList @include(if: $text) {
					in: inOut
					id {
						uid
					}
				}
				number
			}
		}
	`
	
	queryCountF = `
		query MembersCountFlux ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: membersFlux(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	queryCountFPM = `
		query MembersCountFluxPM ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: membersFluxPM(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	queryFirstE = `
		query MembersFirstEntry ($start: Int64, $end: Int64, $text: Boolean! = true) {
			now {
				number
				bct
			}
			events: fECount(start: $start, end: $end) {
				block {
					number @include(if: $text)
					utc0
					bct @include(if: $text)
				}
				idList @include(if: $text) {
					in: inOut
					id {
						uid
					}
				}
				number
			}
		}
	`
	
	queryFEFlux = `
		query MembersFEFlux ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: fEFlux(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	queryFEFluxPM = `
		query MembersFEFluxPM ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: fEFluxPM(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	queryLoss = `
		query MembersLoss ($start: Int64, $end: Int64, $text: Boolean! = true) {
			now {
				number
				bct
			}
			events: lossCount(start: $start, end: $end) {
				block {
					number @include(if: $text)
					utc0
					bct @include(if: $text)
				}
				idList @include(if: $text) {
					in: inOut
					id {
						uid
					}
				}
				number
			}
		}
	`
	
	queryLossFlux = `
		query MembersLossFlux ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: lossFlux(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	queryLossFluxPM = `
		query MembersLossFluxPM ($start: Int64, $end: Int64, $timeUnit: Int64) {
			now {
				number
				bct
			}
			eventsR: lossFluxPM(start: $start, end: $end, timeUnit: $timeUnit) {
				block {
					utc0
				}
				value
			}
		}
	`
	
	html = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<form action="" method="post">
				<p>
					<label for="delay">{{.Delay}}: </label>
					<input type="text" name="delay" id="delay" value="{{.Period}}"/>
				</p>
				<p>
					<input type="submit" value="{{.Submit}}">
				</p>
			</form>
			{{if .List}}
				{{range .List}}
					<p>
						{{.Dates}}
						<blockquote>
							{{range .InOuts}}
								{{.}}
								<br>
							{{end}}
							{{.Number}}
						</blockquote>
					</p>
				{{end}}
			{{end}}
			{{if .ListT}}
				<p>
					{{range .ListT}}
						{{.}}
						<br>
					{{end}}
				</p>
				<p>
				***********************************
				</p>
			{{end}}
			{{if .ListG}}
				<p>
					{{range .ListG}}
						{{.}}
						<br>
					{{end}}
				</p>
			{{end}}
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`
	
	day = 60 * 60 * 24
	month = (365 * 4 + 1) * day / 4 / 12
	
	//scatColor = 00BF00BFH
	//scatShape = HermesScat.point

)

type (
	
	MaxT struct {
		Data struct {
			CountMax struct {
				Utc0 int64
			}
		}
	}
	
	NowT struct {
		Number int
		Bct,
		Utc0 int64
	}
	
	BlockNow struct {
		Data struct {
			Now *NowT
		}
	}
	
	IdT struct {
		In bool
		Id struct {
			Uid string
		}
	}
	
	IdListT []IdT
	
	EventT struct {
		Block *NowT
		IdList IdListT
		Number int
	}
	
	EventsT []EventT
	
	BlockEvents struct {
		Data struct {
			Now *NowT
			Events EventsT
		}
	}
	
	EventRT struct {
		Block *NowT
		Value float64
	}
	
	EventsRT []EventRT
	
	BlockEventsR struct {
		Data struct {
			Now *NowT
			EventsR EventsRT
		}
	}
	
	// Outputs
	
	InOutsT []string
	
	Evt struct {
		Dates string
		InOuts InOutsT
		Number int
	}
	
	ListE []*Evt
	
	ListS []string
	
	Out struct {
		Title,
		Now,
		Delay,
		Period,
		Submit string
		List ListE
		ListT,
		ListG ListS 
	}

)

var (
	
	nowDoc = GS.ExtractDocument(queryNow)
	maxDoc = GS.ExtractDocument(queryCountMax)
	countDoc = GS.ExtractDocument(queryCount)
	countFDoc = GS.ExtractDocument(queryCountF)
	countFPMDoc = GS.ExtractDocument(queryCountFPM)
	firstEDoc = GS.ExtractDocument(queryFirstE)
	fEFluxDoc = GS.ExtractDocument(queryFEFlux)
	fEFluxPMDoc = GS.ExtractDocument(queryFEFluxPM)
	lossDoc = GS.ExtractDocument(queryLoss)
	lossFluxDoc = GS.ExtractDocument(queryLossFlux)
	lossFluxPMDoc = GS.ExtractDocument(queryLossFluxPM)
	
	//tUnit string

)

func printNow (now *NowT) string {
	return fmt.Sprint(SM.Map("#duniterClient:Block"), " ", now.Number, " ", BA.Ts2s(now.Bct))
} //printNow

func printN (now *NowT, title, period string) *Out {
	t := SM.Map(title)
	nowS := printNow(now)
	dy := SM.Map("#duniterClient:Delay")
	s := SM.Map("#duniterClient:OK")
	return &Out{Title: t, Now: nowS, Delay: dy, Period: period, Submit: s}
} //printN

func printT (a EventsT, now *NowT, title, period string) *Out {
	t := SM.Map(title)
	nowS := printNow(now)
	dy := SM.Map("#duniterClient:Delay")
	s := SM.Map("#duniterClient:OK")
	block := SM.Map("#duniterClient:Block");
	actual := SM.Map("#duniterClient:Utc")
	median := SM.Map("#duniterClient:Bct")
	entry := SM.Map("#duniterClient:Entry")
	exit := SM.Map("#duniterClient:Exit")
	la := len(a)
	l := make(ListE, la)
	la--
	for i, ai := range a {
		d := fmt.Sprint(block, ": ", ai.Block.Number,  "    ", actual, ": ", BA.Ts2s(ai.Block.Utc0), "    ", median, ": ", BA.Ts2s(ai.Block.Bct))
		in := make(InOutsT, len(ai.IdList))
		for j, id := range ai.IdList {
			w := new(strings.Builder)
			if id.In {
				fmt.Fprint(w, entry)
			} else {
				fmt.Fprint(w, exit)
			}
			fmt.Fprint(w, id.Id.Uid)
			in[j] = w.String()
		}
		l[la - i] = &Evt{Dates: d, InOuts: in, Number: ai.Number}
	}
	return &Out{Title: t, Now: nowS, Delay: dy, Period: period, Submit: s, List: l}
} //printT

func printG (a EventsT, now *NowT, title, period, label, unit string) *Out {
	t := SM.Map(title)
	nowS := printNow(now)
	dy := SM.Map("#duniterClient:Delay")
	s := SM.Map("#duniterClient:OK")
	var t0 int64
	if len(a) > 0 {
		t0 = a[0].Block.Utc0
	}
	l := make(ListS, len(a))
	for i, ai := range a {
		l[i] = fmt.Sprintf("%21.16f    %v", float64(ai.Block.Utc0 - t0) / float64(month), ai.Number)
	}
	return &Out{Title: t, Now: nowS, Delay: dy, Period: period, Submit: s, ListG: l}
	
	/*
	p := HermesViews.New()
	NEW(ts, m)
	for i := 0 TO m - 1 {
		ts[i].x := (a[i].block.utc0 - t0) / month
		ts[i].y := a[i].number
	}
	tr := HermesScat.Insert(p, ts, scatShape, false, false, false)
	tr.ChangeColor(scatColor)
	tr := HermesAxes.Insert(p, false, true, false, false, true, true, false, false, 't', tUnit, 0, 3)
	tr := HermesAxes.Insert(p, true, true, false, false, true, true, false, false, label, unit, 0, 3)
	Views.OpenAux(p, s)
	*/
} //printG

func printR (a EventsRT, now *NowT, title, period, label, unit string, percent bool) *Out {
	t := SM.Map(title)
	nowS := printNow(now)
	dy := SM.Map("#duniterClient:Delay")
	s := SM.Map("#duniterClient:OK")
	lT := make(ListS, len(a))
	lG := make(ListS, len(a))
	var t0 int64
	if len(a) > 0 {
		t0 = a[0].Block.Utc0
	}
	for i, ai := range a {
		lT[i] = fmt.Sprint(BA.Ts2s(ai.Block.Utc0), "    ", ai.Value)
		lG[i] = fmt.Sprintf("%21.16f    %v", float64(ai.Block.Utc0 - t0) / float64(month), ai.Value)
	}
	return &Out{Title: t, Now: nowS, Delay: dy, Period: period, Submit: s, ListT: lT, ListG: lG}
	
	/*
	p := HermesViews.New()
	NEW(ts, m)
	for i := 0 TO m - 1 {
		ts[i].x := (a[i].block.utc0 - t0) / month
		if percent {
			ts[i].y := 100 * a[i].value
		} else {
			ts[i].y := a[i].value
		}
	}
	tr := HermesScat.Insert(p, ts, scatShape, false, false, false)
	tr.ChangeColor(scatColor)
	tr := HermesAxes.Insert(p, false, true, false, false, true, true, false, false, 't', tUnit, 0, 3)
	tr := HermesAxes.Insert(p, true, true, false, false, true, true, false, false, label, unit, 0, 3)
	Views.OpenAux(p, s)
	*/
} //printR

func end (name string, temp *template.Template, r *http.Request, w http.ResponseWriter) {
	const (
		
		pT = iota
		pG
		pR
	
	)
	
	var (
		
		doc *G.Document
		j J.Json = nil
		m = new(MaxT)
		n = new(BlockNow)
		b = new(BlockEvents)
		br = new(BlockEventsR)
		t string
		pc bool
		print int8
		out *Out
	
	)
	
	switch name {
	case countTName:
		t = "#duniterClient:MembersNb"
	case countGName:
		t = "#duniterClient:MembersNb"
	case countFName:
		t = "#duniterClient:MembersFlux"
	case countFPMName:
		t = "#duniterClient:MembersFluxPM"
	case firstETName:
		t = "#duniterClient:FirstEntries"
	case firstEGName:
		t = "#duniterClient:FirstEntries"
	case fEFluxName:
		t = "#duniterClient:FirstEntriesFlux"
	case fEFluxPMName:
		t = "#duniterClient:FirstEntriesFluxPM"
	case lossTName:
		t = "#duniterClient:Losses"
	case lossGName:
		t = "#duniterClient:Losses"
	case lossFluxName:
		t = "#duniterClient:LossesFlux"
	case lossFluxPMName:
		t = "#duniterClient:LossesFluxPM"
	default:
		M.Halt(name, 100)
	}
	if r.Method == "GET" {
		j = GS.Send(nil, nowDoc)
		J.ApplyTo(j, n)
		out = printN(n.Data.Now, t, defaultDelay)
		temp.ExecuteTemplate(w, name, out)
	} else {
		r.ParseForm()
		delayS := r.PostFormValue("delay")
		mk := J.NewMaker()
		mk.StartObject()
		if delayS != "" {
			delay, err := strconv.Atoi(delayS)
			if err != nil || delay < 0 {
				http.Redirect(w, r, "/" + name, http.StatusFound)
				return
			}
			j = GS.Send(nil, maxDoc)
			J.ApplyTo(j, m)
			start := m.Data.CountMax.Utc0 - int64(delay) * day
			mk.PushInteger(start)
			mk.BuildField("start")
		}
		switch name {
		case countTName:
			doc = countDoc
			print = pT
		case countGName:
			mk.PushBoolean(false)
			mk.BuildField("text")
			doc = countDoc
			print = pG
		case countFName:
			doc = countFDoc
			pc = false
			print = pR
		case countFPMName:
			doc = countFPMDoc
			pc = true
			print = pR
		case firstETName:
			doc = firstEDoc
			print = pT
		case firstEGName:
			mk.PushBoolean(false)
			mk.BuildField("text")
			doc = firstEDoc
			print = pG
		case fEFluxName:
			doc = fEFluxDoc
			pc = false
			print = pR
		case fEFluxPMName:
			doc = fEFluxPMDoc
			pc = true
			print = pR
		case lossTName:
			doc = lossDoc
			print = pT
		case lossGName:
			mk.PushBoolean(false)
			mk.BuildField("text")
			doc = lossDoc
			print = pG
		case lossFluxName:
			doc = lossFluxDoc
			pc = false
			print = pR
		case lossFluxPMName:
			doc = lossFluxPMDoc
			pc = true
			print = pR
		default:
			M.Halt(name, 100)
		}
		mk.BuildObject()
		j = mk.GetJson()
		j = GS.Send(j, doc)
		switch print {
		case pT:
			J.ApplyTo(j, b)
			out = printT(b.Data.Events, b.Data.Now, t, delayS)
		case pG:
			J.ApplyTo(j, b)
			out = printG(b.Data.Events, b.Data.Now, t, delayS, "", "")
		case pR:
			J.ApplyTo(j, br)
			out = printR(br.Data.EventsR, br.Data.Now, t, delayS, "", "", pc)
		}
		temp.ExecuteTemplate(w, name, out)
	}
} //end

func init() {
	W.RegisterPackage(countTName, html, end, true)
	W.RegisterPackage(countGName, html, end, true)
	W.RegisterPackage(countFName, html, end, true)
	W.RegisterPackage(countFPMName, html, end, true)
	W.RegisterPackage(firstETName, html, end, true)
	W.RegisterPackage(firstEGName, html, end, true)
	W.RegisterPackage(fEFluxName, html, end, true)
	W.RegisterPackage(fEFluxPMName, html, end, true)
	W.RegisterPackage(lossTName, html, end, true)
	W.RegisterPackage(lossGName, html, end, true)
	W.RegisterPackage(lossFluxName, html, end, true)
	W.RegisterPackage(lossFluxPMName, html, end, true)
} //init
