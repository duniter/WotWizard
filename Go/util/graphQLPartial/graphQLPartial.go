/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package graphQLPartial

// The module graphQLPartial is an implementation of the GraphQL query interface

import (
	
	A	"util/avl"
	C	"babel/compil"
	F	"path/filepath"
	J	"util/json"
	M	"util/misc"
	S	"strconv"
		"bufio"
		"io"
		"os"
		"util/resources"
		"strings"

)

const (
	
	compDir = "util/graphQLPartial"
	compName = "graphQLPartial.tbl"

)

type (
	
	directory struct {
		r *bufio.Reader
	}
	
	compilation struct {
		*C.Compilation
		r []rune
		size,
		pos int
		err *A.Tree
		sel  SelectionSet
	}

	ErrElem struct {
		Pos,
		Line,
		Col int
		Mes string
	}

	Value interface {
	}

	IntValue struct {
		Int int64
	}

	FloatValue struct {
		Float float64
	}

	StringValue struct {
		S string
	}

	BooleanValue struct {
		Boolean bool
	}

	NullValue struct {
	}

	EnumValue struct {
		Enum string
	}

	ListValue struct {
		List ValueList
	}

	ValueList []Value

	ObjectValue struct {
		Object FieldList
	}

	FieldT struct {
		Name string
		Val Value
	}

	FieldList []FieldT

	Selection struct {
		Alias,
		Name string
		Arguments FieldList
		SelSet SelectionSet
	}

	SelectionSet []Selection

)

var (
	
	compPath = F.Join(compDir, compName)
	wd = resources.FindDir()
	comp *C.Compiler

)

func (e1 *ErrElem) Compare (e2 A.Comparer) A.Comp {
	e := e2.(*ErrElem)
	switch {
		case e1.Pos < e.Pos:
			return A.Lt
		case e1.Pos > e.Pos:
			return A.Gt
		case e1.Mes < e.Mes:
			return A.Lt
		case e1.Mes > e.Mes:
			return A.Gt
		default:
			return A.Eq
	}
}

func (d *directory) ReadInt () int32 {
	m := uint32(0)
	p := uint(0)
	for i := 0; i < 4; i++ {
		n, err := d.r.ReadByte()
		if err != nil {panic(0)}
		m += uint32(n) << p
		p += 8
	}
	return int32(m)
}

func (c *compilation) Read () (ch rune, cLen int) {
	if c.pos >= c.size {
		ch = C.EOF1
	} else {
		ch = c.r[c.pos]
		c.pos++
	}
	cLen = 1
	return
}

func (c *compilation) Pos () int {
	return c.pos
}

func (c *compilation) SetPos (pos int) {
	c.pos = M.Min(pos, c.size)
}

func (co *compilation) Error (p, l, c int, msg string) {
	if co.err != nil {
		_, b, _ := co.err.SearchIns(&ErrElem{Pos: p, Line: l, Col: c, Mes: msg}); M.Assert(!b, p, msg, 100)
	}
}

func (c *compilation) Map (index string) string {
	return index
}

const (
	
	cons = 1
	nilT = 2
	field = 12
	selField = 15

	interpret = 1
	integer = 2
	float = 3
	stringC = 4
	boolean = 5
	null = 6
	enum = 7
	list = 8
	object = 9

)

func makeFieldList (o *C.Object) FieldList {
	n := 0
	oo := o
	for oo.ObjFunc() == cons {
		n++
		oo = oo.ObjTermSon(2)
	}
	var fl FieldList = nil
	if n > 0 {
		fl = make(FieldList, n)
		n := 0
		for o.ObjFunc() == cons {
			oo := o.ObjTermSon(1)
			M.Assert(oo.ObjFunc() == field, 100)
			fl[n].Val = oo.ObjTermSon(2).ObjUser().(Value)
			oo = oo.ObjTermSon(1)
			fl[n].Name = oo.ObjString()
			n++
			o = o.ObjTermSon(2)
		}
	}
	return fl
}

func makeSelectionList (o *C.Object) SelectionSet {
	n := 0
	oo := o
	for oo.ObjFunc() == cons {
		n++
		oo = oo.ObjTermSon(2)
	}
	var sel SelectionSet = nil
	if n > 0 {
		sel = make(SelectionSet, n)
		n = 0
		for o.ObjFunc() == cons {
			oo := o.ObjTermSon(1)
			M.Assert(oo.ObjFunc() == selField, 100)
			ooo := oo.ObjTermSon(1)
			sel[n].Alias = ooo.ObjString()
			ooo = oo.ObjTermSon(2)
			sel[n].Name = ooo.ObjString()
			sel[n].Arguments = makeFieldList(oo.ObjTermSon(3))
			sel[n].SelSet =makeSelectionList(oo.ObjTermSon(5))
			n++
			o = o.ObjTermSon(2)
		}
	}
	return sel
}

