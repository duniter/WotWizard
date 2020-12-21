/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package sentriesPrint



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
		"html/template"

)

const (
	
	sentriesName = "57sentries"
	
	querySentries = `
		query Sentries {
			now {
				number
				bct
			}
			sentryThreshold
			sentries {
				uid
			}
		}
	`
	
	htmlSentries = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">index</a>
			</p>
			<h3>
				{{.Block}}
			</h3>
			<p>
				{{.Threshold}}
			<br>
				{{.Number}}
			</p>
			<p>
				{{range .Ids}}
					{{.}}
					<br>
				{{end}}
			</p>
			<p>
				<a href = "/">index</a>
			</p>
		{{end}}
	`

)

type (
	
	NowT struct {
		Number int
		Bct int64
	}
	
	Identity struct {
		Uid string
	}
	
	Identities []Identity
	
	DataT struct {
		Now *NowT
		SentryThreshold int
		Sentries Identities
	}

	SentriesT struct {
		Data *DataT
	}
	
	Disp0 []string
	
	Disp struct {
		Title,
		Block,
		Threshold,
		Number string
		Ids Disp0
	}

)

var (
	
	sentriesDoc *G.Document

)

func printNow (now *NowT) string {
	return fmt.Sprint(SM.Map("#duniterClient:Block"), " ", now.Number, " ", BA.Ts2s(now.Bct))
} //printNow

func print (sentries *SentriesT) *Disp {
	d := sentries.Data
	t := fmt.Sprint(SM.Map("#duniterClient:Threshold"), " = ", d.SentryThreshold)
	ids := d.Sentries
	n := fmt.Sprint(SM.Map("#duniterClient:SentriesNb"), " = ", len(ids))
	dd := make(Disp0, len(ids))
	for i, id := range(ids) {
		dd[i] = id.Uid
	}
	return &Disp{Title: SM.Map("#duniterClient:Sentries"), Block: printNow(d.Now), Threshold: t, Number: n, Ids: dd}
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	M.Assert(name == sentriesName, name, 100)
	j := GS.Send(nil, sentriesDoc)
	sentries := new(SentriesT)
	J.ApplyTo(j, sentries)
	temp.ExecuteTemplate(w, name, print(sentries))
} //end

func init() {
	sentriesDoc = GS.ExtractDocument(querySentries)
	W.RegisterPackage(sentriesName, htmlSentries, end, true)
} //init
