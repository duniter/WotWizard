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
		"fmt"
		"net/http"
		"io"
		"io/ioutil"
		"os"
		"strings"
		"net/url"

)

const (
	
	typeSystemName = "TypeSystem.txt"
	
	storeSubsFile = "currentSubs.txt"
	
)

type (
	
	action struct {
		es G.ExecSystem
		doc *G.Document
		opName,
		returnAddr string
		varVals J.Json
		variableValues *A.Tree
		w http.ResponseWriter
		c chan bool
	}
	
	readSubsAction struct {
	}
	
	typeSystem struct {
		G.TypeSystem
	}
	
	streamer struct { // G.EventStreamer
		name string
		stream *G.EventStream
		notification G.SourceEventNotification
	}
	
	void struct {
	}
	
	addresses map[string] *void
	
	responseStreamer struct {
		doc *G.Document
		name string // name of the subscription
		varVals J.Json
		stream *G.ResponseStream
		returnAddrs addresses
	}
	
	responseStreamers map[string] *responseStreamer

)

var (
	
	serverAddress = BA.ServerAddress()
	
	storeSubsPath = F.Join(BA.RsrcDir(), storeSubsFile)
	
	lg = BA.Lg
	
	ts *typeSystem
	
	typeSystemPath = F.Join(B.System(), typeSystemName)
	
	responseStreamsByDoc = make(responseStreamers)
	responseStreamsByAddr = make(responseStreamers)

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

func buildResponseStreamerByDocKey (doc *G.Document, opName string, varVals J.Json) string {
	return doc.GetFlatString() + opName + varVals.GetFlatString()
}

func buildResponseStreamerByAddrKey (returnAddr, opName string, varVals J.Json) string {
	return returnAddr + opName + varVals.GetFlatString()
}

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
			key := buildResponseStreamerByDocKey(a.doc, a.opName, a.varVals)
			rs, ok := responseStreamsByDoc[key]; M.Assert(ok, key, 100)
			rs.stream = r.Data
			r.Data.FixResponseStream(rs)
			storeSubs()
			es := rs.stream.SourceStream
			es.EventStreamer.(*streamer).notification(es, nil)
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
	es.notification = notification
} //RecordNotificationProc

func (es *streamer) CloseEvent () {
	B.RemoveUpdateProc(es.name)
} //CloseEvent

func CreateStream (name string) *G.EventStream { // *G.ValMapItem
	s := &streamer{name: name}
	es := G.MakeEventStream(s)
	s.stream = es
	return es
} //CreateStream

func unsubscribe (returnAddr, streamName string, varVals J.Json) {
	key := buildResponseStreamerByAddrKey(returnAddr, streamName, varVals)
	if r, ok := responseStreamsByAddr[key]; ok {
		if _, ok := r.returnAddrs[returnAddr]; ok {
			delete(r.returnAddrs, returnAddr)
			delete(responseStreamsByAddr, key)
			if len(r.returnAddrs) == 0 {
				delete(responseStreamsByDoc, buildResponseStreamerByDocKey(r.doc, streamName, varVals))
				G.Unsubscribe(r.stream)
			}
			storeSubs()
		}
	}
} //unsubscribe

func (rs *responseStreamer) ManageResponseEvent (r G.Response) {
	errors := r.Errors()
	if errors != nil {
		printErrors(errors)
	}
	j := G.ResponseToJson(r)
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushString(rs.name)
	mk.BuildField("operationName")
	mk.PushJson(j)
	mk.BuildField("result")
	mk.BuildObject()
	s := mk.GetJson().GetFlatString()
	reader := strings.NewReader(s)
	for addr := range rs.returnAddrs {
		reader.Seek(0, io.SeekStart)
		r, err := http.Post("http://" + addr, "text/json", reader)
		if err != nil {
			unsubscribe(addr, rs.name, rs.varVals)
			return
		}
		r.Body.Close()
	}
}

func readOpNameVars  (req *http.Request) (varVals J.Json, t *A.Tree, opName, addr string, docS string, err error) {
	buf, error := ioutil.ReadAll(req.Body); M.Assert(error == nil, error, 100)
	j := J.ReadString(string(buf))
	b := j != nil
	var o *J.Object
	if b {
		o, b = j.(*J.Object)
	}
	if !b {
		err = errors.New("Incorrect JSON request")
		return
	}
	opName = J.GetString(o, "operationName")
	addr = J.GetString(o, "returnAddr")
	if addr != "" {
		_, error = url.Parse(addr)
		if error != nil {
			err = errors.New("Incorrect returnAddr value")
			return
		}
	}
	varVals = J.GetJson(o, "variableValues")
	if varVals == nil {
		varVals = J.ReadString("{}")
	}
	obj, b := varVals.(*J.Object)
	if !b {
		err = errors.New("Incorrect variableValues value")
		return
	}
	t = A.New()
	for _, f := range obj.Fields {
		if !G.InsertJsonValue(t, f.Name, f.Value) {
			err = errors.New("Duplicated variable value name " + f.Name)
			return
		}
	}
	docS = J.GetString(o, "graphQL")
	if docS == "" {
		err = errors.New("No graphQL string")
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
		
		j, variableValues, opName, returnAddr, docS, error := readOpNameVars (req)
		if error != nil {
			writeError(error)
			return
		}
		doc, r := G.ReadString(docS)
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
			} else {
				writeError(errors.New("Selected operation name not defined"))
				return
			}
		}
		if es.GetOperation(opName).OpType == G.SubscriptionOp {
			if returnAddr == "" {
				writeError(errors.New("No returnAddr value"))
				return
			}
			key := buildResponseStreamerByDocKey(doc, opName, j)
			var (rs *responseStreamer; ok bool)
			if rs, ok = responseStreamsByDoc[key]; !ok {
				rs = &responseStreamer{doc: doc, name: opName, varVals: j, returnAddrs: make(addresses)}
				responseStreamsByDoc[key] = rs
			}
			rs.returnAddrs[returnAddr] = nil
			responseStreamsByAddr[buildResponseStreamerByAddrKey(returnAddr, opName, j)] = rs
		}
		a := &action{es: es, doc: doc, opName: opName, varVals: j, variableValues: variableValues, w: w, c: make(chan bool)}
		newAction <- a
		<- a.c
	}

} //makeHandler

