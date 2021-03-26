/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

// The module util/graphQL is an implementation of the GraphQL query interface
package graphQL

import (
	
	A	"util/avl"
	C	"babel/compil"
	ES	"util/extStrings"
	F	"path/filepath"
	J	"util/json"
	M	"util/misc"
	R	"util/resources"
	SC	"strconv"
	SM	"util/strMapping"
		"bufio"
		"bytes"
		"fmt"
		"io"
		"io/ioutil"
		"os"
		"strings"
		"sync"

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
	
	// graphQL resource directory, in "rsrc" directory
	graphQLDir = "util/graphQL"
	// Compiler name
	compName = "graphQL.tbl"
	// Introspection definition file name
	introName = "gqIntrospection.txt"

)

const (
	
	// Kinds of TypeDefinition
	
	// null type
	t_NULL = iota
	// scalar type, etc ...
	t_SCALAR
	t_OBJECT
	t_INTERFACE
	t_UNION
	t_ENUM
	t_INPUT_OBJECT
	t_LIST
	t_NON_NULL

)

const (
	
	// DirectiveLocation(s)
	
	// ExecutableDirectiveLocation(s)
	l_QUERY = iota
	l_MUTATION
	l_SUBSCRIPTION
	l_FIELD
	l_FRAGMENT_DEFINITION
	l_FRAGMENT_SPREAD
	l_INLINE_FRAGMENT
	
	// TypeSystemDirectiveLocation(s)
	l_SCHEMA
	l_SCALAR
	l_OBJECT
	l_FIELD_DEFINITION
	l_ARGUMENT_DEFINITION
	l_INTERFACE
	l_UNION
	l_ENUM
	l_ENUM_VALUE
	l_INPUT_OBJECT
	l_INPUT_FIELD_DEFINITION
	
	// Ranks of root operation types in "typeSystem.root"
	
	QueryOp = l_QUERY
	MutationOp = l_MUTATION
	SubscriptionOp = l_SUBSCRIPTION

)

type (
	
	AnyPtr interface {
	}
	
	// Position in compiled text
	PosT struct {
		// Position from start in bytes, line number, column number in bytes
		P, L, C int
	}
	
	// Aggregate of a string and of its position in compiled text
	StrPtr struct {
		P *PosT
		S string
	}
	
	StrArray []string
	
	// Components of "compilation"
	directory struct {
		r *bufio.Reader
	}
	
	compilationer struct {
		// Compiled text
		r []rune
		// Its size
		size,
		// Reading position in "r"
		pos int
		// Resulting "Document"
		doc *Document
		// Set of errors
		err *A.Tree	// *ErrorElem
	}
	
	Location struct {
		Pos,
		Line,
		Column int
	}
	
	Path interface {
		Next () Path
		setNext (p Path)
	}
	
	PathString struct { // Path
		next Path
		S string
	}
	
	PathNb struct { // Path
		next Path
		N int
	}
	
	pathBuilder struct {
		end Path
	}
	
	ErrorElem struct {
		Message string
		Location *Location
		Path Path
	}
	
	namedItem interface {
		name () string
	}
	
	NameMapItem struct {
		Name *StrPtr
	}
	
	ValMapItem struct {
		NameMapItem
		Value Value
	}
	
	defMapItem struct {
		NameMapItem
		def Definition
	}
	
	markMapItem struct {
		NameMapItem
		mark bool
	}
	
	fieldDefMapItem struct {
		NameMapItem
		field *FieldDefinition
	}
	
	fieldList struct {
		next *fieldList
		field *Field
	}
	
	fields struct {
		fieldList *fieldList
	}
	
	fieldMapItem struct {
		NameMapItem
		fields *fields
	}
	
	fieldRing struct {
		next, prev *fieldRing
		field *Field
	}
	
	fieldRingElem struct {
		NameMapItem
		fr *fieldRing
	}
	
	nameQueue struct {
		next *nameQueue
		name *StrPtr
	}
	
	orderedFieldsMap struct {
		t *A.Tree // *fieldRingElem
		q *nameQueue
	}
	
	setMapItem struct {
		NameMapItem
		set *A.Tree // *NameMapItem
	}
	
	varDefMapItem struct {
		NameMapItem
		vd *VariableDefinition
	}
	
	streamMapItem struct {
		NameMapItem
		stream StreamResolver
	}
	
	ScalarCoercer func (ts TypeSystem, v Value, path Path, cV *Value) bool
	
	TypeDirectory interface {
		NewTypeSystem (sc Scalarer) TypeSystem
	}
	
	typeDirectory struct {
	}
	
	errorT struct {
		errors *A.Tree // *ErrorElem
	}
	
	Scalarer interface {
		FixScalarCoercer (scalarName string, sc *ScalarCoercer)
	}
	
	TypeSystem interface {
		Scalarer
		Error (mes, aux1, aux2 string, pos *PosT, path Path)
		GetErrors () *A.Tree // *ErrorElem
		GetTypeDefinition (name string) TypeDefinition
		FixFieldResolver (object, field string, resolver FieldResolver)
		FixStreamResolver (fieldName string, resolver StreamResolver)
		FixAbstractTypeResolver (resolver AbstractTypeResolver)
		InitTypeSystem (doc *Document)
		ExecValidate (doc *Document) ExecSystem
		FixInitialValue (ov *OutputObjectValue)
	}
	
	ExecSystem interface {
		TypeSystem
		ListOperations () StrArray
		GetOperation (operationName string) *OperationDefinition
		executeSubscriptionEvent (subscription *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) Response // *ValMapItem
		Execute (doc *Document, operationName string, variableValues *A.Tree) Response // *ValMapItem
	}
	
	typeSystem struct {
		Scalarer
		errorT
		typeMap, // *defMapItem
		dirMap, // *defMapItem
		streamResolvers *A.Tree // *streamMapItem
		initialValue *OutputObjectValue
		abstractTypeResolver AbstractTypeResolver
		root [3]*ObjectTypeDefinition
		verifNotIntro,
		markUnknowDefError bool
	}
	
	execSystem struct {
		typeSystem
		fragMap, // *defMapItem
		opMap *A.Tree // *defMapItem
	}
	
	// Stream of OutputObjectValue(s)
	
	EventStreamer interface {
		StreamName () string
		RecordNotificationProc (notification SourceEventNotification)
		CloseEvent ()
	}
	
	EventStream struct {
		EventStreamer
		ResponseStreams ResponseStreams
	}
	
	ResponseStreamer interface {
		ManageResponseEvent (r Response)
	}
	
	// Stream of Response(s)
	ResponseStream struct {
		ResponseStreamer
		SourceStream *EventStream
		es ExecSystem
		subscription *OperationDefinition
		variableValues *A.Tree // *ValMapItem
	}
	
	ResponseStreams []*ResponseStream
	
	SourceEventNotification func (es *EventStream, event *OutputObjectValue)
	
	FieldResolver func (rootValue *OutputObjectValue, argumentValues *A.Tree) Value // *ValMapItem
	
	StreamResolver func (rootValue *OutputObjectValue, argumentValues *A.Tree) *EventStream // *ValMapItem
	
	AbstractTypeResolver func (ts TypeSystem, td TypeDefinition, ov *OutputObjectValue) *ObjectTypeDefinition
	
	Response interface {
		Errors () *A.Tree // *ErrorElem
		SetErrors (err *A.Tree) // *ErrorElem
	}
	
	InstantResponse struct { // Response
		errors *A.Tree
		Data *OutputObjectValue
	}
	
	SubscribeResponse struct { // Response
		errors *A.Tree
		Data *ResponseStream
		Name string
	}

)

var (
	
	wd = R.FindDir()
	direc = F.Join(wd, graphQLDir)
	comp *C.Compiler
	lang = SM.NewLanguage("en")
	
	StdDir TypeDirectory = new(typeDirectory)
	Dir = StdDir

)

type (
	
	InputObjectTypeExtension struct {
		TypeExtensionC
		InFieldsDef ArgumentsDefinition
	}
	
	EnumTypeExtension struct {
		TypeExtensionC
		EnumValsDef EnumValuesDefinition
	}
	
	UnionTypeExtension struct {
		TypeExtensionC
		UnionMTypes NamedTypes
	}
	
	InterfaceTypeExtension struct {
		TypeExtensionC
		FieldsDef FieldsDefinition
	}
	
	ObjectTypeExtension struct {
		TypeExtensionC
		ImpInter NamedTypes
		FieldsDef FieldsDefinition
	}
	
	ScalarTypeExtension struct {
		TypeExtensionC
	}
	
	TypeExtensionC struct {
		Name *StrPtr
		Dirs Directives
	}
	
	TypeExtension interface { //TypeSystemExtension
		TypeSystemExtension
		TEM () *TypeExtensionC
	}
	
	SchemaExtension struct { //TypeSystemExtension
		Dirs Directives
		OpTypeDefs OperationTypeDefinitions
	}
	
	TypeSystemExtension interface { //Definition
		isTypeSystemExtension ()
	}
	
	TypeSystemDirectiveLocation struct { // DirectiveLocation
		Loc int
	}
	
	ExecutableDirectiveLocation struct { // DirectiveLocation
		Loc int
	}
	
	DirectiveLocation interface {
		LocM () int
	}
	
	DirectiveLocations []DirectiveLocation
	
	DirectiveDefinition struct {
		listable bool
		Desc *StringValue
		Name *StrPtr
		ArgsDef ArgumentsDefinition
		DirLocs DirectiveLocations // != nil
	}
	
	InputObjectTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		InFieldsDef ArgumentsDefinition //  != nil
	}
	
	EnumValueDefinition struct {
		Desc *StringValue
		EnumVal *EnumValue
		Dirs Directives
		isDeprecated bool
		deprecationReason *StrPtr
	}
	
	EnumValuesDefinition []*EnumValueDefinition
	
	EnumTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		EnumValsDef EnumValuesDefinition // != nil
	}
	
	UnionTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		UnionMTypes NamedTypes
	}
	
	InterfaceTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		FieldsDef FieldsDefinition
		implementedBy NamedTypes
	}
	
	InputValueDefinition struct {
		Desc *StringValue
		Name *StrPtr
		Type Type
		DefVal Value
		Dirs Directives
	}
	
	ArgumentsDefinition []*InputValueDefinition
	
	FieldDefinition struct {
		Desc *StringValue
		Name *StrPtr
		ArgsDef ArgumentsDefinition
		Type Type
		Dirs Directives
		isDeprecated bool
		deprecationReason *StrPtr
		resolver FieldResolver
	}
	
	FieldsDefinition []*FieldDefinition
	
	NamedTypes []*NamedType
	
	ObjectTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		ImpInter NamedTypes
		FieldsDef FieldsDefinition // != nil
	}
	
	ScalarTypeDefinition struct { //TypeDefinition
		TypeDefinitionCommon
		coercer ScalarCoercer
	}
	
	TypeDefinitionCommon struct {
		listable bool
		Desc *StringValue
		Name *StrPtr
		Dirs Directives
	}
	
	TypeDefinition interface {
		TypeSystemDefinition
		TypeDefinitionC () *TypeDefinitionCommon
	}
	
	OperationTypeDefinition struct {
		OpType int
		Type *NamedType
	}
	
	OperationTypeDefinitions []OperationTypeDefinition
	
	SchemaDefinition struct {
		listable bool
		Dirs Directives
		OpTypeDefs OperationTypeDefinitions
	}
	
	TypeSystemDefinition interface { // Definition
		listableM () bool
		setListableM (ok bool)
	}
	
	FragmentDefinition struct { // ExecutableDefinition
		Name *StrPtr
		TypeCond *NamedType // != nil
		Dirs Directives
		SelSet SelectionSet
	}
	
	InlineFragment struct { // Selection
		TypeCond *NamedType // may be nil
		Dirs Directives
		SelSet SelectionSet
	}
	
	FragmentSpread struct { // Selection
		Name *StrPtr
		Dirs Directives
	}
	
	Field struct { // Selection
		Alias,
		Name *StrPtr
		Arguments Arguments
		Dirs Directives
		returnType Type
		parentType *NamedType
		SelSet SelectionSet
	}
	
	Selection interface {
		DirsM () Directives
	}
	
	SelectionSet []Selection
	
	Argument struct {
		Name *StrPtr
		Value Value
	}
	
	Arguments []Argument
	
	Directive struct {
		Name *StrPtr
		Args Arguments
	}
	
	Directives []Directive
	
	ObjectField struct {
		next *ObjectField
		Name *StrPtr
		Value Value
	}
	
	/*
	ObjectValue struct { // Value
		inputFields *A.Tree // *ValMapItem
		outputFields *ObjectField // Queue
		Fields FieldsArray // Array
	}
	*/
	
	InputObjectValue struct { // Value
		inputFields *ObjectField // List
	}
	
	OutputObjectValue struct { // Value
		outputFields *ObjectField // Queue
	}
	
	ValueElem struct {
		next *ValueElem
		Value Value
	}
	
	ListValue struct { // Value
		endL *ValueElem // Queue
	}
	
	AnyValue struct { // Value
		Any interface{}
	}
	
	EnumValue struct { // Value
		Enum *StrPtr
	}
	
	NullValue struct { // Value
	}
	
	BooleanValue struct { // Value
		Boolean bool
	}
	
	StringValue struct { // Value
		String *StrPtr
	}
	
	FloatValue struct { // Value
		Float float64
	}
	
	IntValue struct { // Value
		Int int64
	}
	
	Variable struct { // Value
		Name *StrPtr
	}
	
	Value interface {
		isValue ()
	}
	
	NonNullType struct { // Type
		NullT Type
	}
	
	ListType struct { // Type
		ItemT Type
	}
	
	NamedType struct { // Type
		Name *StrPtr
	}
	
	Type interface {
		isType ()
	}
	
	VariableDefinition struct {
		Var *StrPtr
		Type Type
		DefVal Value
	}
	
	VariableDefinitions []*VariableDefinition
	
	OperationDefinition struct { // ExecutableDefinition
		OpType int
		Name *StrPtr
		VarDefs VariableDefinitions
		Dirs Directives
		SelSet SelectionSet
	}
	
	ExecutableDefinition interface { // Definition
		isExecutableDefinition ()
	}
	
	Definition interface {
	}
	
	Definitions []Definition
	
	Document struct {
		Defs Definitions
		validated bool
	}

)

func (*InputObjectValue) isValue () {}

func (*OutputObjectValue) isValue () {}

func (*ListValue) isValue () {}

func (*AnyValue) isValue () {}

func (*EnumValue) isValue () {}

func (*NullValue) isValue () {}

func (*BooleanValue) isValue () {}

func (*StringValue) isValue () {}

func (*FloatValue) isValue () {}

func (*IntValue) isValue () {}

func (*Variable) isValue () {}

func (*NonNullType) isType () {}

func (*ListType) isType () {}

func (*NamedType) isType () {}

func (d *SchemaExtension) isTypeSystemExtension () {}

func (d *ScalarTypeExtension) isTypeSystemExtension () {}

func (d *ObjectTypeExtension) isTypeSystemExtension () {}

func (d *InterfaceTypeExtension) isTypeSystemExtension () {}

func (d *UnionTypeExtension) isTypeSystemExtension () {}

func (d *EnumTypeExtension) isTypeSystemExtension () {}

func (d *InputObjectTypeExtension) isTypeSystemExtension () {}

func (d *OperationDefinition) isExecutableDefinition () {}

func (d *FragmentDefinition) isExecutableDefinition () {}

