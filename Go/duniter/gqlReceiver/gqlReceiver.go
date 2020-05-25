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
		"errors"
		"os"
		"strings"
		"time"

)

const (
	
	waitDelay = 400 * time.Millisecond
	
	typeSystemName = "TypeSystem.txt"
	
	defaultJsonName = "-Result"
	errorsJsonName	= "-Errors"
	subscriptionSuffix = "-subscription"
	extension = ".json"

)

type (
	
	action struct {
		es G.ExecSystem
		doc *G.Document
		opName string
		jsonNum int
		variableValues *A.Tree
	}
	
	fileName struct {
		next *fileName
		name string
	}
	
	fileQueue struct {
		end *fileName
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
		jsonNum int
	}
	
	responseStreamsT map[string]*G.ResponseStream

)

var (
	
	lg = BA.Lg
	
	ts *typeSystem
	
	typeSystemPath = F.Join(B.System(), typeSystemName)
	qdir = B.Qdir()
	jdir = B.Jdir()
	
	responseStreams = make(responseStreamsT)

)

func json (j J.Json, path string) {
	f, err := M.InstantCreate(path); M.Assert(err == nil, err, 100)
	defer M.InstantClose(f)
	j.Write(f)
} //json

func makeJsonName (opName string, jsonNum int) string {
	name := opName
	if name == "" {
		name = defaultJsonName
	}
	return F.Join(jdir, name + "-" + SC.Itoa(jsonNum) + extension)
}

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
				if p != nil {
					b.WriteString(".")
				}
				if p == nil {break}
			}
		}
		lg.Println(b.String())
		e = errors.Next(e)
	}
} //printErrors

func (a *action) Activate () {
	r := a.es.Execute(a.doc, a.opName, a.variableValues)
	errors := r.Errors()
	if errors != nil {
		printErrors(errors)
	}
	switch r := r.(type) {
	case *G.InstantResponse:
		json(G.ResponseToJson(r), makeJsonName(a.opName, a.jsonNum))
	case *G.SubscribeResponse:
		json(G.ResponseToJson(r), makeJsonName(a.opName + subscriptionSuffix, a.jsonNum))
		if r.Data != nil {
			responseStreams[r.Data.SourceStream.StreamName()] = r.Data
			m := new(responseEventManager)
			m.name = a.opName
			m.jsonNum = a.jsonNum
			r.Data.RecordResponseEventManager(m)
		}
	}
}

func (a *action) Name () string {
	if a.opName == "" {
		return "anonymous (" + defaultJsonName + ")"
	} else {
		return a.opName
	}
}

func updateProc (pars ... interface{}) {
	pars[0].(G.SourceEventNotification)(pars[1].(*streamer).stream, nil)
} //updateProc

func (es *streamer) StreamName () string {
	return es.name
}

func (es *streamer) RecordNotificationProc (notification G.SourceEventNotification) {
	B.AddUpdateProc(es.name, updateProc, notification, es)
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
	}
}

func (m *responseEventManager) ManageResponseEvent (rs *G.ResponseStream, r G.Response) {
	json(G.ResponseToJson(r), makeJsonName(m.name, m.jsonNum))
}

func (q *fileQueue) isEmpty () bool {
	return q.end == nil
} //isEmpty

func (q *fileQueue) present (f string) bool {
	p := q.end
	if p == nil {
		return false
	}
	p = p.next
	for p != q.end {
		if p.name == f {
			return true
		}
		p = p.next
	}
	return p.name == f
} //present

func (q *fileQueue) put (f string) {
	if !q.present(f) {
		if q.end == nil {
			q.end = &fileName{name: f}
			q.end.next = q.end
		} else {
			p := &fileName{name: f, next: q.end.next}
			q.end.next = p
			q.end = p
		}
	}
} //put

func (q *fileQueue) get () string {
	M.Assert(q.end != nil, 100)
	p := q.end.next
	q.end.next = p.next
	if q.end == p {
		q.end = nil
	}
	return p.name
} //get

func readVars (path string) (t *A.Tree, jsonNum int, err error) {
	f, err := os.Open(path); M.Assert(err == nil, err, 100)
	defer f.Close()
	t = A.New()
	jsonNum = 0
	err = nil
	sc := bufio.NewScanner(f)
	ok := sc.Scan()
	var buf []byte
	if ok {
		buf = sc.Bytes()
	}
	for ok && (len(buf) == 0 || buf[0] == '#') {
		if len(buf) >= 3 {
			numStr := string(buf[1:3])
			var n int64
			n, err = SC.ParseInt(numStr, 16, 32)
			if err != nil {
				return
			}
			jsonNum = int(n)
			j := J.ReadString(string(buf[3:]))
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
} //readVars

func loop (newAction chan<- B.Actioner) {
	
	isHidden := func (info os.FileInfo) bool {
		return info.Name()[0] == '.'
	} //isHidden
	
	var fileQ = fileQueue{end: nil}
	for {
		f, err := os.Open(qdir); M.Assert(err == nil, err, 100)
		infos, err := f.Readdir(0); M.Assert(err == nil, err, 101)
		f.Close()
		for _, info := range infos {
			if !info.IsDir() && !isHidden(info) { 
				fileQ.put(info.Name())
			}
		}
		time.Sleep(waitDelay)
		for !fileQ.isEmpty() {
			qf := F.Join(B.Qdir(), fileQ.get())
			variableValues, jsonNum, error := readVars(qf)
			if error != nil {
				m := J.NewMaker()
				m.StartObject()
				m.PushString(error.Error())
				m.BuildField("errors")
				m.BuildObject()
				json(m.GetJson(), makeJsonName(errorsJsonName, 0))
				continue
			}
			doc, r := G.ReadGraphQL(qf)
			BA.SwitchOff(qf)
			if doc == nil {
				err := r.Errors()
				M.Assert(!err.IsEmpty())
				printErrors(err)
				json(G.ResponseToJson(r), makeJsonName(errorsJsonName, jsonNum))
				continue
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
				json(G.ResponseToJson(r), makeJsonName(errorsJsonName, jsonNum))
				continue
			}
			listOp := es.ListOperations()
			for _, op := range listOp {
				newAction <- &action{es, doc, op, jsonNum, variableValues}
			}
		}
	}
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
	os.MkdirAll(B.Qdir(), 0777)
	os.MkdirAll(B.Jdir(), 0777)
	newAction := make(chan B.Actioner)
	go loop(newAction)
	B.Start(newAction)
} //Start

func TS () G.TypeSystem {
	return ts
}

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
}
