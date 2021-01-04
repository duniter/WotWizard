/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package qualitiesPrint

// Print qualities

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
		"html/template"

)

const (
	
	distancesName = "20distances"
	qualitiesName  = "21qualities"
	centralitiesName = "22centralities"
	
	queryDistances = `
		query Distances {
			now {
				number
				bct
			}
			identities(status: MEMBER) {
				uid
				distance {
					value
				}
			}
		}
	`
	
	queryQualities = `
		query Qualities {
			now {
				number
				bct
			}
			identities(status: MEMBER) {
				uid
				quality
			}
		}
	`
	
	queryCentralities = `
		query Centralities {
			now {
				number
				bct
			}
			identities(status: MEMBER) {
				uid
				quality: centrality
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
			<p>
				{{range .Qs}}
					{{.}}
					<br>
				{{end}}
			</p>
			<p>
				{{range .QsId}}
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
		Number int
		Bct int64
	}
	
	IdentityT struct {
		Uid string
		Distance struct {
			Value float64
		}
		Quality float64
	}
	
	IdentitiesT []IdentityT
	
	DataT struct {
		Now *NowT
		Identities IdentitiesT
	}
	
	Qualities struct {
		Data *DataT
	}

	propT struct {
		id string
		prop float64
	}
	
	propsT []*propT
	
	propsSort struct { 
		t propsT
	}

	// Ouputs
	
	Quals []string
	
	QualT struct {
		Title,
		Now string
		Qs,
		QsId Quals
	}

)

var (
	
	distancesDoc = GS.ExtractDocument(queryDistances)
	qualitiesDoc = GS.ExtractDocument(queryQualities)
	centralitiesDoc = GS.ExtractDocument(queryCentralities)

)

func (s *propsSort) Swap (i, j int) {
	s.t[i], s.t[j] = s.t[j], s.t[i]
} //Swap

func (s *propsSort) Less (p1, p2 int) bool {
	return s.t[p1].prop > s.t[p2].prop || s.t[p1].prop == s.t[p2].prop && BA.CompP(s.t[p1].id, s.t[p2].id) == BA.Lt
} //Less

func count (dist bool, ids IdentitiesT) (props, propsId propsT) {
	n := len(ids)
	props = make(propsT, n)
	for i, id := range ids {
		p := new(propT)
		p.id = id.Uid
		if dist {
			p.prop = id.Distance.Value
		} else {
			p.prop = id.Quality
		}
		props[i] = p
	}
	propsId = make(propsT, n)
	copy(propsId, props)
	s := &propsSort{t: props}
	var ts = S.TS{Sorter: s}
	ts.QuickSort(0, n - 1)
	return
} //count

func printNow (now *NowT, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct, lang))
} //printNow

func print (qual string, qualities *Qualities, lang *SM.Lang) *QualT {

/*
const (
	
	scatColor = Ports.red
	scatShape = HermesScat.point

)

var (
	
	i, m: int
	s: Views.Title
	axe: ARRAY 2 { rune
	t: TextModels.Model
	f: TextMappers.Formatter
	p: HermesViews.View
	ts: []HermesUtil.Dot
	tr: HermesViews.Trace
	d: Data
	props, propsId: Props
*/

	d := qualities.Data
	now := printNow(d.Now, lang)
	var t string
	switch qual {
	case distancesName:
		t = lang.Map("#duniterClient:distances")
		//axe := "d"
	case qualitiesName:
		t = lang.Map("#duniterClient:qualities")
		//axe := "q"
	case centralitiesName:
		t = lang.Map("#duniterClient:centralities")
		//axe := "c"
	}
	props, propsId := count(qual == distancesName, d.Identities)
	m := len(props)
	qs := make(Quals, m)
	for i, p := range props {
		qs[i] = fmt.Sprintf("%v    %05.2f    %v", i + 1, p.prop, p.id)
	}
	qsId := make(Quals, m)
	for i, p := range propsId {
		qsId[i] = fmt.Sprintf("%v    %05.2f", p.id, p.prop)
	}
	return &QualT{Title: t, Now: now, Qs: qs, QsId: qsId}
	/*
	if props != nil {
		p := HermesViews.New()
		NEW(ts, m)
		for i := 0 TO m - 1 {
			ts[i].x := i
			ts[i].y := props[i].prop
		}
		tr := HermesScat.Insert(p, ts, scatShape, false, false, false)
		tr.ChangeColor(scatColor)
		tr := HermesAxes.Insert(p, false, true, false, false, true, true, false, false, 'n', '', 0, 3)
		tr := HermesAxes.Insert(p, true, true, false, false, true, true, false, false, axe, '%', 0, 3)
		p.FixY(0., 100.)
	}
	*/
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	var doc *G.Document
	switch name {
	case distancesName:
		doc = distancesDoc
	case qualitiesName:
		doc = qualitiesDoc
	case centralitiesName:
		doc = centralitiesDoc
	default:
		M.Halt(name, 100)
	}
	j := GS.Send(nil, doc)
	quals := new(Qualities)
	J.ApplyTo(j, quals)
	temp.ExecuteTemplate(w, name, print(name, quals, lang))
} //end

func init() {
	W.RegisterPackage(distancesName, html, end, true)
	W.RegisterPackage(qualitiesName, html, end, true)
	W.RegisterPackage(centralitiesName, html, end, true)
} //init
