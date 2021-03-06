/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package gqlSender

import (
	
	BA	"duniterClient/basicPrint"
	G	"util/graphQL"
	J	"util/json"
	M	"util/misc"
		"fmt"
		"net/http"
		"io/ioutil"
		"strings"
		"sync"
		"time"

)

const (
	
	errNoServer = 1
	
	subscriptionReset = `
		subscription ResetQueryMap {
			now {
				number
			}
		}
	`
	
	sendSleepTime = 1 * time.Second

)

type (
	
	askChan chan J.Json
	askChans []chan<- J.Json
	askMap map[string] askChans
	bufferMap map[string] J.Json

)

var (
	
	subAddress = BA.SubAddress()
	
	askM,
	queryM,
	subsM sync.Mutex
	asks = make(askMap)
	queries = make(bufferMap)
	subs = make(bufferMap)

)

func send (request string) J.Json {
	r, err := http.Post("http://" + BA.ServerAddress(), "application/json", strings.NewReader(request))
	for err != nil {
		//M.Assert(strings.Index(err.Error(), "connection refused") >= 0, err, 100)
		time.Sleep(sendSleepTime)
		r, err = http.Post("http://" + BA.ServerAddress(), "application/json", strings.NewReader(request))
	}
	M.Assert(r.StatusCode / 100 == 2, r.StatusCode, 101)
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	M.Assert(err == nil, err, 102)
	return J.ReadString(string(b))
} //send

func Send (j J.Json, doc *G.Document) J.Json {
	
	wait := func (s string) J.Json {
		var j J.Json = nil
		askM.Lock()
		if cs, ok := asks[s]; ok {
			c := make(askChan)
			asks[s] = append(cs, c)
			askM.Unlock()
			j = <-c
		} else {
			askM.Unlock()
		}
		return j
	}
	
	w := new(strings.Builder)
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushString(subAddress)
	mk.BuildField("returnAddr")
	if j != nil {
		mk.PushJson(j)
		mk.BuildField("variables")
		s := j.GetFlatString()
		fmt.Fprint(w, s)
	}
	s := doc.GetFlatString()
	mk.PushString(s)
	mk.BuildField("query")
	fmt.Fprint(w, s)
	s = w.String()
	if j, ok := queries[s]; ok {
		return j
	}
	if j := wait(s); j != nil {
		return j
	}
	askM.Lock()
	asks[s] = make(askChans, 0)
	askM.Unlock()
	mk.BuildObject()
	j = send(mk.GetJson().GetFlatString())
	M.Assert(j != nil && (len(j.(*J.Object).Fields) == 0 || j.(*J.Object).Fields[0].Name != "errors"), 100)
	queryM.Lock()
	queries[s] = j
	queryM.Unlock()
	askM.Lock()
	for _, c := range asks[s] {
		c <- j
	}
	delete(asks, s)
	askM.Unlock()
	return j
} //Send

func startReset () {
	Send(nil, ExtractDocument(subscriptionReset))
}

func subHandler (_ http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body); M.Assert(err == nil, err, 100)
	j := J.ReadString(string(b)); M.Assert(j != nil, 101)
	o, ok := j.(*J.Object); M.Assert(ok, 102)
	opName, _ := J.GetString(o, "operationName"); M.Assert(opName != "", 103)
	if opName == "ResetQueryMap" {
		asks = make(askMap)
		queries = make(bufferMap)
	} else {
		j, ok := J.GetJson(o, "result"); M.Assert(ok,104)
		subsM.Lock()
		subs[opName] = j
		subsM.Unlock()
	}
}

func initReceiver () {
	r := http.NewServeMux()
	r.HandleFunc("/", subHandler)
	server := &http.Server{
		Addr: subAddress,
		Handler: r,
	}
	server.ListenAndServe()
}

func GetSub (opName string) J.Json {
	j, ok := subs[opName]
	if !ok {
		return nil
	}
	return j
} //GetSub

func ExtractDocument (docString string) *G.Document {
	doc, err := G.Compile(strings.NewReader(docString))
	M.Assert(doc != nil, err, 100)
	return doc
} //ExtractDocument

func init () {
	go initReceiver()
	startReset()
}
