/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package tellLimitsPrint

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
		"util/sort"
		"strconv"
		"html/template"

)

const (
	
	limitsMName = "33memLim"
	limitsCName = "34certLim"
	
	day = 24 * 60 * 60
	
	forwardM = 14
	forwardC = 60
	
	queryNow = `
		query Now {
			now {
				number
				bct
			}
		}
	`
	
	queryMemLim = `
		query MemLim {
			now {
				number
				bct
			}
			identities(status: MEMBER) {
				uid
				limitDate
			}
		}
	`
	
	queryCertLim = `
		query CertLim {
			now {
				number
				bct
			}
			idSearch(with: {status_list: [MISSING, MEMBER]}) {
				ids {
					uid
					received_certifications {
						limit
					}
				}
			}
		}
	`
	
	queryAllCertifiers = `
		query AllCertifiers ($uid: String!) {
			idSearch(with: {hint: $uid, status_list: [MISSING, MEMBER]}) {
				ids {
					uid
					all_certifiers {
						uid
					}
				}
			}
		}
	`
	
	html = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">index</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<form action="" method="post">
				<p>
					<label for="forward">{{.Forward}}: </label>
					<input type="text" name="forward" id="forward" value="{{.Fw}}"/>
				</p>
				<p>
					<input type="submit" value="{{.OK}}">
				</p>
			</form>
			{{if .Warnings}}
				{{range .Warnings}}
					<p>
						{{.WTitle}}
					</p>
					{{if .Comp}}
						<p>
							{{.Comp}}
						</p>
					{{end}}
					<p>
						<blockquote>
							{{.Content}}
						</blockquote>
					</p>
					<p>
						{{.Cert}}
						<br>
						{{range .Certs}}
							{{.}}
							<br>
						{{end}}
					</p>
				{{end}}
			{{end}}
			<p>
				<a href = "/">index</a>
			</p>
		{{end}}
	`

)

type (
	
	Block struct {
		Number int
		Bct int64
	}
	
	Identity struct {
		Uid string
		LimitDate int64
		Received_certifications struct {
			Limit int64
		}
		All_certifiers Identities
	}
	
	Identities []Identity
	
	Data struct {
		Now *Block
		Identities Identities
		IdSearch struct {
			Ids Identities
		}
	}
	
	Limits struct {
		Data Data
	}
	
	Certifiers []string
	
	prop struct {
		id string
		aux Certifiers
		prop int64
	}
	
	props []*prop
	
	propsSort struct {
		t props
	}
	
	// Outputs
	
	Warning struct {
		WTitle,
		Comp,
		Content,
		Cert string
		Certs Certifiers
	}
	
	Warnings []*Warning
	
	Out struct {
		Title,
		Now,
		Forward,
		OK string
		Fw int
		Warnings Warnings
	}

)

var (
	
	nowDoc = GS.ExtractDocument(queryNow)
	limitsMDoc = GS.ExtractDocument(queryMemLim)
	limitsCDoc = GS.ExtractDocument(queryCertLim)
	allCertifiersDoc = GS.ExtractDocument(queryAllCertifiers)

)

func next (date int64, fw int) int64 {
	return (date + int64(fw) * day) / day * day
} //next

func (s *propsSort) Swap (i, j int) {
	s.t[i], s.t[j] = s.t[j], s.t[i]
} //Swap

func (s *propsSort) Less (p1, p2 int) bool {
	return s.t[p1].prop < s.t[p2].prop || s.t[p1].prop == s.t[p2].prop && BA.CompP(s.t[p1].id, s.t[p2].id) == BA.Lt
} //Less

func doBlock (b *Block) string {
	return fmt.Sprint(SM.Map("#duniterClient:Block"), " ", b.Number, " ", BA.Ts2s(b.Bct))
} //doBlock

func printN (t string, now *Block, forward string, fw int) *Out {
	tE := SM.Map(t)
	forwardE := SM.Map(forward)
	nowS := fmt.Sprint(doBlock(now), " → ", BA.Ts2s(next(now.Bct, fw)))
	return &Out{Title: tE, Now: nowS, Forward: forwardE, Fw: fw, OK: SM.Map("#duniterClient:OK")}
} //printN

func print (t string, now *Block, forward string, fw int, props props, title, comp, cert string) *Out {
	tE := SM.Map(t)
	titleE := SM.Map(title)
	certE := SM.Map(cert)
	forwardE := SM.Map(forward)
	var compE string
	if comp == "" {
		compE = ""
	} else {
		compE = SM.Map(comp)
	}
	nowS := fmt.Sprint(doBlock(now), " → ", BA.Ts2s(next(now.Bct, fw)))
	ok := SM.Map("#duniterClient:OK")
	warnings := make(Warnings, len(props))
	for i, p := range props {
		content := fmt.Sprint(BA.Ts2s(p.prop), "    ", p.id)
		warnings[i] = &Warning{titleE, compE, content, certE, p.aux}
	}
	return &Out{tE, nowS, forwardE, ok, fw, warnings}
} //print

func makeCount (now *Block, what string, ids Identities, fw int) props {
	
	DoCount := func () props {
		n := len(ids)
		m := n
		if what == limitsCName {
			m = 0
			for _, id := range ids {
				if id.Received_certifications.Limit != 0 {
					m++
				}
			}
		}
		props := make(props, m)
		m = 0
		for _, id := range ids {
			p := new(prop)
			p.id = id.Uid
			if what == limitsCName {
				p.prop = id.Received_certifications.Limit
			} else {
				p.prop = id.LimitDate
			}
			if p.prop != 0 {
				props[m] = p
				m++
			}
		}
		var ts sort.TS
		ts.Sorter = &propsSort{t: props}
		ts.QuickSort(0, m - 1)
		
		t := next(now.Bct, fw)
		i := 0
		for (i < len(props)) && (props[i].prop < t) {
			i++
		}
		if i == len(props) {
			return props[0:0]
		}
		t += day
		k := i
		for (k < len(props)) && (props[k].prop < t) {
			k++
		}
		return props[i:k]
	} //DoCount

	//makeCount
	props := DoCount()
	for k, p := range props {
		mk := J.NewMaker()
		mk.StartObject()
		mk.PushString(p.id)
		mk.BuildField("uid")
		mk.BuildObject()
		j := GS.Send(mk.GetJson(), allCertifiersDoc)
		cs := new(Limits)
		J.ApplyTo(j, cs)
		ids := cs.Data.IdSearch.Ids
		i := 0
		for (i < len(ids)) && (ids[i].Uid != p.id) {
			i++
		}
		M.Assert(i < len(ids), 100)
		certifiers := ids[i].All_certifiers
		c := make(Certifiers, len(certifiers))
		for i, cc := range certifiers {
			c[i] = cc.Uid
		}
		props[k].aux = c
	}
	return props
} //makeCount

func end (name string, temp *template.Template, r *http.Request, w http.ResponseWriter) {

	var (
		
		doc *G.Document
		title,
		comp,
		cert,
		forward string
		fw int
		ids Identities
		
	)
	
	t := "#duniterClient:" + name
	forward = "#duniterClient:Forward"
	switch name {
	case limitsMName:
		title = "#duniterClient:LimitsM"
		comp = "#duniterClient:LMComplement"
		cert = "#duniterClient:LMCertifiers"
		fw = forwardM
		doc = limitsMDoc
	case limitsCName:
		title = "#duniterClient:LimitsC"
		comp = "#duniterClient:LCComplement"
		cert = "#duniterClient:LCCertifiers"
		fw = forwardC
		doc = limitsCDoc
	default:
		M.Halt(name, 100)
	}
	if r.Method == "GET" {
		j := GS.Send(nil, nowDoc)
		n := new(Limits)
		J.ApplyTo(j, n)
		out := printN(t, n.Data.Now, forward, fw)
		err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 100)
	} else {
		r.ParseForm()
		fwS := r.PostFormValue("forward")
		if fwS != "" {
			var err error
			fw, err = strconv.Atoi(fwS)
			if err != nil || fw < 0 {
				http.Redirect(w, r, "/" + name, http.StatusFound)
				return
			}
		}
		j := GS.Send(nil, doc)
		lim := new(Limits)
		J.ApplyTo(j, lim)
		switch name {
		case limitsMName:
			ids = lim.Data.Identities
		case limitsCName:
			ids = lim.Data.IdSearch.Ids
		}
		props := makeCount(lim.Data.Now, name, ids, fw)
		out := print(t, lim.Data.Now, forward, fw, props, title, comp, cert)
		err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 101)
	}
} //end

func init() {
	W.RegisterPackage(limitsMName, html, end, true)
	W.RegisterPackage(limitsCName, html, end, true)
} //init