func (e *ScalarTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (e *ObjectTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (e *InterfaceTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (e *UnionTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (e *EnumTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (e *InputObjectTypeExtension) TEM () *TypeExtensionC {
	return &e.TypeExtensionC
} //TEM

func (dir *DirectiveDefinition) listableM () bool {
	return dir.listable
} //listableM

func (dir *DirectiveDefinition) setListableM (ok bool) {
	dir.listable = ok
} //setListableM

func (i *InputObjectTypeDefinition) listableM () bool {
	return i.listable
} //listableM

func (i *InputObjectTypeDefinition) setListableM (ok bool) {
	i.listable = ok
} //setListableM

func (e *EnumTypeDefinition) listableM () bool {
	return e.listable
} //listableM

func (e *EnumTypeDefinition) setListableM (ok bool) {
	e.listable = ok
} //setListableM

func (u *UnionTypeDefinition) listableM () bool {
	return u.listable
} //listableM

func (u *UnionTypeDefinition) setListableM (ok bool) {
	u.listable = ok
} //setListableM

func (i *InterfaceTypeDefinition) listableM () bool {
	return i.listable
} //listableM

func (i *InterfaceTypeDefinition) setListableM (ok bool) {
	i.listable = ok
} //setListableM

func (o *ObjectTypeDefinition) listableM () bool {
	return o.listable
} //listableM

func (o *ObjectTypeDefinition) setListableM (ok bool) {
	o.listable = ok
} //setListableM

func (s *ScalarTypeDefinition) listableM () bool {
	return s.listable
} //listableM

func (s *ScalarTypeDefinition) setListableM (ok bool) {
	s.listable = ok
} //setListableM

func (s *SchemaDefinition) listableM () bool {
	return s.listable
} //listableM

func (s *SchemaDefinition) setListableM (ok bool) {
	s.listable = ok
} //setListableM

func (ts *errorT) SetErrors (errors *A.Tree) {
	ts.errors = errors
} //SetErrors

func (ts *errorT) GetErrors () *A.Tree { // *ErrorElem
	return ts.errors
} //GetErrors

func MakeEventStream (es EventStreamer) *EventStream {
	e := new(EventStream)
	e.EventStreamer = es
	return e
} //MakeEventStream

func (n *NameMapItem) name () string {
	return n.Name.S
} //name

func (r *InstantResponse) Errors () *A.Tree {
	return r.errors
} //Errors

func (r *InstantResponse) SetErrors (err *A.Tree) {
	r.errors = err
} //SetErrors

func (r *SubscribeResponse) Errors () *A.Tree {
	return r.errors
} //Errors

func (r *SubscribeResponse) SetErrors (err *A.Tree) {
	r.errors = err
} //SetErrors

func (dir TypeSystemDirectiveLocation) LocM () int {
	return dir.Loc
} //LocM

func (dir ExecutableDirectiveLocation) LocM () int {
	return dir.Loc
} //LocM

func (i *InputObjectTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &i.TypeDefinitionCommon
} //TypeDefinitionC

func (e *EnumTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &e.TypeDefinitionCommon
} //TypeDefinitionC

func (u *UnionTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &u.TypeDefinitionCommon
} //TypeDefinitionC

func (i *InterfaceTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &i.TypeDefinitionCommon
} //TypeDefinitionC

func (o *ObjectTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &o.TypeDefinitionCommon
} //TypeDefinitionC

func (s *ScalarTypeDefinition) TypeDefinitionC () *TypeDefinitionCommon {
	return &s.TypeDefinitionCommon
} //TypeDefinitionC

func (f *Field) DirsM () Directives {
	return f.Dirs
} //DirsM

func (f *FragmentSpread) DirsM () Directives {
	return f.Dirs
} //DirsM

func (i *InlineFragment) DirsM () Directives {
	return i.Dirs
} //DirsM

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

func (n1 *NameMapItem) Compare (n2 A.Comparer) A.Comp {
	nn2 := n2.(namedItem)
	if n1.name() < nn2.name() {
		return A.Lt
	}
	if n1.name() > nn2.name() {
		return A.Gt
	}
	return A.Eq
} //Compare

func (n *NameMapItem) Copy () A.Copier {
	return n
} //Copy

func (err1 *ErrorElem) Compare (err2 A.Comparer) A.Comp {
	er2 := err2.(*ErrorElem)
	if err1.Location != nil && er2.Location == nil {
		return A.Lt
	}
	if err1.Location == nil && er2.Location != nil {
		return A.Gt
	}
	if err1.Location != nil && er2.Location != nil {
		if err1.Location.Line < er2.Location.Line {
			return A.Lt
		}
		if err1.Location.Line > er2.Location.Line {
			return A.Gt
		}
		if err1.Location.Column < er2.Location.Column {
			return A.Lt
		}
		if err1.Location.Column > er2.Location.Column {
			return A.Gt
		}
	}
	if err1.Message < er2.Message {
		return A.Lt
	}
	if err1.Message > er2.Message {
		return A.Gt
	}
	return A.Eq
} //Compare

func newEmptyRing () *fieldRing {
	fr := new(fieldRing)
	fr.next = fr
	fr.prev = fr
	return fr
} //newEmptyRing

func newFieldRing (f *Field) *fieldRing {
	fr := new(fieldRing)
	ff := new(fieldRing)
	ff.field = f
	ff.next = fr; ff.prev = fr
	fr.next = ff; fr.prev = ff
	return fr
} //newFieldRing

func appendRing (to, new *fieldRing) {
	new.prev.next = to
	new.next.prev = to.prev
	to.prev.next = new.next
	to.prev = new.prev
} //appendRing

func newOrderedFieldsMap () *orderedFieldsMap {
	return &orderedFieldsMap{t: A.New(), q: nil} // *fieldRingElem
} //newOrderedFieldsMap

func (om *orderedFieldsMap) searchInsFieldRing (name *StrPtr) *fieldRing {
	el := &fieldRingElem{NameMapItem: NameMapItem{Name: name}, fr: newEmptyRing()}
	e, b, _ := om.t.SearchIns(el)
	if !b {
		q := &nameQueue{name: name}
		if om.q == nil {
			q.next = q
			om.q = q
		} else {
			q.next = om.q.next
			om.q.next = q
			om.q = q
		}
	}
	return e.Val().(*fieldRingElem).fr
} //searchInsFieldRing

func MakeAnyValue (any AnyPtr) *AnyValue {
	return &AnyValue{Any: any}
} //MakeAnyValue

func newInputObjectValue () *InputObjectValue {
	return &InputObjectValue{inputFields: nil}
} //newInputObjectValue

func (ov *InputObjectValue) insertInputField (name *StrPtr, value Value) {
	f := &ObjectField{next: ov.inputFields, Name: name, Value: value}
	ov.inputFields = f
} //insertInputField

func (ov *InputObjectValue) First () *ObjectField {
	return ov.inputFields
} //First

func (*InputObjectValue) Next (f *ObjectField) *ObjectField {
	M.Assert(f != nil, 20)
	return f.next
} //Next

func NewOutputObjectValue () *OutputObjectValue {
	return &OutputObjectValue{outputFields: nil}
} //NewOutputObjectValue

func (ov *OutputObjectValue) insertOutputField (name *StrPtr, value Value) *ObjectField {
	f := &ObjectField{Name: name, Value: value}
	if ov.outputFields == nil {
		f.next = f
		ov.outputFields = f
	} else {
		f.next = ov.outputFields.next
		ov.outputFields.next = f
		ov.outputFields = f
	}
	return f
} //insertOutputField

func (ov *OutputObjectValue) InsertOutputField (name string, value Value) *ObjectField {
	return ov.insertOutputField(makeName(name), value)
} //InsertOutputField

func (ov *OutputObjectValue) First () *ObjectField {
	f := ov.outputFields
	if f == nil {
		return nil
	}
	return f.next
} //First

func (ov *OutputObjectValue) Next (f *ObjectField) *ObjectField {
	M.Assert(f != nil, 20)
	if f == ov.outputFields {
		return nil
	}
	return f.next
} //Next

func NewListValue () *ListValue {
	return &ListValue{endL: nil}
} //NewListValue

func (lv *ListValue) Append (value Value) {
	l := &ValueElem{Value: value}
	if lv.endL == nil {
		l.next = l
		lv.endL = l
	} else {
		l.next = lv.endL.next
		lv.endL.next = l
		lv.endL = l
	}
} //Append

func (lv *ListValue) First () *ValueElem {
	l := lv.endL
	if l == nil {
		return nil
	}
	return l.next
} //First

func (lv *ListValue) Next (l *ValueElem) *ValueElem {
	M.Assert(l != nil, 20)
	if l == lv.endL {
		return nil
	}
	return l.next
} //Next

func newPathBuilder () *pathBuilder {
	return &pathBuilder{end: nil}
} //newPathBuilder

func (p *PathString) Next () Path {
	return p.next
} //Next

func (ps *PathString) setNext (p Path) {
	ps.next = p
} //setNext

func (p *PathNb) Next () Path {
	return p.next
} //Next

func (pn *PathNb) setNext (p Path) {
	pn.next = p
} //setNext

func (pb *pathBuilder) copy () *pathBuilder {
	pbc := newPathBuilder()
	p := pb.end
	if p != nil {
		for {
			p = p.Next()
			var pc Path
			switch pp := p.(type) {
			case *PathString:
				pc = &PathString{S: pp.S}
			case *PathNb:
				pc = &PathNb{N: pp.N}
			}
			if pbc.end == nil {
				pc.setNext(pc)
				pbc.end = pc
			} else {
				pc.setNext(pbc.end.Next())
				pbc.end.setNext(pc)
				pbc.end = pc
			}
			if p == pb.end {break}
		}
	}
	return pbc
} //copy

func (pb *pathBuilder) pushPathString (s *StrPtr) *pathBuilder {
	pbc := pb.copy()
	ps := &PathString{S: s.S}
	if pbc.end == nil {
		ps.next = ps
		pbc.end = ps
	} else {
		ps.next = pbc.end.Next()
		pbc.end.setNext(ps)
		pbc.end = ps
	}
	return pbc
} //pushPathString

func (pb *pathBuilder) pushPathNb (n int) *pathBuilder {
	pbc := pb.copy()
	pn := &PathNb{N: n}
	if pbc.end == nil {
		pn.next = pn
		pbc.end = pn
	} else {
		pn.next = pbc.end.Next()
		pbc.end.setNext(pn)
		pbc.end = pn
	}
	return pbc
} //pushPathNb

func (pb *pathBuilder) getPath () Path {
	p := pb.end
	if p == nil {
		return nil
	}
	pc := new(PathString)
	var pe Path
	pe = pc
	for {
		p = p.Next()
		switch pp := p.(type) {
		case *PathString:
			ps := &PathString{S: pp.S}
			pe.setNext(ps)
		case *PathNb:
			pn := &PathNb{N: pp.N}
			pe.setNext(pn)
		}
		pe = pe.Next()
		if p == pb.end {break}
	}
	pe.setNext(nil)
	pe = pc.next
	return pe
} //getPath

func (ts *errorT) Error (mes, aux1, aux2 string, pos *PosT, path Path) {
	
	const
		
		prefix = "#util/graphQL:GQL_"
	
	el := &ErrorElem{Message: lang.Map(prefix + mes, aux1, aux2)}
	if pos == nil {
		el.Location = nil
	} else {
		el.Location = &Location{Pos: pos.P, Line: pos.L, Column: pos.C}
	}
	el.Path = path
	ts.GetErrors().SearchIns(el)
} //Error

func (comp *compilationer) Error (p, li, co int, mes string) {
	if comp.err == nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Error:", li, ":", co, ":", mes)
	} else {
		_, b, _ := comp.err.SearchIns(&ErrorElem{Location:&Location{Pos: p - 1, Line: li, Column: co}, Message: mes})
		M.Assert(!b, li, co, mes, 100)
	}
} //Error

func (c *compilationer) Map (index string) string {
	return lang.Map("#babel:" + index)
} //Map

const (
	
	query = 1
	mutation = 2
	subscription = 3
	cons = 4
	nilC = 5
	void = 6
	opDef = 7
	varC = 8
	namedT = 9
	itemT = 10
	nullT = 11
	field = 12
	dir = 13
	argument = 14
	selField = 15
	frag = 16
	inlF = 17
	cond = 18
	fragDef = 19
	schemDef = 20
	opTDef = 21
	scalarT = 22
	objectT = 23
	fieldDef = 24
	inputDef = 25
	interT = 26
	unionT = 27
	enumT = 28
	enumDef = 29
	inputObjT = 30
	dirDef = 31
	schemExt = 32
	scalarExt = 33
	objectExt = 34
	interExt = 35
	unionExt = 36
	enumExt = 37
	inputObjExt = 38
	arg = 39


	interpret = 101
	integer = 102
	float = 103
	stringC = 104
	boolean = 105
	null = 106
	enum = 107
	list = 108
	object = 109
	varVal = 110

)

func stringO (o *C.Object) *StrPtr {
	M.Assert(o.ObjType() == C.StringObj, 20)
	return &StrPtr{P: &PosT{P: o.Position(), L: o.Line(), C: o.Column()}, S: o.ObjString()}
} //stringO

func makeNameO (o *C.Object) *StrPtr {
	t := o.ObjType()
	M.Assert(t == C.StringObj || t == C.TermObj, 20)
	var sp *StrPtr
	if t == C.TermObj {
		M.Assert(o.ObjFunc() == void, 21)
		sp = &StrPtr{P: nil, S: ""}
	} else {
		sp = stringO(o)
	}
	return sp
} //makeNameO

func makeNamedTypeO (o *C.Object) *NamedType {
	M.Assert(o.ObjFunc() == namedT, 20)
	return &NamedType{Name: makeNameO(o.ObjTermSon(1))}
} //makeNamedTypeO

func makeTypeO (o *C.Object) Type {
	
	MakeListType := func (o *C.Object) *ListType {
		M.Assert(o.ObjFunc() == itemT, 20)
		return &ListType{ItemT: makeTypeO(o.ObjTermSon(1))}
	} //MakeListType
	
	MakeNonNullType := func (o *C.Object) *NonNullType {
		M.Assert(o.ObjFunc() == nullT, 20)
		return &NonNullType{NullT: makeTypeO(o.ObjTermSon(1))}
	} //MakeNonNullType
	
	//makeTypeO
	var t Type
	switch o.ObjFunc() {
	case namedT:
		t = makeNamedTypeO(o)
	case itemT:
		t = MakeListType(o)
	case nullT:
		t = MakeNonNullType(o)
	}
	return t
} //makeTypeO

func makeArgumentsO (o *C.Object) Arguments {
	n := 0
	oo := o
	for oo.ObjFunc() == cons {
		n++
		oo = oo.ObjTermSon(2)
	}
	as := make(Arguments, n)
	n = 0
	for o.ObjFunc() == cons {
		oo := o.ObjTermSon(1)
		M.Assert(oo.ObjFunc() == arg, 100)
		as[n].Name = makeNameO(oo.ObjTermSon(1))
		as[n].Value = oo.ObjTermSon(2).ObjUser().(Value)
		n++
		o = o.ObjTermSon(2)
	}
	return as
} //makeArgumentsO

func makeDirectivesO (o *C.Object) Directives {
	n := 0
	oo := o
	for oo.ObjFunc() == cons {
		n++
		oo = oo.ObjTermSon(2)
	}
	M.Assert(oo.ObjFunc() == nilC, 100)
	d := make(Directives, n)
	n = 0
	for o.ObjFunc() == cons {
		oo := o.ObjTermSon(1)
		M.Assert(oo.ObjFunc() == dir, 101)
		d[n].Name = makeNameO(oo.ObjTermSon(1))
		d[n].Args = makeArgumentsO(oo.ObjTermSon(2))
		n++
		o = o.ObjTermSon(2)
	}
	return d
} //makeDirectivesO

func makeTypeCondO (o *C.Object) *NamedType {
	var tCond *NamedType
	switch o.ObjFunc() {
	case void:
		tCond = nil
	case cond:
		tCond = makeNamedTypeO(o.ObjTermSon(1))
	}
	return tCond
} //makeTypeCondO

func makeSelectionSetO (o *C.Object) SelectionSet {
	
	MakeSelection := func (o *C.Object) Selection {
		var s Selection
		switch o.ObjFunc() {
		case selField:
			s = &Field{
				Alias: makeNameO(o.ObjTermSon(1)),
				Name: makeNameO(o.ObjTermSon(2)),
				Arguments: makeArgumentsO(o.ObjTermSon(3)),
				Dirs: makeDirectivesO(o.ObjTermSon(4)),
				SelSet: makeSelectionSetO(o.ObjTermSon(5)),
			}
		case frag:
			s = &FragmentSpread{
				Name: makeNameO(o.ObjTermSon(1)),
				Dirs: makeDirectivesO(o.ObjTermSon(2)),
			}
		case inlF:
			s = &InlineFragment{
				TypeCond: makeTypeCondO(o.ObjTermSon(1)),
				Dirs: makeDirectivesO(o.ObjTermSon(2)),
				SelSet: makeSelectionSetO(o.ObjTermSon(3)),
			}
		}
		return s
	} //MakeSelection
	
	//makeSelectionSetO
	n := 0
	oo := o
	for oo.ObjFunc() == cons {
		n++
		oo = oo.ObjTermSon(2)
	}
	set := make(SelectionSet, n)
	n = 0
	for o.ObjFunc() == cons {
		set[n] = MakeSelection(o.ObjTermSon(1))
		n++
		o = o.ObjTermSon(2)
	}
	return set
} //makeSelectionSetO

func (c *compilationer) Execution (fNum, parsNb int, pars C.ObjectsList) (o *C.Object, res C.Anyptr, ok bool) {
	
	MakeString := func (s *StrPtr) *StrPtr {
		
		MakeSimpleString := func (s *StrPtr, rs []rune) *StrPtr {
			buf := new(bytes.Buffer)
			pos := 0
			m := len(rs)
			for pos < m {
				if rs[pos] == '\\' {
					pos++
					switch rs[pos] {
					case '"':
						buf.WriteRune(double_quote)
					case '\\':
						buf.WriteRune(reverse_solidus)
					case '/':
						buf.WriteRune(solidus)
					case 'b':
						buf.WriteRune(backspace)
					case 'f':
						buf.WriteRune(form_feed)
					case 'n':
						buf.WriteRune(line_feed)
					case 'r':
						buf.WriteRune(carriage_return)
					case 't':
						buf.WriteRune(horizontal_tab)
					case 'u':
						n := rune(0)
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
						buf.WriteRune(n)
					default:
						M.Halt(100);
					}
				} else {
					buf.WriteRune(rs[pos])
				}
				pos++
			}
			s.S = buf.String()
			return s
		} //MakeSimpleString
		
		MakeBlockString := func (s *StrPtr, rs []rune) *StrPtr {
			
			const (
				
				space = 0x20
				eOL1 = C.EOL1
				eOL2 = C.EOL2
			
			)
			
			type
				
				Lines struct {
					next, prev *Lines
					line string
				}
			
			var
				
				state int
			
			// DFA which detects ends of lines
			AutoEOL := func (currC rune) bool {
				lineStart := false
				loop:
				for {
					switch state {
					case 0:
						switch currC {
						case eOL2:
							state = 1
						case eOL1:
							state = 2
						default:
						}
						break loop
					case 1:
						lineStart = true
						state = 0
					case 2:
						if currC == eOL2 {
							state = 1
							break loop
						}
						lineStart = true
						state = 0
					}
				}
				return lineStart
			} //AutoEOL
		
		NewLines := func () *Lines {
			ll := new(Lines)
			ll.next = ll; ll.prev = ll
			return ll
		} //NewLines
		
		InsertNewLine := func (lines *Lines) *Lines {
			l := &Lines{next: lines, prev: lines.prev}
			lines.prev.next = l
			lines.prev = l
			return l
		} //InsertNewLine
		
		RemoveLine := func (l *Lines) {
			l.next.prev = l.prev
			l.prev.next = l.next
		} //RemoveLine;
			
			//MakeBlockString
			ss := string(rs)
			sps := ""
			i := strings.Index(ss, "\\\"\"\"")
			for i >= 0 {
				sps += ss[0:i] + "\"\"\""
				ss = ss[i + 4:]
				i = strings.Index(ss, "\\\"\"\"")
			}
			sps += ss
			rs = []rune(sps)
			
			ll := NewLines()
			state = 0
			i = 0
			l := InsertNewLine(ll)
			l.line = ""
			for i < len(rs) {
				if AutoEOL(rs[i]) {
					l = InsertNewLine(ll)
					l.line = ""
				}
				if (rs[i] != eOL1) && (rs[i] != eOL2) {
					l.line += string(rs[i])
				}
				i++
			}
			
			commonIndent := M.MaxInt32
			l = ll.next.next
			for l != ll {
				rs := []rune(l.line)
				j := 0
				for rs[j] == horizontal_tab || rs[j] == space {
					j++
				}
				if j < len(rs) && j < commonIndent {
					commonIndent = j
				}
				l = l.next
			}
			if commonIndent != M.MaxInt32 {
				l = ll.next.next
				for l != ll {
					r := []rune(l.line)
					l.line = string(r[M.Min(commonIndent, len(r)):])
					l = l.next
				}
			}
			
			for {
				l = ll.next
				if l == ll {
					break
				}
				rs := []rune(l.line)
				j := 0
				for rs[j] == horizontal_tab || rs[j] == space {
					j++
				}
				if j < len(rs) {
					break
				}
				RemoveLine(l)
			}
			for {
				l = ll.prev
				if l == ll {
					break
				}
				rs := []rune(l.line)
				j := 0
				for rs[j] == horizontal_tab || rs[j] == space {
					j++
				}
				if j < len(rs) {
					break
				}
				RemoveLine(l)
			}
			
			sps = ""
			l = ll.next
			if l != ll {
				sps += l.line
				l = l.next
			}
			for l != ll {
				sps += string(line_feed) + l.line
				l = l.next
			}
			s.S = sps
			return s
		} //MakeBlockString
		
		//MakeString
		sp := new(StrPtr)
		sp.P = new(PosT); *sp.P = *s.P
		rs := []rune(s.S)
		if (len(rs) >= 2 * 3) && (rs[1] == '"') {
			n := len(rs) - 2 * 3
			M.Assert(n >= 0, 100)
			rs = rs[3:n + 3]
			s = MakeBlockString(sp, rs)
		} else {
			n := len(rs) - 2
			M.Assert(n >= 0, 101)
			rs = rs[1:n + 1]
			s = MakeSimpleString(sp, rs)
		}
		return s
	} //MakeString
	
	MakeObjectValue := func (o *C.Object) *InputObjectValue {
		ov := newInputObjectValue()
		for o.ObjFunc() == cons {
			oo := o.ObjTermSon(1)
			M.Assert(oo.ObjFunc() == field, 100)
			ov.insertInputField(makeNameO(oo.ObjTermSon(1)), oo.ObjTermSon(2).ObjUser().(Value))
			o = o.ObjTermSon(2)
		}
		M.Assert(o.ObjFunc() == nilC,101)
		return ov
	} //MakeObjectValue

	MakeDescription := func (o *C.Object) *StringValue {
		t := o.ObjType()
		M.Assert(t == C.UserObj || t == C.TermObj, 20)
		if t == C.TermObj {
			M.Assert(o.ObjFunc() == void, 21)
			return nil
		}
		return o.ObjUser().(*StringValue)
	} //MakeDescription

	MakeValue := func (o *C.Object) Value {
		f := o.ObjFunc()
		M.Assert(o.ObjType() == C.UserObj || f == void, 20)
		var v Value
		switch f {
		case void:
			v = nil
		case varVal:
			v = o.ObjUser().(*Variable)
		case integer:
			v = o.ObjUser().(*IntValue)
		case float:
			v = o.ObjUser().(*FloatValue)
		case stringC:
			v = o.ObjUser().(*StringValue)
		case boolean:
			v = o.ObjUser().(*BooleanValue)
		case null:
			v = o.ObjUser().(*NullValue)
		case enum:
			v = o.ObjUser().(*EnumValue)
		case list:
			v = o.ObjUser().(*ListValue)
		case object:
			v = o.ObjUser().(*InputObjectValue)
		}
		return v
	} //MakeValue
			
	MakeInputValueDefinition := func (o *C.Object) *InputValueDefinition {
		M.Assert(o.ObjFunc() == inputDef, 20)
		d := &InputValueDefinition{
			Desc: MakeDescription(o.ObjTermSon(1)),
			Name: makeNameO(o.ObjTermSon(2)),
			Type: makeTypeO(o.ObjTermSon(3)),
			DefVal: MakeValue(o.ObjTermSon(4)),
			Dirs: makeDirectivesO(o.ObjTermSon(5)),
		}
		return d
	} //MakeInputValueDefinition

	MakeArgumentsDefinition := func (o *C.Object) ArgumentsDefinition {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		d := make(ArgumentsDefinition, n)
		n = 0
		for o.ObjFunc() == cons {
			d[n] = MakeInputValueDefinition(o.ObjTermSon(1))
			n++
			o = o.ObjTermSon(2)
		}
		return d
	} //MakeArgumentsDefinition
	
	MakeDirectiveLocations := func (o *C.Object) DirectiveLocations {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		M.Assert(n > 0, 101)
		ds := make(DirectiveLocations, n)
		n = 0
		for o.ObjFunc() == cons {
			loc := o.ObjTermSon(1).ObjFunc()
			M.Assert(M.In(loc, M.MakeSet(query, mutation, subscription, field, fragDef, frag, inlF, schemDef, scalarT, objectT, fieldDef, argument, interT, unionT, enumT, enumDef, inputObjT, inputDef)), 102)
			switch loc {
			case query, mutation, subscription, field, fragDef, frag, inlF:
				d := new(ExecutableDirectiveLocation)
				switch loc {
				case query:
					d.Loc = l_QUERY
				case mutation:
					d.Loc = l_MUTATION
				case subscription:
					d.Loc = l_SUBSCRIPTION
				case field:
					d.Loc = l_FIELD
				case fragDef:
					d.Loc = l_FRAGMENT_DEFINITION
				case frag:
					d.Loc = l_FRAGMENT_SPREAD
				case inlF:
					d.Loc = l_INLINE_FRAGMENT
				}
				ds[n] = d
			case schemDef, scalarT, objectT, fieldDef, argument, interT, unionT, enumT, enumDef, inputObjT, inputDef:
				d := new(TypeSystemDirectiveLocation)
				switch loc {
				case schemDef:
					d.Loc = l_SCHEMA
				case scalarT:
					d.Loc = l_SCALAR
				case objectT:
					d.Loc = l_OBJECT
				case fieldDef:
					d.Loc = l_FIELD_DEFINITION
				case argument:
					d.Loc = l_ARGUMENT_DEFINITION
				case interT:
					d.Loc = l_INTERFACE
				case unionT:
					d.Loc = l_UNION
				case enumT:
					d.Loc = l_ENUM
				case enumDef:
					d.Loc = l_ENUM_VALUE
				case inputObjT:
					d.Loc = l_INPUT_OBJECT
				case inputDef:
					d.Loc = l_INPUT_FIELD_DEFINITION
				}
				ds[n] = d
			}
			n++
			o = o.ObjTermSon(2)
		}
		return ds
	} //MakeDirectiveLocations
	
	MakeDirectiveDefinition := func (o *C.Object) *DirectiveDefinition {
		M.Assert(o.ObjFunc() == dirDef, 20)
		return &DirectiveDefinition{
			listable: true,
			Desc: MakeDescription(o.ObjTermSon(1)),
			Name: makeNameO(o.ObjTermSon(2)),
			ArgsDef: MakeArgumentsDefinition(o.ObjTermSon(3)),
			DirLocs: MakeDirectiveLocations(o.ObjTermSon(4)),
		}
	} //MakeDirectiveDefinition
	
	CompleteTypeExtension := func (o *C.Object, d TypeExtension) {
		d.TEM().Name = makeNameO(o.ObjTermSon(1))
		d.TEM().Dirs = makeDirectivesO(o.ObjTermSon(2))
	} //CompleteTypeExtension
	
	CompleteTypeDefinition := func (o *C.Object, d TypeDefinition) {
		dc := d.TypeDefinitionC()
		dc.listable = true
		dc.Desc = MakeDescription(o.ObjTermSon(1))
		dc.Name = makeNameO(o.ObjTermSon(2))
		dc.Dirs = makeDirectivesO(o.ObjTermSon(3))
	} //CompleteTypeDefinition
	
	MakeInputObjectTypeExtension := func (o *C.Object) *InputObjectTypeExtension {
		M.Assert(o.ObjFunc() == inputObjExt, 20)
		i := &InputObjectTypeExtension{
			InFieldsDef: MakeArgumentsDefinition(o.ObjTermSon(3)),
		}
		CompleteTypeExtension(o, i)
		return i
	} //MakeInputObjectTypeExtension
	
	MakeInputObjectTypeDefinition := func (o *C.Object) *InputObjectTypeDefinition {
		M.Assert(o.ObjFunc() == inputObjT, 20)
		i := &InputObjectTypeDefinition{
			InFieldsDef: MakeArgumentsDefinition(o.ObjTermSon(4)),
		}
		CompleteTypeDefinition(o, i)
		return i
	} //MakeInputObjectTypeDefinition
	
	MakeEnumValueDefinition := func (o  *C.Object) *EnumValueDefinition {
		M.Assert(o.ObjFunc() == enumDef, 20)
		return &EnumValueDefinition{
			Desc: MakeDescription(o.ObjTermSon(1)),
			EnumVal: o.ObjTermSon(2).ObjUser().(*EnumValue),
			Dirs: makeDirectivesO(o.ObjTermSon(3)),
		}
	} //MakeEnumValueDefinition
	
	MakeEnumValuesDefinition := func (o *C.Object) EnumValuesDefinition {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		d := make(EnumValuesDefinition, n)
		n = 0
		for o.ObjFunc() == cons {
			d[n] = MakeEnumValueDefinition(o.ObjTermSon(1))
			n++
			o = o.ObjTermSon(2)
		}
		return d
	} //MakeEnumValuesDefinition
	
	MakeEnumTypeExtension := func (o *C.Object) *EnumTypeExtension {
		M.Assert(o.ObjFunc() == enumExt, 20)
		e := &EnumTypeExtension{
			EnumValsDef: MakeEnumValuesDefinition(o.ObjTermSon(3)),
		}
		CompleteTypeExtension(o, e)
		return e
	} //MakeEnumTypeExtension
	
	MakeEnumTypeDefinition := func (o *C.Object) *EnumTypeDefinition {
		M.Assert(o.ObjFunc() == enumT, 20)
		e := &EnumTypeDefinition{
			EnumValsDef: MakeEnumValuesDefinition(o.ObjTermSon(4)),
		}
		CompleteTypeDefinition(o, e)
		return e
	} //MakeEnumTypeDefinition
	
	MakeNamedTypes := func (o *C.Object) NamedTypes {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		t := make(NamedTypes, n)
		n = 0
		for o.ObjFunc() == cons {
			t[n] = makeNamedTypeO(o.ObjTermSon(1))
			n++
			o = o.ObjTermSon(2)
		}
		return t
	} //MakeNamedTypes
	
	MakeUnionTypeExtension := func (o *C.Object) *UnionTypeExtension {
		M.Assert(o.ObjFunc() == unionExt, 20)
		u := &UnionTypeExtension{
			UnionMTypes: MakeNamedTypes(o.ObjTermSon(3)),
		}
		CompleteTypeExtension(o, u)
		return u
	} //MakeUnionTypeExtension
	
	MakeUnionTypeDefinition := func (o *C.Object) *UnionTypeDefinition {
		M.Assert(o.ObjFunc() == unionT, 20)
		u := &UnionTypeDefinition{
			UnionMTypes: MakeNamedTypes(o.ObjTermSon(4)),
		}
		CompleteTypeDefinition(o, u)
		return u
	} //MakeUnionTypeDefinition
	
	MakeFieldDefinition := func (o *C.Object) *FieldDefinition {
		M.Assert(o.ObjFunc() == fieldDef, 20)
		return &FieldDefinition{
			Desc: MakeDescription(o.ObjTermSon(1)),
			Name: makeNameO(o.ObjTermSon(2)),
			ArgsDef: MakeArgumentsDefinition(o.ObjTermSon(3)),
			Type: makeTypeO(o.ObjTermSon(4)),
			Dirs: makeDirectivesO(o.ObjTermSon(5)),
			resolver: nil,
		}
	} //MakeFieldDefinition
	
	MakeFieldsDefinition := func (o *C.Object) FieldsDefinition {
			n := 0
			oo := o
			for oo.ObjFunc() == cons {
				n++
				oo = oo.ObjTermSon(2)
			}
			M.Assert(oo.ObjFunc() == nilC, 20)
			d := make(FieldsDefinition, n)
			n = 0
			for o.ObjFunc() == cons {
				d[n] = MakeFieldDefinition(o.ObjTermSon(1))
				n++
				o = o.ObjTermSon(2)
			}
			return d
		} //MakeFieldsDefinition
	
	MakeInterfaceTypeExtension := func (o *C.Object) *InterfaceTypeExtension {
		M.Assert(o.ObjFunc() == interExt, 20)
		i := &InterfaceTypeExtension{
			FieldsDef: MakeFieldsDefinition(o.ObjTermSon(3)),
		}
		CompleteTypeExtension(o, i)
		return i
	} //MakeInterfaceTypeExtension
	
	MakeInterfaceTypeDefinition := func (o *C.Object) *InterfaceTypeDefinition {
		M.Assert(o.ObjFunc() == interT, 20)
		i := &InterfaceTypeDefinition{
			FieldsDef: MakeFieldsDefinition(o.ObjTermSon(4)),
		}
		CompleteTypeDefinition(o, i)
		return i
	} //MakeInterfaceTypeDefinition
	
	MakeObjectTypeExtension := func (o *C.Object) *ObjectTypeExtension {
		M.Assert(o.ObjFunc() == objectExt, 20)
		ob := &ObjectTypeExtension{
			ImpInter: MakeNamedTypes(o.ObjTermSon(3)),
			FieldsDef: MakeFieldsDefinition(o.ObjTermSon(4)),
		}
		CompleteTypeExtension(o, ob)
		return ob
	} //MakeObjectTypeExtension
	
	MakeObjectTypeDefinition := func (o *C.Object) *ObjectTypeDefinition {
		M.Assert(o.ObjFunc() == objectT, 20)
		ob := &ObjectTypeDefinition{
			ImpInter: MakeNamedTypes(o.ObjTermSon(4)),
			FieldsDef: MakeFieldsDefinition(o.ObjTermSon(5)),
		}
		CompleteTypeDefinition(o, ob)
		return ob
	} //MakeObjectTypeDefinition
	
	MakeScalarTypeExtension := func (o *C.Object) *ScalarTypeExtension {
		M.Assert(o.ObjFunc() == scalarExt, 20)
		s := &ScalarTypeExtension{}
		CompleteTypeExtension(o, s)
		return s
	} //MakeScalarTypeExtension
	
	MakeScalarTypeDefinition := func (o *C.Object) *ScalarTypeDefinition {
		M.Assert(o.ObjFunc() == scalarT, 20)
		s := &ScalarTypeDefinition{}
		CompleteTypeDefinition(o, s)
		s.coercer = nil
		return s
	} //MakeScalarTypeDefinition
	
	MakeOperationTypeDefinitions := func (o *C.Object) OperationTypeDefinitions {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		d := make(OperationTypeDefinitions, n)
		n = 0
		for o.ObjFunc() == cons {
			oo := o.ObjTermSon(1)
			M.Assert(oo.ObjFunc() == opTDef, 101)
			d[n].OpType = oo.ObjTermSon(1).ObjFunc() - 1
			d[n].Type = makeNamedTypeO(oo.ObjTermSon(2))
			n++
			o = o.ObjTermSon(2)
		}
		return d
	} //MakeOperationTypeDefinitions
	
	MakeSchemaExtension := func (o *C.Object) *SchemaExtension {
		M.Assert(o.ObjFunc() == schemExt, 20)
		return &SchemaExtension{
			Dirs: makeDirectivesO(o.ObjTermSon(1)),
			OpTypeDefs: MakeOperationTypeDefinitions(o.ObjTermSon(2)),
		}
	} //MakeSchemaExtension
	
	MakeSchemaDefinition := func (o *C.Object) *SchemaDefinition {
		M.Assert(o.ObjFunc() == schemDef, 20)
		return &SchemaDefinition{
			Dirs: makeDirectivesO(o.ObjTermSon(1)),
			OpTypeDefs: MakeOperationTypeDefinitions(o.ObjTermSon(2)),
		}
	} //MakeSchemaDefinition
	
	MakeList := func (o *C.Object) *ListValue {
		l := NewListValue()
		for o.ObjFunc() == cons {
			l.Append(o.ObjTermSon(1).ObjUser().(Value))
			o = o.ObjTermSon(2)
		}
		return l
	} //MakeList
	
	MakeFragmentDefinition := func (o *C.Object) *FragmentDefinition {
		M.Assert(o.ObjFunc() == fragDef, 20)
		return &FragmentDefinition{
			Name: makeNameO(o.ObjTermSon(1)),
			TypeCond: makeTypeCondO(o.ObjTermSon(2)),
			Dirs: makeDirectivesO(o.ObjTermSon(3)),
			SelSet: makeSelectionSetO(o.ObjTermSon(4)),
		}
	} //MakeFragmentDefinition
	
	MakeVarValue := func (o *C.Object) *Variable {
		return &Variable{Name: makeNameO(o)}
	} //MakeVarValue
	
	MakeVariableDefinition := func (o *C.Object) *VariableDefinition {
		M.Assert(o.ObjFunc() == varC, 20)
		return &VariableDefinition{
			Var: makeNameO(o.ObjTermSon(1)),
			Type: makeTypeO(o.ObjTermSon(2)),
			DefVal: MakeValue(o.ObjTermSon(3)),
		}
	} //MakeVariableDefinition
	
	MakeVariableDefinitions := func (o *C.Object) VariableDefinitions {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		v := make(VariableDefinitions, n)
		n = 0
		for o.ObjFunc() == cons {
			v[n] = MakeVariableDefinition(o.ObjTermSon(1))
			n++
			o = o.ObjTermSon(2)
		}
		return v
	} //MakeVariableDefinitions
	
	MakeOperationDefinition := func (o *C.Object) *OperationDefinition {
		return &OperationDefinition{
			OpType: o.ObjTermSon(1).ObjFunc() - 1,
			Name: makeNameO(o.ObjTermSon(2)),
			VarDefs: MakeVariableDefinitions(o.ObjTermSon(3)),
			Dirs: makeDirectivesO(o.ObjTermSon(4)),
			SelSet: makeSelectionSetO(o.ObjTermSon(5)),
		}
	} //MakeOperationDefinition
	
	MakeDefinition := func (o *C.Object) Definition {
		if o.ErrorIn() {
			return nil
		}
		var def Definition
		switch o.ObjFunc() {
		case opDef:
			def = MakeOperationDefinition(o)
		case fragDef:
			def = MakeFragmentDefinition(o)
		case schemDef:
			def = MakeSchemaDefinition(o)
		case scalarT:
			def = MakeScalarTypeDefinition(o)
		case objectT:
			def = MakeObjectTypeDefinition(o)
		case interT:
			def = MakeInterfaceTypeDefinition(o)
		case unionT:
			def = MakeUnionTypeDefinition(o)
		case enumT:
			def = MakeEnumTypeDefinition(o)
		case inputObjT:
			def = MakeInputObjectTypeDefinition(o)
		case dirDef:
			def = MakeDirectiveDefinition(o)
		case schemExt:
			def = MakeSchemaExtension(o)
		case scalarExt:
			def = MakeScalarTypeExtension(o)
		case objectExt:
			def = MakeObjectTypeExtension(o)
		case interExt:
			def = MakeInterfaceTypeExtension(o)
		case unionExt:
			def = MakeUnionTypeExtension(o)
		case enumExt:
			def = MakeEnumTypeExtension(o)
		case inputObjExt:
			def = MakeInputObjectTypeExtension(o)
		}
		return def
	} //MakeDefinition
	
	MakeDefinitions := func (o *C.Object) Definitions {
		n := 0
		oo := o
		for oo.ObjFunc() == cons {
			n++
			oo = oo.ObjTermSon(2)
		}
		M.Assert(oo.ObjFunc() == nilC, 100)
		M.Assert(n > 0, 101)
		defs := make(Definitions, n)
		n = 0
		for o.ObjFunc() == cons {
			defs[n] = MakeDefinition(o.ObjTermSon(1))
			n++
			o = o.ObjTermSon(2)
		}
		return defs
	} //MakeDefinitions
	
	MakeDocument := func (o *C.Object) *Document {
		return &Document{
			Defs: MakeDefinitions(o),
			validated: false,
		}
	} //MakeDocument
	
	//Execution
	if len(pars) > 0 {
		o = C.Parameter(pars, 1)
	}
	switch fNum {
	case interpret:
		c.doc = MakeDocument(o)
	case integer:
		i, err := SC.ParseInt(stringO(o).S, 0, 64); M.Assert(err == nil, 100)
		res = &IntValue{Int: i}
	case float:
		f, err := SC.ParseFloat(stringO(o).S, 64); M.Assert(err == nil, 101)
		res = &FloatValue{Float: f}
	case stringC:
		res = &StringValue{String: MakeString(stringO(o))}
	case boolean:
		res = &BooleanValue{Boolean: stringO(o).S == "true"}
	case null:
		res = &NullValue{}
	case enum:
		res = &EnumValue{Enum: makeNameO(o)}
	case list:
		res = MakeList(o)
	case object:
		res = MakeObjectValue(o)
	case varVal:
		res = MakeVarValue(o)
	}
	ok = true
	return
}

func PrintDocument (w io.Writer, doc *Document) {
	J.BuildJsonFrom(doc).Write(w)
} //PrintDocument

func Compile (r io.Reader) (*Document, Response) {// *ErrorElem
	var b []byte
	if rs, ok := r.(io.ReadSeeker); ok {
		n, err := rs.Seek(0, io.SeekEnd); M.Assert(err == nil, err, 100)
		_, err = rs.Seek(0, io.SeekStart); M.Assert(err == nil, err, 101)
		b = make([]byte, n)
		_, err = io.ReadFull(rs, b); M.Assert(err == nil, err, 102)
	} else {
		var err error
		b, err = ioutil.ReadAll(r); M.Assert(err == nil, err, 103)
	}
	ru := []rune(string(b))
	co := &compilationer{r: ru, size: len(ru), pos: 0, err: A.New(), doc: nil}
	c := C.NewCompilation(co)
	rp := &InstantResponse{errors: co.err}
	if c.Compile(comp, false) {
		return co.doc, rp
	}
	return nil, rp
} //Compile

func ReadString (s string)  (*Document, Response) {// *ErrorElem
	return Compile(strings.NewReader(s))
} //ReadString

type (
	
	// LinkGQL(name) returns the ReadCloser corresponding to name
	LinkGQL func (name string) io.ReadCloser

)

var
	
	lGQL LinkGQL = nil

func ReadGraphQL (name string) (*Document, Response) {// *ErrorElem
	M.Assert(lGQL != nil, 100)
	rsc := lGQL(name)
	defer rsc.Close()
	return Compile(rsc)
} //ReadGraphQL

func DecodeString (s string) string {
	return J.DecodeString(s)
} //DecodeString

// ************** Validation ************

func isIntro (name *StrPtr) bool {
	return strings.Index(name.S, "__") == 0
} //isIntro

func equalTypes (ty1, ty2 Type) bool {
	switch ty1 := ty1.(type) {
	case *NonNullType:
		switch ty2 := ty2.(type) {
		case *NonNullType:
			return equalTypes(ty1.NullT, ty2.NullT)
		default:
		}
	case *ListType:
		switch ty2 := ty2.(type) {
		case *ListType:
			return equalTypes(ty1.ItemT, ty2.ItemT)
		default:
		}
	case *NamedType:
		switch ty2 := ty2.(type) {
		case *NamedType:
			return ty1.Name.S == ty2.Name.S
		default:
		}
	}
	return false
} //equalTypes

func (ts *typeSystem) getTypeDefinition (name *StrPtr) TypeDefinition {
	if el, ok, _ := ts.typeMap.Search(&defMapItem{NameMapItem: NameMapItem{Name: name}}); ok {
		switch def := el.Val().(*defMapItem).def.(type) {
		case TypeDefinition:
			return def
		}
	}
	if ts.markUnknowDefError {
		ts.Error("UnknownType", name.S, "", name.P, nil)
	}
	return nil
} //getTypeDefinition

func (ts *typeSystem) GetTypeDefinition (name string) TypeDefinition {
	return ts.getTypeDefinition(makeName(name))
} //GetTypeDefinition

func (ts *typeSystem) typeOf (ty Type) int {
	M.Assert(ty != nil, 20)
	var t int
	switch ty := ty.(type) {
	case *NonNullType:
		t = t_NON_NULL
	case *ListType:
		t = t_LIST
	case *NamedType:
		d := ts.getTypeDefinition(ty.Name)
		if d == nil {
			t = t_NULL
		} else {
			switch d.(type) {
			case *ScalarTypeDefinition:
				t = t_SCALAR
			case *ObjectTypeDefinition:
				t = t_OBJECT
			case *InterfaceTypeDefinition:
				t = t_INTERFACE
			case *UnionTypeDefinition:
				t = t_UNION
			case *EnumTypeDefinition:
				t = t_ENUM
			case *InputObjectTypeDefinition:
				t = t_INPUT_OBJECT
			}
		}
	}
	return t
} //typeOf

func nameOfType (typ int) string {
	var s string
	switch typ {
	case t_NULL:
		s = "t_NULL"
	case t_SCALAR:
		s = "t_SCALAR"
	case t_OBJECT:
		s = "t_OBJECT"
	case t_INTERFACE:
		s = "t_INTERFACE"
	case t_UNION:
		s = "t_UNION"
	case t_ENUM:
		s = "t_ENUM"
	case t_INPUT_OBJECT:
		s= "t_INPUT_OBJECT"
	case t_LIST:
		s = "t_LIST"
	case t_NON_NULL:
		s = "t_NON_NULL"
	}
	return s
} //nameOfType

func namedTypeOf (typ Type) *NamedType {
	var t *NamedType
	loop:
	for {
		switch ty := typ.(type) {
		case *NonNullType:
			if ty == nil {
				return nil
			}
			typ = ty.NullT
		case *ListType:
			if ty == nil {
				return nil
			}
			typ = ty.ItemT
		case *NamedType:
			if ty == nil {
				return nil
			}
			t = ty
			break loop
		}
	}
	return t
} //namedTypeOf

func getObjectValueInputField (ov *InputObjectValue, name *StrPtr, value *Value) bool {
	var f *ObjectField
	for f = ov.inputFields; f != nil && f.Name.S != name.S; f = ov.Next(f) {
	}
	b := f != nil
	if b {
		*value = f.Value
	}
	return b
} //getObjectValueInputField

func GetObjectValueInputField (ov *InputObjectValue, name string, value *Value) bool {
	return getObjectValueInputField(ov, makeName(name), value)
} //GetObjectValueInputField

func equalValues (v1, v2 Value) bool {
	switch v1 := v1.(type) {
	case *Variable:
		switch v2 := v2.(type) {
		case *Variable:
			return v1.Name.S == v2.Name.S
		default:
		}
	case *IntValue:
		switch v2 := v2.(type) {
		case *IntValue:
			return v1.Int == v2.Int
		default:
		}
	case *FloatValue:
		switch v2 := v2.(type) {
		case *FloatValue:
			return v1.Float == v2.Float
		default:
		}
	case *StringValue:
		switch v2 := v2.(type) {
		case *StringValue:
			return v1.String.S == v2.String.S
		default:
		}
	case *BooleanValue:
		switch v2 := v2.(type) {
		case *BooleanValue:
			return v1.Boolean == v2.Boolean
		default:
		}
	case *NullValue:
		switch v2.(type) {
		case *NullValue:
			return true
		default:
		}
	case *EnumValue:
		switch v2 := v2.(type) {
		case *EnumValue:
			return v1.Enum.S == v2.Enum.S
		default:
		}
	case *ListValue:
		switch v2 := v2.(type) {
		case *ListValue:
			vl1 := v1.First(); vl2 := v2.First()
			for vl1 != nil && vl2 != nil {
				if !equalValues(vl1.Value, vl2.Value) {
					return false
				}
				vl1 = v1.Next(vl1); vl2 = v2.Next(vl2)
			}
			return vl1 == vl2
		default:
		}
	case *InputObjectValue:
		switch v2 := v2.(type) {
		case *InputObjectValue:
			var value Value
			for f1 := v1.First(); f1 != nil; f1 = v1.Next(f1) {
				if !(getObjectValueInputField(v2, f1.Name, &value) && equalValues(f1.Value, value)) {
					return false
				}
			}
			for f2 := v2.First(); f2 != nil; f2 = v2.Next(f2) {
				if !(getObjectValueInputField(v1, f2.Name, &value) && equalValues(f2.Value, value)) {
					return false
				}
			}
			return true
		default:
		}
	}
	return false
} //equalValues

func (ts *typeSystem) getDirectiveDefinition (name *StrPtr) *DirectiveDefinition {
	if e, b, _ := ts.dirMap.Search(&defMapItem{NameMapItem: NameMapItem{name}}); b {
		if d, b := e.Val().(*defMapItem).def.(*DirectiveDefinition); b {
			return d
		}
	}
	if ts.markUnknowDefError {
		ts.Error("UnknownDirective", name.S, "", name.P, nil)
	}
	return nil
} //getDirectiveDefinition

func (es *execSystem) getFragmentDefinition (name *StrPtr) *FragmentDefinition {
	if e, b, _ := es.fragMap.Search(&defMapItem{NameMapItem: NameMapItem{name}}); b {
		if d, b := e.Val().(*defMapItem).def.(*FragmentDefinition); b {
			return d
		}
	}
	es.Error("UnknownFragment", name.S, "", name.P, nil)
	return nil
} //getFragmentDefinition

func getField (f FieldsDefinition, name *StrPtr) *FieldDefinition {
	if f != nil {
		for _, fi := range f {
			if fi.Name.S == name.S {
				return fi
			}
		}
	}
	return nil
} //getField

func getInputField (a ArgumentsDefinition, name *StrPtr, field **InputValueDefinition) bool {
	for _, ai := range a {
		if ai.Name.S == name.S {
			*field = ai
			return true
		}
	}
	return false
} //getInputField

func insertField (t *A.Tree, field *Field) { // *fieldMapItem
	e, _, _ := t.SearchIns(&fieldMapItem{
		NameMapItem{
			field.Alias,
		},
		&fields{
			nil,
		},
	})
	f := e.Val().(*fieldMapItem)
	l := &fieldList{field: field, next: f.fields.fieldList}
	f.fields.fieldList = l
} //insertField

func (ts *typeSystem) isInputType (typ Type) bool {
	switch typ := typ.(type) {
	case *ListType:
		return ts.isInputType(typ.ItemT)
	case *NonNullType:
		return ts.isInputType(typ.NullT)
	case *NamedType:
		def := ts.getTypeDefinition(typ.Name)
		if def != nil {
			switch def.(type) {
			case *ScalarTypeDefinition:
				return true
			case *EnumTypeDefinition:
				return true
			case *InputObjectTypeDefinition:
				return true
			default:
			}
		}
	}
	return false
} //isInputType

func (ts *typeSystem) isOutputType (typ Type) bool {
	switch typ := typ.(type) {
	case *ListType:
		return ts.isOutputType(typ.ItemT)
	case *NonNullType:
		return ts.isOutputType(typ.NullT)
	case *NamedType:
		def := ts.getTypeDefinition(typ.Name)
		if def != nil {
			switch def.(type) {
			case *ScalarTypeDefinition:
				return true
			case *ObjectTypeDefinition:
				return true
			case *InterfaceTypeDefinition:
				return true
			case *UnionTypeDefinition:
				return true
			case *EnumTypeDefinition:
				return true
			default:
			}
		}
	}
	return false
} //isOutputType

func ExecutableDefinitions (doc *Document) bool {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d.(type) {
		case *OperationDefinition:
		case *FragmentDefinition:
		default:
			return false
		}
	}
	return true
} //ExecutableDefinitions

func (ts *typeSystem) validateOperationNameUniqueness (doc *Document) *A.Tree { // *defMapItem
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	t := A.New() // *defMapItem
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			if _, b, _ := t.SearchIns(&defMapItem{NameMapItem{d.Name}, d}); b {
				ts.Error("NotUniqueOperation", d.Name.S, "", d.Name.P, nil)
				return nil
			}
		default:
		}
	}
	return t
} //validateOperationNameUniqueness

func (ts *typeSystem) validateLoneAnonymousOperation (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	var anonymous *OperationDefinition = nil
	opNb := 0
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			opNb++
			if d.Name.S == "" {
				anonymous = d
			}
			if anonymous != nil && opNb > 1 {
				ts.Error("NotLoneAnonOp", "", "", nil, nil)
				return
			}
		default:
		}
	}
} //validateLoneAnonymousOperation

