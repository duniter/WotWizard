package graphQL
	
import (
	
	M	"util/misc"
	S	"util/extStrings"
	SC	"strconv"
		"bytes"
		"fmt"
		"io"

)

func trans (s string) string {
	buf := new(bytes.Buffer)
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
			buf.WriteRune(c)
		}
		buf.WriteRune(c)
	}
	return buf.String()
} //trans

func sep (w S.Writer) {
	w.WriteString(" ")
} //sep

func sepF (w S.Writer, flat bool) {
	if flat {
		w.WriteString(" ")
	}
} //sepF

func sepC (w S.Writer, flat bool) {
	if !flat {
		w.WriteString(" ")
	}
} //sepC

func lnC (w S.Writer, flat bool) {
	if !flat {
		w.WriteChar(line_feed)
	}
} //lnC

func indent (w S.Writer, flat bool, ind int) {
	if !flat {
		for ind > 0 {
			w.WriteChar(horizontal_tab)
			ind--
		}
	}
} //indent

func writeValue (w S.Writer, flat bool, v Value) {
	
	WriteField := func (f *ObjectField) {
		w.WriteString(f.Name.S)
		w.WriteString(":"); sepC(w, flat)
		writeValue(w, flat, f.Value)
	} //WriteField
	
	//writeValue
	if v != nil {
		switch v := v.(type) {
		case *Variable:
			w.WriteString("$")
			w.WriteString(v.Name.S)
		case *IntValue:
			w.WriteString(SC.FormatInt(v.Int, 10))
		case *FloatValue:
			w.WriteString(SC.FormatFloat(v.Float, 'g', -1, 64))
		case *StringValue:
			w.WriteString("\"")
			w.WriteString(trans(v.String.S))
			w.WriteString("\"")
		case *BooleanValue:
			if v.Boolean {
				w.WriteString("true")
			} else {
				w.WriteString("false")
			}
		case *NullValue:
			w.WriteString("null")
		case *EnumValue:
			w.WriteString(v.Enum.S)
		case *ListValue:
			w.WriteString("[")
			if l := v.First(); l != nil {
				writeValue(w, flat, l.Value)
				for l := v.Next(l); l != nil; l = v.Next(l) {
					sep(w)
					writeValue(w, flat, l.Value)
				}
			}
			w.WriteString("]")
		case *OutputObjectValue:
			w.WriteString("{")
			if f := v.First(); f != nil {
				WriteField(f)
				for f := v.Next(f); f != nil; f = v.Next(f) {
					sep(w)
					WriteField(f)
				}
			}
			w.WriteString("}")
		}
	}
} //writeValue

func writeType (w S.Writer, t Type) {
	switch t := t.(type) {
	case *NonNullType:
		writeType(w, t.NullT)
		w.WriteString("!")
	case *ListType:
		w.WriteString("[")
		writeType(w, t.ItemT)
		w.WriteString("]")
	case *NamedType:
		w.WriteString(t.Name.S)
	}
} //writeType

func writeArguments (w S.Writer, flat bool, as Arguments) {
	
	WriteArgument := func (a *Argument) {
		w.WriteString(a.Name.S)
		w.WriteString(":"); sepC(w, flat)
		writeValue(w, flat, a.Value)
	} //WriteArgument
	
	//writeArguments
	if len(as) > 0 {
		w.WriteString("(")
		for i, a := range as {
			if i > 0 {
				sep(w)
			}
			WriteArgument(&a)
		}
		w.WriteString(")")
	}
} //writeArguments

func writeDirectives (w S.Writer, flat bool, dirs Directives) {
	
	WriteDirective := func (dir *Directive) {
		sepC(w, flat); w.WriteString("@")
		w.WriteString(dir.Name.S)
		writeArguments(w, flat, dir.Args)
	} //WriteDirective
	
	//writeDirectives
	for _, dir := range dirs {
		WriteDirective(&dir)
	}
} //writeDirectives

