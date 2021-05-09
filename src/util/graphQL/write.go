package graphQL
	
import (
	
	M	"util/misc"
	S	"util/extStrings"
	SC	"strconv"
		"bytes"
		"fmt"
		"io"

)

type (
	
	ObjectValue interface {
		First () *ObjectField
		Next (f *ObjectField) *ObjectField
	}

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
			buf.WriteRune(reverse_solidus)
		}
		buf.WriteRune(c)
	}
	return buf.String()
} //trans

func sepF (w S.Writer, flat bool, mS bool) {
	if flat && mS {
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

func writeValue (w S.Writer, flat bool, v Value, mS *bool) {
	
	WriteField := func (f *ObjectField, mS *bool) {
		sepF(w, flat, *mS)
		w.WriteString(f.Name.S)
		w.WriteString(":"); sepC(w, flat)
		*mS = false
		writeValue(w, flat, f.Value, mS)
	} //WriteField
	
	//writeValue
	if v != nil {
		switch v := v.(type) {
		case *Variable:
			w.WriteString("$")
			w.WriteString(v.Name.S)
			*mS = true
		case *IntValue:
			if v.Int >=0 {
				sepF(w, flat, *mS)
			}
			w.WriteString(SC.FormatInt(v.Int, 10))
			*mS = true
		case *FloatValue:
			if v.Float >=0 {
				sepF(w, flat, *mS)
			}
			w.WriteString(SC.FormatFloat(v.Float, 'g', -1, 64))
			*mS = true
		case *StringValue:
			w.WriteString("\"")
			w.WriteString(trans(v.String.S))
			w.WriteString("\"")
			*mS = false
		case *BooleanValue:
			sepF(w, flat, *mS)
			if v.Boolean {
				w.WriteString("true")
			} else {
				w.WriteString("false")
			}
			*mS = true
		case *NullValue:
			sepF(w, flat, *mS)
			w.WriteString("null")
			*mS = true
		case *EnumValue:
			sepF(w, flat, *mS)
			w.WriteString(v.Enum.S)
			*mS = true
		case *ListValue:
			w.WriteString("[")
			if l := v.First(); l != nil {
				*mS = false
				writeValue(w, flat, l.Value, mS)
				for l := v.Next(l); l != nil; l = v.Next(l) {
					sepC(w, flat)
					writeValue(w, flat, l.Value, mS)
				}
			}
			w.WriteString("]")
			*mS = false
		case ObjectValue:
			w.WriteString("{")
			if f := v.First(); f != nil {
				*mS = false
				WriteField(f, mS)
				for f := v.Next(f); f != nil; f = v.Next(f) {
					sepC(w, flat)
					WriteField(f, mS)
				}
			}
			w.WriteString("}")
			*mS = false
		}
	}
} //writeValue

func writeType (w S.Writer, flat bool, t Type, mS *bool) {
	switch t := t.(type) {
	case *NonNullType:
		writeType(w, flat, t.NullT, mS)
		w.WriteString("!")
		*mS = false
	case *ListType:
		w.WriteString("[")
		*mS = false
		writeType(w, flat, t.ItemT, mS)
		w.WriteString("]")
		*mS = false
	case *NamedType:
		sepF(w, flat, *mS)
		w.WriteString(t.Name.S)
		*mS = true
	}
} //writeType

func writeArguments (w S.Writer, flat bool, as Arguments, mS *bool) {
	
	WriteArgument := func (a *Argument, mS *bool) {
		sepF(w, flat, *mS)
		w.WriteString(a.Name.S)
		w.WriteString(":"); sepC(w, flat)
		*mS = false
		writeValue(w, flat, a.Value, mS)
	} //WriteArgument
	
	//writeArguments
	if len(as) > 0 {
		w.WriteString("(")
		*mS = false
		for i, a := range as {
			if i > 0 {
				sepC(w, flat)
			}
			WriteArgument(&a, mS)
		}
		w.WriteString(")")
		*mS = false
	}
} //writeArguments

func writeDirectives (w S.Writer, flat bool, dirs Directives, mS *bool) {
	
	WriteDirective := func (dir *Directive, mS *bool) {
		sepC(w, flat); w.WriteString("@")
		w.WriteString(dir.Name.S)
		*mS = true
		writeArguments(w, flat, dir.Args, mS)
	} //WriteDirective
	
	//writeDirectives
	for _, dir := range dirs {
		WriteDirective(&dir, mS)
	}
} //writeDirectives

func writeSelectionSet (w S.Writer, flat bool, set SelectionSet, ind int, mS *bool) {
	
	WriteSelection := func (sel Selection, ind int, mS *bool) {
		
		WriteField := func (f *Field, ind int, mS *bool) {
			indent(w, flat, ind)
			if f.Alias.S != f.Name.S {
				sepF(w, flat, *mS)
				w.WriteString(f.Alias.S)
				w.WriteString(":"); sepC(w, flat)
				*mS = false
			}
			sepF(w, flat, *mS)
			w.WriteString(f.Name.S)
			*mS = true
			writeArguments(w, flat, f.Arguments, mS)
			writeDirectives(w, flat, f.Dirs, mS)
			writeSelectionSet(w, flat, f.SelSet, ind, mS)
		} //WriteField
		
		WriteFragmentSpread := func (f *FragmentSpread, ind int, mS *bool) {
			indent(w, flat, ind)
			w.WriteString("...")
			w.WriteString(f.Name.S)
			*mS = true
			writeDirectives(w, flat, f.Dirs, mS)
		} //WriteFragmentSpread
		
		WriteInlineFragment := func (f *InlineFragment, ind int, mS *bool) {
			indent(w, flat, ind)
			w.WriteString("...")
			*mS = false
			if f.TypeCond != nil {
				w.WriteString("on ")
				w.WriteString(f.TypeCond.Name.S)
				*mS = true
			}
			writeDirectives(w, flat, f.Dirs, mS)
			writeSelectionSet(w, flat, f.SelSet, ind, mS)
		} //WriteInlineFragment
		
		//WriteSelection
		switch sel := sel.(type) {
		case *Field:
			WriteField(sel, ind, mS)
		case *FragmentSpread:
			WriteFragmentSpread(sel, ind, mS)
		case *InlineFragment:
			WriteInlineFragment(sel, ind, mS)
		}
	} //WriteSelection
	
	//writeSelectionSet
	if set != nil && len(set) > 0 {
		sepC(w, flat); w.WriteString("{"); lnC(w, flat)
		*mS = false
		for _, sel := range set {
			WriteSelection(sel, ind + 1, mS); lnC(w, flat)
		}
		indent(w, flat, ind)
		w.WriteString("}")
		*mS = false
	}
} //writeSelectionSet

func (doc *Document) writeCommon (flat bool) string {
	
	WriteDefinition := func (w S.Writer, d Definition, ind int, mS *bool) {
		
		WriteOperationType := func (n int, mS *bool) {
			sepF(w, flat, *mS)
			switch n {
			case QueryOp:
				w.WriteString("query")
			case MutationOp:
				w.WriteString("mutation")
			case SubscriptionOp:
				w.WriteString("subscription")
			}
			*mS = true
		} //WriteOperationType
		
		WriteDefaultValue := func (v Value, mS *bool) {
			if v != nil {
				sepC(w, flat); w.WriteString("="); sepC(w, flat)
				*mS = false
				writeValue(w, flat, v, mS)
			}
		} //WriteDefaultValue
		
		WriteExecutableDefinition := func (d ExecutableDefinition, ind int, mS *bool) {
			
			WriteOperationDefinition := func (d *OperationDefinition, ind int, mS *bool) {
				
				WriteVariableDefinitions := func (vds VariableDefinitions, mS *bool) {
					
					WriteVariableDefinition := func (vd *VariableDefinition, mS *bool) {
						w.WriteString("$")
						w.WriteString(vd.Var.S)
						w.WriteString(":"); sepC(w, flat)
						*mS = false
						writeType(w, flat, vd.Type, mS)
						WriteDefaultValue(vd.DefVal, mS)
					} //WriteVariableDefinition
					
					//WriteVariableDefinitions
					if len(vds) > 0 {
						w.WriteString("(")
						*mS = false
						for _, vd := range vds {
							WriteVariableDefinition(vd, mS)
						}
						w.WriteString(")")
						*mS = false
					}
				} //WriteVariableDefinitions
				
				//WriteOperationDefinition
				indent(w, flat, ind)
				if !(flat && d.OpType == QueryOp && d.Name.S == "" && d.VarDefs == nil && d.Dirs == nil) {
					WriteOperationType(d.OpType, mS)
					if d.Name.S != "" {
						sepF(w, flat, *mS); sepC(w, flat)
						w.WriteString(d.Name.S)
						*mS = true
					}
					WriteVariableDefinitions(d.VarDefs, mS)
					writeDirectives(w, flat, d.Dirs, mS)
				}
				writeSelectionSet(w, flat, d.SelSet, ind, mS)
			} //WriteOperationDefinition
			
			WriteFragmentDefinition := func (d *FragmentDefinition, ind int, mS *bool) {
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString("fragment ")
				w.WriteString(d.Name.S)
				w.WriteString(" on ")
				w.WriteString(d.TypeCond.Name.S)
				*mS = true
				writeDirectives(w, flat, d.Dirs, mS)
				writeSelectionSet(w, flat, d.SelSet, ind, mS)
			} //WriteFragmentDefinition
			
			//WriteExecutableDefinition
			switch d := d.(type) {
			case *OperationDefinition:
				WriteOperationDefinition(d, ind, mS)
			case *FragmentDefinition:
				WriteFragmentDefinition(d, ind, mS)
			}
		} //WriteExecutableDefinition
		
		WriteOperationTypeDefinitions := func (ops OperationTypeDefinitions, ind int, mS *bool) {
			
			WriteOperationTypeDefinition := func (op *OperationTypeDefinition, ind int, mS *bool) {
				indent(w, flat, ind)
				WriteOperationType(op.OpType, mS)
				w.WriteString(":"); sepC(w, flat)
				w.WriteString(op.Type.Name.S)
				*mS = true
			} //WriteOperationTypeDefinition
			
			//WriteOperationTypeDefinitions
			if ops != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat)
				*mS = false
				for _, op := range ops {
					WriteOperationTypeDefinition(&op, ind + 1, mS)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
				*mS = false
			}
		} //WriteOperationTypeDefinitions
		
		WriteImplements := func (nts NamedTypes, mS *bool) {
			if len(nts) > 0 {
				sepF(w, flat, *mS)
				sepC(w, flat); w.WriteString("implements"); sepC(w, flat)
				*mS = true
				for i, nt := range nts {
					if i > 0 {
						sepC(w, flat); w.WriteString("&"); sepC(w, flat)
						*mS = false
					}
					sepF(w, flat, *mS)
					w.WriteString(nt.Name.S)
					*mS = true
				}
			}
		} //WriteImplements
		
		WriteInputValueDefinition := func (ivd *InputValueDefinition, ind int, mS *bool) {
			if ivd.Desc != nil {
				lnC(w, flat)
				indent(w, flat, ind)
				w.WriteString("\"")
				w.WriteString(trans(ivd.Desc.String.S))
				w.WriteString("\"")
				lnC(w, flat)
				*mS = false
			}
			indent(w, flat, ind)
			sepF(w, flat, *mS)
			w.WriteString(ivd.Name.S)
			w.WriteString(":"); sepC(w, flat)
			*mS = false
			writeType(w, flat, ivd.Type, mS)
			WriteDefaultValue(ivd.DefVal, mS)
			writeDirectives(w, flat, ivd.Dirs, mS)
		} //WriteInputValueDefinition
		
		WriteArgumentsDefinition := func (ads ArgumentsDefinition, ind int, mS *bool) {
			if len(ads) > 0 {
				sepC(w, flat); w.WriteString("("); lnC(w, flat)
				*mS = false
				for _, ad := range ads {
					WriteInputValueDefinition(ad, ind + 1, mS)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString(")")
				*mS = false
			}
		} //WriteArgumentsDefinition
		
		WriteFieldsDefinition := func (fds FieldsDefinition, ind int, mS *bool) {
			
			WriteFieldDefinition := func (fd *FieldDefinition, ind int, mS *bool) {
				if fd.Desc != nil {
					lnC(w, flat)
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(fd.Desc.String.S))
					w.WriteString("\"")
					*mS = false
					lnC(w, flat)
				}
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString(fd.Name.S)
				*mS = true
				WriteArgumentsDefinition(fd.ArgsDef, ind, mS)
				w.WriteString(":"); sepC(w, flat)
				*mS = false
				writeType(w, flat, fd.Type, mS)
				writeDirectives(w, flat, fd.Dirs, mS)
			} //WriteFieldDefinition
			
			//WriteFieldsDefinition
			if len(fds) > 0 {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat)
				*mS = false
				for _, fd := range fds {
					WriteFieldDefinition(fd, ind + 1, mS)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
				*mS = false
			}
		} //WriteFieldsDefinition
		
		WriteUnionMemberTypes := func (ns NamedTypes, ind int, mS *bool) {
			if len(ns) > 0 {
				sepC(w, flat); w.WriteString("="); lnC(w, flat)
				for i, n := range ns {
					indent(w, flat, ind + 1)
					if i > 0 || !flat {
						w.WriteString("|"); sepC(w, flat)
					}
					w.WriteString(n.Name.S)
					if i < len(ns) - 1 {
						lnC(w, flat)
					}
				}
				*mS = true
			}
		} //WriteUnionMemberTypes
		
		WriteEnumValuesDefinition := func (es EnumValuesDefinition, ind int, mS *bool) {
			
			WriteEnumValueDefinition := func (e *EnumValueDefinition, ind int, mS *bool) {
				if e.Desc != nil {
					lnC(w, flat)
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(e.Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
					*mS = false
				}
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString(e.EnumVal.Enum.S)
				*mS = true
				writeDirectives(w, flat, e.Dirs, mS)
			} //WriteEnumValueDefinition
			
			//WriteEnumValuesDefinition
			if es != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat)
				*mS = false
				for _, e := range es {
					WriteEnumValueDefinition(e, ind + 1, mS)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
				*mS = false
			}
		} //WriteEnumValuesDefinition
		
		WriteInputFieldsDefinition := func (ads ArgumentsDefinition, ind int, mS *bool) {
			if ads != nil {
				sepC(w, flat); w.WriteString("{"); lnC(w, flat)
				*mS = false
				for _, ad := range ads {
					WriteInputValueDefinition(ad, ind + 1, mS)
					lnC(w, flat)
				}
				indent(w, flat, ind)
				w.WriteString("}")
				*mS = false
			}
		} //WriteInputFieldsDefinition
		
		WriteTypeSystemDefinition := func (d TypeSystemDefinition, ind int, mS *bool) {
			
			WriteSchemaDefinition := func (d *SchemaDefinition, ind int, mS *bool) {
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString("schema")
				*mS = true
				writeDirectives(w, flat, d.Dirs, mS)
				WriteOperationTypeDefinitions(d.OpTypeDefs, ind, mS)
			} //WriteSchemaDefinition
			
			WriteTypeDefinition := func (d TypeDefinition, ind int, mS *bool) {
				
				WriteScalarTypeDefinition := func (d *ScalarTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("scalar ")
					w.WriteString(d.Name.S)
					*mS = true
					writeDirectives(w, flat, d.Dirs, mS)
				} //WriteScalarTypeDefinition
				
				WriteObjectTypeDefinition := func (d *ObjectTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("type ")
					w.WriteString(d.Name.S)
					*mS = true
					WriteImplements(d.ImpInter, mS)
					writeDirectives(w, flat, d.Dirs, mS)
					WriteFieldsDefinition(d.FieldsDef, ind, mS)
				} //WriteObjectTypeDefinition
				
				WriteInterfaceTypeDefinition := func (d *InterfaceTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("interface ")
					w.WriteString(d.Name.S)
					*mS = true
					writeDirectives(w, flat, d.Dirs, mS)
					WriteFieldsDefinition(d.FieldsDef, ind, mS)
				} //WriteInterfaceTypeDefinition
				
				WriteUnionTypeDefinition := func (d *UnionTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("union ")
					w.WriteString(d.Name.S)
					*mS = true
					writeDirectives(w, flat, d.Dirs, mS)
					WriteUnionMemberTypes(d.UnionMTypes, ind, mS)
				} //WriteUnionTypeDefinition
				
				WriteEnumTypeDefinition := func (d *EnumTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("enum ")
					w.WriteString(d.Name.S)
					*mS = true
					writeDirectives(w, flat, d.Dirs, mS)
					WriteEnumValuesDefinition(d.EnumValsDef, ind, mS)
				} //WriteEnumTypeDefinition
				
				WriteInputObjectTypeDefinition := func (d *InputObjectTypeDefinition, ind int, mS *bool) {
					indent(w, flat, ind)
					sepF(w, flat, *mS)
					w.WriteString("input ")
					w.WriteString(d.Name.S)
					*mS = true
					writeDirectives(w, flat, d.Dirs, mS)
					WriteInputFieldsDefinition(d.InFieldsDef, ind, mS)
				} //WriteInputObjectTypeDefinition
				
				//WriteTypeDefinition
				if d.TypeDefinitionC().Desc != nil {
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(d.TypeDefinitionC().Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
					*mS = false
				}
				switch d := d.(type) {
				case *ScalarTypeDefinition:
					WriteScalarTypeDefinition(d, ind, mS)
				case *ObjectTypeDefinition:
					WriteObjectTypeDefinition(d, ind, mS)
				case *InterfaceTypeDefinition:
					WriteInterfaceTypeDefinition(d, ind, mS)
				case *UnionTypeDefinition:
					WriteUnionTypeDefinition(d, ind, mS)
				case *EnumTypeDefinition:
					WriteEnumTypeDefinition(d, ind, mS)
				case *InputObjectTypeDefinition:
					WriteInputObjectTypeDefinition(d, ind, mS)
				}
			} //WriteTypeDefinition
			
			WriteDirectiveDefinition := func (d *DirectiveDefinition, ind int, mS *bool) {
				
				WriteDirectiveLocations := func (dls DirectiveLocations, ind int, mS *bool) {
					lnC(w, flat)
					indent(w, flat, ind + 1)
					sepF(w, flat, *mS)
					w.WriteString("on")
					lnC(w, flat)
					for _, dl := range dls {
						indent(w, flat, ind + 2)
						w.WriteString("|"); sepC(w, flat)
						w.WriteString(locationNameOf(dl))
						lnC(w, flat)
					}
					*mS = true
				} //WriteDirectiveLocations
				
				//WriteDirectiveDefinition
				if d.Desc != nil {
					indent(w, flat, ind)
					w.WriteString("\"")
					w.WriteString(trans(d.Desc.String.S))
					w.WriteString("\"")
					lnC(w, flat)
					*mS = false
				}
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString("directive")
				sepC(w, flat); w.WriteString("@")
				w.WriteString(d.Name.S)
				*mS = true
				WriteArgumentsDefinition(d.ArgsDef, ind, mS)
				WriteDirectiveLocations(d.DirLocs, ind, mS)
			} //WriteDirectiveDefinition
			
			//WriteTypeSystemDefinition
			switch d := d.(type) {
			case *SchemaDefinition:
				WriteSchemaDefinition(d, ind, mS)
			case TypeDefinition:
				WriteTypeDefinition(d, ind, mS)
			case *DirectiveDefinition:
				WriteDirectiveDefinition(d, ind, mS)
			}
		} //WriteTypeSystemDefinition
		
		WriteTypeSystemExtension := func (e TypeSystemExtension, ind int, mS *bool) {
			
			WriteSchemaExtension := func (e *SchemaExtension, ind int, mS *bool) {
				indent(w, flat, ind)
				sepF(w, flat, *mS)
				w.WriteString("schema")
				*mS = true
				writeDirectives(w, flat, e.Dirs, mS)
				WriteOperationTypeDefinitions(e.OpTypeDefs, ind, mS)
			} //WriteSchemaExtension
			
			WriteTypeExtension := func (e TypeExtension, ind int, mS *bool) {
				
				WriteScalarTypeExtension := func (e *ScalarTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("scalar ")
					w.WriteString(e.Name.S)
					*mS = true
					writeDirectives(w, flat, e.Dirs, mS)
				} //WriteScalarTypeExtension
				
				WriteObjectTypeExtension := func (e *ObjectTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("type ")
					w.WriteString(e.Name.S)
					*mS = true
					WriteImplements(e.ImpInter, mS)
					writeDirectives(w, flat, e.Dirs, mS)
					WriteFieldsDefinition(e.FieldsDef, ind, mS)
				} //WriteObjectTypeExtension
				
				WriteInterfaceTypeExtension := func (e *InterfaceTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("interface ")
					w.WriteString(e.Name.S)
					*mS = true
					writeDirectives(w, flat, e.Dirs, mS)
					WriteFieldsDefinition(e.FieldsDef, ind, mS)
				} //WriteInterfaceTypeExtension
				
				WriteUnionTypeExtension := func (e *UnionTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("union ")
					w.WriteString(e.Name.S)
					*mS = true
					writeDirectives(w, flat, e.Dirs, mS)
					WriteUnionMemberTypes(e.UnionMTypes, ind, mS)
				} //WriteUnionTypeExtension
				
				WriteEnumTypeExtension := func (e *EnumTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("enum ")
					w.WriteString(e.Name.S)
					*mS = true
					writeDirectives(w, flat, e.Dirs, mS)
					WriteEnumValuesDefinition(e.EnumValsDef, ind, mS)
				} //WriteEnumTypeExtension
				
				WriteInputObjectTypeExtension := func (e *InputObjectTypeExtension, ind int, mS *bool) {
					sepF(w, flat, *mS)
					w.WriteString("input ")
					w.WriteString(e.Name.S)
					*mS = true
					writeDirectives(w, flat, e.Dirs, mS)
					WriteInputFieldsDefinition(e.InFieldsDef, ind, mS)
				} //WriteInputObjectTypeExtension
				
				//WriteTypeExtension
				switch e := e.(type) {
				case *ScalarTypeExtension:
					WriteScalarTypeExtension(e, ind, mS)
				case *ObjectTypeExtension:
					WriteObjectTypeExtension(e, ind, mS)
				case *InterfaceTypeExtension:
					WriteInterfaceTypeExtension(e, ind, mS)
				case *UnionTypeExtension:
					WriteUnionTypeExtension(e, ind, mS)
				case *EnumTypeExtension:
					WriteEnumTypeExtension(e, ind, mS)
				case *InputObjectTypeExtension:
					WriteInputObjectTypeExtension(e, ind, mS)
				}
			} //WriteTypeExtension
			
			//WriteTypeSystemExtension
			indent(w, flat, ind)
			w.WriteString("extend"); sepC(w, flat)
			*mS = true
			switch e := e.(type) {
			case *SchemaExtension:
				WriteSchemaExtension(e, ind, mS)
			case TypeExtension:
				WriteTypeExtension(e, ind, mS)
			}
		} //WriteTypeSystemExtension
		
		//WriteDefinition
		switch d := d.(type) {
		case ExecutableDefinition:
			WriteExecutableDefinition(d, ind, mS)
		case TypeSystemDefinition:
			WriteTypeSystemDefinition(d, ind, mS)
		case TypeSystemExtension:
			WriteTypeSystemExtension(d, ind, mS)
		}
	} //WriteDefinition
	
	//writeCommon
	M.Assert(doc != nil, 20)
	es := S.Dir().New()
	w := es.NewWriter()
	var mS bool
	for i, d := range doc.Defs {
		WriteDefinition(w, d, 0, &mS)
		lnC(w, flat)
		if i < len(doc.Defs) - 1 {
			lnC(w, flat)
		}
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
} //WriteFlat