func loop (newAction chan<- B.Actioner) {
	newAction <- new(readSubsAction)
	r := http.NewServeMux()
	r.HandleFunc("/", makeHandler(newAction))
	server := &http.Server{
		Addr: serverAddress,
		Handler: r,
	}
	s := fmt.Sprint("Listening on ", serverAddress, " ...")
	lg.Println(s)
	fmt.Println(s)
	server.ListenAndServe()
} //loop

func storeSubs () {
	f, err := os.Create(storeSubsPath); M.Assert(err == nil, err, 100)
	defer f.Close()
	fmt.Fprintln(f, len(responseStreamsByDoc))
	for _, rs := range responseStreamsByDoc {
		fmt.Fprintln(f, rs.doc.GetFlatString())
		fmt.Fprintln(f, rs.name)
		fmt.Fprintln(f, rs.varVals.GetFlatString())
		fmt.Fprintln(f, len(rs.returnAddrs))
		for addr := range rs.returnAddrs {
			fmt.Fprintln(f, addr)
		}
	}
} //storeSubs

func readSubs () {
	f, err := os.Open(storeSubsPath)
	if err != nil {
		lg.Println(err)
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return
	}
	n, err := SC.Atoi(sc.Text()); M.Assert(err == nil, err, 101)
	for ; n > 0; n-- {
		ok := sc.Scan(); M.Assert(ok, 102)
		doc, _ := G.ReadString(sc.Text()); M.Assert(doc != nil, 103)
		ok = sc.Scan(); M.Assert(ok, 104)
		opName := sc.Text()
		ok = sc.Scan(); M.Assert(ok, 105)
		j := J.ReadString(sc.Text()); M.Assert(j != nil, 106)
		ok = sc.Scan(); M.Assert(ok, 107)
		returnAddrs := make(addresses)
		rs := &responseStreamer{doc: doc, name: opName, varVals: j, returnAddrs: returnAddrs}
		m, err := SC.Atoi(sc.Text()); M.Assert(err == nil, err, 108)
		for ; m > 0; m-- {
			ok := sc.Scan(); M.Assert(ok, 109)
			returnAddr := sc.Text()
			returnAddrs[returnAddr] = nil
			responseStreamsByAddr[buildResponseStreamerByAddrKey(returnAddr, opName, j)] = rs
		}
		responseStreamsByDoc[buildResponseStreamerByDocKey(doc, opName, j)] = rs
		t := A.New()
		for _, f := range j.(*J.Object).Fields {
			ok := G.InsertJsonValue(t, f.Name, f.Value); M.Assert(ok, 110)
		}
		M.Assert(G.ExecutableDefinitions(doc), 111)
		es := ts.ExecValidate(doc); M.Assert(es.GetErrors().IsEmpty(), 112)
		r := es.Execute(doc, opName, t); M.Assert(r.Errors() == nil, 113)
		rr, ok := r.(*G.SubscribeResponse); M.Assert(ok, 114)
		M.Assert(rr.Data != nil, 115)
		rs.stream = rr.Data
		rr.Data.FixResponseStream(rs)
		ess := rs.stream.SourceStream
		ess.EventStreamer.(*streamer).notification(ess, nil)
	}
} //readSubs

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

func (a *readSubsAction) Activate () {
	readSubs()
}

func (a *readSubsAction) Name () string {
	return "readSubs"
}

func Start () {
	newAction := make(chan B.Actioner)
	go loop(newAction)
	B.Start(newAction)
} //Start

func TS () G.TypeSystem {
	return ts
} //TS

func stopSubR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var (v G.Value; returnAddr, name, s string)
	if G.GetValue(argumentValues, "returnAddr", &v) {
		switch v := v.(type) {
		case *G.StringValue:
			returnAddr = v.String.S
		default:
			M.Halt(v, 100)
		}
	} else {
		M.Halt(v, 101)
	}
	if G.GetValue(argumentValues, "name", &v) {
		switch v := v.(type) {
		case *G.StringValue:
			name = v.String.S
		default:
			M.Halt(v, 102)
		}
	} else {
		M.Halt(v, 103)
	}
	if G.GetValue(argumentValues, "varVals", &v) {
		switch v := v.(type) {
		case *G.StringValue:
			s = v.String.S
		default:
			M.Halt(v, 104)
		}
	} else {
		s = "{}"
	}
	varVals := J.ReadString(s); M.Assert(varVals != nil, 105)
	unsubscribe(returnAddr, name, varVals)
	return nil
} //stopSubR

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
	ts.FixFieldResolver("Mutation", "stopSubscription", stopSubR)
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