func writeSelectionSet (w S.Writer, flat bool, set SelectionSet, ind int) {
	
	WriteSelection := func (sel Selection, ind int) {
		
		WriteField := func (f *Field, ind int) {
			indent(w, flat, ind)
			if f.Alias.S != f.Name.S {
				w.WriteString(f.Alias.S)
				w.WriteString(":"); sepC(w, flat)
			}
			w.WriteString(f.Name.S)
			writeArguments(w, flat, f.Arguments)
			writeDirectives(w, flat, f.Dirs)
			writeSelectionSet(w, flat, f.SelSet, ind)
		} //WriteField
		
		WriteFragmentSpread := func (f *FragmentSpread, ind int) {
			indent(w, flat, ind)
			w.WriteString("...")
			w.WriteString(f.Name.S)
			writeDirectives(w, flat, f.Dirs)
		} //WriteFragmentSpread
		
		WriteInlineFragment := func (f *InlineFragment, ind int) {
			indent(w, flat, ind)
			w.WriteString("...")
			if f.TypeCond != nil {
				w.WriteString("on ")
				w.WriteString(f.TypeCond.Name.S)
			}
			writeDirectives(w, flat, f.Dirs)
			writeSelectionSet(w, flat, f.SelSet, ind)
		} //WriteInlineFragment
		
		//WriteSelection
		switch sel := sel.(type) {
		case *Field:
			WriteField(sel, ind)
		case *FragmentSpread:
			WriteFragmentSpread(sel, ind)
		case *InlineFragment:
			WriteInlineFragment(sel, ind)
		}
	} //WriteSelection
	
	//writeSelectionSet
	if set != nil {
		sepC(w, flat); w.WriteString("{"); lnC(w, flat)
		for i, sel := range set {
			if i > 0 {
				sepF(w, flat)
			}
			WriteSelection(sel, ind + 1); lnC(w, flat)
		}
		indent(w, flat, ind)
		w.WriteString("}")
	}
} //writeSelectionSet