func (ts *typeSystem) doesFragmentTypeApply (objectType *ObjectTypeDefinition, fragmentType *NamedType) bool {

	Implements := func (ot *ObjectTypeDefinition, it *InterfaceTypeDefinition) bool {
		for _, i := range ot.ImpInter {
			if i.Name.S == it.Name.S {
				return true
			}
		}
		return false
	} //Implements

	InUnion := func (ot *ObjectTypeDefinition, ut *UnionTypeDefinition) bool {
		for _, u := range ut.UnionMTypes {
			if u.Name.S == ot.Name.S {
				return true
			}
		}
		return false
	} //InUnion

	//doesFragmentTypeApply
	ft := ts.getTypeDefinition(fragmentType.Name)
	if ft != nil {
		switch ft := ft.(type) {
		case *ObjectTypeDefinition:
			if objectType == ft {
				return true
			}
		case *InterfaceTypeDefinition:
			if Implements(objectType, ft) {
				return true
			}
		case *UnionTypeDefinition:
			if InUnion(objectType, ft) {
				return true
			}
		}
	}
	return false
} //doesFragmentTypeApply

func getVarValue (variableValues *A.Tree, name *StrPtr, value *Value) bool { // *ValMapItem
	if variableValues != nil {
		if e, b, _ := variableValues.Search(&ValMapItem{NameMapItem: NameMapItem{name}}); b {
			*value = e.Val().(*ValMapItem).Value
			return true
		}
	}
	return false
} //getVarValue

func (es *execSystem) collectFields (objectType *ObjectTypeDefinition, selectionSet SelectionSet, variableValues, visitedFragments *A.Tree) *orderedFieldsMap { // *ValMapItem, *NameMapItem
	
	If := func (a Arguments, name *StrPtr) bool {
		if len(a) != 1 || a[0].Name.S != "if" {
			if len(a) == 0 {
				es.Error("NoArgumentIn", name.S, "", name.P, nil)
			} else {
				es.Error("IncorrectArgumentIn", a[0].Name.S, name.S, name.P, nil)
			}
			return false
		}
		switch val := a[0].Value.(type) {
		case *BooleanValue:
			return val.Boolean
		case *Variable:
			var valV Value
			if variableValues != nil && getVarValue(variableValues, val.Name, &valV) {
				switch valV := valV.(type) {
				case *BooleanValue:
					return valV.Boolean
				default:
					es.Error("IncorrectVarValue", name.S, "", name.P, nil)
				}
			}
		default:
			es.Error("IncorrectValueType", name.S, "", name.P, nil)
		}
		return false
	} //If
	
	NewVisitedFragment := func (visitedFragments *A.Tree, name *StrPtr) bool { // *NameMapItem
		_, b, _ := visitedFragments.SearchIns(&NameMapItem{name})
		return !b
	} //NewVisitedFragment
	
	//collectFields
	if visitedFragments == nil {
		visitedFragments = A.New()
	}
	groupedFields := newOrderedFieldsMap()
	for _, selection := range selectionSet {
		ok := true
		ds := selection.DirsM()
		for _, d := range ds {
			if d.Name.S == "skip" {
				ok = !If(d.Args, d.Name)
			} else if d.Name.S == "include" {
				ok = If(d.Args, d.Name)
			}
			if !ok {break}
		}
		if ok {
			switch selection := selection.(type) {
			case *Field:
				responseKey := selection.Alias
				groupForResponseKey := groupedFields.searchInsFieldRing(responseKey)
				appendRing(groupForResponseKey, newFieldRing(selection))
			case *FragmentSpread:
				fragmentSpreadName := selection.Name
				if NewVisitedFragment(visitedFragments, fragmentSpreadName) {
					fragment := es.getFragmentDefinition(fragmentSpreadName)
					if fragment != nil {
						fragmentType := fragment.TypeCond
						if es.doesFragmentTypeApply(objectType, fragmentType) {
							fragmentSelectionSet := fragment.SelSet
							fragmentGroupedFieldSet := es.collectFields(objectType, fragmentSelectionSet, variableValues, visitedFragments)
							if fragmentGroupedFieldSet.q != nil {
								q := fragmentGroupedFieldSet.q
								for {
									q = q.next
									responseKey := q.name
									fragmentGroup := fragmentGroupedFieldSet.searchInsFieldRing(responseKey)
									groupForResponseKey := groupedFields.searchInsFieldRing(responseKey)
									appendRing(groupForResponseKey, fragmentGroup)
									if q == fragmentGroupedFieldSet.q {break}
								}
							}
						}
					}
				}
			case *InlineFragment:
				fragmentType := selection.TypeCond
				if fragmentType == nil || es.doesFragmentTypeApply(objectType, fragmentType) {
					fragmentSelectionSet := selection.SelSet
					fragmentGroupedFieldSet := es.collectFields(objectType, fragmentSelectionSet, variableValues, visitedFragments)
					if fragmentGroupedFieldSet.q != nil {
						q := fragmentGroupedFieldSet.q
						for {
							q = q.next
							responseKey := q.name
							fragmentGroup := fragmentGroupedFieldSet.searchInsFieldRing(responseKey)
							groupForResponseKey := groupedFields.searchInsFieldRing(responseKey)
							appendRing(groupForResponseKey, fragmentGroup)
							if q == fragmentGroupedFieldSet.q {break}
						}
					}
				}
			}
		}
	}
	return groupedFields
} //collectFields

