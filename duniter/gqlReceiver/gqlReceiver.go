/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package gqlReceiver

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	F	"path/filepath"
	G	"util/graphQL"
	J	"util/json"
	M	"util/misc"
	SC	"strconv"
	W	"duniter/wotWizard"
		"bufio"
		"bytes"
		"errors"
		"net/http"
		"io"
		"io/ioutil"
		"github.com/gorilla/mux"
		"strings"

)

const (
	
	typeSystemName = "TypeSystem.txt"
	
)

type (
	
	action struct {
		es G.ExecSystem
		doc *G.Document
		opName string
		variableValues *A.Tree
		w http.ResponseWriter
		c chan bool
	}
	
	typeSystem struct {
		G.TypeSystem
	}
	
	streamer struct { // G.EventStreamer
		name string
		stream *G.EventStream
		rootValue *G.OutputObjectValue
		argumentValues *A.Tree
	}
	
	responseEventManager struct { // G.ResponseEventManager
		name string // name of the subscription
	}
	
	subscriptionData struct {
		
		response J.Json
	}
	
	responseStreamsT map[string] *G.ResponseStream
	subscriptionsDataT map[string] J.Json
	updateT struct {
		notification G.SourceEventNotification
		param *G.EventStream
	}
	notificationsT map[string] *updateT

)

var (
	
	serverAddress = BA.ServerAddress()
	
	lg = BA.Lg
	
	ts *typeSystem
	
	typeSystemPath = F.Join(B.System(), typeSystemName)
	
	responseStreams = make(responseStreamsT)
	subscriptionsData = make(subscriptionsDataT)
	notifications = make(notificationsT)

)

func printErrors (errors *A.Tree) {
	e := errors.Next(nil)
	for e != nil {
		el := e.Val().(*G.ErrorElem)
		var b strings.Builder
		b.WriteString("***ERROR*** " + el.Message)
		if el.Location != nil {
			b.WriteString(" at ")
			b.WriteString(SC.Itoa(el.Location.Line))
			b.WriteString(":")
			b.WriteString(SC.Itoa(el.Location.Column))
		}
		p := el.Path
		if p != nil {
			b.WriteString(" in ")
			for {
				switch p := p.(type) {
				case *G.PathString:
					b.WriteString(p.S)
				case *G.PathNb:
					b.WriteString(SC.Itoa(p.N))
				}
				p = p.Next()
				if p == nil {
					break
				}
				b.WriteString(".")
			}
		}
		lg.Println(b.String())
		e = errors.Next(e)
	}
} //printErrors

func (a *action) Activate () {
	a.w.Header().Set("Content-Type", "application/json")
	a.w.WriteHeader(http.StatusCreated)
	r := a.es.Execute(a.doc, a.opName, a.variableValues)
	errors := r.Errors()
	if errors != nil {
		printErrors(errors)
	}
	switch r := r.(type) {
	case *G.InstantResponse:
		G.ResponseToJson(r).Write(a.w)
	case *G.SubscribeResponse:
		G.ResponseToJson(r).Write(a.w)
		if r.Data != nil {
			m := new(responseEventManager)
			m.name = r.Data.SourceStream.StreamName()
			responseStreams[m.name] = r.Data
			r.Data.RecordResponseEventManager(m)
		}
	}
	a.c <- true
} //Activate

func (a *action) Name () string {
	if a.opName == "" {
		return "anonymous"
	} else {
		return a.opName
	}
} //Name

func updateProc (pars ... interface{}) {
	pars[0].(G.SourceEventNotification)(pars[1].(*streamer).stream, nil)
} //updateProc

func (es *streamer) StreamName () string {
	return es.name
} //StreamName

func (es *streamer) RecordNotificationProc (notification G.SourceEventNotification) {
	B.AddUpdateProc(es.name, updateProc, notification, es)
	notifications[es.name] = &updateT{notification: notification, param: es.stream}
} //RecordNotificationProc

func (es *streamer) CloseEvent () {
	B.RemoveUpdateProc(es.name)
} //CloseEvent

func CreateStream (name string, rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	s := &streamer{name: name, rootValue: rootValue, argumentValues: argumentValues}
	es := G.MakeEventStream(s)
	s.stream = es
	return es
} //CreateStream

