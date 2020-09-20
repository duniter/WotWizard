/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package parameters

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
		"unicode"

)

func allParametersR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	o := B.ParsJ().(*J.Object)
	l := G.NewListValue()
	for _, f := range o.Fields {
		rs := []rune(f.Name)
		rs[0] = unicode.ToLower(rs[0])
		name := string(rs)
		l.Append(GQ.Wrap(name, G.TransVal(f.Value)))
	}
	return l
} //allParametersR

func parameterR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	o := B.ParsJ().(*J.Object)
	var v G.Value
	if !G.GetValue(argumentValues, "name", &v) {
		return G.MakeNullValue()
	}
	switch v := v.(type) {
	case *G.EnumValue:
		name0 := v.Enum.S
		rs := []rune(name0)
		rs[0] = unicode.ToUpper(rs[0])
		name := string(rs)
		for _, f := range o.Fields {
			if f.Name == name {
				return GQ.Wrap(name0, G.TransVal(f.Value))
			}
		}
		return G.MakeNullValue()
	case *G.NullValue:
		return v
	default:
		M.Halt(v, 100)
		return nil
	}
} //parameterR

func parameterNameR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch name := GQ.Unwrap(rootValue, 0).(type) {
	case string:
		return G.MakeEnumValue(name)
	case *G.NullValue:
		return name
	default:
		M.Halt(name, 100)
		return nil
	}
} //parameterNameR

func parameterCommentR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch name := GQ.Unwrap(rootValue, 0).(type) {
	case string:
		d := GQ.TS().GetTypeDefinition("ParameterName")
		ok := d != nil
		var e *G.EnumTypeDefinition
		if ok {
			e, ok = d.(*G.EnumTypeDefinition)
		}
		if ok {
			for _, v := range e.EnumValsDef {
				if v.EnumVal.Enum.S == name {
					return v.Desc
				}
			}
		}
		return G.MakeNullValue()
	case *G.NullValue:
		return name
	default:
		M.Halt(name, 100)
		return nil
	}
} //parameterCommentR

func parameterValueR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	return GQ.Unwrap(rootValue, 1).(G.Value)
} //parameterValueR

func parameterTypeR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch name := GQ.Unwrap(rootValue, 0).(type) {
	case string:
		var t string
		switch name {
		case "ud0", "sigStock", "sigQty", "stepMax", "medianTimeBlocks", "dtDiffEval":
			t = "INTEGER"
		case "c", "xpercent", "percentRot":
			t = "FLOAT"
		case "dt", "sigPeriod", "sigWindow", "sigValidity", "sigReplay", "idtyWindow", "msWindow", "msPeriod", "msValidity", "avgGenTime", "dtReeval", "txWindow":
			t = "DURATION"
		case "udTime0", "udReevalTime0":
			t = "DATE"
		default:
			M.Halt(name, 100)
		}
		return G.MakeEnumValue(t)
	case *G.NullValue:
		return name
	default:
		M.Halt(name, 101)
		return nil
	}
} //parameterTypeR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "allParameters", allParametersR)
	ts.FixFieldResolver("Query", "parameter", parameterR)
	ts.FixFieldResolver("Parameter", "name", parameterNameR)
	ts.FixFieldResolver("Parameter", "comment", parameterCommentR)
	ts.FixFieldResolver("Parameter", "value", parameterValueR)
	ts.FixFieldResolver("Parameter", "par_type", parameterTypeR)
} //fixFieldResolvers

func init () {
	fixFieldResolvers(GQ.TS())
} //init