func (es *execSystem) validateSubscriptionSingleRootField (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	subscriptionType := es.root[SubscriptionOp]
	isSubscriptionRoot := true
	var subscriptionName *StrPtr
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			if d.OpType == SubscriptionOp {
				subscription := d
				if subscriptionType == nil {
					isSubscriptionRoot = false
					subscriptionName = d.Name
				} else {
					selectionSet := subscription.SelSet
					groupedFieldSet := es.collectFields(subscriptionType, selectionSet, nil, nil)
					if groupedFieldSet.t.NumberOfElems() != 1 {
						es.Error("NotSingleSubscriptRoot", d.Name.S, "", d.Name.P, nil)
					}
				}
			}
		default:
		}
	}
	if !isSubscriptionRoot {
		es.Error("NoSubscriptionRoot", subscriptionName.S, "", subscriptionName.P, nil)
	}
} //validateSubscriptionSingleRootField

func (ts *typeSystem) proceedSelection (typ *NamedType, sel Selection) {
	switch sel := sel.(type) {
	case *Field:
		sel.parentType = typ
		d := ts.getTypeDefinition(typ.Name)
		if d != nil {
			switch d := d.(type) {
			case *ObjectTypeDefinition:
				f := getField(d.FieldsDef, sel.Name)
				if f == nil {
					ts.Error("IsNotFieldFor", sel.Name.S, typ.Name.S, sel.Name.P, nil)
				} else {
					sel.returnType = f.Type
					ts.walkProcSel(namedTypeOf(sel.returnType), sel.SelSet)
				}
			case *InterfaceTypeDefinition:
				f := getField(d.FieldsDef, sel.Name)
				if f == nil {
					ts.Error("IsNotFieldFor", sel.Name.S, typ.Name.S, sel.Name.P, nil)
				} else {
					sel.returnType = f.Type
					ts.walkProcSel(namedTypeOf(sel.returnType), sel.SelSet)
				}
			case *UnionTypeDefinition:
				if sel.Name.S != "__typename" {
					ts.Error("IsNotFieldFor", sel.Name.S, typ.Name.S, sel.Name.P, nil)
				}
				sel.returnType = nil
			default:
				ts.Error("IsNotFieldFor", sel.Name.S, typ.Name.S, sel.Name.P, nil)
				sel.returnType = nil
			}
		}
	case *InlineFragment:
		if sel.TypeCond == nil {
			ts.walkProcSel(typ, sel.SelSet)
		} else {
			ts.walkProcSel(sel.TypeCond, sel.SelSet)
		}
	default:
	}
} //proceedSelection

func (ts *typeSystem) walkProcSel (typ *NamedType, set SelectionSet) {
	for _, sel := range set {
		ts.proceedSelection(typ, sel)
	}
} //walkProcSel

func (ts *typeSystem) validateScopedTargetFields (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			for _, sel := range d.SelSet {
				switch sel := sel.(type) {
				case *Field:
					root := ts.root[d.OpType]
					if root == nil {
						ts.Error("NotDefinedInScope", sel.Name.S, "", sel.Name.P, nil)
					} else {
						ts.proceedSelection(&NamedType{root.Name}, sel)
					}
				case *InlineFragment:
					ts.proceedSelection(nil, sel)
				default:
				}
			}
		case *FragmentDefinition:
			ts.walkProcSel(d.TypeCond, d.SelSet)
		default:
		}
	}
} //validateScopedTargetFields

func (es *execSystem) buildSetMap (setMap *A.Tree, set SelectionSet) { // *fieldMapItem
	for _, sel := range set {
		switch sel := sel.(type) {
		case *Field:
			insertField(setMap, sel)
		case *FragmentSpread:
			frag := es.getFragmentDefinition(sel.Name)
			if frag != nil {
				es.buildSetMap(setMap, frag.SelSet)
			}
		case *InlineFragment:
			es.buildSetMap(setMap, sel.SelSet)
		}
	}
} //buildSetMap

func (es *execSystem) sameResponseShape (fieldA, fieldB *Field) bool {
	typeA := fieldA.returnType
	typeB := fieldB.returnType
	if typeA == nil || typeB == nil {
		return false
	}
	for {
		repeat := false
		nA, bA := typeA.(*NonNullType)
		nB, bB := typeB.(*NonNullType)
		if bA || bB {
			if !bA || !bB {
				return false
			}
			typeA = nA.NullT
			typeB = nB.NullT
			repeat = true
		}
		lA, bA := typeA.(*ListType)
		lB, bB := typeB.(*ListType)
		if bA || bB {
			if !bA || !bB {
				return false
			}
			typeA = lA.ItemT
			typeB = lB.ItemT
			repeat = true
		}
		if !repeat {break}
	}
	tyA := es.typeOf(typeA); tyB := es.typeOf(typeB)
	if tyA == t_SCALAR || tyA == t_ENUM || tyB == t_SCALAR || tyB == t_ENUM {
		return equalTypes(typeA, typeB)
	}
	if !(tyA == t_OBJECT || tyA == t_INTERFACE) || !(tyB == t_OBJECT || tyB == t_INTERFACE) {
		return false
	}
	mergedSet := A.New() // *fieldMapItem
	es.buildSetMap(mergedSet, fieldA.SelSet)
	es.buildSetMap(mergedSet, fieldB.SelSet)
	e := mergedSet.Next(nil)
	for e != nil {
		fieldsForName := e.Val().(*fieldMapItem).fields.fieldList
		sA := fieldsForName
		for sA != nil {
			subFieldA := sA.field
			sB := sA.next
			for sB != nil {
				subFieldB := sB.field
				if !es.sameResponseShape(subFieldA, subFieldB) {
					return false
				}
				sB = sB.next
			}
			sA = sA. next
		}
		e = mergedSet.Next(e)
	}
	return true
} //sameResponseShape

func (es *execSystem) fieldsInSetCanMerge (set *A.Tree) { // *fieldMapItem
	
	EqualArgs := func (argsA, argsB Arguments) bool {
		if len(argsA) != len(argsB) {
			return false
		}
		for i := 0; i < len(argsA); i++ {
			j := 0
			for j < len(argsA) && !(argsA[i].Name.S == argsB[j].Name.S && equalValues(argsA[i].Value, argsB[j].Value)) {
				j++
			}
			if j == len(argsA) {
				return false
			}
		}
		return true
	} //EqualArgs

	//fieldsInSetCanMerge
	e := set.Next(nil)
	for e != nil {
		fieldsForName := e.Val().(*fieldMapItem).fields.fieldList
		sA := fieldsForName
		for sA != nil {
			fieldA := sA.field
			sB := sA.next
			for sB != nil {
				fieldB := sB.field
				if !es.sameResponseShape(fieldA, fieldB) {
					es.Error("FieldsCantMerge", fieldA.Name.S, fieldB.Name.S, fieldA.Name.P, nil)
					return
				}
				tyA := fieldA.parentType; tyB := fieldB.parentType
				if equalTypes(tyA, tyB) || es.typeOf(tyA) != t_OBJECT || es.typeOf(tyB) != t_OBJECT {
					if (fieldA.Name.S != fieldB.Name.S) || !EqualArgs(fieldA.Arguments, fieldB.Arguments) {
						es.Error("FieldsCantMerge", fieldA.Name.S, fieldB.Name.S, fieldA.Name.P, nil)
						return
					}
					mergedSet := A.New() // *fieldMapItem
					es.buildSetMap(mergedSet, fieldA.SelSet)
					es.buildSetMap(mergedSet, fieldB.SelSet)
					es.fieldsInSetCanMerge(mergedSet)
				}
				sB = sB.next
			}
			sA = sA. next
		}
		e = set.Next(e)
	}
} //fieldsInSetCanMerge

func (es *execSystem) allFieldsWalk (set SelectionSet) {
	if len(set)  > 0 {
		fieldM := A.New() // *fieldMapItem
		es.buildSetMap(fieldM, set)
		es.fieldsInSetCanMerge(fieldM)
		for _, sel := range set {
			switch sel := sel.(type) {
			case *Field:
				es.allFieldsWalk(sel.SelSet)
			case *InlineFragment:
				es.allFieldsWalk(sel.SelSet)
			default:
			}
		}
	}
} //allFieldsWalk

func (es *execSystem) validateAllFieldsCanMerge (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			es.allFieldsWalk(d.SelSet)
		case *FragmentDefinition:
			es.allFieldsWalk(d.SelSet)
		default:
		}
	}
} //validateAllFieldsCanMerge

func (ts *typeSystem) leafFieldsWalk (set SelectionSet) {
	for _, sel := range set {
		switch sel := sel.(type) {
		case *Field:
			typ := sel.returnType
			if typ != nil {
				typ = namedTypeOf(typ)
				selectionType := ts.typeOf(typ)
				switch selectionType {
				case t_SCALAR, t_ENUM:
					if len(sel.SelSet) > 0 {
						ts.Error("SubselectionOfScalarOrEnum", sel.Name.S, "", sel.Name.P, nil)
					}
				case t_OBJECT, t_INTERFACE, t_UNION:
					if len(sel.SelSet) == 0 {
						ts.Error("NoSubselOfObjInterOrUnion", sel.Name.S, "", sel.Name.P, nil)
					}
				default:
					ts.Error("IncorrectReturnTypeOf", nameOfType(selectionType), sel.Name.S, sel.Name.P, nil)
				}
			}
			ts.leafFieldsWalk(sel.SelSet)
		case *InlineFragment:
			ts.leafFieldsWalk(sel.SelSet)
		default:
		}
	}
} //leafFieldsWalk

func (ts *typeSystem) validateLeafFieldSelections (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			ts.leafFieldsWalk(d.SelSet)
		case *FragmentDefinition:
			ts.leafFieldsWalk(d.SelSet)
		default:
		}
	}
} //validateLeafFieldSelections

func areTypesCompatible (vt, lt Type) bool {
	ltn, bn := lt.(*NonNullType)
	if bn {
		vtn, bn := vt.(*NonNullType)
		if !bn {
			return false
		}
		return areTypesCompatible(vtn.NullT, ltn.NullT)
	}
	vtn, bn := vt.(*NonNullType)
	if bn {
		return areTypesCompatible(vtn.NullT, lt)
	}
	ltl, bl := lt.(*ListType)
	if bl {
		vtl, bl := vt.(*ListType)
		if !bl {
			return false
		}
		return areTypesCompatible(vtl.ItemT, ltl.ItemT)
	}
	_, bl = vt.(*ListType)
	if bl {
		return false
	}
	return vt.(*NamedType).Name.S == lt.(*NamedType).Name.S
} //areTypesCompatible

func (ts *typeSystem) getCoercer (name *StrPtr, sc *ScalarCoercer) bool {
	d := ts.getTypeDefinition(name)
	ok := d != nil
	var sd *ScalarTypeDefinition
	if ok {
		sd, ok = d.(*ScalarTypeDefinition)
	}
	if ok {
		*sc = sd.coercer
		ok = *sc != nil
	}
	return ok
} //getCoercer

func (ts *typeSystem) isConstInputValueCoercibleToType (value, defValue Value, typ Type, vd VariableDefinitions) bool {
	if value != nil {
		if val, isVar := value.(*Variable); isVar {
			var (v Value = nil; vt Type = nil)
			for _, vdi := range vd {
				if vdi.Var.S == val.Name.S {
					v = vdi.DefVal
					vt = vdi.Type
					break
				}
			}
			if v == nil {
				v = defValue
			}
			if v == nil && vt != nil {
				return areTypesCompatible(vt, typ)
			}
			value = v
		}
	}
	ok := value == nil
	if !ok {
		_, ok = value.(*NullValue)
	}
	if ok {
		value = defValue
	}
	switch typ := typ.(type) {
	case *NonNullType:
		if value == nil {
			return false
		}
		switch value := value.(type) {
		case *NullValue:
		default:
			return ts.isConstInputValueCoercibleToType(value, nil, typ.NullT, vd)
		}
	case *ListType:
		if value == nil {
			return true
		}
		switch value := value.(type) {
		case *ListValue:
			for vl := value.First(); vl != nil; vl = value.Next(vl) {
				if !ts.isConstInputValueCoercibleToType(vl.Value, nil, typ.ItemT, vd) {
					return false
				}
			}
			return true
		default:
			return ts.isConstInputValueCoercibleToType(value, nil, typ.ItemT, vd)
		}
	case *NamedType:
		d := ts.getTypeDefinition(typ.Name)
		if d == nil {
			return false
		}
		ok := value == nil
		if !ok {
			_, ok = value.(*NullValue)
		}
		if ok {
			return true
		}
		switch d := d.(type) {
		case *ScalarTypeDefinition:
			switch d.Name.S {
			case "Int":
				switch value := value.(type) {
				case *IntValue:
					return (M.MinInt32 <= value.Int) && (value.Int <= M.MaxInt32)
				default:
				}
			case "Float":
				switch value.(type) {
				case *IntValue, *FloatValue:
					return true
				default:
				}
			case "String", "ID":
				_, ok := value.(*StringValue)
				return ok
			case "Boolean":
				_, ok := value.(*BooleanValue)
				return ok
			default:
				var (sc ScalarCoercer; v Value)
				return ts.getCoercer(d.Name, &sc) && sc(ts, value, nil, &v)
			}
		case *EnumTypeDefinition:
			switch value := value.(type) {
			case *EnumValue:
				s := value.Enum.S
				en := d.EnumValsDef
				for _, e := range en {
					if e.EnumVal.Enum.S == s {
						return true
					}
				}
			default:
			}
		case *InputObjectTypeDefinition:
			switch value := value.(type) {
			case *InputObjectValue:
				as := d.InFieldsDef
				for _, a := range as {
					var v Value
					if !getObjectValueInputField(value, a.Name, &v) {
						v = a.DefVal
						if v == nil {
							if _, ok := a.Type.(*NonNullType); ok {
								return false
							}
						}
					}
				}
				var (t Type; defVal Value)
				ok := false
				for f := value.First(); f != nil; f = value.Next(f) {
					name := f.Name
					for _, a := range as {
						if a.Name.S == name.S {
							ok = true
							t = a.Type
							defVal = a.DefVal
							break
						}
					}
					if !ok {
						return false
					}
					v := f.Value
					if _, ok := v.(*NullValue); ok {
						if _, ok := t.(*NonNullType); ok {
							return false
						}
					} else if _, ok := v.(*Variable); !ok {
						if !ts.isConstInputValueCoercibleToType(v, nil, t, vd) {
							return false
						}
					} else { // v IS *Variable
						name := v.(*Variable).Name
						if vd == nil {
							return false
						}
						ok := false
						for _, vdi := range vd {
							if vdi.Var.S == name.S {
								ok = true
								v = vdi.DefVal
								break
							}
						}
						if !ok {
							return false
						}
						if v == nil {
							v = defVal
						}
						_, b := t.(*NonNullType)
						if b {
							b = v == nil
							if !b {
								_, b = v.(*NullValue)
							}
						}
						if b {
							return false
						}
					}
				}
				return true
			default:
			}
		default:
		}
	}
	return false
} //isConstInputValueCoercibleToType

func (ts *typeSystem) validateArgumentValuesCoercion (argumentValues Arguments, argumentDefinitions ArgumentsDefinition, defParentName *StrPtr, vd VariableDefinitions) {
	for _, argumentValue := range argumentValues {
		value := argumentValue.Value
		name := argumentValue.Name
		var (typ Type; defVal Value)
		ok := false
		for _, a := range argumentDefinitions {
			if a.Name.S == name.S {
				ok = true
				typ = a.Type
				defVal = a.DefVal
				break
			}
		}
		if !ok {
			ts.Error("IsNotInputFieldFor", name.S, defParentName.S, name.P, nil)
		} else {
			if !ts.isConstInputValueCoercibleToType(value, defVal, typ, vd) {
				ts.Error("UnableToCoerce", name.S, "", name.P, nil)
			}
		}
	}
	for _, a := range argumentDefinitions {
		typ := a.Type
		name := a.Name
		ok := false
		for _, argumentValue := range argumentValues {
			if argumentValue.Name.S == name.S {
				ok = true
				break
			}
		}
		if !ok && !ts.isConstInputValueCoercibleToType(nil, a.DefVal, typ, vd) {
			ts.Error("NoArgForNonNullType", name.S, defParentName.S, defParentName.P, nil)
		}
	}
} //validateArgumentValuesCoercion

func (ts *typeSystem) validateArgs (asf ArgumentsDefinition, as Arguments, parentName *StrPtr) {
	for i := 0; i < len(as); i++ {
		name := as[i].Name
		var f *InputValueDefinition
		if !getInputField(asf, name, &f) {
			ts.Error("IsNotInputFieldFor", name.S, parentName.S, name.P, nil)
		}
		for j := i + 1; j < len(as); j++ {
			if name.S == as[j].Name.S {
				ts.Error("DupArgName", as[j].Name.S, "", as[j].Name.P, nil)
			}
		}
	}
} //validateArgs

func (ts *typeSystem) validateDirectivesS (dirs Directives) {
	for _, dir := range dirs {
		d := ts.getDirectiveDefinition(dir.Name)
		if d != nil {
			ts.validateArgs(d.ArgsDef, dir.Args, d.Name)
		}
	}
} //validateDirectivesS

func (ts *typeSystem) argumentNamesSetWalkS (set SelectionSet) {
	for _, sel := range set {
		ts.validateDirectivesS(sel.DirsM())
		switch sel := sel.(type) {
		case *Field:
			if sel.parentType != nil {
				d := ts.getTypeDefinition(sel.parentType.Name)
				if d == nil {
					return
				}
				var fs FieldsDefinition = nil
				switch d := d.(type) {
				case *ObjectTypeDefinition:
					fs = d.FieldsDef
				case *InterfaceTypeDefinition:
					fs = d.FieldsDef
				default:
				}
				f := getField(fs, sel.Name)
				if f == nil {
					_, ok := d.(*UnionTypeDefinition)
					if !(ok && sel.Name.S == "__typename") {
						ts.Error("IsNotFieldFor", sel.Name.S, sel.parentType.Name.S, sel.Name.P, nil)
					}
				} else {
					ts.validateArgs(f.ArgsDef, sel.Arguments, sel.Name)
				}
				ts.argumentNamesSetWalkS(sel.SelSet)
			}
		case *InlineFragment:
			ts.argumentNamesSetWalkS(sel.SelSet)
		default:
		}
	}
} //argumentNamesSetWalkS

func (ts *typeSystem) simpleArgValidation (doc *Document) {
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			ts.validateDirectivesS(d.Dirs)
			ts.argumentNamesSetWalkS(d.SelSet)
		case *FragmentDefinition:
			ts.validateDirectivesS(d.Dirs)
			ts.argumentNamesSetWalkS(d.SelSet)
		default:
		}
	}
} //simpleArgValidation

func (ts *typeSystem) validateDirectivesR (dirs Directives, vd VariableDefinitions) {
	for _, dir := range dirs {
		d := ts.getDirectiveDefinition(dir.Name)
		if d != nil {
			ts.validateArgumentValuesCoercion(dir.Args, d.ArgsDef, dir.Name, vd)
		}
	}
} //validateDirectivesR

func (es *execSystem) argumentNamesSetWalkR (set SelectionSet, vd VariableDefinitions) {
	for _, sel := range set {
		es.validateDirectivesR(sel.DirsM(), vd)
		switch sel := sel.(type) {
		case *Field:
			if sel.parentType != nil {
				d := es.getTypeDefinition(sel.parentType.Name)
				if d == nil {
					return
				}
				var fs FieldsDefinition = nil
				switch d := d.(type) {
				case *ObjectTypeDefinition:
					fs = d.FieldsDef
				case *InterfaceTypeDefinition:
					fs = d.FieldsDef
				default:
				}
				f := getField(fs, sel.Name)
				if f == nil {
					_, ok := d.(*UnionTypeDefinition)
					if !(ok && sel.Name.S == "__typename") {
						es.Error("IsNotFieldFor", sel.Name.S, sel.parentType.Name.S, sel.Name.P, nil)
					}
				} else {
					es.validateArgumentValuesCoercion(sel.Arguments, f.ArgsDef, sel.Name, vd)
				}
				es.argumentNamesSetWalkR(sel.SelSet, vd)
			}
		case *FragmentSpread:
			fd := es.getFragmentDefinition(sel.Name)
			if fd != nil {
				es.validateDirectivesR(fd.Dirs, vd)
				es.argumentNamesSetWalkR(fd.SelSet, vd)
			}
		case *InlineFragment:
			es.argumentNamesSetWalkR(sel.SelSet, vd)
		}
	}
} //argumentNamesSetWalkR

func (es *execSystem) validateRequiredArguments (doc *Document) {
	
	ValidateVariableDefinitions := func (vd VariableDefinitions) {
		for _, v := range vd {
			value := v.DefVal
			if (value != nil) && !es.isConstInputValueCoercibleToType(value, nil, v.Type, nil) {
				es.Error("DefValIncoercibleToTypeIn", v.Var.S, "", v.Var.P, nil)
			}
		}
	} //ValidateVariableDefinitions

	//validateRequiredArguments
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			es.validateDirectivesR(d.Dirs, d.VarDefs)
			ValidateVariableDefinitions(d.VarDefs)
			es.argumentNamesSetWalkR(d.SelSet, d.VarDefs)
		default:
		}
	}
} //validateRequiredArguments

func (es *execSystem) validateArguments (doc *Document) {
	M.Assert(doc != nil, 20)
	es.simpleArgValidation(doc)
	es.validateRequiredArguments(doc)
} //validateArguments

func addName (t *A.Tree, name *StrPtr) bool { // *NameMapItem
	_, ok, _ := t.SearchIns(&NameMapItem{name})
	return ok
} //addName

func walkDC (s Selection, t *A.Tree) { // *NameMapItem
	switch s := s.(type) {
	case *Field:
		for _, sel := range s.SelSet {
			walkDC(sel, t)
		}
	case *FragmentSpread:
		addName(t, s.Name)
	case *InlineFragment:
		for _, sel := range s.SelSet {
			walkDC(sel, t)
		}
	}
} //walkDC

func (es *execSystem) detectCycles (fragmentDefinition *FragmentDefinition, visited *A.Tree) bool { // *NameMapItem

	Descendants := func (fd *FragmentDefinition, t *A.Tree) { // *NameMapItem
		for _, sel := range fd.SelSet {
			walkDC(sel, t)
		}
	} //Descendants

	//detectCycles
	spreads := A.New() // *NameMapItem
	Descendants(fragmentDefinition, spreads)
	e := spreads.Next(nil)
	for e != nil {
		v := visited.Copy()
		el := e.Val().(*NameMapItem)
		if addName(v, el.Name) {
			pathB := newPathBuilder()
			pathB = pathB.pushPathString(fragmentDefinition.Name)
			es.Error("CycleThrough", el.Name.S, "", el.Name.P, pathB.getPath())
			return true
		}
		if es.detectCycles(es.getFragmentDefinition(el.Name), v) {
			return true
		}
		e = spreads.Next(e)
	}
	return false
} //detectCycles