func (c *compilation) Execution (fNum, parsNb int, pars C.ObjectsList) (o *C.Object, res C.Anyptr, ok bool) {
	
	// Replace all "\/" occurences by "/"
	makeString := func (s string) string {
		s = strings.Replace(s, `\"`, `"`, -1)
		s = strings.Replace(s, `\\`, `\`, -1)
		s = strings.Replace(s, `\/`, "/", -1)
		s = strings.Replace(s, `\b`, "\u0008", -1)
		s = strings.Replace(s, `\f`, "\u000C", -1)
		s = strings.Replace(s, `\n`, "\u000A", -1)
		s = strings.Replace(s, `\r`, "\u000D", -1)
		s = strings.Replace(s, `\t`, "\u0009", -1)
		pos := strings.Index(s, `\u`)
		for pos >= 0 {
			M.Assert(len(s) >= pos + 6, 100)
			sub := []byte(s)[pos:pos + 6]
			n, err := S.ParseInt(string(sub[2:6]), 16, 32); M.Assert(err == nil, err, 101)
			s = strings.Replace(s, string(sub), string(n), 1)
			pos = strings.Index(s, `\u`)
		}
		return s
	}
	
	if pars == nil {
		ok = false
		return
	}
	ok = true
	o = C.Parameter(pars, 1)
	switch fNum {
	case interpret:
		M.Assert(o.ObjFunc() == cons, 100)
		o = o.ObjTermSon(1)
		c.sel = makeSelectionList(o)
	case integer:
		n, err := S.ParseInt(o.ObjString(), 0, 64); M.Assert(err == nil, err, 101)
		res = &IntValue{Int: n}
	case float:
		f, err := S.ParseFloat(o.ObjString(), 64); M.Assert(err == nil, err, 101)
		res = &FloatValue{Float: f}
	case stringC:
		s := o.ObjString()
		res = &StringValue{S: makeString(string([]byte(s)[1:len(s) - 1]))}
	case boolean:
		res = &BooleanValue{Boolean: o.ObjString() == "true"}
	case null:
		res = &NullValue{}
	case enum:
		res = &EnumValue{Enum: makeString(o.ObjString())}
	case list:
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		var vl ValueList = nil
		if n > 0 {
			vl = make(ValueList, n)
			n = 0
			for o.ObjFunc() == cons {
				vl[n] = o.ObjTermSon(1).ObjUser().(Value)
				n++
				o = o.ObjTermSon(2)
			}
		}
		res = &ListValue{List: vl}
	case object:
		res = &ObjectValue{Object: makeFieldList(o)}
	}
	return
}

func buildValueList (l ValueList, mk *J.Maker) {
	mk.StartArray()
	if l != nil {
		for v := range l {
			buildValue(v, mk)
		}
	}
	mk.BuildArray()
}

func buildField (f *FieldT, mk *J.Maker) {
	mk.StartObject()
	mk.PushString(f.Name)
	mk.BuildField("name")
	buildValue(f.Val, mk)
	mk.BuildField("value")
	mk.BuildObject()
}

func buildFieldList (fl FieldList, mk *J.Maker) {
	mk.StartArray()
	if fl != nil {
		for _, f := range fl {
			buildField(&f, mk)
		}
	}
	mk.BuildArray()
}

func buildValue (val Value, mk *J.Maker) {
	switch v := val.(type) {
	case *IntValue:
		mk.PushInteger(v.Int)
	case *FloatValue:
		mk.PushFloat(v.Float)
	case *StringValue:
		mk.PushString(v.S)
	case *BooleanValue:
		mk.PushBoolean(v.Boolean)
	case *NullValue:
		mk.PushNull()
	case *EnumValue:
		mk.PushString(v.Enum)
	case *ListValue:
		buildValueList(v.List, mk)
	case *ObjectValue:
		buildFieldList(v.Object, mk)
	}
}

func buildSel (sel *Selection, mk *J.Maker) {
	mk.StartObject()
	mk.PushString(sel.Alias)
	mk.BuildField("alias")
	mk.PushString(sel.Name)
	mk.BuildField("name")
	buildFieldList(sel.Arguments, mk)
	mk.BuildField("arguments")
	buildSet(sel.SelSet, mk)
	mk.BuildField("selSet")
	mk.BuildObject()
}

func buildSet (set SelectionSet, mk *J.Maker) {
	if set == nil {
		mk.PushNull()
	} else {
		mk.StartArray()
		for _, sel := range set {
			buildSel(&sel, mk)
		}
		mk.BuildArray()
	}
}

func PrintDocument (w io.Writer, set SelectionSet) {
	M.Assert(set != nil, 100)
	mk := J.NewMaker()
	buildSet(set, mk)
	J.Fprint(w, mk.GetJson())
}

func Compile (rs io.ReadSeeker) (SelectionSet, *A.Tree) {
	n, err := rs.Seek(0, io.SeekEnd); M.Assert(err == nil, err, 100)
	_, err = rs.Seek(0, io.SeekStart); M.Assert(err == nil, err, 101)
	b := make([]byte, n)
	_, err = io.ReadFull(rs, b); M.Assert(err == nil, err, 102)
	r := []rune(string(b))
	co := &compilation{r: r, size: len(r), pos: 0, err: A.New(), sel: nil}
	c := C.NewCompilation(co)
	co.Compilation = c
	if co.Compile(comp, false) {
		return co.sel, nil
	}
	return nil, co.err
}

type (
	
	ReadSeekerCloser interface {
		io.ReadCloser
		io.Seeker
	}
	LinkGQL func (name string) ReadSeekerCloser

)

var (
	
	lGQL LinkGQL

)

func ReadGraphQL (name string) (SelectionSet, *A.Tree) {
	rsc := lGQL(name)
	defer rsc.Close()
	return Compile(rsc)
}

func SetLGQL (lG LinkGQL) {
	lGQL = lG
}

func SetRComp (rComp io.Reader) {
	comp = C.NewDirectory(&directory{r: bufio.NewReader(rComp)}).ReadCompiler()
}

func linkFile (name string) ReadSeekerCloser {
	f, err := os.Open(name); M.Assert(err == nil, err, 100)
	return f
}

func init () {
	SetLGQL(linkFile)
	f, err := os.Open(F.Join(wd, compPath))
	if err == nil {
		defer f.Close()
		SetRComp(f)
	}
}
