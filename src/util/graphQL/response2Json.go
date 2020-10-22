package graphQL

import (
	
	A	"util/avl"
	J	"util/json"

)

func makeErrors (mk *J.Maker, errors *A.Tree) {
	if errors != nil && !errors.IsEmpty() {
		mk.StartArray()
		e := errors.Previous(nil)
		for e != nil {
			el := e.Val().(*ErrorElem)
			mk.StartObject()
			mk.PushString(el.Message)
			mk.BuildField("message")
			if el.Location != nil {
				mk.StartObject()
				mk.PushInteger(int64(el.Location.Pos))
				mk.BuildField("position")
				mk.PushInteger(int64(el.Location.Line))
				mk.BuildField("line")
				mk.PushInteger(int64(el.Location.Column))
				mk.BuildField("column")
				mk.BuildObject()
				mk.BuildField("location")
			}
			p := el.Path
			if p != nil {
				mk.StartArray()
				for {
					switch p := p.(type) {
					case *PathString:
						mk.PushString(p.S)
					case *PathNb:
						mk.PushInteger(int64(p.N))
					}
					p = p.Next()
					if p == nil {break}
				}
				mk.BuildArray()
				mk.BuildField("path")
			}
			mk.BuildObject()
			e = errors.Previous(e)
		}
		mk.BuildArray()
		mk.BuildField("errors")
	}
} //makeErrors

func makeValue (mk *J.Maker, value Value) {
	
	makeList := func (lv *ListValue) {
		mk.StartArray()
		for l := lv.First(); l != nil; l = lv.Next(l) {
			makeValue(mk, l.Value)
		}
		mk.BuildArray()
	} //makeList
	
	//makeValue
	switch value := value.(type) {
	case *IntValue:
		mk.PushInteger(value.Int)
	case *FloatValue:
		mk.PushFloat(value.Float)
	case *StringValue:
		if value == nil {
			mk.PushNull()
		} else {
			mk.PushString(value.String.S)
		}
	case *BooleanValue:
		mk.PushBoolean(value.Boolean)
	case *NullValue:
		mk.PushNull()
	case *EnumValue:
		mk.PushString(value.Enum.S)
	case *ListValue:
		makeList(value)
	case *OutputObjectValue:
		makeObject(mk, value)
	}
} //makeValue

func makeObject (mk *J.Maker, ov *OutputObjectValue) {
	
	makeField := func (f *ObjectField) {
		makeValue(mk, f.Value)
		mk.BuildField(f.Name.S)
	} //makeField
	
	//makeObject
	mk.StartObject()
	for f := ov.First(); f != nil; f = ov.Next(f) {
		makeField(f)
	}
	mk.BuildObject()
} //makeObject

func buildInstant (mk *J.Maker, data *OutputObjectValue) {
	if data != nil {
		makeObject(mk, data)
		mk.BuildField("data")
	}
} //buildInstant

func ResponseToJson (r Response) J.Json {
	mk := J.NewMaker()
	mk.StartObject()
	makeErrors(mk, r.Errors())
	switch r := r.(type) {
	case *InstantResponse:
		buildInstant(mk, r.Data)
	case *SubscribeResponse:
	}
	mk.BuildObject()
	return mk.GetJson()
} //ResponseToJson
