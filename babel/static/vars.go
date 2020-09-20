/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package static

// Package strMapping implements a mapping between keys and longer strings.

var (
	
	strEn = `STRINGS

AAlreadyDef	Already defined identifier.
AIncreasing	It must be in increasing order.
AUnknownIdent	Unknown identifier.
AOneOnly	One and only one character.
ANonTTwiceDef	Twice defined nonterminal.
AUnknownTerm	Unknown terminal.
AWrong	Wrong number.
AUnknownAtt	Unknown attribute.
ANoAttrib	Terminals have no attribute.
AMandatoryAtt	Mandatory attribute for a nonterminal.
AUnknownNonT	Unknown nonterminal.
AHFuncTwiceDef	Twice defined hard function.
ASFuncTwiceDef	Twice defined soft function.
AUnknownFunc	Unknown function.
ANullFuncNumber	Null function number.
AAttTwiceDef	Twice defined attribute.
ATwoBig	Too big number.
AUnknownErr	Unknown error.

AnError	*** ERROR! ***
AnWarning	*** WARNING! ***
AnRemark	*** REMARK! ***

BNoCreate	Cannot create tables file.
BOk	OK.
BRem	Remarks.
BWarning	Warnings.
BError	Errors.
BBabel	Babel:
BNotFound	Not found.
BSee	See

CForbidden	Unexpected character.
COpenedComment	Opened comment at the end of text.
CNoComment	Unexpected end of comment.
Cor	or
Cexpected	expected.
Cread	read

ENothing	nothing
Eunnamed	unnamed
Eused	used
Eunused	unused
EusedVal	used value
EunusedVal	unused value
Ecopy	copy function
Esoft	soft function
Ehard	hard function

EBCircImportWith	Circular Imports with ^0
EBNoAxiom	No Axiom
EBNotDefined	Not Defined: ^0

ILine	Line
ICol	Column
lPos	Position

LEOF	end of text

MCompile	Compile
MCompile_list	Compile List
MFind	Find position...
MInstall	Install tbl converter
MTest	Test grammar...
MAllTest	Test grammar on all files...

SNoRule	The nonterminal ^0 is defined by no rule.
SSyntAndHer	The attribute ^0 of nonterminal ^1 is both synthesized and inherited.
SNeverCalc	The attribute ^0 of nonterminal ^1 is never calculated.
SNeverUsedAtt	The attribute ^0 of nonterminal ^1 is never used.
SNoCalc	An instance of the ^0 attribute ^1 of nonterminal ^2 is not calculated in rule:
SLocSynt	locally synthesized
SGlobSynt	globally synthesized
SLocHer	locally inherited
SGlobHer	globally inherited
SSeveralCalc	An instance of the attribute ^0 of nonterminal ^1 is calculated several times in rule:
SCyclicAtt	Calculus of attributes is cyclic. Indeed:
SAttNonT	The attribute ^0 of nonterminal ^1
SDerives	derives from the attribute ^0 of nonterminal ^1
Swhich	, which
Sor	or
SHaveRead	When you have read:
SIfRead	if you read:
SMustRead	should you read:
SMustUnderstand	should you understand:
SFirst	The first solution has been selected.
SNeverUsedNonT	The nonterminal ^0 is never used.
SResumeAfter	If a syntactic error happens when reading the marked nonterminal ^0, compilation resumes only after reading ^1
SorAfter	, or ^0

WBrowse	Browse
WDir	Directories:
WFilesInError	Files in Error
WFind	Find
WFirstLevel	First Level
WGrammar	Grammar:
WReset	Reset Grammar
WRoot	Root:
WSndLevel	Second Level
WTest	Test
`
	strFr = `STRINGS

AAlreadyDef	Identificateur déjà défini.
AIncreasing	L'ordre doit être croissant.
AUnknownIdent	Identificateur inconnu.
AOneOnly	Un et un seul caractère.
ANonTTwiceDef	Non-terminal défini deux fois.
AUnknownTerm	Terminal inconnu.
AWrong	Numéro incorrect.
AUnknownAtt	Attribut inconnu.
ANoAttrib	Les terminaux n'ont pas d'attributs.
AMandatoryAtt	Attribut obligatoire pour un non-terminal.
AUnknownNonT	Non-terminal inconnu.
AHFuncTwiceDef	Fonction dure définie deux fois.
ASFuncTwiceDef	Fonction douce définie deux fois.
AUnknownFunc	Fonction inconnue.
ANullFuncNumber	Numéro de fonction nul.
AAttTwiceDef	Attribut défini deux fois.
ATwoBig	Nombre trop grand.
AUnknownErr	Erreur inconnue.

AnError	*** ERREUR ! ***
AnWarning	*** ATTENTION ! ***
AnRemark	*** REMARQUE ! ***

BNoCreate	Impossible de créer le fichier tables.
BOk	OK.
BRem	Remarques.
BWarning	Attention.
BError	Erreurs.
BBabel	Babel :
BNotFound	N'existe pas.
BSee	Voir

CForbidden	Caractère interdit ici.
COpenedComment	Commentaire ouvert à la fin du texte.
CNoComment	Fin de commentaire inattendu.
Cor	ou
Cexpected	attendu(e).
Cread	lu

ENothing	rien
Eunnamed	sans nom
Eused	utile
Eunused	inutile
EusedVal	valeur utilisée
EunusedVal	valeur non utilisée
Ecopy	fonction de copie
Esoft	fonction douce
Ehard	fonction dure

EBCircImportWith	Importation circulaire avec ^0
EBNoAxiom	Il n'y a pas d'axiome
EBNotDefined	Non défini : ^0

ILine	Ligne
ICol	Colonne
lPos	Position

LEOF	fin de texte

MCompile	Compiler
MCompile_list	Compiler la liste
MFind	Chercher une position...
MInstall	Installer le convertisseur de tbl
MTest	Tester une grammaire...
MAllTest	Tester une grammaire sur tous fichiers...

SNoRule	Le non-terminal ^0  n'est défini par aucune règle.
SSyntAndHer	L'attribut ^0 du non-terminal ^1 est à la fois synthétisé et hérité.
SNeverCalc	L'attribut ^0 du non-terminal ^1 n'est jamais calculé.
SNeverUsedAtt	L'attribut ^0 du non-terminal ^1 n'est jamais utilisé.
SNoCalc	Une occurrence de l'attribut ^0 ^1 du non-terminal ^2 n'est pas calculée dans la règle :
SLocSynt	synthétisé local
SGlobSynt	synthétisé global
SLocHer	hérité local
SGlobHer	hérité global
SSeveralCalc	Une occurrence de l'attribut ^0 du non-terminal ^1 est calculée plusieurs fois dans la règle :
SCyclicAtt	Le calcul des attributs est cyclique. En effet :
SAttNonT	L'attribut ^0 du non-terminal ^1
SDerives	dépend de l'attribut ^0 du non-terminal ^1
Swhich	, lequel
Sor	ou
SHaveRead	Lorsqu'on a lu :
SIfRead	si on lit :
SMustRead	doit-on avoir lu :
SMustUnderstand	doit-on avoir compris :
SFirst	C'est la première solution qui est retenue.
SNeverUsedNonT	Le non-terminal ^0 n'est jamais utilisé.
SResumeAfter	Si une erreur syntaxique se produit lors de la lecture du non-terminal marqué ^0, la compilation ne reprend qu'après la lecture de ^1
SorAfter	, ou de ^0

WBrowse	Parcourir...
WDir	Répertoires :
WFilesInError	Fichiers en erreur
WFind	Chercher
WFirstLevel	Premier niveau
WGrammar	Grammaire :
WReset	Mise à jour
WRoot	Racine :
WSndLevel	Second niveau
WTest	Test
`

)
