/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package web

import (
	
	A	"util/avl"
	BA	"duniterClient/basicPrint"
	F	"path/filepath"
	M	"util/misc"
	R	"util/resources"
	SM	"util/strMapping"
		"errors"
		"fmt"
		"net/http"
		"os"
		"text/scanner"
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
		{{define "head"}}<title>{{Map "Index"}}</title>{{end}}
		{{define "body"}}
			<h1>{{Map "Index"}}</h1>
			{{range $name, $temp := .}}
				<p>
					<a href="/{{$name}}">{{Map $name}}</a> 
				</p>
			{{end}}
		{{end}}
	`
	
	authorizationsName = "Authorizations.txt"

)

type (
	
	executeFunc func (name string, temp *template.Template, r *http.Request, w http.ResponseWriter)
	
	pack struct {
		temp *template.Template
		call executeFunc
	}
	
	stringComp string

)

var (
	
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

func (s1 stringComp) Compare (s2 A.Comparer) A.Comp {
	ss2 := s2.(stringComp)
	switch {
	case s1 < ss2:
		return A.Lt
	case s1 > ss2:
		return A.Gt
	default:
		return A.Eq
	}
}

func initAuthorizations () {
	name := F.Join(R.FindDir(), "duniterClient", authorizationsName)
	f, err := os.Open(name)
	if err == nil {
		defer f.Close()
		s := new(scanner.Scanner)
		s.Init(f)
		s.Error = func(s *scanner.Scanner, msg string) {panic(errors.New("File " + name + " incorrect"))}
		s.Mode = scanner.ScanStrings
		auth := make(map[string] *int)
		for s.Scan() != scanner.EOF {
			ss := s.TokenText()
			M.Assert(ss[0] == '"' && ss[len(ss) - 1] == '"', ss, 101)
			auth[ss[1:len(ss) - 1]] = nil
		}
		for view, _ := range packagesD {
			if _, ok := auth[view]; !ok {
				delete(packages, view)
				delete(packagesD, view)
			}
		}
	} else {
		f, err := os.Create(name)
		M.Assert(err == nil, err, 102)
		defer f.Close()
		t := A.New()
		for view, _ := range packagesD {
			t.SearchIns(stringComp(view))
		}
		e := t.Next(nil)
		for e != nil {
			fmt.Fprintln(f, "\"" + e.Val().(stringComp) + "\"")
			e = t.Next(e)
		}
	}
}

func getHandler (name string, p *pack) http.HandlerFunc {
	
	return func (w http.ResponseWriter, r *http.Request) {
		p.call(name, p.temp, r, w)
	}
	
}

func Start () {
	initAuthorizations()
	r := http.NewServeMux()
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