func Unsubscribe (streamName string) {
	if r := responseStreams[streamName]; r != nil {
		G.Unsubscribe(r)
		delete(responseStreams, streamName)
		delete(subscriptionsData, streamName)
	}
} //Unsubscribe

func (m *responseEventManager) ManageResponseEvent (rs *G.ResponseStream, r G.Response) {
	subscriptionsData[m.name] = G.ResponseToJson(r)
} //ManageResponseEvent

func readOpNameVars  (r io.Reader) (t *A.Tree, opName string, err error) {
	t = A.New()
	err = nil
	sc := bufio.NewScanner(r)
	ok := sc.Scan()
	var buf []byte
	if ok {
		buf = sc.Bytes()
	}
	for ok && (len(buf) == 0 || buf[0] == '#') {
		if len(buf) >= 1 {
			i := bytes.IndexRune(buf, '{')
			var s string
			if i < 0 {
				opName = string(buf[1:])
				s = ""
			} else {
				opName = string(buf[1:i])
				s = string(buf[i:])
			}
			j := J.ReadString(s)
			b := j != nil
			var o *J.Object
			if b {
				o, b = j.(*J.Object)
			}
			if b {
				for _, f := range o.Fields {
					if !G.InsertJsonValue(t, f.Name, f.Value) {
						err = errors.New("Duplicated variable value name " + f.Name)
						return
					}
				}
				break
			}
		}
		ok = sc.Scan()
		if ok {
			buf = sc.Bytes()
		}
	}
	return
} //readOpNameVars 

func makeHandler (newAction chan<- B.Actioner) func (w http.ResponseWriter, req *http.Request) {
	return func (w http.ResponseWriter, req *http.Request) {
		
		writeError := func (err error) {
			m := J.NewMaker()
			m.StartObject()
			m.PushString(err.Error())
			m.BuildField("errors")
			m.BuildObject()
			m.GetJson().Write(w)
			lg.Println("***ERROR*** ", err)
		}
		
		getReadCloser := func (b []byte) io.ReadCloser {
			return ioutil.NopCloser(strings.NewReader(string(b)))
		}
		
		b, er := ioutil.ReadAll(req.Body); M.Assert(er == nil, er, 100)
		if b[0] == '@' {
			name := string(b[1:])
			j, ok := subscriptionsData[name]
			if !ok {
				update, ok := notifications[name]
				if !ok {
					writeError(errors.New("Unknown subscription " + name))
					return
				}
				update.notification(update.param, nil)
				j, ok = subscriptionsData[name]; M.Assert(ok, 101)
			}
			j.Write(w)
			return
		}
		variableValues, opName, error := readOpNameVars (getReadCloser(b))
		if error != nil {
			writeError(error)
			return
		}
		doc, r := G.Compile(getReadCloser(b))
		if doc == nil {
			err := r.Errors()
			M.Assert(!err.IsEmpty(), 102)
			printErrors(err)
			G.ResponseToJson(r).Write(w)
			return
		}
		if !G.ExecutableDefinitions(doc) {
			ts.Error("NotExecDefs", "", "", nil, nil)
		}
		es := ts.ExecValidate(doc)
		err := es.GetErrors()
		if !err.IsEmpty() {
			printErrors(err)
			r := new(G.InstantResponse)
			r.SetErrors(err)
			G.ResponseToJson(r).Write(w)
			return
		}
		if opName == "" {
			if opList := es.ListOperations(); len(opList) == 1 {
				opName = opList[0]
			}
		}
		a := &action{es, doc, opName, variableValues, w, make(chan bool)}
		newAction <- a
		<- a.c
	}
} //makeHandler

func loop (newAction chan<- B.Actioner) {
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", makeHandler(newAction)).Methods("POST")
	server := &http.Server{
		Addr: serverAddress,
		Handler: r,
	}
	lg.Println("Listening...")
	server.ListenAndServe()
} //loop

func coerceInt64 (ts G.TypeSystem, v G.Value, path G.Path, cV *G.Value) bool {
	switch v := v.(type) {
	case *G.IntValue:
		*cV = v
		return true
	default:
		return false
	}
} //coerceInt64