func (ts *typeSystem) getPossibleTypes (ty Type) *A.Tree { // *NameMapItem
	set := A.New() // *NameMapItem
	if ty != nil {
		nty := namedTypeOf(ty)
		d := ts.getTypeDefinition(nty.Name)
		if d != nil {
			switch d := d.(type) {
			case *ObjectTypeDefinition:
				addName(set, d.Name)
			case *InterfaceTypeDefinition:
				for _, ib := range d.implementedBy {
					addName(set, ib.Name)
				}
			case *UnionTypeDefinition:
				for _, u := range d.UnionMTypes {
					addName(set, u.Name)
				}
			default:
			}
		}
	}
	return set
} //getPossibleTypes

func mark (used *A.Tree, name *StrPtr, isUsed bool) { // *markMapItem
	e, found, _ := used.SearchIns(&markMapItem{NameMapItem: NameMapItem{name}})
	el := e.Val().(*markMapItem)
	if found {
		if isUsed {
			el.mark = true
		}
	} else {
		el.mark = isUsed
	}
} //mark

func (ts *typeSystem) validateFragmentType (ty Type) {
	namedType := namedTypeOf(ty)
	if namedType != nil {
		switch ts.typeOf(namedType) {
		case t_UNION, t_INTERFACE, t_OBJECT:
		case t_NULL:
			ts.Error("UnknownType", namedType.Name.S, "", namedType.Name.P, nil)
		default:
			ts.Error("IncorrectFragmentType", namedType.Name.S, "", namedType.Name.P, nil)
		}
	}
} //validateFragmentType

func (es *execSystem) walkFragments (used *A.Tree, set SelectionSet, parentType Type) { // *markMapItem

	FragmentSpreadIsPossible := func (fragmentType *NamedType, parentType Type) {
		set1 := es.getPossibleTypes(fragmentType)
		set2 := es.getPossibleTypes(parentType)
		if set1.NumberOfElems() > set2.NumberOfElems() {
			set1, set2 = set2, set1
		}
		e := set1.Next(nil)
		for e != nil {
			if _, found, _ := set2.Search(e.Val().(*NameMapItem)); found {
				return
			}
			e = set1.Next(e)
		}
		es.Error("FragSpreadImpossible", fragmentType.Name.S, "", fragmentType.Name.P, nil)
	} //FragmentSpreadIsPossible

	//walkFragments
	for _, sel := range set {
		switch sel := sel.(type) {
		case *Field:
			es.walkFragments(used, sel.SelSet, sel.returnType)
		case *InlineFragment:
			es.validateFragmentType(sel.TypeCond)
			if sel.TypeCond != nil {
				FragmentSpreadIsPossible(sel.TypeCond, parentType)
			}
			es.walkFragments(used, sel.SelSet, sel.TypeCond)
		case *FragmentSpread:
			d := es.getFragmentDefinition(sel.Name)
			if d != nil {
				mark(used, d.Name, true)
				FragmentSpreadIsPossible(d.TypeCond, parentType)
			}
		}
	}
} //walkFragments

func (es *execSystem) validateFragments (doc *Document) bool {

	VerifyAllFragmentsUsed := func (used *A.Tree) { // *NameMapItem
		e := used.Next(nil)
		for e != nil {
			el := e.Val().(*markMapItem)
			if !el.mark {
				es.Error("FragmentNotUsed", el.Name.S, "", el.Name.P, nil)
			}
			e = used.Next(e)
		}
	} //VerifyAllFragmentsUsed
	
	//validateFragments
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	used := A.New() // *markMapItem
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			root := es.root[d.OpType]
			if root == nil {
				es.Error("RootNotDefinedFor", d.Name.S, "", d.Name.P, nil)
			} else {
				rootType := &NamedType{root.Name}
				es.walkFragments(used, d.SelSet, rootType)
			}
		case *FragmentDefinition:
			visited := A.New() // *NameMapItem
			if es.detectCycles(d, visited) {
				return false
			}
			es.validateFragmentType(d.TypeCond)
			mark(used, d.Name, false)
			es.walkFragments(used, d.SelSet, d.TypeCond)
		default:
		}
	}
	VerifyAllFragmentsUsed(used)
	return true
} //validateFragments

func (ts *typeSystem) validateValueIV (name *StrPtr, value, defVal Value, typ Type, vd VariableDefinitions) {
	
	ValidateInputObject := func (value *InputObjectValue, d *InputObjectTypeDefinition, name *StrPtr, vd VariableDefinitions) {

		ValidateObjectFields := func (ov *InputObjectValue, ds ArgumentsDefinition, parentName *StrPtr, vd VariableDefinitions, set *A.Tree) { // *NameMapItem

			ValidateObjectField := func (f *ObjectField, defVal Value, typ Type, vd VariableDefinitions) {
				ts.validateValueIV(f.Name, f.Value, defVal, typ, vd)
			} //ValidateObjectField
			
			//ValidateObjectFields
			var (f  *ObjectField; d *InputValueDefinition)
			for fi := ov.First(); fi != nil; fi = ov.Next(fi) {
				name := fi.Name
				b := false
				for _, di := range ds {
					if di.Name.S == name.S {
						b = true
						f = fi
						d = di
						break
					}
				}
				if !b {
					ts.Error("IsNotInputFieldFor", name.S, parentName.S, name.P, nil)
				} else {
					ValidateObjectField(f, d.DefVal, d.Type, vd)
				}
				if set != nil {
					if _, found, _ := set.SearchIns(&NameMapItem{name}); found {
						ts.Error("DupFieldName", name.S, "", name.P, nil)
					}
				}
			}
		} //ValidateObjectFields
		
		//ValidateInputObject
		set := A.New() // *NameMapItem
		fieldDefinitions := d.InFieldsDef
		ValidateObjectFields(value, fieldDefinitions, name, vd, set)
		for _, fieldDefinition := range fieldDefinitions {
			typ := fieldDefinition.Type
			defaultValue := fieldDefinition.DefVal
			_, ok := typ.(*NonNullType)
			if ok && defaultValue == nil {
				fieldName := fieldDefinition.Name
				var val Value
				if !getObjectValueInputField(value, fieldName, &val) {
					ts.Error("NoArgForNonNullType", fieldName.S, name.S, name.P, nil)
				} else if _, ok := val.(*NullValue); ok {
					ts.Error("NullArgForNonNullType", name.S, fieldName.S, name.P, nil)
				}
			}
		}
	} //ValidateInputObject
	
	InstantiateVariable := func (value Value, vd VariableDefinitions) Value {
		if value != nil {
			switch value := value.(type) {
			case *Variable:
				var v *VariableDefinition
				ok := false
				for _, vi := range vd {
					if vi.Var.S == value.Name.S {
						ok = true
						v = vi
						break
					}
				}
				if ok {
					return v.DefVal
				} else {
					return nil
				}
			default:
			}
		}
		return value
	} //InstantiateVariable
	
	//validateValueIV
	if !ts.isConstInputValueCoercibleToType (value, defVal, typ, vd) {
		ts.Error("UnableToCoerce", name.S, "", name.P, nil)
	}
	d := ts.getTypeDefinition(namedTypeOf(typ).Name)
	if d != nil {
		switch d := d.(type) {
		case *InputObjectTypeDefinition:
			value = InstantiateVariable(value, vd)
			if value != nil {
				switch value := value.(type) {
				case *InputObjectValue:
					ValidateInputObject(value, d, name, vd)
				default:
					ts.Error("IncorrectValueType", name.S, "", name.P, nil)
				}
			}
		default:
		}
	}
} //validateValueIV

func (ts *typeSystem) validateArgumentsIV (as Arguments, ds ArgumentsDefinition, parentName *StrPtr, vd VariableDefinitions, set *A.Tree) { // *NameMapItem

	ValidateArgument := func (a *Argument, defVal Value, typ Type, vd VariableDefinitions) {
		ts.validateValueIV(a.Name, a.Value, defVal, typ, vd)
	} //ValidateArgument
	
	//validateArgumentsIV
	var (a *Argument; d *InputValueDefinition)
	for _, ai := range as {
		name := ai.Name
		ok := false
		for _, di := range ds {
			if di.Name.S == name.S {
				ok = true
				a = &ai
				d = di
				break
			}
		}
		if !ok {
			ts.Error("IsNotInputFieldFor", name.S, parentName.S, name.P, nil)
		} else {
			ValidateArgument(a, d.DefVal, d.Type, vd)
		}
		if set != nil {
			if _, found, _ := set.SearchIns(&NameMapItem{name}); found {
				ts.Error("DupArgName", name.S, "", name.P, nil)
			}
		}
	}
} //validateArgumentsIV

func (ts *typeSystem) walkDirsIV (ds Directives, vd VariableDefinitions) {
	for _, dir := range ds {
		d := ts.getDirectiveDefinition(dir.Name)
		if d == nil {
			ts.Error("UnknownDirective", dir.Name.S, "", dir.Name.P, nil)
		} else {
			ts.validateArgumentsIV(dir.Args, d.ArgsDef, dir.Name, vd, nil)
		}
	}
} //walkDirsIV

func (es *execSystem) walkSetsIV (set SelectionSet, parentType TypeDefinition, vd VariableDefinitions) {
	for _, sel := range set {
		es.walkDirsIV(sel.DirsM(), vd)
		var fd *FieldDefinition = nil
		switch sel := sel.(type) {
		case *Field:
			switch parentType := parentType.(type) {
			case *ObjectTypeDefinition:
				fd = getField(parentType.FieldsDef, sel.Name)
			case *InterfaceTypeDefinition:
				fd = getField(parentType.FieldsDef, sel.Name)
			default:
			}
			if fd == nil {
				_, ok := parentType.(*UnionTypeDefinition)
				if !(ok && sel.Name.S == "__typename") {
					es.Error("IsNotFieldFor", sel.Name.S, parentType.TypeDefinitionC().Name.S, sel.Name.P, nil)
				}
			} else {
				es.validateArgumentsIV(sel.Arguments, fd.ArgsDef, sel.Name, vd, nil)
				d := es.getTypeDefinition(namedTypeOf(fd.Type).Name)
				if d != nil {
					es.walkSetsIV(sel.SelSet, d, vd)
				}
			}
		case *FragmentSpread:
			frd := es.getFragmentDefinition(sel.Name)
			if frd != nil {
				es.walkDirsIV(frd.Dirs, vd)
				d := es.getTypeDefinition(frd.TypeCond.Name)
				if d != nil {
					es.walkSetsIV(frd.SelSet, d, vd)
				}
			}
		case *InlineFragment:
			if sel.TypeCond == nil {
				es.walkSetsIV(sel.SelSet, parentType, vd)
			} else {
				d := es.getTypeDefinition(sel.TypeCond.Name)
				if d != nil {
					es.walkSetsIV(sel.SelSet, d, vd)
				}
			}
		}
	}
} //walkSetsIV

func (es *execSystem) validateInputValues (doc *Document) {

	ValidateVariableDefinitions := func (vds VariableDefinitions) {
		for _, vd := range vds {
			if vd.DefVal != nil {
				es.validateValueIV(vd.Var, vd.DefVal, nil, vd.Type, nil)
			}
		}
	} //ValidateVariableDefinitions

	//validateInputValues
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			rootType := es.root[d.OpType]
			if rootType == nil {
				es.Error("RootNotDefinedFor", d.Name.S, "", d.Name.P, nil)
			} else {
				es.walkDirsIV(d.Dirs, d.VarDefs)
				ValidateVariableDefinitions(d.VarDefs)
				es.walkSetsIV(d.SelSet, rootType, d.VarDefs )
			}
		default:
		}
	}
} //validateInputValues

func (ts *typeSystem) verifyLocation (ds Directives, loc int) {
	set := A.New() // *NameMapItem
	for _, dir := range ds {
		if _, found, _ := set.SearchIns(&NameMapItem{dir.Name}); found {
			ts.Error("DupDirName", dir.Name.S, "", dir.Name.P, nil)
		}
		d := ts.getDirectiveDefinition(dir.Name)
		if d == nil {
			ts.Error("UnknownDirective", dir.Name.S, "", dir.Name.P, nil)
		} else {
			ok := false
			for _, l := range d.DirLocs {
				if l.LocM() == loc {
					ok = true
					break
				}
			}
			if !ok {
				ts.Error("DirectiveInWrongLocation", dir.Name.S, "", dir.Name.P, nil)
			}
		}
	}
} //verifyLocation

func (ts *typeSystem) walkSetsVD (set SelectionSet) {
	for _, sel := range set {
		switch sel := sel.(type) {
		case *Field:
			ts.verifyLocation(sel.Dirs, l_FIELD)
			ts.walkSetsVD(sel.SelSet)
		case *FragmentSpread:
			ts.verifyLocation(sel.Dirs, l_FRAGMENT_SPREAD)
		case *InlineFragment:
			ts.verifyLocation(sel.Dirs, l_INLINE_FRAGMENT)
			ts.walkSetsVD(sel.SelSet)
		}
	}
} //walkSetsVD

func (ts *typeSystem) validateDirectives (doc *Document) {
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			ts.verifyLocation(d.Dirs, d.OpType + 1)
			ts.walkSetsVD(d.SelSet)
		case *FragmentDefinition:
			ts.verifyLocation(d.Dirs, l_FRAGMENT_DEFINITION)
			ts.walkSetsVD(d.SelSet)
		default:
		}
	}
} //validateDirectives

func isVariableUsageAllowed (variableDefinition *VariableDefinition, variableUsage *InputValueDefinition) bool {
	variableType := variableDefinition.Type
	variableDefaultValue := variableDefinition.DefVal
	locationType := variableUsage.Type
	locationDefaultValue := variableUsage.DefVal
	hasNonNullVariableDefaultValue := variableDefaultValue != nil
	if hasNonNullVariableDefaultValue {
		_, b := variableDefaultValue.(*NullValue)
		hasNonNullVariableDefaultValue = !b
	}
	hasLocationDefaultValue := locationDefaultValue != nil
	_, bl := locationType.(*NonNullType)
	_, bv := variableType.(*NonNullType)
	if bl && !bv {
		if !hasNonNullVariableDefaultValue && !hasLocationDefaultValue {
			return false
		}
		nullableLocationType := locationType.(*NonNullType).NullT
		return areTypesCompatible(variableType, nullableLocationType)
	}
	return areTypesCompatible(variableType, locationType)
} //isVariableUsageAllowed

func (ts *typeSystem) validateValueVV (ivd *InputValueDefinition, value Value, opName *StrPtr, varDefs, vars *A.Tree) { // *varDefMapItem & *NameMapItem
	
	ValidateInputFields := func (ads ArgumentsDefinition, ov *InputObjectValue) {
		for f := ov.First(); f != nil; f = ov.Next(f) {
			j := 0
			for j < len(ads) && ads[j].Name.S != f.Name.S {
				j++
			}
			if j < len(ads) {
				ts.validateValueVV(ads[j], f.Value, opName, varDefs, vars)
			} else {
				ts.Error("UnknownArgument", f.Name.S, "", f.Name.P, nil);
			}
		}
	} //ValidateInputFields
	
	if value != nil {
		switch value := value.(type) {
		case *Variable:
			vars.SearchIns(&NameMapItem{value.Name})
			if e, found, _ := varDefs.Search(&NameMapItem{value.Name}); !found {
				ts.Error("UnknownVariable", value.Name.S, opName.S, value.Name.P, nil)
			} else if !isVariableUsageAllowed(e.Val().(*varDefMapItem).vd, ivd) {
				ts.Error("VarUsageNotAllowed", value.Name.S, "", value.Name.P, nil)
			}
		case *ListValue:
			for vl := value.First(); vl != nil; vl = value.Next(vl) {
				ts.validateValueVV(ivd, vl.Value, opName, varDefs, vars)
			}
		case *InputObjectValue:
			d := ts.getTypeDefinition(namedTypeOf(ivd.Type).Name)
			ok := d != nil
			var id *InputObjectTypeDefinition
			if ok {
				id, ok = d.(*InputObjectTypeDefinition)
			}
			M.Assert(ok, 100)
			ValidateInputFields(id.InFieldsDef, value)
		default:
		}
	}
} //validateValueVV

func (ts *typeSystem) validateArgumentsVV (ads ArgumentsDefinition, as Arguments, opName *StrPtr, varDefs, vars *A.Tree) { // *varDefMapItem & *NameMapItem
	if len(as) > len(ads) {
		ts.Error("IncorrectArgsNb", opName.S, "", opName.P, nil)
	} else {
		for _, a := range as {
			j := 0
			for j < len(ads) && ads[j].Name.S != a.Name.S {
				j++
			}
			if j < len(ads) {
				ts.validateValueVV(ads[j], a.Value, opName, varDefs, vars)
			} else {
				ts.Error("UnknownArgument", a.Name.S, "", a.Name.P, nil);
			}
		}
	}
} //validateArgumentsVV

func (ts *typeSystem) validateDirectivesVV (ds Directives, opName *StrPtr, varDefs, vars *A.Tree) { // *varDefMapItem & *NameMapItem
	for _, dsi := range ds {
		name := dsi.Name
		d := ts.getDirectiveDefinition(name)
		if d != nil {
			ts.validateArgumentsVV(d.ArgsDef, dsi.Args, opName, varDefs, vars)
		}
	}
} //validateDirectivesVV

func (es *execSystem) walkSelectionSet (parentType TypeDefinition, set SelectionSet, opName *StrPtr, varDefs, vars *A.Tree) { // *varDefMapItem & *NameMapItem
	for _, sel := range set {
		es.validateDirectivesVV(sel.DirsM(), opName, varDefs, vars)
		switch sel := sel.(type) {
		case *Field:
			var f *FieldDefinition = nil
			switch parentType := parentType.(type) {
			case *ObjectTypeDefinition:
				f = getField(parentType.FieldsDef, sel.Name)
			case *InterfaceTypeDefinition:
				f = getField(parentType.FieldsDef, sel.Name)
			default:
			}
			if f == nil {
				_, b := parentType.(*UnionTypeDefinition)
				if !(b && sel.Name.S == "__typename") {
					es.Error("IsNotFieldFor", sel.Name.S, parentType.TypeDefinitionC().Name.S, sel.Name.P, nil)
				}
			} else {
				es.validateArgumentsVV(f.ArgsDef, sel.Arguments, opName, varDefs, vars)
				d := es.getTypeDefinition(namedTypeOf(f.Type).Name)
				if d != nil {
					es.walkSelectionSet(d, sel.SelSet, opName, varDefs, vars)
				}
			}
		case *FragmentSpread:
			fd := es.getFragmentDefinition(sel.Name)
			if fd != nil {
				es.validateDirectivesVV(fd.Dirs, opName, varDefs, vars)
				d := es.getTypeDefinition(fd.TypeCond.Name)
				if d != nil {
					es.walkSelectionSet(d, fd.SelSet, opName, varDefs, vars)
				}
			}
		case *InlineFragment:
			if sel.TypeCond == nil {
				es.walkSelectionSet(parentType, sel.SelSet, opName, varDefs, vars)
			} else {
				d := es.getTypeDefinition(sel.TypeCond.Name)
				if d != nil {
					es.walkSelectionSet(d, sel.SelSet, opName, varDefs, vars)
				}
			}
		}
	}
} //walkSelectionSet

func (es *execSystem) validateVariables (doc *Document) {

	VerifyVariables := func (vds VariableDefinitions, varDefs *A.Tree) { // *varDefMapItem
		for _, vd := range vds {
			if _, found, _ := varDefs.SearchIns(&varDefMapItem{NameMapItem{vd.Var}, vd}); found {
				es.Error("DupVarName", vd.Var.S, "", vd.Var.P, nil)
			}
			if !es.isInputType(vd.Type) {
				es.Error("NotInputType", vd.Var.S, "", vd.Var.P, nil)
			}
		}
	} //VerifyVariables

	//validateVariables
	M.Assert(doc != nil, 20)
	ds := doc.Defs
	for _, d := range ds {
		switch d := d.(type) {
		case *OperationDefinition:
			varDefs := A.New() // *varDefMapItem
			VerifyVariables(d.VarDefs, varDefs)
			 vars := A.New() // *NameMapItem
			es.validateDirectivesVV(d.Dirs, d.Name, varDefs, vars)
			rootType := es.root[d.OpType]
			if rootType == nil {
				es.Error("RootNotDefinedFor", d.Name.S, "", d.Name.P, nil)
			} else {
				es.walkSelectionSet(rootType, d.SelSet, d.Name, varDefs, vars)
			}
			e := varDefs.Next(nil)
			for e != nil {
				ee := varDefs.Next(e)
				if _, found, _ := vars.Search(e.Val().(*varDefMapItem)); !found {
					el := e.Val().(*varDefMapItem)
					es.Error("VariableNotUsed", el.Name.S, "", el.Name.P, nil)
				}
				e = ee
			}
		default:
		}
	}
} //validateVariables

// ******************* /Validation *****************

// ******************** Execution *****************

func makeName (s string) *StrPtr {
	return &StrPtr{S: s, P: nil}
} //makeName
	
func GetValue (argumentValues *A.Tree, name string, v *Value) bool { // *ValMapItem
	e, ok, _ := argumentValues.Search(&ValMapItem{NameMapItem: NameMapItem{makeName(name)}})
	if ok {
		*v = e.Val().(*ValMapItem).Value
	}
	return ok
} //GetValue

func MakeIntValue (n int) *IntValue {
	return &IntValue{int64(n)}
} //MakeIntValue

func MakeInt64Value (n int64) *IntValue {
	return &IntValue{n}
} //MakeInt64Value

func MakeFloat32Value (f float32) *FloatValue {
	return &FloatValue{float64(f)}
} //MakeFloat32Value

func MakeFloat64Value (f float64) *FloatValue {
	return &FloatValue{f}
} //MakeFloat64Value

func makeStringValue (s *StrPtr) *StringValue {
	return &StringValue{s}
} //makeStringValue

func MakeStringValue (s string) *StringValue {
	return &StringValue{makeName(s)}
} //MakeStringValue

func MakeEnumValue (s string) *EnumValue {
	return &EnumValue{makeName(s)}
} //MakeEnumValue

func MakeNullValue () *NullValue {
	return &NullValue{}
} //MakeNullValue

func MakeBooleanValue (b bool) *BooleanValue {
	return &BooleanValue{b}
} //MakeBooleanValue

func locationNameOf (dl DirectiveLocation) string {
	var name string
	switch dl.LocM() {
	case l_QUERY:
		name = "QUERY"
	case l_MUTATION:
		name = "MUTATION"
	case l_SUBSCRIPTION:
		name = "SUBSCRIPTION"
	case l_FIELD:
		name = "FIELD"
	case l_FRAGMENT_DEFINITION:
		name = "FRAGMENT_DEFINITION"
	case l_FRAGMENT_SPREAD:
		name = "FRAGMENT_SPREAD"
	case l_INLINE_FRAGMENT:
		name = "INLINE_FRAGMENT"
	case l_SCHEMA:
		name = "SCHEMA"
	case l_SCALAR:
		name = "SCALAR"
	case l_OBJECT:
		name = "OBJECT"
	case l_FIELD_DEFINITION:
		name = "FIELD_DEFINITION"
	case l_ARGUMENT_DEFINITION:
		name = "ARGUMENT_DEFINITION"
	case l_INTERFACE:
		name = "INTERFACE"
	case l_UNION:
		name = "UNION"
	case l_ENUM:
		name = "ENUM"
	case l_ENUM_VALUE:
		name = "ENUM_VALUE"
	case l_INPUT_OBJECT:
		name = "INPUT_OBJECT"
	case l_INPUT_FIELD_DEFINITION:
		name = "INPUT_FIELD_DEFINITION"
	}
	return name
} //locationNameOf

