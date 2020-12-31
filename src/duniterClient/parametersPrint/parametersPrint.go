/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package parametersPrint

import (
	
	G	"util/graphQL"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"strings"
		"html/template"
		"time"
	
)

const (
	
	allParametersName = "50allParameters"
		
	queryAllParameters = `
		query AllParameters {
			allParameters {
				name
				par_type
				comment
				value
			}
		}
	`
	
	htmlAllParameters = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			{{range .Pars}}
				<p>
				{{range .}}
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
	
	min = 60
	hour = 60 * min
	day = 24 * hour
	year = 365 * day + day / 4
	month = year / 12

)

type (
	
	Parameter0 struct {
		Name,
		Par_type,
		Comment string
		Value float64
	}
	
	AllParameters0 []Parameter0
	
	AllParametersT struct {
		Data struct {
			AllParameters AllParameters0 "allParameters"
		}
	}
	
	Disp0 [2]string
	
	Disp1 []Disp0
	
	Disp struct {
		Title string
		Pars Disp1
	}

)

var (
	
	allParametersDoc *G.Document
	
	times = [...]int64{min, hour, day, month, year}
	tNames [6]string

)

func norm (t int64) (tN int, tName string) {
	i := 4
	for t < times[i] {
		i--
	}
	if i >= 0 {
		tN = int(t / times[i])
	}
	tName = tNames[i + 1]
	return
} //norm

func print (allParameters *AllParametersT) *Disp {
	pars := allParameters.Data.AllParameters
	d := make(Disp1, len(pars))
	for i, p := range pars {
		d[i][0] = p.Comment
		w := new(strings.Builder)
		fmt.Fprint(w, p.Name, " = ")
		if p.Par_type == "FLOAT" {
			fmt.Fprint(w, p.Value, " (", 100 * p.Value, "%)")
		} else {
			val := int64(p.Value + 0.5)
			fmt.Fprint(w, val)
			if p.Par_type == "DURATION" {
				tN, tName := norm(val)
				fmt.Fprint(w, " s (", tN, " ",  tName, ")")
			} else if p.Par_type == "DATE" {
				dt := time.Unix(val, 0)
				fmt.Fprint(w, " (", dt, ")")
			}
		}
		d[i][1] = w.String()
	}
	return &Disp{Title: SM.Map("#duniterClient:Parameters"), Pars: d}
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	j := GS.Send(nil, allParametersDoc)
	if name == allParametersName {
		allParameters := new(AllParametersT)
		J.ApplyTo(j, allParameters)
		temp.ExecuteTemplate(w, name, print(allParameters))
	}
} //end

func init() {
	tNames = [...]string{SM.Map("#duniterClient:second"), SM.Map("#duniterClient:minute"), SM.Map("#duniterClient:hour"), SM.Map("#duniterClient:day"), SM.Map("#duniterClient:month"), SM.Map("#duniterClient:year")}
	allParametersDoc = GS.ExtractDocument(queryAllParameters)
	W.RegisterPackage(allParametersName, htmlAllParameters, end, true)
} //init
