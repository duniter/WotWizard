/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package web

import (
	
	BA	"duniterClient/basicPrint"
	M	"util/misc"
	R	"util/resources"
	SM	"util/strMapping"
		"fmt"
		"net/http"
		"github.com/gorilla/mux"
		"html/template"

)

const (
	
	base = `
		<html>
			<head>{{template "head" .}}</head>
			<body>{{template "body" .}}</body>
		</html>
	`
	
	htmlIndex = `
		{{define "head"}}<title>Index</title>{{end}}
		{{define "body"}}
			<h1>Index</h1>
			{{range $name, $temp := .}}
				<p>
					<a href="/{{$name}}">{{Map $name}}</a> 
				</p>
			{{end}}
		{{end}}
	`

)

type (
	
	executeFunc func (name string, temp *template.Template, r *http.Request, w http.ResponseWriter)
	
	pack struct {
		temp *template.Template
		call executeFunc
	}

)

var (
	
	wd = R.FindDir()
	
	packages = make(map[string] *pack)
	packagesD = make(map[string] *pack)

)

func RegisterPackage (name, temp string, call executeFunc, displayed bool) {
	funcMap := make(template.FuncMap)
	funcMap["Map"] = func (name string) string {return SM.Map("#duniterClient:" + name)}
	p := new(pack)
	p.temp = template.New(name)
	p.temp = p.temp.Funcs(funcMap)
	p.temp = template.Must(p.temp.Parse(temp))
	p.temp = template.Must(p.temp.Parse(base))
	p.call = call
	packages[name] = p
	if displayed {
		packagesD[name] = p
	}
}

func getHandler (name string, p *pack) http.HandlerFunc {
	
	return func (w http.ResponseWriter, r *http.Request) {
		p.call(name, p.temp, r, w)
	}
	
}

func Start () {
	r := mux.NewRouter().StrictSlash(false)
	for name, p := range packages {
		if name == "index" {
			r.HandleFunc("/", getHandler(name, p))
		} else {
			r.HandleFunc("/" + name, getHandler(name, p))
		}
	}
	htmlAddress := BA.HtmlAddress()
	server := &http.Server{
		Addr: htmlAddress,
		Handler: r,
	}
	fmt.Println("Listening on", htmlAddress, "...")
	server.ListenAndServe()
}

func manageIndex (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	err := temp.ExecuteTemplate(w, name, packagesD)
	M.Assert(err == nil, err, 100)
}

func init () {
	RegisterPackage("index", htmlIndex, manageIndex, false)
}