func (ts *typeSystem) FixFieldResolver (object, field string, resolver FieldResolver) {
	d := ts.GetTypeDefinition(object)
	b := d != nil
	var od *ObjectTypeDefinition
	if b {
		od, b = d.(*ObjectTypeDefinition)
	}
	if !b {
		ts.Error("UnknownObject", object, "", nil, nil)
		return
	}
	f := getField(od.FieldsDef, makeName(field))
	if f == nil {
		ts.Error("IsNotFieldFor", field, object, nil, nil)
		return
	}
	if f.resolver != nil {
		ts.Error("ResolverAlrdyDefined", object, field, nil, nil)
		return
	}
	f.resolver = resolver
} //FixFieldResolver

func (ts *typeSystem) FixStreamResolver (fieldName string, resolver StreamResolver) {
	if e, found, _ := ts.streamResolvers.SearchIns(&streamMapItem{NameMapItem{makeName(fieldName)}, resolver}); found {
		ts.Error("StreamResolverAlrdyDefined", e.Val().(*streamMapItem).Name.S, "", nil, nil)
	}
} //FixStreamResolver

func (ts *typeSystem) FixAbstractTypeResolver (resolver AbstractTypeResolver) {
	if ts.abstractTypeResolver != nil {
		ts.Error("AbsTypeResolverAlrdyDefined", "", "", nil, nil)
		return
	}
	ts.abstractTypeResolver = resolver
} //FixAbstractTypeResolver

func (ts *typeSystem) coerceValue (value Value, targetType Type, cValue *Value, variableDefinitions VariableDefinitions, variableValues *A.Tree, pathB *pathBuilder) bool { // *ValMapItem
	if Var, b := value.(*Variable); b {
		name := Var.Name
		if !getVarValue(variableValues, name, &value) {
			var v *VariableDefinition = nil
			for _, vdi := range variableDefinitions {
				if vdi.Var.S == name.S {
					v = vdi
					break
				}
			}
			if v == nil {
				ts.Error("NoValueForVar", name.S, "", nil, pathB.getPath())
				return false
			}
			value = v.DefVal
		}
		if value == nil {
			value = MakeNullValue()
		}
	}
	switch targetType := targetType.(type) {
	case *NonNullType:
		var v Value
		if !ts.coerceValue(value, targetType.NullT, &v, variableDefinitions, variableValues, pathB) {
			*cValue = MakeNullValue()
			return false
		}
		*cValue = v
		if _, b := v.(*NullValue); b {
			ts.Error("NullValueWithNonNullType", "", "", nil, pathB.getPath())
			return false
		}
		return true
	case *ListType:
		if _, b := value.(*ListValue); !b {
			ts.Error("NonListValueForListType", namedTypeOf(targetType).Name.S, "", nil, pathB.getPath())
			*cValue = MakeNullValue()
			return false
		}
		lv := NewListValue()
		val := value.(*ListValue)
		var v Value
		for vl := val.First(); vl != nil; vl = val.Next(vl) {
			if !ts.coerceValue(vl.Value, targetType.ItemT, &v, variableDefinitions, variableValues, pathB) {
				lv.Append(MakeNullValue())
			} else {
				lv.Append(v)
			}
		}
		*cValue = lv
		return true
	case *NamedType:
		if _, ok := value.(*NullValue); ok {
			*cValue = value
			return true
		}
		d := ts.getTypeDefinition(targetType.Name)
		M.Assert(d != nil, 100)
		switch d := d.(type) {
		case *ScalarTypeDefinition:
			scalarName := d.Name
			switch scalarName.S {
			case "Int":
				switch value := value.(type) {
				case *IntValue:
					if M.MinInt32 <= value.Int && value.Int <= M.MaxInt32 {
						*cValue = value
						return true
					}
					ts.Error("IntOutOfRange", "", "", nil, pathB.getPath())
				default:
					ts.Error("NonIntValueForIntType", "", "", nil, pathB.getPath())
				}
			case "Float":
				switch value := value.(type) {
				case *IntValue:
					*cValue = &FloatValue{float64(value.Int)}
					return true
				case *FloatValue:
					*cValue = value
					return true
				default:
					ts.Error("NonFloatOrIntValueForFloatType", "", "", nil, pathB.getPath())
				}
			case "String":
				switch value := value.(type) {
				case *StringValue:
					*cValue = value
					return true
				default:
					ts.Error("NonStringValueForStringType", "", "", nil, pathB.getPath())
				}
			case "Boolean":
				switch value := value.(type) {
				case *BooleanValue:
					*cValue = value
					return true
				default:
					ts.Error("NonBoolValueForBoolType", "", "", nil, pathB.getPath())
				}
			case "ID":
				switch value := value.(type) {
				case *StringValue:
					*cValue = value
					return true
				default:
					ts.Error("NonStringValueForIDType", "", "", nil, pathB.getPath())
				}
			default:
				var sc ScalarCoercer
				if ts.getCoercer(scalarName, &sc) {
					return sc(ts, value, pathB.getPath(), cValue)
				} else {
					ts.Error("NonScalarValueForScalarType", d.Name.S, "", nil, pathB.getPath())
					return false
				}
			}
		case *ObjectTypeDefinition:
			ov := NewOutputObjectValue()
			switch value := value.(type) {
			case *OutputObjectValue:
				for of := value.First(); of != nil; of = value.Next(of) {
					fd := getField(d.FieldsDef, of.Name)
					var v Value
					if (fd == nil) || !ts.coerceValue(of.Value, fd.Type, &v, variableDefinitions, variableValues, pathB.pushPathString(fd.Name)) {
						ov.insertOutputField(of.Name, MakeNullValue())
					} else {
						ov.insertOutputField(of.Name, v)
					}
				}
				*cValue = ov
				return true
			default:
				ts.Error("NonObjValForObjType", "", "", nil, pathB.getPath())
			}
		case *EnumTypeDefinition:
			switch value := value.(type) {
			case *EnumValue, *StringValue:
				var s *StrPtr
				switch value := value.(type) {
				case *EnumValue:
					s = value.Enum
					*cValue = value
				case *StringValue:
					s = value.String
					*cValue = MakeEnumValue(s.S)
				}
				en := d.EnumValsDef
				for _, e := range en {
					if e.EnumVal.Enum.S == s.S {
						return true
					}
				}
				*cValue = MakeNullValue()
				ts.Error("WrongEnumValueForEnumType", s.S, d.Name.S, s.P, pathB.getPath())
			default:
				ts.Error("NonEnumValueForEnumType", d.Name.S, "", nil, pathB.getPath())
			}
		case *InputObjectTypeDefinition:
			ov := newInputObjectValue()
			switch val := value.(type) {
			case *InputObjectValue:
				argumentDefinitions := d.InFieldsDef
				for _, argumentDefinition := range argumentDefinitions {
					argumentName := argumentDefinition.Name
					argumentType := argumentDefinition.Type
					defaultValue := argumentDefinition.DefVal
					pathBB := pathB.pushPathString(argumentName)
					var argumentValue Value
					hasValue := getObjectValueInputField(val, argumentName, &argumentValue)
					var value, coercedValue Value
					if hasValue {
						if v, ok := argumentValue.(*Variable); ok {
							variableName := v.Name
							hasValue = getVarValue(variableValues, variableName, &value)
						} else {
							value = argumentValue
						}
					}
					if !hasValue && defaultValue != nil {
						ov.insertInputField(argumentName, defaultValue)
					} else {
						_, nul := argumentType.(*NonNullType)
						if nul {
							nul = !hasValue
							if !nul {
								_, nul = value.(*NullValue)
							}
						}
						if nul {
							ts.Error("NullValueWithNonNullType", "", "", nil, pathBB.getPath())
						} else if hasValue {
							_, b := value.(*NullValue)
							if !b {
								_, b = argumentValue.(*Variable)
							}
							if b {
								ov.insertInputField(argumentName, value)
							} else if !ts.coerceValue(value, argumentType, &coercedValue, variableDefinitions, variableValues, pathBB) {
								ts.Error("UnableToCoerce", "", "", nil, pathBB.getPath())
							} else {
								ov.insertInputField(argumentName, coercedValue)
							}
						}
					}
				}
				*cValue = ov
				return true
			default:
				ts.Error("NonObjValForInputObjType", "", "", nil, pathB.getPath())
			}
		default:
			panic(100)
		}
	}
	return false
} //coerceValue

func getArgValue (a Arguments, name *StrPtr, val *Value) bool {
	for _, ai := range a {
		if ai.Name.S == name.S {
			*val = ai.Value
			return true
		}
	}
	return false
} //getArgValue

func newEntry (t *A.Tree, name *StrPtr, value Value) { // *ValMapItem
	_, b, _ := t.SearchIns(&ValMapItem{NameMapItem{name}, value}); M.Assert(!b, 100)
} //newEntry

func (ts *typeSystem) coerceArgumentValues (objectType *ObjectTypeDefinition, field *Field, variableDefinitions VariableDefinitions, variableValues *A.Tree, pathB *pathBuilder) *A.Tree { // *ValMapItem & *ValMapItem
	coercedValues := A.New() // *ValMapItem
	argumentValues := field.Arguments
	fieldName := field.Name
	var argumentDefinitions ArgumentsDefinition = nil
	fd := objectType.FieldsDef
	for _, f := range fd {
		if f.Name.S == fieldName.S {
			argumentDefinitions = f.ArgsDef
			break
		}
	}
	M.Assert(argumentDefinitions != nil, 100)
	for _, argumentDefinition := range argumentDefinitions {
		argumentName := argumentDefinition.Name
		argumentType := argumentDefinition.Type
		defaultValue := argumentDefinition.DefVal
		pathBB := pathB.pushPathString(argumentName)
		var argumentValue Value
		hasValue := getArgValue(argumentValues, argumentName, &argumentValue)
		var value, coercedValue Value
		if hasValue {
			if v, ok := argumentValue.(*Variable); ok {
				variableName := v.Name
				hasValue = getVarValue(variableValues, variableName, &value)
			} else {
				value = argumentValue
			}
		}
		if !hasValue && defaultValue != nil {
			newEntry(coercedValues, argumentName, defaultValue)
		} else {
			_, nul := argumentType.(*NonNullType)
			if nul {
				nul = !hasValue
				if !nul {
					_, nul = value.(*NullValue)
				}
			}
			if nul {
				ts.Error("NullValueWithNonNullType", "", "", nil, pathBB.getPath())
			} else if hasValue {
				_, b := value.(*NullValue)
				if !b {
					_, b = argumentValue.(*Variable)
				}
				if b {
					newEntry(coercedValues, argumentName, value)
				} else if !ts.coerceValue(value, argumentType, &coercedValue, variableDefinitions, variableValues, pathBB) {
					ts.Error("UnableToCoerce", "", "", nil, pathBB.getPath())
				} else {
					newEntry(coercedValues, argumentName, coercedValue)
				}
			}
		}
	}
	return coercedValues
} //coerceArgumentValues

func (ts *execSystem) coerceVariableValues (operation *OperationDefinition, variableValues *A.Tree) *A.Tree { // *ValMapItem & *ValMapItem
	coercedValues := A.New()
	variableDefinitions := operation.VarDefs
	if len(variableDefinitions) > 0 {
		pathB := newPathBuilder()
		if operation.Name.S != "" {
			pathB = pathB.pushPathString(operation.Name)
		}
		for _, variableDefinition := range variableDefinitions {
			variableName := variableDefinition.Var
			variableType := variableDefinition.Type
			M.Assert(ts.isInputType(variableType), 100)
			pathBB := pathB.pushPathString(variableName)
			defaultValue := variableDefinition.DefVal
			var value Value
			hasValue := getVarValue(variableValues, variableName, &value)
			if !hasValue && defaultValue != nil {
				newEntry(coercedValues, variableName, defaultValue)
			} else {
				_, nul := variableType.(*NonNullType)
				if nul {
					nul = !hasValue || value == nil
					if !nul {
						_, nul = value.(*NullValue)
					}
				}
				if nul {
					ts.Error("NullVarWithNonNullType", variableName.S, "", nil, pathBB.getPath())
				} else if hasValue {
					var coercedValue Value
					if _, nul := value.(*NullValue); nul {
						newEntry(coercedValues, variableName, value)
					} else if ts.coerceValue(value, variableType, &coercedValue, variableDefinitions, variableValues, pathBB) {
						newEntry(coercedValues, variableName, coercedValue)
					}
				}
			}
		}
	}
	return coercedValues
} //coerceVariableValues

func recValue (w ES.Writer, value Value) {
	if value != nil {
		switch value := value.(type) {
		case *IntValue:
			w.WriteString(SC.FormatInt(value.Int, 10))
		case *FloatValue:
			w.WriteString(SC.FormatFloat(value.Float, 'G', -1, 64))
		case *StringValue:
			w.WriteString("\"")
			w.WriteString(value.String.S)
			w.WriteString("\"")
		case *BooleanValue:
			if value.Boolean {
				w.WriteString("true")
			} else {
				w.WriteString("false")
			}
		case *NullValue:
			w.WriteString("null")
		case *EnumValue:
			w.WriteString(value.Enum.S)
		case *ListValue:
			w.WriteString("[")
			if vl := value.First(); vl != nil {
				recValue(w, vl.Value)
				for vl := value.Next(vl); vl != nil; vl = value.Next(vl) {
					w.WriteString(", ")
					recValue(w, vl.Value)
				}
			}
			w.WriteString("]")
		case *InputObjectValue:
			w.WriteString("{")
			if f := value.First(); f != nil {
				w.WriteString(value.inputFields.Name.S)
				w.WriteString(": ")
				recValue(w, f.Value)
				for f := value.Next(f); f != nil; f = value.Next(f) {
					w.WriteString(", ")
					w.WriteString(f.Name.S)
					w.WriteString(": ")
					recValue(w, f.Value)
				}
			}
			w.WriteString("}")
		case *OutputObjectValue:
			w.WriteString("{")
			if f := value.First(); f != nil {
				w.WriteString(f.Name.S)
				w.WriteString(": ")
				recValue(w, f.Value)
				for f := value.Next(f); f != nil; f = value.Next(f) {
					w.WriteString(", ")
					w.WriteString(f.Name.S)
					w.WriteString(": ")
					recValue(w, f.Value)
				}
			}
			w.WriteString("}")
		}
	}
} //recValue

func recType (w ES.Writer, typ Type) {
	switch typ := typ.(type) {
	case *NonNullType:
		recType(w, typ.NullT)
		w.WriteString("!")
	case *ListType:
		w.WriteString("[")
		recType(w, typ.ItemT)
		w.WriteString("]")
	case *NamedType:
		w.WriteString(typ.Name.S)
	}
} //recType

