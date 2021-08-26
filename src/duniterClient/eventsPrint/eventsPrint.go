/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package eventsPrint

import (
	
	BA	"duniterClient/basicPrint"
	G	"util/graphQL"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	S	"util/sort"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"strings"
		"html/template"

)

const (
	
	limitsMemberName = "30limitsMember"
	limitsMissingName = "31limitsMissing"
	limitsCertsName = "32limitsCerts"
	
	queryLimitsM = `
		query LimitsM ($status: Identity_Status!) {
			now {
				number
				bct
			}
			identities(status: $status) {
				uid
				limitDate
			}
		}
	`
	
	queryLimitsC = `
		query LimitsC {
			now {
				number
				bct
			}
			idSearch(with: {status_list: [MISSING, MEMBER]}) {
				ids {
					uid
					status
					received_certifications {
						limit
					}
				}
			}
		}
	`
	
	htmlLimits = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<p>
				{{range .Ps}}
					{{.}}
					<br>
				{{end}}
			</p>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`

)

type (
	
	NowT struct {
		Number int32
		Bct int64
	}
	
	IdentityT struct {
		Uid,
		Status string
		LimitDate int64
		Received_certifications struct {
			Limit int64
		}
	}
	
	IdentitiesT []IdentityT
	
	DataT struct {
		Now *NowT
		Identities IdentitiesT
		IdSearch struct {
			Ids IdentitiesT
		}
	}
	
	Limits struct {
		Data *DataT
	}

	propT struct {
		id string
		aux bool
		prop int64
	}
	
	propsT []*propT
	
	propsSort struct {
		t propsT
	}
	
	// Outputs
	
	PropList []string
	
	Disp struct {
		Now,
		Title string
		Ps PropList
	}

)

var (
	
	limitsMDoc = GS.ExtractDocument(queryLimitsM)
	limitsCDoc = GS.ExtractDocument(queryLimitsC)

)

func (s *propsSort) Swap (i, j int) {
	s.t[i], s.t[j] = s.t[j], s.t[i]
} //Swap

func (s *propsSort) Less (p1, p2 int) bool {
	return s.t[p1].prop < s.t[p2].prop || s.t[p1].prop == s.t[p2].prop && BA.CompP(s.t[p1].id, s.t[p2].id) == BA.Lt
} //Less

func count (limits *Limits, what string, ids IdentitiesT) (props propsT) {
	n := len(ids)
	var m int
	if what == limitsCertsName {
		m = 0
		for _, id := range ids {
			if id.Received_certifications.Limit != 0 {
				m++
			}
		}
	} else {
		m = n
	}
	props = make(propsT, m)
	m = 0
	for _, id := range ids {
		p := new(propT)
		p.id = id.Uid
		if what == limitsCertsName {
			p.prop = id.Received_certifications.Limit
			p.aux = id.Status == "MEMBER"
		} else {
			p.prop = id.LimitDate
			p.aux = true
		}
		if p.prop != 0 {
			props[m] = p
			m++
		}
	}
	s := &propsSort{t: props}
	var ts = S.TS{Sorter: s}
	ts.QuickSort(0, m - 1)
	return
} //count

func printNow (now *NowT, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct, lang))
} //printNow

func print (limits *Limits, what, title string, lang *SM.Lang) *Disp {
	d := limits.Data
	now := printNow(d.Now, lang)
	t := lang.Map(title)
	var ids IdentitiesT
	if what == limitsCertsName {
		ids = d.IdSearch.Ids
	} else {
		ids = d.Identities
	}
	props := count(limits, what, ids)
	t = fmt.Sprint(t, " (", len(props), ")")
	ps := make(PropList, len(props))
	for i, p := range props {
		w := new(strings.Builder)
		fmt.Fprint(w, BA.Ts2s(p.prop, lang), BA.SpL)
		if p.aux {
			fmt.Fprint(w, BA.SpS, BA.SpS)
		} else {
			fmt.Fprint(w, BA.OldIcon)
		}
		fmt.Fprint(w, p.id, BA.SpS)
		ps[i] = w.String()
	}
	return &Disp{Now: now, Title: t, Ps: ps}
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	var (j J.Json; doc *G.Document; title string)
	m := J.NewMaker()
	switch name {
	case limitsMemberName:
		title = "#duniterClient:limitsMember"
		m.StartObject()
		m.PushString("MEMBER")
		m.BuildField("status")
		m.BuildObject()
		j = m.GetJson()
		doc = limitsMDoc
	case limitsMissingName:
		title = "#duniterClient:limitsMissing"
		m.StartObject()
		m.PushString("MISSING")
		m.BuildField("status")
		m.BuildObject()
		j = m.GetJson()
		doc = limitsMDoc
	case limitsCertsName:
		title = "#duniterClient:limitsCerts"
		j = nil
		doc = limitsCDoc
	default:
		M.Halt(name, 100)
	}
	j = GS.Send(j, doc)
	limits := new(Limits)
	J.ApplyTo(j, limits)
	temp.ExecuteTemplate(w, name, print(limits, name, title, lang))
} //end

func init() {
	W.RegisterPackage(limitsMemberName, htmlLimits, end, true)
	W.RegisterPackage(limitsMissingName, htmlLimits, end, true)
	W.RegisterPackage(limitsCertsName, htmlLimits, end, true)
} //init
