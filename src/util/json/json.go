/* 
Util: Utility tools.

Copyright (C) 2017…2019 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package json

import (
	
	C	"babel/compil"
	F	"path/filepath"
	J	"encoding/json"
	M	"util/misc"
	R	"reflect"
	RS	"util/resources"
	SC	"strconv"
		"bufio"
		"bytes"
		"fmt"
		"io"
		"os"
		"strings"

)

const (
	
	double_quote = 0x22
	reverse_solidus = 0x5C
	solidus = 0x2F
	backspace = 0x08
	form_feed = 0x0C
	line_feed = 0x0A
	carriage_return = 0x0D
	horizontal_tab = 0x09
	
	jsonDir = "util/json" // Directory of JSON resources files
	compName = "json.tbl" // Json compiler

)

type (
	
	// For compiler internalizing
	directory struct { //C.Directory
		r *bufio.Reader
	}
	
	compilationer struct {
		r []rune
		size,
		pos int
		json Json
	}

)

type (
	
	// Outputs of ReadString, ReadText && ReadFile
	
	// A Json may be an 'Object' or an 'Array'
	Json interface {
		// 'GetString' returns a formatted string for reading
		GetString () string
		
		// 'GetFlatString' returns a compact string
		GetFlatString () string
		
		// 'Write' writes 'GetString()' to 'w'
		Write (w io.Writer)
		
		// 'WriteFlat' writes 'GetFlatString' to 'w'
		WriteFlat (w io.Writer)
	}
	
	Object struct { //Json
		Fields Fields
	}
	
	Fields []Field
	
	Field struct {
		Name string
		Value Value
	}
	
	Array struct { //Json
		Elements Values
	}
	
	Values []Value
	
	Value interface {
		// Do-nothing func used to filter a 'Value'
		isValue ()
	}
	
	String struct { //Value
		S string
	}
	
	Integer struct { //Value
		N int64
	}
	
	Float struct { //Value
		F float64
	}
	
	JsonVal struct { //Value
		Json Json
	}
	
	Bool struct { //Value
		Bool bool
	}
	
	Null struct { //Value
	}

)

// Stack used for building Json objects
type (
	
	stacker interface {
		next () stacker
		setNext (s stacker)
	}
	
	stack struct {
		nxt stacker
	}
	
	startObj struct {
		stack
	}
	
	startArr struct {
		stack
	}
	
	stValue struct {
		stack
		val Value
	}
	
	stField struct {
		stack
		name string
		value Value
	}
	
	// A 'Maker' builds Json by stack manipulation
	Maker struct {
		stk stacker
	}

)

var (
		
	wd = F.Join(RS.FindDir(), jsonDir)
	compPath = F.Join(wd, compName)
	comp *C.Compiler // Compiler of json texts

)

func (*String) isValue () {}

func (*Integer) isValue () {}

func (*Float) isValue () {}

func (*JsonVal) isValue () {}

func (*Bool) isValue () {}

func (*Null) isValue () {}

func GetField (o *Object, name string) *Field {
	for _, f := range o.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func GetString (o *Object, name string) (string, bool) {
	f := GetField(o, name)
	if f == nil {
		return "", false
	}
	if f.Value == nil {
		return "", false
	}
	s, ok := f.Value.(*String)
	if !ok {
		return "", false
	}
	return s.S, true
}

func GetInt (o *Object, name string) (int64, bool) {
	f := GetField(o, name)
	if f == nil {
		return 0, false
	}
	if f.Value == nil {
		return 0, false
	}
	i, ok := f.Value.(*Integer)
	if !ok {
		return 0, false
	}
	return i.N, true
}

func GetFloat (o *Object, name string) (float64, bool) {
	f := GetField(o, name)
	if f == nil {
		return 0.0, false
	}
	if f.Value == nil {
		return 0.0, false
	}
	ff, ok := f.Value.(*Float)
	if !ok {
		return 0, false
	}
	return ff.F, true
}

func GetJson (o *Object, name string) (Json, bool) {
	f := GetField(o, name)
	if f == nil {
		return nil, false
	}
	if f.Value == nil {
		return nil, false
	}
	j, ok := f.Value.(*JsonVal)
	if !ok {
		return nil, false
	}
	return j.Json, true
}

func GetBool (o *Object, name string) (bool, bool) {
	f := GetField(o, name)
	if f == nil {
		return false, false
	}
	if f.Value == nil {
		return false, false
	}
	b, ok := f.Value.(*Bool)
	if !ok {
		return false, false
	}
	return b.Bool, true
}

func GetNull (o *Object, name string) bool {
	f := GetField(o, name)
	if f == nil {
		return false
	}
	if f.Value == nil {
		return false
	}
	_, ok := f.Value.(*Null)
	return ok
}

//*********** Json -> Go ***********

// Implementation of the standard procedures of Babel

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
} //ReadInt

func (c *compilationer) Read () (ch rune, cLen int) {
	if c.pos >= c.size {
		ch = C.EOF1
	} else {
		ch = c.r[c.pos]
		c.pos++
	}
	cLen = 1
	return
} //Read

func (c *compilationer) Pos () int {
	return c.pos
} //Pos

func (c *compilationer) SetPos (pos int) {
	c.pos = M.Min(pos, c.size)
} //SetPos

func (c *compilationer) Error (pos, line, col int, msg string) {
}

func (c *compilationer) Map (index string) string {
	return index
}

const (
	
	object = 1
	array = 2
	nilC = 3
	cons = 4
	
	json = 1
	stringC = 2
	number = 3
	trueC = 4
	falseC = 5
	null = 6

)

func eval (o *C.Object) Json {
	
	EvalVal := func (o *C.Object) Value {
		var v Value
		switch o.ObjType() {
		case C.UserObj:
			v = o.ObjUser().(Value)
		case C.TermObj:
			v = &JsonVal{eval(o)}
		}
		return v
	} //EvalVal

	EvalObj := func (o *C.Object) Fields {
		oo := o; n := 0
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(3)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		f := make(Fields, n)
		for i := 0; i < n; i++ {
			oo := o.ObjTermSon(1)
			s := oo.ObjString()
			f[i].Name = s[1:len(s) - 1]
			f[i].Value = EvalVal(o.ObjTermSon(2))
			o = o.ObjTermSon(3)
		}
		return f
	} //EvalObj

	EvalArr := func (o *C.Object) Values {
		oo := o; n := 0
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		v := make(Values, n)
		for i := 0; i < n; i++ {
			v[i] = EvalVal(o.ObjTermSon(1))
			o = o.ObjTermSon(2)
		}
		return v
	} //EvalArr

	//eval
	var j Json
	M.Assert(o.ObjType() == C.TermObj, 20)
	switch o.ObjFunc() {
	case object:
		j = &Object{EvalObj(o.ObjTermSon(1))}
	case array:
		j = &Array{EvalArr(o.ObjTermSon(1))}
	}
	return j
} //eval

func (c *compilationer) Execution (fNum, parsNb int, pars C.ObjectsList) (o *C.Object, res C.Anyptr, ok bool) {

	MakeString :=  func (rs []rune) string {
		sp := ""
		m := len(rs)
		pos := 0
		for pos < m {
			if rs[pos] == '\\' {
				pos++
				switch rs[pos] {
				case '"':
					sp += string(double_quote)
				case '\\':
					sp += string(reverse_solidus)
				case '/':
					sp += string(solidus)
				case 'b':
					sp += string(backspace)
				case 'f':
					sp += string(form_feed)
				case 'n':
					sp += string(line_feed)
				case 'r':
					sp += string(carriage_return)
				case 't':
					sp += string(horizontal_tab)
				case 'u':
					n := rune(0)
					for j := 1; j <= 4; j++ {
						var x rune
						pos++
						switch c := rs[pos]; {
						case '0' <= c && c <= '9':
							x = c - '0'
						case 'A' <= c && c <= 'F':
							x = c - 'A' + 0xA
						case 'a' <= c && c <= 'f':
							x = c - 'a' + 0xA
						}
						n = n * 0x10 + x
					}
					sp += string(n)
				default:
					M.Halt(100);
				}
			} else {
				sp += string(rs[pos])
			}
			pos++
		}
		return sp
	} //MakeString
	
	//Execution
	switch fNum {
	case json:
		o = pars[0]
		if o.ErrorIn() {
			ok = false
			return
		}
		c.json = eval(o)
	case stringC:
		o = pars[0]
		if o.ErrorIn() {
			ok = false
			return
		}
		r := []rune(o.ObjString())
		res = &String{MakeString(r[1:len(r) - 1])}
	case number:
		o = pars[0]
		if o.ErrorIn() {
			ok = false
			return
		}
		s := o.ObjString()
		i, err := SC.ParseInt(s, 0, 64)
		if err == nil {
			res = &Integer{i}
		} else {
			f, err := SC.ParseFloat(s, 64); M.Assert(err == nil, 100)
			res = &Float{f}
		}
	case trueC, falseC:
		res = &Bool{fNum == trueC}
	case null:
		res = &Null{}
	}
	ok = true
	return
} //Execution

// 'Compile' builds a Json from an 'io.ReadSeeker'
func Compile (rs io.ReadSeeker) Json {
	n, err := rs.Seek(0, io.SeekEnd); M.Assert(err == nil, err, 100)
	_, err = rs.Seek(0, io.SeekStart); M.Assert(err == nil, err, 101)
	b := make([]byte, n)
	_, err = io.ReadFull(rs, b); M.Assert(err == nil, err, 102)
	r := []rune(string(b))
	co := &compilationer{r: r, size: len(r), pos: 0, json: nil}
	c := C.NewCompilation(co)
	if c.Compile(comp, true) {
		return co.json
	}
	return nil
} //Compile

type
	
	ReadSeekCloser interface {
		io.ReadCloser
		io.Seeker
	}

// string -> Json
func ReadString (s string) Json {
	return Compile(strings.NewReader(s))
} //ReadString

// ReadSeekCloser -> Json
func ReadRSC (rsc ReadSeekCloser) Json {
	defer rsc.Close()
	return Compile(rsc)
} //ReadRSC

// os.File -> Json
func ReadFile (path string ) Json {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return ReadRSC(f)
} //ReadFile

// ApplyTo modifies obj according to j; warning: obj must be a pointer to a global, exported and modifiable record variable of its module, and only its exported and modifiable fields are taken into account; in obj, records must not be extended (i.e. their static types must match their dynamic types), but pointers to arrays, if their lengths are not already correct, are allocated with the same length as in j, provided that their BaseTyp is in {(Meta) boolTyp, sCharTyp, charTyp, byteTyp, sIntTyp, intTyp, sRealTyp, realTyp, longTyp, recTyp} (in particular, no pointers; if you want pointers, replace them by records with a pointer field).
func ApplyTo (j Json, obj interface{}) {
	buf := new(strings.Builder)
	j.Write(buf)
	J.Unmarshal([]byte(buf.String()), obj)
} //ApplyTo

func writeJson (w io.Writer, j Json, flat bool, indent int) {
	
	lnC := func () {
		if !flat {
			fmt.Fprintln(w)
		}
	} //lnC
	
	Indent := func (indent int) {
		if !flat {
			for indent > 0 {
				fmt.Fprint(w, "\t")
				indent--
			}
		}
	} //Indent
	
	WriteString := func (s string) {
		
		JsonFilter := func (s string) string {
			ss := ""
			for _, r := range s {
				escape := true
				var c rune
				switch r {
				case double_quote, reverse_solidus:
					c = r
				case backspace:
					c = 'b'
				case form_feed:
					c = 'f'
				case line_feed:
					c = 'n'
				case carriage_return:
					c = 'r'
				case horizontal_tab:
					c = 't'
				default:
					escape = false
					c = r
				}
				if escape {
					ss += "\\"
				}
				ss += string(c)
			}
			return ss
		} //JsonFilter
		
		//WriteString
		fmt.Fprintf(w, "%s", "\"")
		fmt.Fprintf(w, "%s", JsonFilter(s))
		fmt.Fprintf(w, "%s", "\"")
	} //WriteString
	
	WriteVal := func (v Value, indent int) {
		if _, b := v.(*JsonVal); !b {
			Indent(indent)
		}
		switch v := v.(type) {
		case *String:
			WriteString(v.S)
		case *Integer:
			fmt.Fprintf(w, "%s", SC.FormatInt(v.N, 10))
		case *Float:
			fmt.Fprintf(w, "%s", SC.FormatFloat(v.F, 'g', -1, 64))
		case *JsonVal:
			writeJson(w, v.Json, flat, indent)
		case *Bool:
			if v.Bool {
				fmt.Fprintf(w, "%s", "true")
			} else {
				fmt.Fprintf(w, "%s", "false")
			}
		case *Null:
			fmt.Fprintf(w, "%s", "null")
		default:
			M.Halt("Incorrect value type", 100)
		}
	} //WriteVal
	
	WriteArr := func (vs Values, indent int) {
		if len(vs) > 0 {
			for i, v := range vs {
				if i > 0 {
					fmt.Fprintf(w, "%s", ",")
					lnC()
				}
				WriteVal(v, indent)
			}
			lnC()
		}
	} //WriteArr
	
	WriteObj := func (fs Fields, indent int) {
		
		WriteField := func (f *Field, indent int) {
			Indent(indent)
			WriteString(f.Name)
			fmt.Fprintf(w, "%s", ":")
			lnC()
			WriteVal(f.Value, indent + 1)
		} //WriteField
		
		//WriteObj
		if len(fs) > 0 {
			for i, f := range fs {
				if i > 0 {
					fmt.Fprintf(w, "%s", ",")
					lnC()
				}
				WriteField(&f, indent)
			}
			lnC()
		}
	} //WriteObj

	//writeJson
	switch j := j.(type) {
	case *Object:
		Indent(indent)
		fmt.Fprintf(w, "%s", "{")
		lnC()
		WriteObj(j.Fields, indent + 1)
		Indent(indent)
		fmt.Fprintf(w, "%s", "}")
	case *Array:
		Indent(indent)
		fmt.Fprintf(w, "%s", "[")
		lnC()
		WriteArr(j.Elements, indent + 1)
		Indent(indent)
		fmt.Fprintf(w, "%s", "]")
	}
} //writeJson

func (o *Object) GetString () string {
	M.Assert(o != nil, 20)
	var buf = new(bytes.Buffer)
	writeJson(buf, o, false, 0)
	return buf.String()
} //GetString

// Write writes a with w
func (a *Array) GetString () string {
	M.Assert(a != nil, 20)
	var buf = new(bytes.Buffer)
	writeJson(buf, a, false, 0)
	return buf.String()
} //GetString

// Write writes o with w
func (o *Object) GetFlatString () string {
	M.Assert(o != nil, 20)
	var buf = new(bytes.Buffer)
	writeJson(buf, o, true, 0)
	return buf.String()
} //GetFlatString

// Write writes a with w
func (a *Array) GetFlatString () string {
	M.Assert(a != nil, 20)
	var buf = new(bytes.Buffer)
	writeJson(buf, a, true, 0)
	return buf.String()
} //GetFlatString

// Write writes o with w
func (o *Object) Write (w io.Writer) {
	writeJson(w, o, false, 0)
} //Write

// Write writes a with w
func (a *Array) Write (w io.Writer) {
	writeJson(w, a, false, 0)
} //Write

// WriteFlat writes o with w
func (o *Object) WriteFlat (w io.Writer) {
	writeJson(w, o, true, 0)
} //WriteFlat

// WriteFlat writes a with w
func (a *Array) WriteFlat (w io.Writer) {
	writeJson(w, a, true, 0)
} //WriteFlat

func DecodeString (s string) string {
	
	const (
		
		backspace = string(0x8)
		form_feed = string(0xC)
		line_feed = string(0xA)
		carriage_return = string(0xD)
		horizontal_tab = string(0x9)
	
	)
	
	rs := []rune(s)
	m := len(rs)
	sp := ""
	pos := 0
	for pos < m {
		if rs[pos] == '\\' {
			pos++
			switch rs[pos] {
			case '"':
				sp += "\""
			case '\\':
				sp += "\\"
			case '/':
				sp += "/"
			case 'b':
				sp += backspace
			case 'f':
				sp += form_feed
			case 'n':
				sp += line_feed
			case 'r':
				sp += carriage_return
			case 't':
				sp += horizontal_tab
			case 'u':
				var n rune = 0
				for j := 1; j <= 4; j++ {
					pos++
					var x rune
					switch c := rs[pos]; {
						case '0' <= c && c <= '9':
							x = c - '0'
						case 'A' <= c && c <= 'F':
							x = c - 'A' + 0xA
						case 'a' <= c && c <= 'f':
							x = c - 'a' + 0xA
					}
					n = n * 0x10 + x
				}
				sp += string(n)
			}
		} else {
			sp += string(rs[pos])
		}
		pos++
	}
	return sp
} //DecodeString

//*********** Go -> Json ***********

// stack procedures

func (s *stack) next () stacker {
	return s.nxt
}

func (s *stack) setNext (ss stacker) {
	s.nxt = ss
}

func (m *Maker) push (s stacker) {
	s.setNext(m.stk)
	m.stk = s
}

func (m *Maker) pull () stacker {
	M.Assert(m.stk != nil, 20);
	s := m.stk
	m.stk = s.next()
	return s
}

//StartObject begins a json object
func (m *Maker) StartObject () {
	m.push(new(startObj))
}

// StartArray begins a json array
func (m *Maker) StartArray () {
	m.push(new(startArr));
}

// PushString pushes the string s on the stack
func (m *Maker) PushString (s string) {
	m.push(&stValue{val: &String{s}})
}

// PushInteger pushes the integer n on the stack
func (m *Maker) PushInteger (n int64) {
	m.push(&stValue{val: &Integer{n}})
}

// PushReal pushes the float f on the stack
func (m *Maker) PushFloat (f float64) {
	m.push(&stValue{val: &Float{f}})
}

// PushBoolean pushes the boolean b on the stack
func (m *Maker) PushBoolean (b bool) {
	m.push(&stValue{val: &Bool{b}})
}

// PushNull pushes a json null value on the stack
func (m *Maker) PushNull () {
	m.push(&stValue{val: new(Null)})
}

// PushValue pushes the value v on the stack
func (m *Maker) PushValue (v Value) {
	m.push(&stValue{val: v})
}

// PushJson pushes the Json j on the stack
func (m *Maker) PushJson (j Json) {
	m.push(&stValue{val: &JsonVal{j}})
}

// PushField pushes the field f on the stack
func (m *Maker) PushField (f *Field) {
	m.push(&stField{name: f.Name, value: f.Value})
}

// BuildField builds the json object field whose name is name and whose value is the last element on the stack and replaces this element with the built field
func (m *Maker) BuildField (name string) {
	s, ok := m.pull().(*stValue)
	M.Assert(ok, 100)
	m.push(&stField{name: name, value: s.val})
}

// BuildObject builds a json object from all the fields stacked from the last StartObject call and stacks it
func (m *Maker) BuildObject () {
	s := m.stk
	n := 0
	for {
		ok := s != nil
		if ok {
			_, ok = s.(*stField)
		}
		if !ok {break}
		n++
		s = s.next()
	}
	ok := s != nil
	if ok {
		_, ok = s.(*startObj)
	}
	M.Assert(ok, 100)
	o := new(Object)
	if n == 0 {
		o.Fields = nil
	} else {
		o.Fields = make(Fields, n)
		for i := n - 1; i >= 0; i-- {
			sf := m.pull().(*stField)
			o.Fields[i] = Field{sf.name, sf.value}
		}
	}
	m.pull()
	m.push(&stValue{val: &JsonVal{o}})
}

// BuildArray builds a json array from all the values stacked from the last StartArray call and stacks it
func (m *Maker) BuildArray () {
	s := m.stk
	n := 0
	for {
		ok := s != nil
		if ok {
			_, ok = s.(*stValue)
		}
		if !ok {break}
		n++
		s = s.next()
	}
	ok := s != nil
	if ok {
		_, ok = s.(*startArr)
	}
	M.Assert(ok, 100)
	a := new(Array)
	if n == 0 {
		a.Elements = nil
	} else {
		a.Elements = make(Values, n)
		for i := n - 1; i >= 0; i-- {
			a.Elements[i] = m.pull().(*stValue).val
		}
	}
	m.pull()
	m.push(&stValue{val: &JsonVal{a}})
}

func (m *Maker) Roll (n int) {
	M.Assert(n >= 1, 20)
	s := m.stk
	for s != nil && n > 2 {
		s = s.next(); n--
	}
	M.Assert(s != nil && s.next() != nil, 21)
	if n >= 2 {
		ss := s.next()
		s.setNext(ss.next())
		ss.setNext(m.stk)
		m.stk = ss
	}
}

func (m *Maker) RollD (n int) {
	M.Assert(n >= 1, 20)
	s := m.stk
	for s != nil && n > 1 {
		s = s.next(); n--
	}
	M.Assert(s != nil, 21)
	if n >= 2 {
		ss := m.stk
		m.stk = m.stk.next()
		ss.setNext(s.next())
		s.setNext(ss)
	}
}

func (m *Maker) Swap () {
	m.Roll(2)
}

// GetJson builds and returns a Json built from the last object on the stack
func (m *Maker) GetJson () Json {
	ok := m.stk != nil
	var (sv *stValue; jv *JsonVal)
	if ok {
		sv, ok = m.stk.(*stValue)
	}
	if ok {
		jv, ok = sv.val.(*JsonVal)
	}
	M.Assert(ok, 20)
	return jv.Json
}

func NewMaker () *Maker {
	return &Maker{stk: (*stack)(nil)}
}

func process (v R.Value, m *Maker) {
	switch v.Kind() {
	case R.Bool:
		m.PushBoolean(v.Bool())
	case R.Int32:
		if v.Type().Name() == "rune" {
			m.PushString(string(rune(v.Int())))
		} else {
			m.PushInteger(v.Int())
		}
	case R.Int, R.Int8, R.Int16, R.Int64:
		m.PushInteger(v.Int())
	case R.Uint, R.Uint8, R.Uint16, R.Uint32, R.Uint64, R.Uintptr:
		m.PushInteger(int64(v.Int()))
	case R.Float32, R.Float64:
		m.PushFloat(v.Float())
	case R.Complex64, R.Complex128:
		c := v.Complex()
		m.StartObject()
		m.PushFloat(real(c))
		m.BuildField("real")
		m.PushFloat(imag(c))
		m.BuildField("imag")
		m.BuildObject()
	case R.Array, R.Slice:
		m.StartArray()
		for k := 0; k < v.Len(); k++ {
			process(v.Index(k), m)
		}
		m.BuildArray()
	case R.Interface:
		p := v.Elem()
		process(p, m)
	case R.Ptr:
		if v.IsNil() {
			m.PushNull()
		} else {
			p := v.Elem()
			process(p, m)
		}
	case R.String:
		m.PushString(v.String())
	case R.Struct:
		t := v.Type()
		n := t.NumField()
		m.StartObject()
		for k := 0; k < n; k++ {
			process(v.Field(k), m)
			m.BuildField(t.Field(k).Name)
		}
		m.BuildObject()
	default:
		M.Halt("Incorrect value type", 100)
	}
} //process

// BuildJsonFrom builds and returns a Json describing obj; warning: obj must be nil or a pointer to a global exported struct variable of its package, and only its exported fields are taken into account
func BuildJsonFrom (obj interface{}) Json {
	if obj == nil {
		return nil
	}
	v := R.ValueOf(obj)
	m := NewMaker()
	process(v, m)
	return m.GetJson()
} //BuildJsonFrom

func SetRComp (rComp io.Reader) {
	comp = C.NewDirectory(&directory{r: bufio.NewReader(rComp)}).ReadCompiler()
}

func init () {
	// Compilers reading
	f, err := os.Open(compPath)
	if err == nil {
		defer f.Close()
		SetRComp(f)
	}
} //init
