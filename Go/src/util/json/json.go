package json

import (
	
	J	"encoding/json"
	M	"util/misc"
			"fmt"
			"io"
			"os"

)

const (
	
	//compName = "util/json/json.tbl"

)

type (
	
	Json interface {
	}
	
	Object struct {
		F Fields
	}
	
	Fields []Field
	
	Field struct {
		N string
		V Value
	}
	
	Array struct {
		V Values
	}
	
	Values []Value
	
	Value interface {
	}
	
	JsonVal struct {
		J Json
	}
	
	Null struct {
	}

)

func indent (w io.Writer, ind int) {
	for i := 0; i < ind; i++ {
		fmt.Fprint(w, "\t")
	}
}

func printVal (w io.Writer, v Value, ind int) {
	switch val := v.(type) {
	case *JsonVal:
		printJson(w, val.J, ind)
	case *Null:
		indent(w, ind)
		fmt.Fprint(w, "null")
	case string:
		indent(w, ind)
		fmt.Fprint(w, "\"", val, "\"")
	default:
		indent(w, ind)
		fmt.Fprint(w, val)
	}
}

func printField (w io.Writer, f *Field, ind int) {
	indent(w, ind)
	fmt.Fprint(w, "\"", f.N, "\":\n")
	printVal(w, f.V, ind + 1)
}

func printObj (w io.Writer, o *Object, ind int) {
	if o.F != nil {
		printField(w, &o.F[0], ind)
		for i := 1; i < len(o.F); i++ {
			fmt.Fprint(w, ",\n")
			printField(w, &o.F[i], ind)
		}
		fmt.Fprint(w, "\n")
	}
}

func printArr (w io.Writer, a *Array, ind int) {
	if a.V != nil {
		printVal(w, a.V[0], ind)
		for i := 1; i < len(a.V); i++ {
			fmt.Fprint(w, ",\n")
			printVal(w, a.V[i], ind)
		}
		fmt.Fprint(w, "\n")
	}
}

func printJson (w io.Writer, j Json, ind int) {
	switch js := j.(type) {
	case *Object:
		indent(w, ind)
		fmt.Fprint(w, "{\n")
		printObj(w, js, ind + 1)
		indent(w, ind)
		fmt.Fprint(w, "}")
	case *Array:
		indent(w, ind)
		fmt.Fprint(w, "[\n")
		printArr(w, js, ind + 1)
		indent(w, ind)
		fmt.Fprint(w, "]")
	}
}

func Fprint (w io.Writer, j Json) {
	printJson(w, j, 0)
}

//************ Json -> Go ************

func ReadString (s string, v interface{}) (ok bool) {
	return J.Unmarshal([]byte(s), v) == nil
}

func ReadReader (r io.Reader, n int, v interface{}) (ok bool) {
	M.Assert(n >=0, 20)
	buf := make([]byte, n)
	m, err := r.Read(buf)
	if !(err == nil || err == io.EOF) {
		return false
	}
	err = J.Unmarshal(buf[: m], v)
	return err == nil
}

func ReadFile (path string, v interface{}) (ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return ReadReader(f, int(fi.Size()), v)
}

//************ Go -> Json ************

// Stack

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
	
	Maker struct {
		stk stacker
	}

)

func (s *stack) next () stacker {
	return s.nxt
}

func (s *stack) setNext (ss stacker) {
	s.nxt = ss
}

// Stack procedures

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
	m.push(&stValue{val: s})
}

// PushInteger pushes the integer n on the stack
func (m *Maker) PushInteger (n int64) {
	m.push(&stValue{val: n})
}

// PushReal pushes the float f on the stack
func (m *Maker) PushFloat (f float64) {
	m.push(&stValue{val: f})
}

// PushBoolean pushes the boolean b on the stack
func (m *Maker) PushBoolean (b bool) {
	m.push(&stValue{val: b})
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
	m.push(&stValue{val: &JsonVal{J: j}})
}

// PushField pushes the field f on the stack
func (m *Maker) PushField (f *Field) {
	m.push(&stField{name: f.N, value: f.V})
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
		o.F = nil
	} else {
		o.F = make(Fields, n)
		for i := n - 1; i >= 0; i-- {
			sf := m.pull().(*stField)
			o.F[i] = Field{N: sf.name, V: sf.value}
		}
	}
	m.pull()
	m.push(&stValue{val: &JsonVal{J: o}})
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
		a.V = nil
	} else {
		a.V = make(Values, n)
		for i := n - 1; i >= 0; i-- {
			a.V[i] = m.pull().(*stValue).val
		}
	}
	m.pull()
	m.push(&stValue{val: &JsonVal{J: a}})
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
	return jv.J
}

func NewMaker () *Maker {
	return &Maker{stk: (*stack)(nil)}
}

func BuildJsonStringFrom (v interface{}) (string, bool) {
	buf, err := J.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", false
	}
	return string(buf), true
}

func FprintJsonOf (w io.Writer, v interface{}) (ok bool) {
	s, ok := BuildJsonStringFrom(v)
	if !ok {
		return
	}
	_, err := fmt.Fprint(w, s)
	return err == nil
}
