/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package static

var (
	
	errEn = `STRINGS

GQL_ArgNotRequiredByInterface	The argument ^0 is not required by the interface
GQL_AbsTypeResolverAlrdyDefined	The abstract type resolver is already defined
GQL_AbsTypeNotResolved	The abstract type ^0 has not been resolved
GQL_AbsTypeResolverNotDefined	The abstract type resolver is not defined
GQL_CycleThrough	Cycle through ^0
GQL_DefValIncoercibleToTypeIn	In ^0, the default value is incoercible to its type
GQL_DirectiveInWrongLocation	The directive ^0 is in a wrong location
GQL_DupArgName	^0 is a duplicated argument name
GQL_DupDirName	^0 is a duplicated directive name
GQL_DupEnumValue	^0 is a duplicated enum value
GQL_DupFieldName	^0 is a duplicated field name
GQL_DupInterface	^0 is a duplicated interface name
GQL_DupName	^0 is a duplicated name
GQL_DupTypeName	^0 is a duplicated type name
GQL_DupVarName	^0 is a duplicated variable name
GQL_EnumExtWoDef	^0: enum extension without definition
GQL_FieldNotAvailableFor	^0: this field is not available for the type ^1
GQL_FieldsCantMerge	Fields ^0 and ^1 can't merge
GQL_FragmentNotUsed	Fragment ^0 is not used
GQL_FragSpreadImpossible	Type ^0 is impossible here
GQL_IncorrectArgsNb	Incorrect number of arguments in ^0
GQL_IncorrectArgumentIn	^0: incorrect argument of ^1
GQL_IncorrectFragmentType	^0: incorrect fragment type
GQL_IncorrectReturnTypeOf	^0: incorrect return type of ^1
GQL_IncorrectValueType	^0: incorrect value type
GQL_IncorrectVarValue	^0: incorrect variable value
GQL_InitialValueAlrdyFixed	The initial value was already fixed
GQL_InitialValueNotFixed	The initial value was never fixed
GQL_InputObjExtWoDef	^0: input object extension without definition
GQL_IntOutOfRange	Int out of range
GQL_IntroName	^0 begins with __
GQL_IsNotFieldFor	^0 is not a field of ^1
GQL_IsNotInputFieldFor	^0 is not an input argument of ^1
GQL_MisusedName	^0: misused predeclared name
GQL_NotAnInputObject	^0 is not an input object
GQL_NoArgForNonNullType	No argument for non-null ^0 argument in ^1
GQL_NoArgumentIn	No argument in ^0
GQL_NonBoolValueForBoolType	Non Boolean value for Boolean type
GQL_NonEnumValueForEnumType	Non Enum value for ^0 Enum type
GQL_NonFloatOrIntValueForFloatType	Non Float or Int value for Float type
GQL_NonIntValueForIntType	Non Int value for Int type
GQL_NonListValueForListType	Non list value for ^0 list type
GQL_NotObjectTypeInUnion	In union ^1, ^0 is not an object type
GQL_NonObjValForInputObjType	Non Object value for Input Object type
GQL_NonObjValForObjType	Non Object value for Object type
GQL_NonScalarValueForScalarType	Non scalar value for ^0 scalar type
GQL_NonStringValueForIDType	Non String value for ID type
GQL_NonStringValueForStringType	Non String value for String type
GQL_NoSubscriptionRoot	^0: the subscription root is not defined
GQL_NoSubselOfObjInterOrUnion	Missing subselection of ^0 (OBJECT, INTERFACE or UNION)
GQL_NotDefinedInScope	^0 is not defined in scope
GQL_NotDefinedOp	The operation to be executed is not defined
GQL_NotImplementedIn	The field ^0 is not implemented in the interface ^1
GQL_NotInputField	^0 is not an input field
GQL_NotInputType	^0 has not an input type
GQL_NotLoneAnonOp	Anonymous operation is not alone
GQL_NotOutputType	^0 has not an output type
GQL_NotSingleSubscriptRoot	The subscription root ^0 is not single
GQL_NotSubtype	The type of ^0 is not a subtype of ^1
GQL_NotUniqueOperation	The operation ^0 is not unique
GQL_NoTypeSystemInDoc	No type system in the definitions document
GQL_NoValueForVar	^0: variable with no value
GQL_NullArgForNonNullType	Null argument ^0 for non-null type in ^1
GQL_NullValueWithNonNullType	Null value with non-null type
GQL_NullVarWithNonNullType	^0: Null variable with non-null type
GQL_ObjectExtWoDef	^0: object extension without definition
GQL_QueryRootNotDefined	The query root is not defined
GQL_ResolverAlrdyDefined	The resolver for ^1 in ^0 is already defined
GQL_ResolverNotDefined	The resolver for ^1 in ^0 is not defined
GQL_RootAlrdyDefined	The root ^0 is already defined
GQL_RootNotDefinedFor	The root for ^0 operation is not defined
GQL_ScalarCoercerAlrdyDefined	The scalar coercer for ^0 is already defined
GQL_ScalarExtWoDef	^0: scalar extension without definition
GQL_SchemaExtWoDefinition	Schema extension without definition
GQL_StreamResolverAlrdyDefined	The stream resolver for ^0 is already defined
GQL_StreamResolverNotDefined	The stream resolver for ^0 is not defined
GQL_SubselectionOfScalarOrEnum	Incorrect subselection of ^0 (SCALAR or ENUM)
GQL_UnableToCoerce	Unable to coerce ^0 to its type
GQL_UnionExtWoDef	^0: union extension without definition
GQL_UnknownArgument	^0: Unknown argument
GQL_UnknownDirective	^0: Unknown directive
GQL_UnknownFragment	^0: Unknown fragment
GQL_UnknownObject	^0: Unknown object type
GQL_UnknownOperation	^0: Unknown operation
GQL_UnknownScalarType	^0: Unknown scalar type
GQL_UnknownType	^0: Unknown type
GQL_UnknownVariable	$^0: Unknown variable in ^1
GQL_VariableNotUsed	$^0: this variable is not used
GQL_VarUsageNotAllowed	Incorrect usage of $^0
GQL_WrongArgumentType	^0 has a wrong argument type
GQL_WrongRootType	The root ^0 is not an object type
`

)
