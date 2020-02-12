/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package gqlReceiver
import (
	
	B	"duniter/blockchain"
	BA	"duniter/basic"
	F	"path/filepath"
	G	"util/graphQLPartial"
	J	"util/json"
	M	"util/misc"
	R	"reflect"
		"errors"
		"fmt"
		"os"
		"time"

)

const (
	
	waitDelay = 400 * time.Millisecond

)

type (
	
	Function interface{}
	Arguments []string
	
	action struct {
		function Function
		arguments Arguments
	}
	
	actionsT map[string] *action
	
	fileName struct {
		next *fileName
		name string
	}
	
	fileQueue struct {
		end *fileName
	}

)

var (
	
	actions = make(actionsT)
	
	lg = BA.Lg

)

func Json (j J.Json, name string) {
	f, err := M.InstantCreate(name); M.Assert(err == nil, err, 100)
	defer M.InstantClose(f)
	J.Fprint(f, j)
}

func (q *fileQueue) isEmpty () bool {
	return q.end == nil
}

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
}

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
}

func (q *fileQueue) get () string {
	M.Assert(q.end != nil, 100)
	p := q.end.next
	q.end.next = p.next
	if q.end == p {
		q.end = nil
	}
	return p.name
}

func AddAction (name string, f Function, a Arguments) {
	lg.Println("Adding action", name)
	actions[name] = &action{function: f, arguments: a}
}

func getAction (name string) (f Function, a Arguments, ok bool) {
	var act *action 
	if act, ok = actions[name]; ok {
		f = act.function
		a = act.arguments
	}
	return
}

func sortArguments (fields G.FieldList, args Arguments) (vL G.ValueList, err error) {
	err = nil
	if len(fields) != len(args) {
		err = errors.New("Wrong number of arguments")
		return
	}
	vL = make(G.ValueList, len(args))
	for i, a := range args {
		ok := false
		for _, f := range fields {
			if f.Name == a {
				vL[i] = f.Val
				ok = true
				break
			}
		}
		if !ok {
			err = errors.New("Argument " + a + " is missing")
			return
		}
	}
	return
}

func evalArguments (vL G.ValueList, nL []string, f R.Value, output string, newAction chan<- B.Actioner) (rV []R.Value, err error) {
	M.Assert(f.Kind() == R.Func, 100)
	err = nil
	tf := f.Type()
	if f.Type().NumIn() != len(vL) + 3 {
		err = errors.New("The function must have three parameters more than the number of arguments")
		return
	}
	rV = make([]R.Value, len(vL) + 3)
	for i, v := range vL {
		t := tf.In(i)
		switch val := v.(type) {
		case *G.BooleanValue:
			if t.Kind() != R.Bool {
				err = errors.New("boolean expected")
				return
			}
			rV[i] = R.ValueOf(val.Boolean)
		case *G.EnumValue:
			if t.Kind() != R.String {
				err = errors.New("enum expected")
				return
			}
			rV[i] = R.ValueOf(val.Enum)
		case *G.FloatValue:
			if t.Kind() != R.Float64 {
				err = errors.New("float64 expected")
				return
			}
			rV[i] = R.ValueOf(val.Float)
		case *G.IntValue:
			if t.Kind() != R.Int64 {
				err = errors.New("int64 expected")
				return
			}
			rV[i] = R.ValueOf(val.Int)
		case *G.NullValue:
			// M.Assert(t.Kind() == R.Ptr, 107)
			rV[i] = R.ValueOf(nil)
		case *G.StringValue:
			if t.Kind() != R.String {
				err = errors.New("string expected")
				return
			}
			rV[i] = R.ValueOf(val.S)
		default:
			err = errors.New("unknown type")
			return
		}
	}
	rV[len(vL)] = R.ValueOf(output)
	rV[len(vL) + 1] = R.ValueOf(newAction)
	rV[len(vL) + 2] = R.ValueOf(nL)
	return
}

func callAction (sel G.Selection, newAction chan<- B.Actioner) error {
	f, a, ok := getAction(sel.Name)
	if !ok {
		lg.Println("***Error***", sel.Name, ": Unknown action")
		fmt.Fprintln(os.Stderr, sel.Name, ": Unknown action")
		fmt.Fprintln(os.Stderr, "List of actions:")
		for name, _ := range actions {
			fmt.Fprintln(os.Stderr, "\t", name)
		}
		fmt.Fprintln(os.Stderr)
		return errors.New(sel.Name + ": Unknown action")
	}
	var err error = nil
	fv := R.ValueOf(f)
	if fv.Kind() != R.Func {
		return errors.New("The action " + sel.Name + " contains no function")
	}
	var vL G.ValueList
	nL := make([]string, len(sel.SelSet))
	for i, s := range sel.SelSet {
		nL[i] = s.Name
	}
	vL, err = sortArguments(sel.Arguments, a)
	if err != nil {
		return err
	}
	rV, err := evalArguments(vL, nL, fv, F.Join(B.Jdir(), sel.Alias + ".json"), newAction)
	if err == nil {
		fv.CallSlice(rV)
	}
	return err
}

func loop (newAction chan<- B.Actioner) {
	var fileQ = fileQueue{end: nil}
	for {
		f, err := os.Open(B.Qdir()); M.Assert(err == nil, err, 100)
		infos, err := f.Readdir(0); M.Assert(err == nil, err, 101)
		f.Close()
		for _, info := range infos {
			if !info.IsDir() && info.Name()[0] != '.' {
				fileQ.put(info.Name())
			}
		}
		time.Sleep(waitDelay)
		for !fileQ.isEmpty() {
			qf := F.Join(B.Qdir(), fileQ.get())
			set, err := G.ReadGraphQL(qf)
			BA.SwitchOff(qf)
			if err != nil {
				lg.Println("***Errors*** in", qf)
				fmt.Fprintln(os.Stderr, "Errors in", qf)
				e := err.Next(nil)
				for e != nil {
					el := e.Val().(*G.ErrElem)
					lg.Println("\t", el.Line, ":", el.Col, ":", el.Mes)
					fmt.Fprintln(os.Stderr, "\t", el.Line, ":", el.Col, ":", el.Mes)
					e = err.Next(e)
				}
				fmt.Fprintln(os.Stderr)
				continue
			}
			for _, sel := range set {
				if err := callAction(sel, newAction); err != nil {
					BA.Lg.Println(err)
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}
	}
}

func Start () {
	os.MkdirAll(B.Qdir(), 0777)
	os.MkdirAll(B.Jdir(), 0777)
	newAction := make(chan B.Actioner)
	go loop(newAction)
	B.Start(newAction)
}
