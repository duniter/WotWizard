/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package language

import (
	
	M	"util/misc"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"net/http"
		"html/template"

)

const (
	
	languageName = "language"
	
	html = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<form action="" method="post">
				<p>
					<select name="language" id="language" >
						{{range .Languages}}
							<option value="{{.Code}}"{{.Selected}}>{{.Name}}</option>
						{{end}}
					</select>
					<label for="language">{{.Select}}</label>
				</p>
				<p>
					<input type="submit" value="{{.OK}}">
				</p>
			</form>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`

)

type (
	
	stringList []string
	
	// Outputs
	
	lang struct {
		Code,
		Selected,
		Name string
	}
	
	langList []lang
	
	Out struct {
		Title,
		Select,
		OK string
		Languages langList
	}

)

var (
	
	languages = stringList{"en", "fr"}

)

func end (name string, temp *template.Template, r *http.Request, w http.ResponseWriter) {
	M.Assert(name == languageName, name, 100)
	if r.Method != "GET" {
		r.ParseForm()
		newLang := r.PostFormValue("language")
		SM.Reinit(newLang)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	t := SM.Map("#duniterClient:language")
	sel := SM.Map("#duniterClient:Select")
	ok := SM.Map("#duniterClient: OK")
	l := make(langList, len(languages))
	language := SM.Language()
	for i, lang := range languages {
		l[i].Code = lang
		if lang == language {
			l[i].Selected = "selected"
		} else {
			l[i].Selected = ""
		}
		l[i].Name = SM.Map("#duniterClient:" + lang)
	}
	out := &Out{t, sel, ok, l}
	err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 101)
} //end

func init() {
	W.RegisterPackage(languageName, html, end, true)
} //init