func (ts *typeSystem) resolveFieldValue (objectType *ObjectTypeDefinition, objectValue  *OutputObjectValue, fieldName *StrPtr, argumentValues *A.Tree, pathB *pathBuilder) Value { // *ValMapItem

	ProcessIntrospection := func () Value {
		
		MakeStringFromValue := func (value Value) *StrPtr {
			s := ES.Dir().New()
			recValue(s.NewWriter(), value)
			return makeName(s.Convert())
		} //MakeStringFromValue
		
		MakeStringFromType := func (typ Type) *StrPtr {
			s := ES.Dir().New()
			recType(s.NewWriter(), typ)
			return makeName(s.Convert())
		} //MakeStringFromType
		
		ProcessTypename := func () *StringValue {
			return makeStringValue(objectType.Name)
		} //ProcessTypename
		
		ProcessRoot := func () Value {
			switch fieldName.S {
			case "__schema":
				return MakeAnyValue(nil)
			case "__type":
				var value Value
				ok := GetValue(argumentValues, "name", &value); M.Assert(ok, 100)
				name := value.(*StringValue).String
				d := ts.getTypeDefinition(name)
				if d == nil {
					ts.Error("UnknownType", name.S, "", nil, pathB.getPath())
					return MakeAnyValue(nil)
				}
				return MakeAnyValue(d)
			case "__typename":
				return ProcessTypename()
			default:
				ts.Error("IsNotFieldFor", fieldName.S, objectType.Name.S, nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessRoot
		
		ProcessSchema := func () Value {
			switch fieldName.S {
			case "types":
				listValue := NewListValue()
				e := ts.typeMap.Next(nil)
				for e != nil {
					el := e.Val().(*defMapItem)
					if el.def.(TypeDefinition).listableM() {
						listValue.Append(MakeAnyValue(el.def))
					}
					e = ts.typeMap.Next(e)
				}
				return listValue
			case "queryType":
				return MakeAnyValue(ts.root[QueryOp])
			case "mutationType":
				if ts.root[MutationOp] == nil {
					return MakeAnyValue(nil)
				}
				return MakeAnyValue(ts.root[MutationOp])
			case "subscriptionType":
				if ts.root[SubscriptionOp] == nil {
					return MakeAnyValue(nil)
				}
				return MakeAnyValue(ts.root[SubscriptionOp])
			case "directives":
				listValue := NewListValue()
				e := ts.dirMap.Next(nil)
				for e != nil {
					el := e.Val().(*defMapItem)
					if el.def.(*DirectiveDefinition).listable {
						listValue.Append(MakeAnyValue(el.def))
					}
					e = ts.dirMap.Next(e)
				}
				return listValue
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__Schema", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessSchema
		
		ProcessType := func () Value {
			a := objectValue.outputFields.Value.(*AnyValue).Any
			if aa, ok := a.(*NamedType); ok {
				a = ts.getTypeDefinition(aa.Name)
			}
			switch fieldName.S {
			case "kind":
				var name string
				switch a.(type) {
				case *ScalarTypeDefinition:
					name = "SCALAR"
				case *ObjectTypeDefinition:
					name = "OBJECT"
				case *InterfaceTypeDefinition:
					name = "INTERFACE"
				case *UnionTypeDefinition:
					name = "UNION"
				case *EnumTypeDefinition:
					name = "ENUM"
				case *InputObjectTypeDefinition:
					name = "INPUT_OBJECT"
				case *NonNullType:
					name = "NON_NULL"
				case *ListType:
					name = "LIST"
				}
				return MakeEnumValue(name)
			case "name":
				var name *StrPtr
				switch a := a.(type) {
				case TypeDefinition:
					name = a.TypeDefinitionC().Name
				case Type:
					name = MakeStringFromType(a)
				}
				return makeStringValue(name)
			case "description":
				var value Value
				switch a := a.(type) {
				case TypeDefinition:
					value = a.TypeDefinitionC().Desc
				case Type:
					value = ts.getTypeDefinition(namedTypeOf(a).Name).TypeDefinitionC().Desc
				}
				return value
			case "fields":
				var fs FieldsDefinition
				switch a := a.(type) {
				case *ObjectTypeDefinition:
					fs = a.FieldsDef
				case *InterfaceTypeDefinition:
					fs = a.FieldsDef
				default:
					return MakeNullValue()
				}
				var value Value
				ok := GetValue(argumentValues, "includeDeprecated", &value); M.Assert(ok, 100)
				includeDeprecated := value.(*BooleanValue).Boolean
				listValue := NewListValue()
				for _, f := range fs {
					if (includeDeprecated || !f.isDeprecated) && !isIntro(f.Name) {
						listValue.Append(MakeAnyValue(f))
					}
				}
				return listValue
			case "interfaces":
				var nts NamedTypes
				switch a := a.(type) {
				case *ObjectTypeDefinition:
					nts = a.ImpInter
				default:
					return MakeNullValue()
				}
				listValue := NewListValue()
				for _, nt := range nts {
					d := ts.getTypeDefinition(nt.Name); M.Assert(d != nil, 101)
					listValue.Append(MakeAnyValue(d))
				}
				return listValue
			case "possibleTypes":
				var nts NamedTypes
				switch a := a.(type) {
				case *InterfaceTypeDefinition:
					nts = a.implementedBy
				case *UnionTypeDefinition:
					nts = a.UnionMTypes
				default:
					return MakeNullValue()
				}
				listValue := NewListValue()
				for _, nt := range nts {
					d := ts.getTypeDefinition(nt.Name); M.Assert(d != nil, 102)
					listValue.Append(MakeAnyValue(d))
				}
				return listValue
			case "enumValues":
				var evds EnumValuesDefinition
				switch a := a.(type) {
				case *EnumTypeDefinition:
					evds = a.EnumValsDef
				default:
					return MakeNullValue()
				}
				var value Value
				ok := GetValue(argumentValues, "includeDeprecated", &value); M.Assert(ok, 103)
				includeDeprecated := value.(*BooleanValue).Boolean
				listValue := NewListValue()
				for _, evd := range evds {
					if includeDeprecated || !evd.isDeprecated {
						listValue.Append(MakeAnyValue(evd))
					}
				}
				return listValue
			case "inputFields":
				var as ArgumentsDefinition
				switch a := a.(type) {
				case *InputObjectTypeDefinition:
					as = a.InFieldsDef
				default:
					return MakeNullValue()
				}
				listValue := NewListValue()
				for _, ivd := range as {
					listValue.Append(MakeAnyValue(ivd))
				}
				return listValue
			case "ofType":
				var ty Type
				switch a := a.(type) {
				case *NonNullType:
					ty = a.NullT
				case *ListType:
					ty = a.ItemT
				default:
					return MakeNullValue()
				}
				return MakeAnyValue(ty)
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__Type", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessType
		
		ProcessField := func () Value {
			f := objectValue.outputFields.Value.(*AnyValue).Any.(*FieldDefinition)
			switch fieldName.S {
			case "name":
				return makeStringValue(f.Name)
			case "description":
				return f.Desc
			case "args":
				listValue := NewListValue()
				for _, a := range f.ArgsDef {
					listValue.Append(MakeAnyValue(a))
				}
				return listValue
			case "type":
				return MakeAnyValue(f.Type)
			case "isDeprecated":
				return MakeBooleanValue(f.isDeprecated)
			case "deprecationReason":
				if f.deprecationReason == nil {
					return MakeStringValue("")
				}
				return makeStringValue(f.deprecationReason)
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__Field", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessField
		
		ProcessInputValue := func () Value {
			ivd := objectValue.outputFields.Value.(*AnyValue).Any.(*InputValueDefinition)
			switch fieldName.S {
			case "name":
				return makeStringValue(ivd.Name)
			case "description":
				return ivd.Desc
			case "type":
				return MakeAnyValue(ivd.Type)
			case "defaultValue":
				return makeStringValue(MakeStringFromValue(ivd.DefVal))
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__InputValue", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessInputValue
		
		ProcessEnumValue := func () Value {
			enum := objectValue.outputFields.Value.(*AnyValue).Any.(*EnumValueDefinition)
			switch fieldName.S {
			case "name":
				return makeStringValue(enum.EnumVal.Enum)
			case "description":
				return enum.Desc
			case "isDeprecated":
				return MakeBooleanValue(enum.isDeprecated)
			case "deprecationReason":
				if enum.deprecationReason == nil {
					return MakeStringValue("")
				}
				return makeStringValue(enum.deprecationReason)
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__EnumValue", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessEnumValue
		
		ProcessDirective := func () Value {
			dir := objectValue.outputFields.Value.(*AnyValue).Any.(*DirectiveDefinition)
			switch fieldName.S {
			case "name":
				return makeStringValue(dir.Name)
			case "description":
				return dir.Desc
			case "locations":
				listValue := NewListValue()
				for _, l := range dir.DirLocs {
					listValue.Append(MakeEnumValue(locationNameOf(l)))
				}
				return listValue
			case "args":
				listValue := NewListValue()
				for _, a := range dir.ArgsDef {
					listValue.Append(MakeAnyValue(a))
				}
				return listValue
			default:
				ts.Error("IsNotFieldFor", fieldName.S, "__Directive", nil, pathB.getPath())
				return MakeNullValue()
			}
		} //ProcessDirective
		
		//ProcessIntrospection
		if objectType == ts.root[QueryOp] {
			return ProcessRoot()
		} else {
			switch objectType.Name.S {
			case "__Schema":
				return ProcessSchema()
			case "__Type":
				return ProcessType()
			case "__Field":
				return ProcessField()
			case "__InputValue":
				return ProcessInputValue()
			case "__EnumValue":
				return ProcessEnumValue()
			case "__Directive":
				return ProcessDirective()
			default:
				if fieldName.S == "__typename" {
					return ProcessTypename()
				}
				ts.Error("IsNotFieldFor", fieldName.S, objectType.Name.S, nil, pathB.getPath())
				return MakeNullValue()
			}
		}
	} //ProcessIntrospection
	
	//resolveFieldValue
	if isIntro(objectType.Name) || isIntro(fieldName) {
		return ProcessIntrospection()
	}
	f := getField(objectType.FieldsDef, fieldName)
	M.Assert(f != nil, 100)
	resolver := f.resolver
	if resolver == nil {
		ts.Error("ResolverNotDefined", objectType.Name.S, fieldName.S, nil, pathB.getPath())
		return nil
	}
	return resolver(objectValue, argumentValues)
} //resolveFieldValue

func (ts *typeSystem) resolveAbstractType (abstractType TypeDefinition, objectValue *OutputObjectValue) *ObjectTypeDefinition {
	if ts.abstractTypeResolver == nil {
		ts.Error("AbsTypeResolverNotDefined", "", "", nil, nil)
		return nil
	}
	d := ts.abstractTypeResolver(ts, abstractType, objectValue)
	if d == nil {
		ts.Error("AbsTypeNotResolved", abstractType.TypeDefinitionC().Name.S, "", nil, nil)
		return nil
	}
	return d
} //resolveAbstractType

func (ts *typeSystem) mergeSelectionSets (fields *fieldRing) SelectionSet {
	n := 0
	fr := fields.next
	for fr != fields {
		fieldSelectionSet := fr.field.SelSet
		n += len(fieldSelectionSet)
		fr = fr.next
	}
	selectionSet := make(SelectionSet, n)
	n = 0
	fr = fields.next
	for fr != fields {
		fieldSelectionSet := fr.field.SelSet
		for _, sel := range fieldSelectionSet {
			selectionSet[n] = sel
			n++
		}
		fr = fr.next
	}
	return selectionSet
} //mergeSelectionSets

func (es *execSystem) completeValue (fieldType Type, fields *fieldRing, result Value, variableDefinitions VariableDefinitions, variableValues *A.Tree, pathB *pathBuilder) Value { // *ValMapItem
	
	MakeObject := func (name string, value Value) *OutputObjectValue {
		ov := NewOutputObjectValue()
		ov.InsertOutputField(name, value)
		return ov
	} //MakeObject

	//completeValue
	if f, ok := fieldType.(*NonNullType); ok {
		innerType := f.NullT
		completedResult := es.completeValue(innerType, fields, result, variableDefinitions, variableValues, pathB)
		nul := completedResult == nil
		if !nul {
			_, nul = completedResult.(*NullValue)
		}
		if nul {
			es.Error("NullValueWithNonNullType", "", "", nil, pathB.getPath())
		}
		return completedResult
	}
	nul := result == nil
	if !nul {
		_, nul = result.(*NullValue)
	}
	if nul {
		return MakeNullValue()
	}
	if f, ok := fieldType.(*ListType); ok {
		if _, ok := result.(*ListValue); !ok {
			es.Error("NonListValueForListType", namedTypeOf(f).Name.S, "", nil, pathB.getPath())
			return MakeNullValue()
		}
		innerType := f.ItemT
		listValue := NewListValue()
		res := result.(*ListValue)
		i := 0
		for vl := res.First(); vl != nil; vl = res.Next(vl) {
			resultItem := vl.Value
			listValue.Append(es.completeValue(innerType, fields, resultItem, variableDefinitions, variableValues, pathB.pushPathNb(i)))
			i++
		}
		return listValue
	}
	ty := es.typeOf(fieldType)
	if M.In(ty, M.MakeSet(t_SCALAR, t_ENUM)) {
		var completedResult Value
		if !es.coerceValue(result, fieldType, &completedResult, variableDefinitions, variableValues, pathB) {
			return MakeNullValue()
		}
		return completedResult
	}
	M.Assert(M.In(ty, M.MakeSet(t_OBJECT, t_INTERFACE, t_UNION)), 100)
	_, b := result.(*OutputObjectValue)
	if !b {
		_, b = result.(*AnyValue)
	}
	M.Assert(b, 101)
	d := es.getTypeDefinition(fieldType.(*NamedType).Name)
	M.Assert(d != nil, 102)
	var obj *OutputObjectValue
	if _, ok := result.(*AnyValue); ok {
		obj = MakeObject("object", result)
	} else {
		obj = result.(*OutputObjectValue)
	}
	var objectType *ObjectTypeDefinition
	if ty == t_OBJECT {
		objectType = d.(*ObjectTypeDefinition)
	} else {
		objectType = es.resolveAbstractType(d, obj)
		if objectType == nil {
			return MakeNullValue()
		}
	}
	subSelectionSet := es.mergeSelectionSets(fields)
	return es.executeSelectionSet(subSelectionSet, objectType, obj, variableDefinitions, variableValues, pathB, true)
} //completeValue

func (es *execSystem) executeField (objectType *ObjectTypeDefinition, objectValue *OutputObjectValue, fieldType Type, fields *fieldRing, variableDefinitions VariableDefinitions, variableValues *A.Tree, pathB *pathBuilder) Value { // *ValMapItem
	field := fields.next.field; M.Assert(fields.next != fields, 100)
	fieldName := field.Name
	pathBB := pathB.pushPathString(field.Alias)
	argumentValues := es.coerceArgumentValues(objectType, field, variableDefinitions, variableValues, pathBB)
	resolvedValue := es.resolveFieldValue(objectType, objectValue, fieldName, argumentValues, pathBB)
	return es.completeValue(fieldType, fields, resolvedValue, variableDefinitions, variableValues, pathBB)
} //executeField

func (es *execSystem) executeSelectionSet (selectionSet SelectionSet, objectType *ObjectTypeDefinition, objectValue *OutputObjectValue, variableDefinitions VariableDefinitions, variableValues *A.Tree, pathB *pathBuilder, parallel bool) *OutputObjectValue { // *ValMapItem
	
	executeSelection := func (groupedFieldSet *orderedFieldsMap, responseKey *StrPtr, fieldPtr *ObjectField) {
		e, b, _ := groupedFieldSet.t.Search(&fieldRingElem{NameMapItem: NameMapItem{responseKey}}); M.Assert(b, 100)
		fields := e.Val().(*fieldRingElem).fr
		fieldName := fields.next.field.Name; M.Assert(fields.next != fields, 101)
		field := getField(objectType.FieldsDef, fieldName)
		if field != nil {
			fieldType := field.Type
			responseValue := es.executeField(objectType, objectValue, fieldType, fields, variableDefinitions, variableValues, pathB)
			fieldPtr.Value = responseValue
		}
	} //executeSelection
	
	//executeSelectionSet
	visitedFragments := A.New()
	groupedFieldSet := es.collectFields(objectType, selectionSet, variableValues, visitedFragments)
	resultMap := NewOutputObjectValue()
	q := groupedFieldSet.q
	if q != nil {
		if parallel { // Concurrent if parallel
			n := 0
			for { // Concurrent if parallel
				q = q.next
				n++
				if q == groupedFieldSet.q {break}
			}
			var wg sync.WaitGroup
			wg.Add(n)
			for {
				q = q.next
				responseKey := q.name
				f := resultMap.insertOutputField(responseKey, nil)
				go func (wg *sync.WaitGroup, groupedFieldSet *orderedFieldsMap, responseKey *StrPtr, f *ObjectField) {
					defer wg.Done()
					executeSelection(groupedFieldSet, responseKey, f)
				}(&wg, groupedFieldSet, responseKey, f)
				if q == groupedFieldSet.q {break}
			}
			wg.Wait()
		} else {
			for {
				q = q.next
				responseKey := q.name
				f := resultMap.insertOutputField(responseKey, nil)
				executeSelection(groupedFieldSet, responseKey, f)
				if q == groupedFieldSet.q {break}
			}
		}
	}
	return resultMap
} //executeSelectionSet

func (es *execSystem) executeSubscriptionEvent (subscription *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) Response { // *ValMapItem
	newES := new(execSystem)
	*newES = *es
	newES.SetErrors(A.New())
	pathB := newPathBuilder()
	subscriptionType := newES.root[SubscriptionOp]
	selectionSet := subscription.SelSet
	data := newES.executeSelectionSet(selectionSet, subscriptionType, initialValue, subscription.VarDefs, variableValues, pathB.pushPathString(subscriptionType.Name), true)
	return &InstantResponse{newES.GetErrors(), data}
} //executeSubscriptionEvent

func (ts *typeSystem) getStreamResolver (fieldName *StrPtr) StreamResolver {
	if e, found, _ := ts.streamResolvers.Search(&streamMapItem{NameMapItem: NameMapItem{fieldName}}); found {
		return e.Val().(*streamMapItem).stream
	}
	return nil
} //getStreamResolver

func (ts *typeSystem) resolveFieldEventStream (rootValue *OutputObjectValue, fieldName *StrPtr, argumentValues *A.Tree, pathB *pathBuilder) *EventStream { // *ValMapItem
	resolver := ts.getStreamResolver(fieldName)
	if resolver == nil {
		ts.Error("StreamResolverNotDefined", fieldName.S, "", nil, pathB.getPath())
		return nil
	}
	return resolver(rootValue, argumentValues)
} //resolveFieldEventStream

func (es *execSystem) createSourceEventStream (subscription *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) *EventStream { // *ValMapItem
	pathB := newPathBuilder()
	subscriptionType := es.root[SubscriptionOp]
	pathB = pathB.pushPathString(subscriptionType.Name)
	selectionSet := subscription.SelSet
	visitedFragments := A.New() // *NameMapItem
	groupedFieldSet :=  es.collectFields(subscriptionType, selectionSet, variableValues, visitedFragments)
	M.Assert(groupedFieldSet.t.NumberOfElems() == 1, 100)
	fields := groupedFieldSet.t.Next(nil).Val().(*fieldRingElem).fr
	field := fields.next.field; M.Assert(fields.next != fields, 100)
	fieldName := field.Name
	pathB = pathB.pushPathString(field.Alias)
	argumentValues := es.coerceArgumentValues(subscriptionType, field, subscription.VarDefs, variableValues, pathB)
	fieldStream := es.resolveFieldEventStream(initialValue, fieldName, argumentValues, pathB)
	fieldStream.ResponseStreams = make(ResponseStreams, 0)
	return fieldStream
} //createSourceEventStream

// Event sent by context: acknowledged by a Response
func getSourceEvent (es *EventStream, event *OutputObjectValue) {
	for _, rs := range es.ResponseStreams {
		rs.ManageResponseEvent (rs.es.executeSubscriptionEvent(rs.subscription, rs.variableValues, event))
	}
} //getSourceEvent

func (rs *ResponseStream) FixResponseStream (s ResponseStreamer) {
	rs.ResponseStreamer = s
} //FixResponseStream

func (es *execSystem) mapSourceToResponseEvent (sourceStream *EventStream, subscription *OperationDefinition, variableValues *A.Tree) *ResponseStream { // *ValMapItem
	responseStream := &ResponseStream{SourceStream: sourceStream, es: es, subscription: subscription, variableValues: variableValues}
	sourceStream.ResponseStreams = append(sourceStream.ResponseStreams, responseStream)
	sourceStream.RecordNotificationProc(getSourceEvent)
	return responseStream
} //mapSourceToResponseEvent

func (es *execSystem) executeQuery (query *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) Response { // *ValMapItem
	pathB := newPathBuilder()
	queryType := es.root[QueryOp]
	selectionSet := query.SelSet
	variableDefinitions := query.VarDefs
	data := es.executeSelectionSet(selectionSet, queryType, initialValue, variableDefinitions, variableValues, pathB.pushPathString(queryType.Name), true)
	return &InstantResponse{Data: data}
} //executeQuery

func (es *execSystem) executeMutation (mutation *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) Response { // *ValMapItem
	pathB := newPathBuilder()
	mutationType := es.root[MutationOp]
	selectionSet := mutation.SelSet
	variableDefinitions := mutation.VarDefs
	data := es.executeSelectionSet(selectionSet, mutationType, initialValue, variableDefinitions, variableValues, pathB.pushPathString(mutationType.Name), false)
	return &InstantResponse{Data: data}
} //executeMutation

func (es *execSystem) subscribe (subscription *OperationDefinition, variableValues *A.Tree, initialValue *OutputObjectValue) Response { // *ValMapItem
	sourceStream := es.createSourceEventStream(subscription, variableValues, initialValue)
	var responseStream *ResponseStream = nil
	if sourceStream != nil {
		responseStream = es.mapSourceToResponseEvent(sourceStream, subscription, variableValues)
	}
	return &SubscribeResponse{Data: responseStream, Name: subscription.Name.S}
} //subscribe

func Unsubscribe (responseStream *ResponseStream) {
	responseStream.SourceStream.CloseEvent()
} //Unsubscribe

func (es *execSystem) ListOperations () StrArray {
	m := es.opMap
	if m == nil || m.IsEmpty() {
		return nil
	}
	l := make(StrArray, m.NumberOfElems())
	i := 0
	e := m.Next(nil)
	for e != nil {
		l[i] = e.Val().(*defMapItem).Name.S
		i++
		e = m.Next(e)
	}
	return l
} //ListOperations

func (es *execSystem) getOperation (operationName *StrPtr) *OperationDefinition {
	if operationName.S == "" {
		if es.opMap.NumberOfElems() == 1 {
			return es.opMap.Next(nil).Val().(*defMapItem).def.(*OperationDefinition)
		}
		es.Error("NotDefinedOp", "", "", nil, nil)
		return nil
	}
	if e, ok, _ := es.opMap.Search(&defMapItem{NameMapItem: NameMapItem{operationName}}); ok {
		return e.Val().(*defMapItem).def.(*OperationDefinition)
	}
	es.Error("UnknownOperation", operationName.S, "", nil, nil)
	return nil
} //getOperation

func (es *execSystem) GetOperation (operationName string) *OperationDefinition {
	return es.getOperation(makeName(operationName))
} //GetOperation

func (es *execSystem) executeRequest (operationName *StrPtr, variableValues *A.Tree, initialValue *OutputObjectValue) Response { // *ValMapItem
	operation := es.getOperation(operationName)
	if operation == nil {
		return &InstantResponse{es.GetErrors(), nil}
	}
	coercedVariableValues := es.coerceVariableValues(operation, variableValues)
	if !es.GetErrors().IsEmpty() {
		return &InstantResponse{es.GetErrors(), nil}
	}
	var r Response
	switch operation.OpType {
	case QueryOp:
		r = es.executeQuery(operation, coercedVariableValues, initialValue)
	case MutationOp:
		r = es.executeMutation(operation, coercedVariableValues, initialValue)
	case SubscriptionOp:
		r = es.subscribe(operation, coercedVariableValues, initialValue)
	}
	r.SetErrors(nil)
	if !es.GetErrors().IsEmpty() {
		r.SetErrors(es.GetErrors())
	}
	return r
} //executeRequest

// ************** /Execution ************

// ************** Type System ************

func (ts *typeSystem) verifyNotIntro (name *StrPtr) {
	if ts.verifNotIntro && isIntro(name) {
		ts.Error("IntroName", name.S, "", name.P, nil)
	}
} //verifyNotIntro

func (ts *typeSystem) validateTypesDirectives (dirs Directives) {

	ValidateArgs := func (asf ArgumentsDefinition, as Arguments, parentName *StrPtr) {
		for i := 0; i < len(as); i++ {
			for j := i + 1; j < len(as); j++ {
				if as[i].Name.S == as[j].Name.S {
					ts.Error("DupArgName", as[j].Name.S, "", as[j].Name.P, nil)
				}
			}
		}
		ts.validateArgumentValuesCoercion(as, asf, parentName, nil)
	} //ValidateArgs

	//validateTypesDirectives
	for _, dir := range dirs {
		d := ts.getDirectiveDefinition(dir.Name)
		M.Assert(d != nil, 100)
		ValidateArgs(d.ArgsDef, dir.Args, dir.Name)
	}
} //validateTypesDirectives

func (ts *typeSystem) validateArgumentsDefinition (ds ArgumentsDefinition) {

	ValidateArgumentDefinition := func (d *InputValueDefinition) {
		ts.verifyNotIntro(d.Name)
		if !ts.isInputType(d.Type) {
			ts.Error("NotInputField", d.Name.S, "", d.Name.P, nil)
		}
		if d.DefVal != nil && !ts.isConstInputValueCoercibleToType(d.DefVal, nil, d.Type, nil) {
			ts.Error("DefValIncoercibleToTypeIn", d.Name.S, "", d.Name.P, nil)
		}
		ts.validateTypesDirectives(d.Dirs)
		ts.verifyLocation(d.Dirs, l_ARGUMENT_DEFINITION)
	} //ValidateArgumentDefinition

	//validateArgumentsDefinition
	t := A.New()
	for _, d := range ds {
		name := d.Name
		if _, b, _ := t.SearchIns(&NameMapItem{name}); b {
			ts.Error("DupArgName", name.S, "", name.P, nil)
		}
		ValidateArgumentDefinition(d)
	}
} //validateArgumentsDefinition

func (ts *typeSystem) subtype (ty1, ty2 Type) bool {
	if equalTypes(ty1, ty2) {
		return true
	}
	switch ty1 := ty1.(type) {
	case *NamedType:
		d1 := ts.getTypeDefinition(ty1.Name)
		if d1 != nil {
			switch d1 := d1.(type) {
			case *ObjectTypeDefinition:
				switch ty2 := ty2.(type) {
				case *NamedType:
					d2 := ts.getTypeDefinition(ty2.Name)
					if d2 != nil {
						switch d2 := d2.(type) {
						case *InterfaceTypeDefinition:
							name := d2.Name
							for _, nt := range d1.ImpInter {
								if nt.Name.S == name.S {
									return true
								}
							}
						case *UnionTypeDefinition:
							name := d1.Name
							for _, nt := range d2.UnionMTypes {
								if nt.Name.S == name.S {
									return true
								}
							}
						default:
						}
					}
				}
			default:
			}
		}
	case *ListType:
		switch ty2 := ty2.(type) {
		case *ListType:
			return ts.subtype(ty1.ItemT, ty2.ItemT)
		default:
		}
	case *NonNullType:
		switch ty2 := ty2.(type) {
		case *NonNullType:
			return ts.subtype(ty1.NullT, ty2.NullT)
		default:
			return ts.subtype(ty1.NullT, ty2)
		}
	default:
	}
	return false
} //subtype

func (ts *typeSystem) validateTypeDefinition (d TypeDefinition) {

	FixDeprecation := func (dirs Directives) (isDeprecated bool, deprecationReason *StrPtr) {
		isDeprecated = false
		deprecationReason = nil
		for _, dir := range dirs {
			if dir.Name.S == "deprecated" {
				isDeprecated = true
				M.Assert(len(dir.Args) <= 1, 100)
				if len(dir.Args) == 1 {
					deprecationReason = dir.Args[0].Value.(*StringValue).String
				}
				break
			}
		}
		return
	} //FixDeprecation
	
	ValidateFieldDefinition := func (t *A.Tree, d *FieldDefinition) { // *fieldDefMapItem
		if _, b, _ := t.SearchIns(&fieldDefMapItem{NameMapItem{d.Name}, d}); b {
			ts.Error("DupFieldName", d.Name.S, "", d.Name.P, nil)
		}
		ts.verifyNotIntro(d.Name)
		if !ts.isOutputType(d.Type) {
			ts.Error("NotOutputType", d.Name.S, "", d.Name.P, nil)
		}
		ts.validateArgumentsDefinition(d.ArgsDef)
		ts.validateTypesDirectives(d.Dirs)
		ts.verifyLocation(d.Dirs, l_FIELD_DEFINITION)
		d.isDeprecated, d.deprecationReason = FixDeprecation(d.Dirs)
	} //ValidateFieldDefinition

	ValidateObjectDefinition := func (d *ObjectTypeDefinition) {
	
		ValidateImplementation := func (t *A.Tree, ty *NamedType) { // *fieldDefMapItem
		
			ValidateArguments := func (a, ai ArgumentsDefinition) {
				for _, aai := range ai {
					var ivd *InputValueDefinition
					ok := false
					for _, aa := range a {
						if aa.Name.S == aai.Name.S {
							ok = true
							ivd = aa
							break
						}
					}
					if !ok {
						ts.Error("UnknownArgument", aai.Name.S, "", aai.Name.P, nil)
					} else if !equalTypes(ivd.Type, aai.Type) {
						ts.Error("WrongArgumentType", aai.Name.S, "", aai.Name.P, nil)
					}
				}
				for _, aa := range a {
					ts.verifyNotIntro(aa.Name)
					ok := false
					for _, aai := range ai {
						if aai.Name.S == aa.Name.S {
							ok = true
							break
						}
					}
					if !ok {
						if _, b := aa.Type.(*NonNullType); b {
							ts.Error("ArgNotRequiredByInterface", aa.Name.S, "", aa.Name.P, nil)
						}
					}
				}
			} //ValidateArguments
		
			//ValidateImplementation
			di := ts.getTypeDefinition(ty.Name)
			if di != nil {
				ddi := di.(*InterfaceTypeDefinition)
				fi := ddi.FieldsDef
				for _, ffi := range fi {
					if e, b, _ := t.Search(&fieldDefMapItem{NameMapItem: NameMapItem{ffi.Name}}); !b {
						ts.Error("NotImplementedIn", ffi.Name.S, d.Name.S, ffi.Name.P, nil)
					} else {
						f := e.Val().(*fieldDefMapItem).field
						if !ts.subtype(f.Type, ffi.Type) {
							ts.Error("NotSubtype", f.Name.S, ffi.Name.S, f.Name.P, nil)
						}
						ValidateArguments(f.ArgsDef, ffi.ArgsDef)
					}
				}
			}
		} //ValidateImplementation
	
		//ValidateObjectDefinition
		t := A.New() // *fieldDefMapItem
		for _, fd := range d.FieldsDef {
			ValidateFieldDefinition(t, fd)
		}
		ti := A.New() // *NameMapItem
		for _, ii := range d.ImpInter {
			if _, b, _ := ti.SearchIns(&NameMapItem{ii.Name}); b {
				ts.Error("DupInterface", ii.Name.S, "", ii.Name.P, nil)
			}
			ValidateImplementation(t, ii)
		}
	} //ValidateObjectDefinition

	ValidateInterfaceDefinition := func (d *InterfaceTypeDefinition) {
		t := A.New() // *fieldDefMapItem
		for _, fd := range d.FieldsDef {
			ValidateFieldDefinition(t, fd)
		}
	} //ValidateInterfaceDefinition

	ValidateUnionDefinition := func (d *UnionTypeDefinition) {
		t := A.New() // *NameMapItem
		for _, u := range d.UnionMTypes {
			name := u.Name
			if _, b, _ := t.SearchIns(&NameMapItem{name}); b {
				ts.Error("DupTypeName", name.S, "", name.P, nil)
			}
			do := ts.getTypeDefinition(name)
			if do != nil {
				if _, b := do.(*ObjectTypeDefinition); !b {
					ts.Error("NotObjectTypeInUnion", name.S, d.Name.S, name.P, nil)
				}
			}
		}
	} //ValidateUnionDefinition

	ValidateEnumDefinition := func (d *EnumTypeDefinition) {
		t := A.New() // *NameMapItem
		evs := d.EnumValsDef
		for _, ev := range evs {
			name := ev.EnumVal.Enum
			if _, b, _ := t.SearchIns(&NameMapItem{name}); b {
				ts.Error("DupEnumValue", name.S, "", name.P, nil)
			}
			ts.validateTypesDirectives(d.Dirs)
			ts.verifyLocation(ev.Dirs, l_ENUM_VALUE)
			ev.isDeprecated, ev.deprecationReason = FixDeprecation(ev.Dirs)
		}
	} //ValidateEnumDefinition

	ValidateInputObjectDefinition := func (d *InputObjectTypeDefinition) {
		ts.validateArgumentsDefinition(d.InFieldsDef)
	} //ValidateInputObjectDefinition

	//validateTypeDefinition
	ts.validateTypesDirectives(d.TypeDefinitionC().Dirs)
	switch d := d.(type) {
	case *ScalarTypeDefinition:
		ts.verifyLocation(d.Dirs, l_SCALAR)
	case *ObjectTypeDefinition:
		ValidateObjectDefinition(d)
		ts.verifyLocation(d.Dirs, l_OBJECT)
	case *InterfaceTypeDefinition:
		ValidateInterfaceDefinition(d)
		ts.verifyLocation(d.Dirs, l_INTERFACE)
	case *UnionTypeDefinition:
		ValidateUnionDefinition(d)
		ts.verifyLocation(d.Dirs, l_UNION)
	case *EnumTypeDefinition:
		ValidateEnumDefinition(d)
		ts.verifyLocation(d.Dirs, l_ENUM)
	case *InputObjectTypeDefinition:
		ValidateInputObjectDefinition(d)
		ts.verifyLocation(d.Dirs, l_INPUT_OBJECT)
	}
} //validateTypeDefinition

func (ts *typeSystem) walkType (ty Type, visited *A.Tree) { // *NameMapItem
	
	WalkTypeDefinition := func (d TypeDefinition) {
		
		WalkArgumentsDefinition := func (as ArgumentsDefinition) {
			for _, a := range as {
				ts.walkType(a.Type, visited)
				ts.walkDirs(a.Dirs, visited)
			}
		} //WalkArgumentsDefinition
	
		//WalkTypeDefinition
		ts.walkDirs(d.TypeDefinitionC().Dirs, visited)
		switch d := d.(type) {
		case *ObjectTypeDefinition:
			for _, i := range d.ImpInter {
				ts.walkType(i, visited)
			}
			for _, f := range d.FieldsDef {
				WalkArgumentsDefinition(f.ArgsDef)
				ts.walkType(f.Type, visited)
			}
		case *InterfaceTypeDefinition:
			for _, f := range d.FieldsDef {
				WalkArgumentsDefinition(f.ArgsDef)
				ts.walkType(f.Type, visited)
			}
		case *UnionTypeDefinition:
			for _, u := range d.UnionMTypes {
				ts.walkType(u, visited)
			}
		case *EnumTypeDefinition:
			for _, e := range d.EnumValsDef {
				ts.walkDirs(e.Dirs, visited)
			}
		case *InputObjectTypeDefinition:
			WalkArgumentsDefinition(d.InFieldsDef)
		default:
		}
	} //WalkTypeDefinition

	//walkType
	switch ty := ty.(type) {
	case *NonNullType:
		ts.walkType(ty.NullT, visited)
	case *ListType:
		ts.walkType(ty.ItemT, visited)
	case *NamedType:
		td := ts.getTypeDefinition(ty.Name)
		if td != nil {
			WalkTypeDefinition(td)
		}
	}
} //walkType

func (ts *typeSystem) verifyNoCycle (d *DirectiveDefinition, visited *A.Tree) { // *NameMapItem
	if _, b, _ := visited.SearchIns(&NameMapItem{d.Name}); b {
		ts.Error("CycleThrough", d.Name.S, "", d.Name.P, nil)
	}
	as := d.ArgsDef
	for _, a := range as {
		ts.walkType(a.Type, visited)
		ts.walkDirs(a.Dirs, visited)
	}
} //verifyNoCycle

func (ts *typeSystem) walkDirs (ds Directives, visited *A.Tree) { // *NameMapItem
	for _, dir := range ds {
		d := ts.getDirectiveDefinition(dir.Name)
		if d == nil {
			ts.Error("UnknownDirective", dir.Name.S, "", dir.Name.P, nil)
		} else {
			ts.verifyNoCycle(d, visited)
		}
	}
} //walkDirs

func (ts *typeSystem) validateDirectiveDefinition (d *DirectiveDefinition) {
	ts.verifyNotIntro(d.Name)
	ts.validateArgumentsDefinition(d.ArgsDef)
	visited := A.New() // *NameMapItem
	ts.verifyNoCycle(d, visited)
} //validateDirectiveDefinition

func (ts *typeSystem) collectInfos (ds Definitions) {

	CopyDirs := func (from, to *Directives) {
		dirs := make(Directives, len(*from) + len(*to))
		k := 0
		for _, dir := range *to {
			dirs[k] = dir
			k++
		}
		for _, dir := range *from {
			dirs[k] = dir
			k++
		}
		*to = dirs
		*from = nil
	} //CopyDirs
	
	FixRoot := func (root **ObjectTypeDefinition, name *StrPtr) {
		if *root != nil {
			ts.Error("AlrdyDefinedRoot", name.S, "", name.P, nil)
		}
		ty := ts.getTypeDefinition(name)
		if ty != nil {
			if typ, ok := ty.(*ObjectTypeDefinition); ok {
				*root = typ
			} else {
				ts.Error("WrongRootType", name.S, "", name.P, nil)
			}
		}
	} //FixRoot

	//collectInfos
	// Move extensions into definitions
	for _, d := range ds {
		switch d := d.(type) {
		case *SchemaExtension:
			for _, d2 := range ds {
				if dd, ok := d2.(*SchemaDefinition); ok {
					CopyDirs(&d.Dirs, &dd.Dirs)
					opTypeDefs := make(OperationTypeDefinitions, len(d.OpTypeDefs) + len(dd.OpTypeDefs))
					k := 0
					for _, o := range dd.OpTypeDefs {
						opTypeDefs[k] = o
						k++
					}
					for _, o := range d.OpTypeDefs {
						opTypeDefs[k] = o
						k++
					}
					dd.OpTypeDefs = opTypeDefs
					d.OpTypeDefs = nil
					break
				}
			}
			ts.Error("SchemaExtWoDefinition", "", "", nil, nil)
		case *ScalarTypeExtension:
			dd := ts.getTypeDefinition(d.Name)
			if dd != nil {
				if s, b := dd.(*ScalarTypeDefinition); !b {
					ts.Error("ScalarExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &s.Dirs)
				}
			}
		case *ObjectTypeExtension:
			d2 := ts.getTypeDefinition(d.Name)
			if d2 != nil {
				if dd , b := d2.(*ObjectTypeDefinition); !b {
					ts.Error("ObjectExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &dd.Dirs)
					impInter := make(NamedTypes, len(d.ImpInter) + len(dd.ImpInter))
					k := 0
					for _, i := range dd.ImpInter {
						impInter[k] = i
						k++
					}
					for _, i := range d.ImpInter {
						impInter[k] = i
						k++
					}
					dd.ImpInter = impInter
					d.ImpInter = nil
					fieldsDef := make(FieldsDefinition, len(d.FieldsDef) + len(dd.FieldsDef))
					k = 0
					for _, f := range dd.FieldsDef {
						fieldsDef[k] = f
						k++
					}
					for _, f := range d.FieldsDef {
						fieldsDef[k] = f
						k++
					}
					dd.FieldsDef = fieldsDef
					d.FieldsDef = nil
				}
			}
		case *InterfaceTypeExtension:
			d2 := ts.getTypeDefinition(d.Name)
			if d2 != nil {
				if dd, b := d2.(*InterfaceTypeDefinition); !b {
					ts.Error("InterfaceExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &dd.Dirs)
					fieldsDef := make(FieldsDefinition, len(d.FieldsDef) + len(dd.FieldsDef))
					k := 0
					for _, f := range dd.FieldsDef {
						fieldsDef[k] = f
						k++
					}
					for _, f := range d.FieldsDef {
						fieldsDef[k] = f
						k++
					}
					dd.FieldsDef = fieldsDef
					d.FieldsDef = nil
				}
			}
		case *UnionTypeExtension:
			d2 := ts.getTypeDefinition(d.Name)
			if d2 != nil {
				if dd, b := d2.(*UnionTypeDefinition); !b {
					ts.Error("UnionExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &dd.Dirs)
					unionMTypes := make(NamedTypes, len(d.UnionMTypes) + len(dd.UnionMTypes))
					k := 0
					for _, u := range dd.UnionMTypes {
						unionMTypes[k] = u
						k++
					}
					for _, u := range d.UnionMTypes {
						unionMTypes[k] = u
						k++
					}
					dd.UnionMTypes = unionMTypes
					d.UnionMTypes = nil
				}
			}
		case *EnumTypeExtension:
			d2 := ts.getTypeDefinition(d.Name)
			if d2 != nil {
				if dd, b := d2.(*EnumTypeDefinition); !b {
					ts.Error("EnumExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &dd.Dirs)
					enumValsDef := make(EnumValuesDefinition, len(d.EnumValsDef) + len(dd.EnumValsDef))
					k := 0
					for _, e := range dd.EnumValsDef {
						enumValsDef[k] = e
						k++
					}
					for _, e := range d.EnumValsDef {
						enumValsDef[k] = e
						k++
					}
					dd.EnumValsDef = enumValsDef
					d.EnumValsDef = nil
				}
			}
		case *InputObjectTypeExtension:
			d2 := ts.getTypeDefinition(d.Name)
			if d2 != nil {
				if dd, b := d2.(*InputObjectTypeDefinition); !b {
					ts.Error("InputObjExtWoDef", d.Name.S, "", d.Name.P, nil)
				} else {
					CopyDirs(&d.Dirs, &dd.Dirs)
					inFieldsDef := make(ArgumentsDefinition, len(d.InFieldsDef) + len(dd.InFieldsDef))
					k := 0
					for _, i := range dd.InFieldsDef {
						inFieldsDef[k] = i
						k++
					}
					for _, i := range d.InFieldsDef {
						inFieldsDef[k] = i
						k++
					}
					dd.InFieldsDef = inFieldsDef
					d.InFieldsDef = nil
				}
			}
		default:
		}
	}
	// Fix root objects
	for _, d := range ds {
		switch d := d.(type) {
		case *SchemaDefinition:
			ops := d.OpTypeDefs
			for _, op := range ops {
				FixRoot(&ts.root[op.OpType], op.Type.Name)
			}
		case *ObjectTypeDefinition:
			switch d.Name.S {
			case "Query":
				FixRoot(&ts.root[QueryOp], d.Name)
			case "Mutation":
				FixRoot(&ts.root[MutationOp], d.Name)
			case "Subscription":
				FixRoot(&ts.root[SubscriptionOp], d.Name)
			}
		default:
		}
	}
	if ts.root[QueryOp] == nil {
		ts.Error("QueryRootNotDefined", "", "", nil, nil)
	}
} //collectInfos

func makeNamedType (s string) Type {
	return &NamedType{makeName(s)}
} //makeNamedType

func makeNonNullType (t Type) Type {
	return &NonNullType{t}
} //makeNonNullType

func makeArgument (name string, typ Type) *InputValueDefinition {
	return &InputValueDefinition{Name: makeName(name), Type: typ}
} //makeArgument

func makeFieldDefinition (name string, argsDef ArgumentsDefinition, typ Type) *FieldDefinition {
	return &FieldDefinition{Name: makeName(name), ArgsDef: argsDef, Type: typ}
} //makeFieldDefinition

func (ts *typeSystem) insertDefinition (t *A.Tree, name *StrPtr, d Definition) { // *defMapItem
	ts.verifyNotIntro(name)
	if _, b, _ := t.SearchIns(&defMapItem{NameMapItem{name}, d}); b {
		ts.Error("DupName", name.S, "", name.P, nil)
	}
} //insertDefinition

func (ts *typeSystem) collectNames (ds Definitions) {
	
	InsertBuiltInScalar := func (s string) {
		
		MakeScalarTypeDefinition := func (name *StrPtr) *ScalarTypeDefinition {
			return &ScalarTypeDefinition{TypeDefinitionCommon{Name: name, Dirs: make(Directives, 0)}, nil}
		} //MakeScalarTypeDefinition
	
		//InsertBuiltInScalar
		name := makeName(s)
		ts.markUnknowDefError = false
		td := ts.getTypeDefinition(name)
		ts.markUnknowDefError = true
		if td == nil {
			ts.insertDefinition(ts.typeMap, name, MakeScalarTypeDefinition(name))
		} else if _, b := td.(*ScalarTypeDefinition); !b {
			ts.Error("MisusedName", s, "", name.P, nil)
		}
	} //InsertBuiltInScalar
	
	TestDirective := func (name *StrPtr) bool {
		ts.markUnknowDefError = false
		dd := ts.getDirectiveDefinition(name)
		ts.markUnknowDefError = true
		return dd != nil
	} //TestDirective

	MakeDirective := func (name *StrPtr, argName string, argType Type, argDefVal Value, locs DirectiveLocations) *DirectiveDefinition {
		
		MakeInputValueDefinition := func (name string, typ Type, defVal Value) *InputValueDefinition {
			return &InputValueDefinition{Name: makeName(name), Type: typ, DefVal: defVal}
		} //MakeInputValueDefinition
		
		//MakeDirective
		return &DirectiveDefinition{Name: name, ArgsDef: ArgumentsDefinition{MakeInputValueDefinition(argName, argType, argDefVal)}, DirLocs: locs}
	} //MakeDirective
	
	InsertBuiltInDirective := func (dd *DirectiveDefinition) {
		ts.insertDefinition(ts.dirMap, dd.Name, dd)
	} //InsertBuiltInDirective

	//collectNames
	for _, d := range ds {
		switch d := d.(type) {
		case TypeDefinition:
			ts.insertDefinition(ts.typeMap, d.TypeDefinitionC().Name, d)
			switch d := d.(type) {
			case *ScalarTypeDefinition:
				if d.coercer == nil {
					ts.FixScalarCoercer(d.Name.S, &d.coercer);
					if d.coercer == nil {
						ts.Error("CoercerNotDefined", d.Name.S, "", d.Name.P, nil);
					}
				}
			default:
			}
		case *DirectiveDefinition:
			ts.insertDefinition(ts.dirMap, d.Name, d)
		default:
		}
	}
	InsertBuiltInScalar("Int")
	InsertBuiltInScalar("Float")
	InsertBuiltInScalar("String")
	InsertBuiltInScalar("Boolean")
	InsertBuiltInScalar("ID")
	name := makeName("skip")
	if !TestDirective(name) {
		InsertBuiltInDirective(MakeDirective(name, "if", makeNonNullType(makeNamedType("Boolean")), nil, DirectiveLocations{&ExecutableDirectiveLocation{l_FIELD}, &ExecutableDirectiveLocation{l_FRAGMENT_SPREAD}, &ExecutableDirectiveLocation{l_INLINE_FRAGMENT}}))
	}
	name = makeName("include")
	if !TestDirective(name) {
		InsertBuiltInDirective(MakeDirective(name, "if", makeNonNullType(makeNamedType("Boolean")), nil, DirectiveLocations{&ExecutableDirectiveLocation{l_FIELD}, &ExecutableDirectiveLocation{l_FRAGMENT_SPREAD}, &ExecutableDirectiveLocation{l_INLINE_FRAGMENT}}))
	}
	name = makeName("deprecated")
	if !TestDirective(name) {
		InsertBuiltInDirective(MakeDirective(name, "reason", makeNamedType("String"), MakeStringValue("No longer supported"), DirectiveLocations{&ExecutableDirectiveLocation{l_FIELD_DEFINITION}, &ExecutableDirectiveLocation{l_ENUM_VALUE}}))
	}
} //collectNames

func (ts *typeSystem) validateDefinitions (ds Definitions) {
	for _, d := range ds {
		switch d := d.(type) {
		case *SchemaDefinition:
			ts.validateTypesDirectives(d.Dirs)
			ts.verifyLocation(d.Dirs, l_SCHEMA)
		case TypeDefinition:
			ts.validateTypeDefinition(d)
		case *DirectiveDefinition:
			ts.validateDirectiveDefinition(d)
		default:
		}
	}
} //validateDefinitions

func (ts *typeSystem) conjugateObjectsAndInterfaces (ds Definitions) {

	Insert :=  func (t *A.Tree, name1, name2 *StrPtr) { // *setMapItem
		e, _, _ := t.SearchIns(&setMapItem{NameMapItem{name1}, A.New()})
		e.Val().(*setMapItem).set.SearchIns(&NameMapItem{name2})
	} //Insert

	//conjugateObjectsAndInterfaces
	conj := A.New() // *setMapItem
	for _, d := range ds {
		switch d := d.(type) {
		case *ObjectTypeDefinition:
			tys := d.ImpInter
			for _, ty := range tys {
				Insert(conj, ty.Name, d.Name)
			}
		default:
		}
	}
	e1 := conj.Next(nil)
	for e1 != nil {
		el1 := e1.Val().(*setMapItem)
		di := ts.getTypeDefinition(el1.Name)
		M.Assert(di != nil, 100)
		ddi, ok := di.(*InterfaceTypeDefinition)
		M.Assert(ok, 101)
		ddi.implementedBy = make(NamedTypes, el1.set.NumberOfElems())
		i := 0
		e2 := el1.set.Next(nil)
		for e2 != nil {
			el2 := e2.Val().(*NameMapItem)
			ddi.implementedBy[i] = &NamedType{el2.Name}
			i++
			e2 = el1.set.Next(e2)
		}
		e1 = conj.Next(e1)
	}
} //conjugateObjectsAndInterfaces

func SetDir (td TypeDirectory) {
	Dir = td
} //SetDir

func (dir *typeDirectory) NewTypeSystem (sc Scalarer) TypeSystem {
	ts := new(typeSystem)
	ts.Scalarer = sc
	ts.errorT = errorT{A.New()}
	return ts
} //NewTypeSystem

func (ts *typeSystem) InitTypeSystem (doc *Document) {
	
	AddTypeName := func () {
		
		NewFieldDef := func () *FieldDefinition {
			
			NewStringType := func () Type {
				return &NonNullType{&NamedType{makeName("String")}}
			} //NewStringType
			
			//NewFieldDef
			return &FieldDefinition{Name: makeName("__typename"), Type: NewStringType()}
		} //NewFieldDef
		
		AddField := func (fs *FieldsDefinition, f *FieldDefinition) {
			newFs := make(FieldsDefinition, len(*fs) + 1)
			i := 0
			for _, fi := range *fs {
				newFs[i] = fi
				i++
			}
			newFs[len(newFs) - 1] = f
			*fs = newFs
		} //AddField
	
		//AddTypeName
		f := NewFieldDef()
		for _, d := range doc.Defs {
			switch d := d.(type) {
			case *ObjectTypeDefinition:
				AddField(&d.FieldsDef, f)
			case *InterfaceTypeDefinition:
				AddField(&d.FieldsDef, f)
			default:
			}
		}
	} //AddTypeName
	
	MakeNotListable := func (doc *Document) {
		defs := doc.Defs
		for _, d := range defs {
			switch d := d.(type) {
			case TypeSystemDefinition:
				d.setListableM(false)
			default:
			}
		}
	} //MakeNotListable

	//InitTypeSystem
	M.Assert(doc != nil, 20)
	ts.verifNotIntro = true
	ts.markUnknowDefError = true
	ts.SetErrors(A.New())
	ts.root[QueryOp] = nil
	ts.root[MutationOp] = nil
	ts.root[SubscriptionOp] = nil
	ts.typeMap = A.New()
	ts.dirMap = A.New()
	ts.collectNames (doc.Defs)
	ts.collectInfos (doc.Defs)
	ts.typeMap = A.New()
	ts.dirMap = A.New()
	ts.collectNames(doc.Defs)
	ts.validateDefinitions(doc.Defs)
	if ts.GetErrors().IsEmpty() {
		
		ts.conjugateObjectsAndInterfaces(doc.Defs)
		ts.streamResolvers = A.New()
		ts.abstractTypeResolver = nil
		ts.initialValue = nil
		
		AddTypeName()
		introDoc, _ := ReadGraphQL(F.Join(direc, introName))
		M.Assert(introDoc != nil, 100)
		MakeNotListable(introDoc)
		ts.verifNotIntro = false
		ts.collectNames(introDoc.Defs)
		ts.verifNotIntro = true
		defs := make(Definitions, len(doc.Defs) + len(introDoc.Defs))
		j := 0
		for _, def := range doc.Defs {
			defs[j] = def
			j++
		}
		for _, def := range introDoc.Defs {
			defs[j] = def
			j++
		}
		doc.Defs = defs
		if ts.root[QueryOp] != nil {
			fs := make(FieldsDefinition, len(ts.root[QueryOp].FieldsDef) + 2)
			i := 0;
			for _, f := range ts.root[QueryOp].FieldsDef {
				fs[i] = f
				i++
			}
			fs[len(fs) - 2] = makeFieldDefinition("__schema", make(ArgumentsDefinition, 0), makeNonNullType(makeNamedType("__Schema")))
			fs[len(fs) - 1] = makeFieldDefinition("__type", ArgumentsDefinition{makeArgument("name", makeNonNullType(makeNamedType("String")))}, makeNamedType("__Type"))
			ts.root[QueryOp].FieldsDef = fs
		}
	}
} //InitTypeSystem

func (es *execSystem) collectFragments (ds Definitions) {
	es.fragMap = A.New()
	for _, d := range ds {
		switch d := d.(type) {
		case *FragmentDefinition:
			es.insertDefinition(es.fragMap, d.Name, d)
		default:
		}
	}
} //collectFragments

func (ts *typeSystem) ExecValidate (doc *Document) ExecSystem {
	M.Assert(doc != nil, 20)
	doc.validated = false
	es := &execSystem{*ts, A.New(), A.New()}
	es.SetErrors(A.New())
	es.opMap = es.validateOperationNameUniqueness(doc)
	es.collectFragments(doc.Defs)
	es.validateLoneAnonymousOperation(doc)
	es.validateSubscriptionSingleRootField(doc)
	es.validateScopedTargetFields(doc)
	if !es.validateFragments(doc) {
		return es
	}
	es.validateAllFieldsCanMerge(doc)
	es.validateLeafFieldSelections(doc)
	es.validateArguments(doc)
	es.validateInputValues(doc)
	es.validateDirectives(doc)
	es.validateVariables(doc)
	if es.GetErrors().IsEmpty() {
		doc.validated = true
	}
	return es
} //ExecValidate

func (es *execSystem) Execute (doc *Document, operationName string, variableValues *A.Tree) Response { // *ValMapItem
	M.Assert(doc != nil, 20)
	M.Assert(ExecutableDefinitions(doc), 21)
	// *** For execution of concurrent operations ***
	newES := new(execSystem)
	*newES = *es
	newES.SetErrors(A.New())
	//  ************************************
	if newES.initialValue == nil {
		newES.Error("InitialValueNotFixed", "", "", nil, nil)
	}
	M.Assert(doc.validated, 100)
	if newES.GetErrors().IsEmpty() {
		opName := makeName(operationName)
		return newES.executeRequest(opName, variableValues, newES.initialValue)
	}
	return &InstantResponse{errors: es.GetErrors()}
} //Execute

// ************** /Type System ************

// JSON -> graphQL
func TransVal (value J.Value) Value {
	var v Value
	switch value := value.(type) {
	case *J.Bool:
		v = &BooleanValue{value.Bool}
	case *J.Integer:
		v = &IntValue{value.N}
	case *J.Null:
		v = new(NullValue)
	case *J.Float:
		v = &FloatValue{value.F}
	case *J.String:
		v = MakeStringValue(value.S)
	case *J.JsonVal:
		switch j := value.Json.(type) {
		case *J.Array:
			lv := NewListValue()
			for _, e := range j.Elements {
				lv.Append(TransVal(e))
			}
			v = lv
		case *J.Object:
			ov := newInputObjectValue()
			for _, o := range j.Fields {
				name := o.Name
				M.Assert(name != "", 100)
				ov.insertInputField(makeName(name), TransVal(o.Value))
			}
			v = ov
		}
	}
	return v
} //TransVal

func InsertJsonValue (t *A.Tree, name string, value J.Value) bool { // *ValMapItem
	if name == "" {
		return false
	}
	_, b, _ := t.SearchIns(&ValMapItem{NameMapItem{makeName(name)}, TransVal(value)})
	return !b
} //InsertJsonValue

func (ts *typeSystem) FixInitialValue (ov *OutputObjectValue) {
	ts.initialValue = ov
} //FixInitialValue

// SetLGQL sets the current LinkGQL function and returns the previous one
func SetLGQL (lG LinkGQL) LinkGQL {
	l := lGQL
	lGQL = lG
	return l
} //SetLGQL

func SetRComp (rComp io.Reader) {
	comp = C.NewDirectory(&directory{r: bufio.NewReader(rComp)}).ReadCompiler()
} //SetRComp

func linkFile (name string) io.ReadCloser {
	f, err := os.Open(name); M.Assert(err == nil, err, 100)
	return f
} //linkFile

func init () {
	SetLGQL(linkFile)
	f, err := os.Open(F.Join(direc, compName))
	if err == nil {
		defer f.Close()
		SetRComp(f)
	}
} //init