func (doc *Document) writeCommon (flat bool) string {
	
	WriteDefinition := func (w S.Writer, d Definition, ind int) {
		
		WriteOperationType := func (n int) {
			switch n {
			case queryOp:
				w.WriteString("query")
			case mutationOp:
				w.WriteString("mutation")
			case subscriptionOp:
				w.WriteString("subscription")
			}
		} //WriteOperationType
		
		WriteDefaultValue := func (v Value) {
			if v != nil {
				sepC(w, flat); w.WriteString("="); sepC(w, flat)
				writeValue(w, flat, v)
			}
		} //WriteDefaultValue
		
		WriteExecutableDefinition := func (d ExecutableDefinition, ind int) {
			
			WriteOperationDefinition := func (d *OperationDefinition, ind int) {
				
				WriteVariableDefinitions := func (vds VariableDefinitions) {
					
					WriteVariableDefinition := func (vd *VariableDefinition) {
						w.WriteString("$")
						w.WriteString(vd.Var.S)
						w.WriteString(":"); sepC(w, flat)
						writeType(w, vd.Type)
						WriteDefaultValue(vd.DefVal)
					} //WriteVariableDefinition
					
					//WriteVariableDefinitions
					if len(vds) > 0 {
						w.WriteString("(")
						for _, vd := range vds {
							WriteVariableDefinition(vd)
						}
						w.WriteString(")")
					}
				} //WriteVariableDefinitions
				
				//WriteOperationDefinition
				indent(w, flat, ind)
				if !(flat && (d.OpType == queryOp) && (d.Name.S == "") && (d.VarDefs == nil) && (d.Dirs == nil)) {
					WriteOperationType(d.OpType)
					if d.Name.S != "" {
						sep(w)
						w.WriteString(d.Name.S)
					}
					WriteVariableDefinitions(d.VarDefs)
					writeDirectives(w, flat, d.Dirs)
				}
				writeSelectionSet(w, flat, d.SelSet, ind)
			} //WriteOperationDefinition
			
			WriteFragmentDefinition := func (d *FragmentDefinition, ind int) {
				indent(w, flat, ind)
				w.WriteString("fragment ")
				w.WriteString(d.Name.S)
				w.WriteString(" on ")
				w.WriteString(d.TypeCond.Name.S)
				writeDirectives(w, flat, d.Dirs)
				writeSelectionSet(w, flat, d.SelSet, ind)
			} //WriteFragmentDefinition
			
			//WriteExecutableDefinition
			switch d := d.(type) {
			case *OperationDefinition:
				WriteOperationDefinition(d, ind)
			case *FragmentDefinition:
				WriteFragmentDefinition(d, ind)
			}
		} //WriteExecutableDefinition
		
		WriteOperationTypeDefinitions := func (ops OperationTypeDefinitions, ind int) {
			
			WriteOperationTypeDefinition := func (op *OperationTypeDefinition, ind int) {
				indent(w, flat, ind)
				WriteOperationType(op.OpType)
				w.WriteString(":"); sepC(w, flat)
				w.WriteString(op.Type.Name.S)
			} //WriteOperationTypeDefinition
			
			//WriteOperationTypeDefinitions
			if ops != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat)
				for _, op := range ops {
					WriteOperationTypeDefinition(&op, ind + 1)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
			}
		} //WriteOperationTypeDefinitions
		
		WriteImplements := func (nts NamedTypes) {
			if len(nts) > 0 {
				sep(w); w.WriteString("implements"); sep(w)
				for i, nt := range nts {
					if i > 0 {
						sepC(w, flat); w.WriteString("&"); sepC(w, flat)
					}
					w.WriteString(nt.Name.S)
				}
			}
		} //WriteImplements
		
		WriteInputValueDefinition := func (ivd *InputValueDefinition, ind int) {
			if ivd.Desc != nil {
				indent(w, flat, ind)
				w.WriteString("\"")
				w.WriteString(trans(ivd.Desc.String.S))
				w.WriteString("\"")
				lnC(w, flat)
			}
			indent(w, flat, ind)
			w.WriteString(ivd.Name.S)
			w.WriteString(":"); sepC(w, flat)
			writeType(w, ivd.Type)
			WriteDefaultValue(ivd.DefVal)
			writeDirectives(w, flat, ivd.Dirs)
		} //WriteInputValueDefinition
		
		WriteArgumentsDefinition := func (ads ArgumentsDefinition, ind int) {
			if len(ads) > 0 {
				sepC(w, flat); w.WriteString("("); lnC(w, flat)
				for i, ad := range ads {
					if i > 0 {
						sepF(w, flat)
					}
					WriteInputValueDefinition(ad, ind + 1)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString(")")
			}
		} //WriteArgumentsDefinition
		
		WriteFieldsDefinition := func (fds FieldsDefinition, ind int) {
			
			WriteFieldDefinition := func (fd *FieldDefinition, ind int) {
				if fd.Desc != nil {
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(fd.Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString(fd.Name.S)
				WriteArgumentsDefinition(fd.ArgsDef, ind)
				w.WriteString(":"); sepC(w, flat)
				writeType(w, fd.Type)
				writeDirectives(w, flat, fd.Dirs)
			} //WriteFieldDefinition
			
			//WriteFieldsDefinition
			sepC(w, flat); w.WriteString("{"); lnC(w, flat); lnC(w, flat)
			for i, fd := range fds {
				if i > 0 {
					sepF(w, flat)
				}
				WriteFieldDefinition(fd, ind + 1)
				lnC(w, flat); lnC(w, flat)
			}
			indent(w, flat, ind)
			w.WriteString("}")
		} //WriteFieldsDefinition
		
		WriteUnionMemberTypes := func (ns NamedTypes, ind int) {
			sepC(w, flat); w.WriteString("="); lnC(w, flat)
			for i, n := range ns {
				indent(w, flat, ind + 1)
				if i > 0 || !flat {
					sepC(w, flat); w.WriteString("|"); sepC(w, flat)
				}
				w.WriteString(n.Name.S)
				lnC(w, flat)
			}
		} //WriteUnionMemberTypes
		
		WriteEnumValuesDefinition := func (es EnumValuesDefinition, ind int) {
			
			WriteEnumValueDefinition := func (e *EnumValueDefinition, ind int) {
				indent(w, flat, ind)
				if e.Desc != nil {
					w.WriteString("\"")
					w.WriteString(trans(e.Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString(e.EnumVal.Enum.S)
				writeDirectives(w, flat, e.Dirs)
			} //WriteEnumValueDefinition
			
			//WriteEnumValuesDefinition
			if es != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat); lnC(w, flat)
				for i, e := range es {
					if i > 0 {
						sepF(w, flat)
					}
					WriteEnumValueDefinition(e, ind + 1)
					lnC(w, flat); lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
			}
		} //WriteEnumValuesDefinition
		
		WriteInputFieldsDefinition := func (ads ArgumentsDefinition, ind int) {
			if ads != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat); lnC(w, flat)
				for i, ad := range ads {
					if i > 0 {
						sepF(w, flat)
					}
					WriteInputValueDefinition(ad, ind + 1)
					lnC(w, flat); lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
			}
		} //WriteInputFieldsDefinition
		
		WriteTypeSystemDefinition := func (d TypeSystemDefinition, ind int) {
			
			WriteSchemaDefinition := func (d *SchemaDefinition, ind int) {
				indent(w, flat, ind)
				w.WriteString("schema")
				writeDirectives(w, flat, d.Dirs)
				WriteOperationTypeDefinitions(d.OpTypeDefs, ind)
			} //WriteSchemaDefinition
			
			WriteTypeDefinition := func (d TypeDefinition, ind int) {
				
				WriteScalarTypeDefinition := func (d *ScalarTypeDefinition, ind int) {
					indent(w, flat, ind)
					w.WriteString("scalar ")
					w.WriteString(d.Name.S)
					writeDirectives(w, flat, d.Dirs)
				} //WriteScalarTypeDefinition
				
				WriteObjectTypeDefinition := func (d *ObjectTypeDefinition, ind int) {
					indent(w, flat, ind)
					w.WriteString("type ")
					w.WriteString(d.Name.S)
					WriteImplements(d.ImpInter)
					writeDirectives(w, flat, d.Dirs)
					WriteFieldsDefinition(d.FieldsDef, ind)
				} //WriteObjectTypeDefinition
				
				WriteInterfaceTypeDefinition := func (d *InterfaceTypeDefinition, ind int) {
					w.WriteString("interface ")
					w.WriteString(d.Name.S)
					writeDirectives(w, flat, d.Dirs)
					WriteFieldsDefinition(d.FieldsDef, ind)
				} //WriteInterfaceTypeDefinition
				
				WriteUnionTypeDefinition := func (d *UnionTypeDefinition, ind int) {
					w.WriteString("union ")
					w.WriteString(d.Name.S)
					writeDirectives(w, flat, d.Dirs)
					WriteUnionMemberTypes(d.UnionMTypes, ind)
				} //WriteUnionTypeDefinition
				
				WriteEnumTypeDefinition := func (d *EnumTypeDefinition, ind int) {
					indent(w, flat, ind)
					w.WriteString("enum ")
					w.WriteString(d.Name.S)
					writeDirectives(w, flat, d.Dirs)
					WriteEnumValuesDefinition(d.EnumValsDef, ind)
				} //WriteEnumTypeDefinition
				
				WriteInputObjectTypeDefinition := func (d *InputObjectTypeDefinition, ind int) {
					indent(w, flat, ind)
					w.WriteString("input ")
					w.WriteString(d.Name.S)
					writeDirectives(w, flat, d.Dirs)
					WriteInputFieldsDefinition(d.InFieldsDef, ind)
				} //WriteInputObjectTypeDefinition
				
				//WriteTypeDefinition
				indent(w, flat, ind)
				if d.TypeDefinitionC().Desc != nil {
					w.WriteString("\"")
					w.WriteString(trans(d.TypeDefinitionC().Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
				}
				switch d := d.(type) {
				case *ScalarTypeDefinition:
					WriteScalarTypeDefinition(d, ind)
				case *ObjectTypeDefinition:
					WriteObjectTypeDefinition(d, ind)
				case *InterfaceTypeDefinition:
					WriteInterfaceTypeDefinition(d, ind)
				case *UnionTypeDefinition:
					WriteUnionTypeDefinition(d, ind)
				case *EnumTypeDefinition:
					WriteEnumTypeDefinition(d, ind)
				case *InputObjectTypeDefinition:
					WriteInputObjectTypeDefinition(d, ind)
				}
			} //WriteTypeDefinition
			
			WriteDirectiveDefinition := func (d *DirectiveDefinition, ind int) {
				
				WriteDirectiveLocations := func (dls DirectiveLocations, ind int) {
					lnC(w, flat)
					for _, dl := range dls {
						indent(w, flat, ind + 1)
						w.WriteString("case "); sepC(w, flat)
						w.WriteString(locationNameOf(dl))
						lnC(w, flat)
					}
				} //WriteDirectiveLocations
				
				//WriteDirectiveDefinition
				if d.Desc != nil {
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(d.Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("directive")
				sepC(w, flat); w.WriteString("@")
				w.WriteString(d.Name.S)
				WriteArgumentsDefinition(d.ArgsDef, ind)
				WriteDirectiveLocations(d.DirLocs, ind)
			} //WriteDirectiveDefinition
			
			//WriteTypeSystemDefinition
			switch d := d.(type) {
			case *SchemaDefinition:
				WriteSchemaDefinition(d, ind)
			case TypeDefinition:
				WriteTypeDefinition(d, ind)
			case *DirectiveDefinition:
				WriteDirectiveDefinition(d, ind)
			}
		} //WriteTypeSystemDefinition
		
		WriteTypeSystemExtension := func (e TypeSystemExtension, ind int) {
			
			WriteSchemaExtension := func (e *SchemaExtension, ind int) {
				indent(w, flat, ind)
				w.WriteString("schema")
				writeDirectives(w, flat, e.Dirs)
				WriteOperationTypeDefinitions(e.OpTypeDefs, ind)
			} //WriteSchemaExtension
			
			WriteTypeExtension := func (e TypeExtension, ind int) {
				
				WriteScalarTypeExtension := func (e *ScalarTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("scalar ")
					w.WriteString(e.Name.S)
					writeDirectives(w, flat, e.Dirs)
				} //WriteScalarTypeExtension
				
				WriteObjectTypeExtension := func (e *ObjectTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("type ")
					w.WriteString(e.Name.S)
					WriteImplements(e.ImpInter)
					writeDirectives(w, flat, e.Dirs)
					WriteFieldsDefinition(e.FieldsDef, ind)
				} //WriteObjectTypeExtension
				
				WriteInterfaceTypeExtension := func (e *InterfaceTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("interface ")
					w.WriteString(e.Name.S)
					writeDirectives(w, flat, e.Dirs)
					WriteFieldsDefinition(e.FieldsDef, ind)
				} //WriteInterfaceTypeExtension
				
				WriteUnionTypeExtension := func (e *UnionTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("union ")
					w.WriteString(e.Name.S)
					writeDirectives(w, flat, e.Dirs)
					WriteUnionMemberTypes(e.UnionMTypes, ind)
				} //WriteUnionTypeExtension
				
				WriteEnumTypeExtension := func (e *EnumTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("enum ")
					w.WriteString(e.Name.S)
					writeDirectives(w, flat, e.Dirs)
					WriteEnumValuesDefinition(e.EnumValsDef, ind)
				} //WriteEnumTypeExtension
				
				WriteInputObjectTypeExtension := func (e *InputObjectTypeExtension, ind int) {
					indent(w, flat, ind)
					w.WriteString("input ")
					w.WriteString(e.Name.S)
					writeDirectives(w, flat, e.Dirs)
					WriteInputFieldsDefinition(e.InFieldsDef, ind)
				} //WriteInputObjectTypeExtension
				
				//WriteTypeExtension
				switch e := e.(type) {
				case *ScalarTypeExtension:
					WriteScalarTypeExtension(e, ind)
				case *ObjectTypeExtension:
					WriteObjectTypeExtension(e, ind)
				case *InterfaceTypeExtension:
					WriteInterfaceTypeExtension(e, ind)
				case *UnionTypeExtension:
					WriteUnionTypeExtension(e, ind)
				case *EnumTypeExtension:
					WriteEnumTypeExtension(e, ind)
				case *InputObjectTypeExtension:
					WriteInputObjectTypeExtension(e, ind)
				}
			} //WriteTypeExtension
			
			//WriteTypeSystemExtension
			w.WriteString("extend ")
			switch e := e.(type) {
			case *SchemaExtension:
				WriteSchemaExtension(e, ind)
			case TypeExtension:
				WriteTypeExtension(e, ind)
			}
		} //WriteTypeSystemExtension
		
		//WriteDefinition
		switch d := d.(type) {
		case ExecutableDefinition:
			WriteExecutableDefinition(d, ind)
		case TypeSystemDefinition:
			WriteTypeSystemDefinition(d, ind)
		case TypeSystemExtension:
			WriteTypeSystemExtension(d, ind)
		}
	} //WriteDefinition
	
	//writeCommon
	M.Assert(doc != nil, 20)
	es := S.Dir().New()
	w := es.NewWriter()
	for i, d := range doc.Defs {
		WriteDefinition(w, d, 0)
		if i < len(doc.Defs) - 1 {
			sepF(w, flat); lnC(w, flat)
		}
		lnC(w, flat)
	}
	return es.Convert()
} //writeCommon

func (doc *Document) GetString() string {
	return doc.writeCommon(false)
} //GetString

func (doc *Document) GetFlatString () string {
	return doc.writeCommon(true)
} //GetFlatString

func (doc *Document) Write (w io.Writer) {
	fmt.Fprintf(w, "%s", doc.GetString())
} //Write

func (doc *Document) WriteFlat (w io.Writer) {
	fmt.Fprintf(w, "%s", doc.GetFlatString())
} //Write
