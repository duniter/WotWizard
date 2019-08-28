package json

import (
	
		"testing"
		"os"
		"fmt"

)

func TestJ1 (t *testing.T) {
	var (
		
		v0, v1, v2 Value
		j Json
	
	)
	v0 = new(Null)
	v1 = 42
	v2 = "OK"
	j = &Array{V: Values{v0, v1, v2}}
	Fprint(os.Stdout, j)
	fmt.Fprint(os.Stdout, "\n")
}

func TestJ2 (t *testing.T) {
	mk := NewMaker()
	mk.StartObject()
	mk.PushBoolean(true)
	mk.BuildField("truth")
	mk.PushBoolean(false)
	mk.BuildField("falseness")
	mk.BuildObject()
	j := mk.GetJson()
	Fprint(os.Stdout, j)
	fmt.Fprint(os.Stdout, "\n")
}

func TestJ3 (t *testing.T) {
	mk := NewMaker()
	mk.StartArray()
	mk.PushNull()
	mk.PushInteger(42)
	mk.PushString("OK\"OK\"")
	mk.StartObject()
	mk.PushBoolean(true)
	mk.BuildField("truth")
	mk.PushBoolean(false)
	mk.BuildField("falseness")
	mk.BuildObject()
	mk.PushFloat(3.14159265)
	mk.BuildArray()
	j := mk.GetJson()
	Fprint(os.Stdout, j)
	fmt.Fprint(os.Stdout, "\n")
}

func TestJ4 (t *testing.T) {
	mk := NewMaker()
	mk.StartObject()
	mk.PushBoolean(true)
	mk.BuildField("truth")
	mk.StartArray()
	mk.PushNull()
	mk.PushInteger(42)
	mk.PushString("OK\"OK\"")
	mk.PushFloat(3.14159265)
	mk.BuildArray()
	mk.BuildField("list")
	mk.PushBoolean(false)
	mk.BuildField("falseness")
	mk.BuildObject()
	j := mk.GetJson()
	Fprint(os.Stdout, j)
	fmt.Fprint(os.Stdout, "\n")
}