func coerceHash (ts G.TypeSystem, v G.Value, path G.Path, cV *G.Value) bool {
	switch v := v.(type) {
	case *G.StringValue:
		*cV = v
		return true
	default:
		return false
	}
} //coerceHash

func coercePubkey (ts G.TypeSystem, v G.Value, path G.Path, cV *G.Value) bool {
	switch v := v.(type) {
	case *G.StringValue:
		*cV = v
		return true
	default:
		return false
	}
} //coercePubkey

func coerceVoid (ts G.TypeSystem, v G.Value, path G.Path, cV *G.Value) bool {
	*cV = nil
	return true
} //coerceVoid

func coerceNumber (ts G.TypeSystem, v G.Value, path G.Path, cV *G.Value) bool {
	switch v := v.(type) {
	case *G.IntValue:
		*cV = v
		return true
	case *G.FloatValue:
		*cV = v
		return true
	default:
		return false
	}
} //coerceNumber

func (ts *typeSystem) FixScalarCoercer (scalarName string, sc *G.ScalarCoercer) {
	switch scalarName {
	case "Int64":
		*sc = coerceInt64
	case "Hash":
		*sc = coerceHash
	case "Pubkey":
		*sc = coercePubkey
	case "Void":
		*sc = coerceVoid
	case "Number":
		*sc = coerceNumber
	}
} //FixScalarCoercer

func abstractTypeResolver (ts G.TypeSystem, td G.TypeDefinition, ov *G.OutputObjectValue) *G.ObjectTypeDefinition {
	var name string
	switch td.TypeDefinitionC().Name.S {
	case "CertifOrDossier":
		switch Unwrap(ov, 0).(type) {
		case *W.Certif:
			name = "MarkedDatedCertification"
		case *W.Dossier:
			name = "MarkedDossier"
		default:
			return nil
		}
	case "File":
		name = "FileS"
	case "WWResult":
		name = "WWResultS"
	default:
		return nil
	}
	d := ts.GetTypeDefinition(name)
	ok := d != nil
	var od *G.ObjectTypeDefinition
	if ok {
		od, ok = d.(*G.ObjectTypeDefinition)
	}
	if !ok {
		return nil
	}
	return od
} //abstractTypeResolver

func Start () {
	newAction := make(chan B.Actioner)
	go loop(newAction)
	B.Start(newAction)
} //Start

func TS () G.TypeSystem {
	return ts
} //TS

func initAll () {

	const (
		
		rootName = "root"

	)

	doc, r := G.ReadGraphQL(typeSystemPath)
	err := r.Errors()
	if !err.IsEmpty() {
		printErrors(err)
		M.Halt(100)
	}
	M.Assert(doc != nil, 101)
	if G.ExecutableDefinitions(doc) {
		ts.Error("NoTypeSystemInDoc", "", "", nil, nil)
		err := ts.GetErrors()
		printErrors(err)
		M.Halt(102)
	}
	tsRead := false
	ts = new(typeSystem)
	ts.TypeSystem = G.Dir.NewTypeSystem(ts)
	ts.InitTypeSystem(doc)
	initialValue := G.NewOutputObjectValue()
	initialValue.InsertOutputField(rootName, nil)
	ts.FixInitialValue(initialValue)
	ts.FixAbstractTypeResolver(abstractTypeResolver)
	tsRead = ts.GetErrors().IsEmpty()
	if !tsRead {
		err := ts.GetErrors()
		printErrors(err)
		M.Halt(103)
	}
} //initAll

func init () {
	initAll()
} //init

func Wrap (i ...interface{}) *G.OutputObjectValue {
	o := G.NewOutputObjectValue()
	for j, in := range i {
		o.InsertOutputField("data" + SC.Itoa(j), G.MakeAnyValue(in))
	}
	return o
} //Wrap

func Unwrap (rootValue *G.OutputObjectValue, n int) interface{} {
	f := rootValue.First()
	for i := 0; i < n && f != nil; i++ {
		f = rootValue.Next(f)
	}
	M.Assert(f != nil, 20)
	switch v := f.Value.(type) {
	case *G.AnyValue:
		return v.Any
	case *G.NullValue:
		return v
	default:
		M.Halt(100)
		return nil
	}
} //Unwrap
