/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package gqlSender

import (
	
	F	"path/filepath"
	G	"util/graphQL"
	J	"util/json"
	M	"util/misc"
	R	"util/resources"
		"errors"
		"fmt"
		"net/http"
		"io/ioutil"
		"os"
		"text/scanner"
		"strings"

)

const (
	
	serverDefaultAddress = ":8080"
	serverAddressName = "serverAddress.txt"
	
	errNoServer = 1

)

var (
	
	wd = R.FindDir()
	serverAddress = serverDefaultAddress

)

func send (request string) J.Json {
	r, err := http.Post("http://" + serverAddress, "text/plain", strings.NewReader(request))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in duniterClient/gqlSender.send:", err)
		os.Exit(errNoServer)
	}
	M.Assert(r.StatusCode / 100 == 2, r.StatusCode, 101)
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	M.Assert(err == nil, err, 102)
	return J.ReadString(string(b))
} //send

func Send (j J.Json, doc *G.Document) J.Json {
	w := new(strings.Builder)
	if j != nil {
		fmt.Fprint(w, "#")
		j.WriteFlat(w)
		fmt.Fprintln(w)
	}
	doc.WriteFlat(w)
	j = send(w.String())
	M.Assert(j != nil && (len(j.(*J.Object).Fields) == 0 || j.(*J.Object).Fields[0].Name != "errors"), 100)
	return j
} //Send

func SendSub (name string) J.Json {
	return send("@" + name)
} //SendSub

func ExtractDocument (docString string) *G.Document {
	doc, err := G.Compile(strings.NewReader(docString))
	M.Assert(doc != nil, err, 100)
	return doc
} //ExtractDocument

func fixServerAddress () {
	dir := F.Join(wd, "duniterClient")
	err := os.MkdirAll(dir, 0777); M.Assert(err == nil, err, 100)
	name := F.Join(dir, serverAddressName)
	f, err := os.Open(name)
	if err == nil {
		defer f.Close()
		s := new(scanner.Scanner)
		s.Init(f)
		s.Error = func(s *scanner.Scanner, msg string) {panic(errors.New("File " + name + " incorrect"))}
		s.Mode = scanner.ScanStrings
		s.Scan()
		ss := s.TokenText()
		M.Assert(ss[0] == '"' && ss[len(ss) - 1] == '"', ss, 101)
		serverAddress = ss[1:len(ss) - 1]
	} else {
		f, err := os.Create(name)
		M.Assert(err == nil, err, 102)
		defer f.Close()
		fmt.Fprint(f, "\"" + serverAddress + "\"")
	}
}

func init () {
	fixServerAddress()
}
