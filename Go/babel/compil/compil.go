/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package compil

/* The module BabelCompil (Compil for short) is the online part of the Babel subsystem, a compiler compiler. The module BabelBabel converts a grammar written in a text definition document into tables written into a binary file (.tbl extension). Compil is able to read this .tbl file and to parse texts in accordance with provided grammar rules. There are three objects defined in Compil:
		-	Compiler (a compiler, which represents the content of a .tbl file);
		-	Compilation (a compilation instance, that one must provide with a compiler and several methods (Read for the reading of text, Execution for the implementation of hard functions ---see below---, Map, optional, for translations of messages, and Error, optional, for errors handling);
		-	Object (objects produced by the compilation process: strings extracted from text; results of hard functions; trees of Object, recursively produced by soft functions ---see below).
	Hard and soft functions are defined in grammar definition documents. A hard function is implemented in the Compilation.Execution method and produces an userObj result; a soft function doesn't require any method: it produces a tree (termObj) whose root is the function and whose subtrees are its arguments. TermeObj trees may be used later by hard functions. */

const (
	eOS rune = 0x0	// End of string 
	
	EOF1 rune = 0x0
	EOF2 rune = 0x1a	// End of file = EOF1 or EOF2 
	
	EOL1 rune = 0x0d
	EOL2 rune = 0x0a	// End of line = EOL1, or EOL2, or EOL1 followed by EOL2 
)
	
const (
	StringObj = iota	// Points out an Object containing a terminal string of the parsed text,... 
	UserObj	// ... the result of a hard function,... 
	TermObj	// ... a tree of Object, result of a soft function,... 
	NulObj	// ... or an empty Object, when an syntax error has occurred. 
)
	
const
	tBS = 64
	
const (
	deplaceS = iota	// Shift action 
	reduitS	// Reduce action 
	accepteS	// Accepts action 
	erreurS	// Error 
)
	
const (
	lexFin = iota
	debComment
	finComment
)

type
	
	Anyptr interface {}
	
	
	// Definitions for the lexical analyzer 

type (
	
	tokLex struct {	// A token 
		nom string
		utile,
		valUt bool
	}
	
	toksLexT []tokLex

)

type (
	
	gotoLex interface { // *gotoLexC or *gotoLexT
	}
	
	gotoLexC struct {
		goTo int
		premCar,
		derCar rune
	}
	
	gotoLexT struct {
		goTo,
		transit int
	}
	
	transLex []gotoLex

)

type (
	
	card []int
	
	etatLex struct {
		recon,
		nbTrans,
		nbEps int
		transL transLex
	}
	
	etatsLexT []etatLex

)

// Definitions for the parser 

type (
	
	actionS struct {
		quoi int8
		premTerm,
		derTerm,
		aux int
	}
	
	actSynt []actionS

)

type (
	
	actionSyntT struct {
		nbT int
		actions actSynt
	}
	
	actionsSynt []actionSyntT

)

type (
	
	gtS struct {
		depart,
		arrivee int
	}
	
	gotoS []gtS
	
	gotoSyntT struct {
		nbAtts int
		typsAt card
		nbE int
		gotos gotoS
	}
	
	gotosSyntT []gotoSyntT

)

// Definitions for the semantic analyser 

type (
	
	param struct {
		profD,
		attD int
	}
	
	params []param
	
	actionSem struct {
		sOrH bool
		profG,
		attG,
		fonc,
		nbPars int
		pars params
	}
	
	actionsSem []actionSem
	
	regleSyntT struct{
		longueur,
		nonTerm,
		nbAct int
		act actionsSem
	}
	
	reglesSynt []regleSyntT
	
	recEtatT struct{
		etatDep,
		nTGoto int
	}
	
	recEtats []recEtatT
	
	recTerm struct {
		numTerm,
		nbEtats int
		recEtat recEtats
	}
	
	recTermsT []recTerm

)

// Compiler

type
	
	Compiler struct {
		nbToksLex int
		toksLex toksLexT
		nbEtatsLex,
		profEtatsL int
		etatsLex etatsLexT
		nbEtatsCom,
		profEtatsC int
		etatsCom etatsLexT
		nbEtatsSynt int
		actionSynt actionsSynt
		nbNonTSynt int
		gotoSynt gotosSyntT
		nbRegleSynt int
		regleSynt reglesSynt
		nbRecTerms int
		recTerms recTermsT
	}

// Objects

type (
	
	ObjectsList []*Object	// A list of Object, hard function parameters
	
	suiteObjets struct {
		suivant *suiteObjets
		obj *Object
	}
	
	Object struct {	// An Object, string extracted from the text or result of a hard or soft function
		o objetRefer	// *objetNC, *objetCC, *objetCOChaineObj, *objetCOUtilObj, *objetCOTermeObj, *objetCONulObj
	}
	
	objetRefer interface {
		coords () (lig, col, pos int)
	}
	
	objetRef struct {
		ligO,
		colO,
		posO int
	}
	
	objetNC struct {
		objetRef
		numRSynt,
		numRSem,
		numNC int
		params ObjectsList
		declic *suiteObjets
	}
	
	objetC struct {
		objetRef
	}
	
	objetCC struct {
		objetC
		subst *Object
	}
	
	objetCOer interface {
		objetRefer
		err () bool
	}
		
	objetCO struct {
		objetC
		errO bool
	}
	
	objetCOChaineObj struct {
		objetCO
		valC string
	}
	
	objetCOUtilObj struct {
		objetCO
		numU,
		foncU int
		valU Anyptr
	}
	
	objetCOTermeObj struct {
		objetCO
		numT,
		foncT int
		filsT ObjectsList
	}
	
	objetCONulObj struct {
		objetCO
	}

)

// Compilation

type (
	
	pileT struct{	// Stack 
		suivant *pileT
		etat int
		attrib ObjectsList
	}
	
	ensBS []uint64
	
	ensEtatT struct {
		pEtat card
		sommet int
		ensemble ensBS
	}
	
	ensEtats [2]ensEtatT
	
	ensEtatsTab []ensEtats
	
	Compilationer interface {
		
		// Text reading method: ch is the next character read. When reading after the end of text, ch must return EOF1 or EOF2.
		Read () (ch rune, cLen int)
		
		// Returns the current position in the input stream of chars (origin can have any value).
		Pos () int
		
		// Moves the current position in the input stream of chars to pos (origin must be the same than for Pos). pos may extend until the position which follows the end of the text.
		SetPos (pos int)
		
		// Execution of hard functions: fNum is the index of the hard function, parsNb is the number of input parameters; pars is the list of input parameters; objPos is an Object whose position in text will be the position of res; res is the result of the hard function.
		Execution (fNum, parsNb int, pars ObjectsList) (objPos *Object, res Anyptr, ok bool)
		
		// Lexical and syntactic errors notification: pos position, line line number and col column number of the error; msg is the error message. pos, line and col are numbered from 1.
		Error (pos, line, col int, msg string)
		
		// Maps an index text to a more lengthy one. Used to make explicit or to translate error messages.
		Map (index string) string
	
	}
	
	Compilation struct {	// A compilation
		Compilationer
		compil *Compiler
		ensEtatLex,
		ensEtatCom ensEtatsTab
		bSE,
		yAErreurI,
		yAErreurE,
		yAEuErreur,
		stopCompil,
		arret bool
		etatEOL,
		ligne,
		colonne,
		pos,
		forward int
		cCour rune
		cLen int
		pile *pileT
	}
)

// Directory

type (
	
	Directorier interface {
		ReadInt () int32	// Reads next integer in the binary file built by module BabelBabel
	}
	
	Directory struct {	// Compiler factory
		Directorier
	}

)

func (o *objetRef) coords () (lig, col, pos int) {
	lig = o.ligO
	col = o.colO
	pos = o.posO
	return
}

func (o *objetCO) err () bool {
	return o.errO
}

func NewCompilation (c Compilationer) *Compilation {
	comp := new(Compilation)
	comp.Compilationer = c
	return comp
}

// Beginning of Errors handling 

// Handling of lexical errors 

func (c *Compilation) erreurLex (n, li, co, p int, r ... rune) {
	
	err := func (mes string, r ... rune) {
		if !c.stopCompil {
			mes = c.Map(mes)
			if r != nil {
				mes += " (" + string(r[0]) + ")"
			}
			c.Error(p, li, co, mes)
			c.yAErreurI = true
		}
	}
	
	switch n {
		case 1: err("CForbidden", r ...)
		case 2: err("COpenedComment")
		case 3: err("CNoComment")
	}
}

// Handling of syntactic errors 

func (c *Compilation) erreurSynt (li, co, p int, s string) {
	if !c.stopCompil {
		s0 := c.Map("Cor")
		s1 := c.Map("Cexpected")
		s2 := c.Map("Cread")
		com := c.compil
		aSy := com.actionSynt[c.pile.etat]
		b := aSy.actions[aSy.nbT - 1].quoi != erreurS
		mes := ""
		if s !="" {
			mes = "\"" + s + "\" " + s2 + ": "
		}
		cont := false
		j := 1;
		for i := 0; i <= aSy.nbT - 2; i++ {
			aS := aSy.actions[i]
			if b {
				for k := j; k <= aS.premTerm; k++ {
					if cont {
						mes = mes + " " + s0 + " "
					} else {
						cont = true
					}
					mes = mes + com.toksLex[k - 1].nom
				}
			}
			if aS.quoi != erreurS {
				for k := aS.premTerm; k <= aS.derTerm; k++ {
					if cont {
						mes = mes + " " + s0 + " "
					} else {
						cont = true
					}
					mes = mes + com.toksLex[k].nom
				}
			}
			j = aS.derTerm + 2
		}
		if b {
			for k := j; k <= com.nbToksLex; k++ {
				if cont {
					mes = mes + " " + s0 + " "
				} else {
					cont = true
				}
				mes = mes + com.toksLex[k - 1].nom
			}
		}
		mes = mes + " " + s1
		//mes = strings.Title(mes)
		c.Error(p, li, co, mes)
		c.yAErreurI = true
	}
}

// End of Errors handling 

// Beginning of lexical part 

// DFA which detects ends of lines 
func (c *Compilation) autoEOL () {
	c.pos++
	c.colonne++
	loop:
	for {
		switch c.etatEOL {
			case 0:
				switch c.cCour {
					case EOL2:
							c.etatEOL = 1
					case EOL1:
							c.etatEOL = 2
					default:
				}
				break loop
			case 1:
				c.etatEOL = 0
				c.ligne++
				c.colonne = 1
			case 2:
				if c.cCour == EOL2 {
					c.etatEOL = 1
					break loop
				} else {
					c.etatEOL = 0
					c.ligne++
					c.colonne = 1
				}
		}
	}
}

func (c *Compilation) initLex () {
	c.arret = false
	c.forward = c.Pos()
	c.ligne = 1
	c.colonne = 0
	c.pos = 0
	c.etatEOL = 0
	c.cCour, c.cLen = c.Read()
	c.autoEOL()
}

func in (ens uint64, i int) bool {
	return ens & (1 << uint(i)) != 0
}

func incl (ens *uint64, i int) {
	*ens |= 1 << uint(i)
}

func excl (ens *uint64, i int) {
	*ens &^= 1 << uint(i)
}

// Lexical analyzer 

func avance (c *Compilation, err1 *bool) {
	*err1 = false
	c.arret = c.cCour == EOF1 || c.cCour == EOF2
	c.forward += c.cLen
	if !c.arret {
		c.cCour, c.cLen = c.Read()
		c.autoEOL()
	}
	return
}

func corrige (c *Compilation, lexBegin int, err1 *bool) {
	if !c.arret && c.forward == lexBegin {
		avance(c, err1)
	}
}

func pousseL (e *ensEtatT, etat int) {
	if !in(e.ensemble[etat / tBS], etat % tBS) {
		incl(&e.ensemble[etat / tBS], etat % tBS)
		e.sommet++
		e.pEtat[e.sommet] = etat
	}
}

func valEtat (e *ensEtatT, n int) (etat int, ok bool) {
	if n > e.sommet {
		ok = false
	} else {
		etat = e.pEtat[n]
		ok = true
	}
	return
}

func estDans (e *ensEtatT, etat int) bool {
	return in(e.ensemble[etat / tBS], etat % tBS)
}

func tireL (e *ensEtatT) (etat int, ok bool) {
	if e.sommet < 0 {
		ok = false
	} else {
		etat = e.pEtat[e.sommet]
		e.sommet--
		excl(&e.ensemble[etat / tBS], etat % tBS)
		ok = true
	}
	return
}

func vide (e *ensEtatT) {
	_, ok := tireL(e)
	for ok {
		_, ok = tireL(e)
	}
}

func suit (c *Compilation, ensEtat *ensEtats, depart, prof, position int, ch rune, cL int, stop bool, eL etatsLexT, ensTab ensEtatsTab) bool {
	b := false
	arrete := stop
	etatCour := 0
	pousseL(&ensEtat[etatCour], depart)
	loop:
	for {
		complete(c, &ensEtat[etatCour], prof + 1, position, ch, cL, arrete, eL, ensTab)
		j := 0
		k, ok := valEtat(&ensEtat[etatCour], j)
		for ok {
			if eL[k].recon == 0 {
				b = true
				break loop
			}
			j++
			k, ok = valEtat(&ensEtat[etatCour], j)
		}
		if arrete {
			break loop
		}
		n, ok := tireL(&ensEtat[etatCour])
		for ok {
			i := eL[n].nbEps
			j := eL[n].nbTrans
			for i < j {
				k := (i + j) / 2
				if eL[n].transL[k].(*gotoLexC).derCar < ch {
					i = k + 1
				} else {
					j = k
				}
			}
			if j < eL[n].nbTrans && eL[n].transL[j].(*gotoLexC).premCar <= ch {
				pousseL(&ensEtat[1 - etatCour], eL[n].transL[j].(*gotoLexC).goTo)
			}
			n, ok = tireL(&ensEtat[etatCour])
		}
		etatCour = 1 - etatCour
		if n, ok = valEtat(&ensEtat[etatCour], 0); !ok {
			break loop
		}
		arrete = ch == EOF1 || ch == EOF2
		position += cL
		if !arrete {
			ch, cL = c.Read()
		}
	}
	vide(&ensEtat[etatCour])
	return b
}

func complete (c *Compilation, e *ensEtatT, prof, position int, ch rune, cL int, stop bool, eL etatsLexT, ensTab ensEtatsTab) {
	i := 0
	n, ok := valEtat(e, i)
	for ok {
		for j := 0; j < eL[n].nbEps; j++ {
			if !estDans(e, eL[n].transL[j].(*gotoLexT).goTo) && suit(c, &ensTab[prof], eL[n].transL[j].(*gotoLexT).transit, prof, position, ch, cL, stop, eL, ensTab) {
				if !stop {
					c.SetPos(position + cL)
					c.cCour, c.cLen = ch, cL
				}
				pousseL(e, eL[n].transL[j].(*gotoLexT).goTo)
			}
		}
		i++
		n, ok = valEtat(e, i)
	}
}

func tourne (c *Compilation, nT int, eL etatsLexT, ensEtat *ensEtats, ensTab ensEtatsTab, inC bool, err1 *bool) (tok, li, co, p, lexBegin int, err bool) {
	var (
		chE rune
		cLE int
		arrE bool
		lexEnd,
		pE,
		liE,
		coE,
		eEE int
	)
	etatCour := 0
	loop1:
	for {
		err = false // Est-ce bien la bonne place pour initialiser err ?
		if !inC {
			li = c.ligne
			co = c.colonne
			p = c.pos
		}
		lexBegin = c.forward
		if c.arret {
			tok = lexFin
			lexBegin -= c.cLen
		} else {
			tok = nT
			pousseL(&ensEtat[etatCour], 0)
			loop2:
			for {
				complete(c, &ensEtat[etatCour], 1, c.forward, c.cCour, c.cLen, c.arret, eL, ensTab)
				i := nT
				j := 0
				k, ok := valEtat(&ensEtat[etatCour], j)
				for ok {
					if eL[k].recon < i {
						i = eL[k].recon
					}
					j++
					k, ok = valEtat(&ensEtat[etatCour], j)
				}
				if i < nT {
					tok = i
					lexEnd = c.forward
					chE = c.cCour
					cLE = c.cLen
					arrE = c.arret
					liE = c.ligne
					coE = c.colonne
					pE = c.pos
					eEE = c.etatEOL
				}
				if c.arret {
					break loop2
				}
				n, ok := tireL(&ensEtat[etatCour])
				for ok {
					i := eL[n].nbEps
					j := eL[n].nbTrans
					for i < j {
						k := (i + j) / 2
						if eL[n].transL[k].(*gotoLexC).derCar < c.cCour {
							i = k + 1
						} else {
							j = k
						}
					}
					if j < eL[n].nbTrans && eL[n].transL[j].(*gotoLexC).premCar <= c.cCour {
						pousseL(&ensEtat[1 - etatCour], eL[n].transL[j].(*gotoLexC).goTo)
					}
					n, ok = tireL(&ensEtat[etatCour])
				}
				etatCour = 1 - etatCour
				if _, ok = valEtat(&ensEtat[etatCour], 0); !ok {
					break loop2
				}
				avance(c, err1)
			}
			vide(&ensEtat[etatCour])
			if tok < nT {
				c.forward = lexEnd
				c.arret = arrE
				if !arrE {
					c.SetPos(lexEnd + cLE)
					c.cCour, c.cLen = chE, cLE
				}
				c.ligne = liE
				c.colonne = coE
				c.pos = pE
				c.etatEOL = eEE
			}
		}
		if inC {
			if tok == lexFin {
				c.erreurLex(2, li, co, p)
				err = true
			}
			if tok < nT {
				break loop1
			}
		} else {
			if c.forward == lexBegin {
				if !*err1 {
					c.erreurLex(1, c.ligne, c.colonne, c.pos, c.cCour)
					*err1 = true
				}
				err = true
			}
			switch tok {
				case lexFin:
					 break loop1
				case debComment:
					tok, lexBegin = inComment(c, err1)
					if tok == lexFin {
						break loop1
					}
				case finComment:
					 c.erreurLex(3, li, co, p)
					 err = true
				default:
					 if tok == nT {
						 if !*err1 {
							 c.erreurLex(1, c.ligne, c.colonne, c.pos)
							 *err1 = true
						 }
						 err = true
					 } else if c.compil.toksLex[tok].utile {
						 break loop1
					 }
			}
		}
		corrige(c, lexBegin, err1)
	}
	return
}

func inComment (c *Compilation, err1 *bool) (tok, lexBegin int) {
	const nbToksCom = 3
	profComment := 1
	for {
		tok, _, _, _, lexBegin, _ = tourne(c, nbToksCom, c.compil.etatsCom, &c.ensEtatCom[0], c.ensEtatCom, true, err1)
		if tok == debComment {
			profComment++
		}else if tok == finComment {
			profComment--
		}
		if tok == lexFin || profComment == 0 {
			break
		}
	}
	if tok != lexFin {
		corrige(c, lexBegin, err1)
	}
	return
}

func (c *Compilation) lex () (p, li, co int, valStr string, valUt, err bool, tok int) {
	err1 := false
	tok, li, co, p, lexBegin, err := tourne(c, c.compil.nbToksLex, c.compil.etatsLex, &c.ensEtatLex[0], c.ensEtatLex, false, &err1)
	if c.compil.toksLex[tok].valUt {
		valUt = true
		lVal := c.forward - lexBegin
		c.SetPos(lexBegin)
		valStr = ""
		var (ch rune; cL int)
		for j := 0; j < lVal; j += cL {
			ch, cL = c.Read()
			valStr += string(ch)
		}
		if !c.arret {
			c.SetPos(c.forward + c.cLen)
		}
	} else {
		valUt = false
		valStr = ""
	}
	corrige(c, lexBegin, &err1)
	return
}

// End of lexical part 

// Beginning of syntaxic part 

func (o *Object) subst () *Object {
	oC, ok := o.o.(*objetCC)
	for ok {
		o = oC.subst
		oC, ok = o.o.(*objetCC)
	}
	return o
}

func pousseS (pile **pileT, x int, a ObjectsList) {
	*pile = &pileT{suivant: *pile, etat: x, attrib: a}
}

func tireS (pile **pileT) {
	*pile = (*pile).suivant
}

func videPile (pile **pileT) {
	*pile = nil
}

func (c *Compilation) initSynt () {
	c.pile = &pileT{etat: 0}
}

func initT (p, li, co int, s string, err bool) (attrib ObjectsList) {
	attrib = make(ObjectsList, 1)
	attrib[0] = new(Object)
	attrib[0].o = &objetCOChaineObj{objetCO{objetC{objetRef{ligO: li, colO: co, posO: p}}, err}, s}
	return
}

func initNT (nbA int, typesA card) (a ObjectsList) {
	a = make(ObjectsList, nbA)
	for i := 0; i < nbA; i++ {
		a[i] = new(Object)
		a[i].o = &objetNC{objetRef: objetRef{0, 0, 0}, numNC: typesA[i]}
	}
	return
}

func initErr (nbA int) (a ObjectsList) {
	a = make(ObjectsList, nbA)
	for i := 0; i < nbA; i++ {
		a[i] = new(Object)
		a[i].o = &objetCONulObj{objetCO{objetC: objetC{objetRef{0, 0, 0}}, errO: true}}
	}
	return
}

func agit (c *Compilation, o *Object) {
	loop:
	for {
		var (
			oo *objetNC
			ok bool
		)
		if oo, ok = o.o.(*objetNC); !ok {
			break loop
		}
		aS := c.compil.regleSynt[oo.numRSynt].act[oo.numRSem]
		ok = true
		i := 0
		for ok && i < aS.nbPars {
			_, ok = oo.params[i].subst().o.(*objetNC);
			ok = !ok
			i++
		}
		if ok {
			d := oo.declic
			if aS.sOrH {
				p := oo.params
				oT := new(objetCOTermeObj)
				oT.ligO = oo.ligO
				oT.colO = oo.colO
				oT.posO = oo.posO
				oT.numT = oo.numNC
				oT.foncT = aS.fonc
				oT.filsT = p
				oT.errO = false
				i := 0
				for !oT.errO && i < aS.nbPars {
					oT.errO = p[i].subst().o.(objetCOer).err()
					i++
				}
				o.o = oT
			} else {
				var (
					obj *Object
					a Anyptr
				)
				ok := !c.yAErreurE
				if ok {
					obj, a, ok = c.Execution(aS.fonc, aS.nbPars, oo.params)
				}
				var oR objetRef
				if obj != nil {
					lig, col, pos := obj.o.coords()
					oR = objetRef{ligO: lig, colO: col, posO: pos}
				} else {
					oR = objetRef{ligO: oo.ligO, colO: oo.colO, posO: oo.posO}
				}
				oU := &objetCOUtilObj{objetCO: objetCO{objetC: objetC{oR}, errO: !ok}, numU: oo.numNC, foncU: aS.fonc, valU: a}
				o.o = oU
				c.yAEuErreur = c.yAEuErreur || oU.errO
			}
			if d == nil {
				break loop
			}
			for {
				o = d.obj
				d = d.suivant
				if d == nil {
					break
				}
				agit(c, o)
			}
		} else {
			break loop
		}
	}
}

func (c *Compilation) execute (nSynt int, gauche ObjectsList) {
	
	trouve := func (prof int) ObjectsList {
		if prof == 0 {
			return gauche
		}
		p := c.pile
		for i := 2; i <= prof; i++ {
			p = p.suivant
		}
		return p.attrib
	}

	extrait := func (prof, att int) *Object {
		return trouve(prof)[att - 1].subst()
	}

	insereDeclic := func (decl **suiteObjets, o *Object) {
		s := *decl
		for (s != nil) && (s.obj != o) {
			s = s.suivant
		}
		if s == nil {
			*decl = &suiteObjets{suivant: *decl, obj: o}
		}
	}

	rS := c.compil.regleSynt[nSynt]
	for i := 0 ; i < rS.nbAct; i++ {
		aS := rS.act[i]
		o := extrait(aS.profG, aS.attG)
		if oo, ok := o.o.(*objetNC); ok { // Dans le cas contraire, o.o IS ObjetCONulObj 
			if aS.sOrH && aS.fonc == 0 {
				p := aS.pars[0]
				oC := &objetCC{subst: extrait(p.profD, p.attD)}
				o.o = oC
				e := oo.declic
				var oN *objetNC
				oN, ok = oC.subst.o.(*objetNC)
				if !ok {
					for e != nil {
						agit(c, e.obj)
						e = e.suivant
					}
				} else {
					for e != nil {
						insereDeclic(&oN.declic, e.obj)
						e = e.suivant
					}
				}
			} else {
				oo.numRSynt = nSynt
				oo.numRSem = i
				if aS.nbPars > 0 {
					oo.params = make(ObjectsList, aS.nbPars)
					for j := 0; j < aS.nbPars; j++ {
						p := aS.pars[j]
						oS := extrait(p.profD, p.attD)
						oo.params[j] = oS
						if oN, ok := oS.o.(*objetNC); ok {
							insereDeclic(&oN.declic, o)
						}
					}
				}
				agit(c, o)
			}
		}
	}
}

// Parser 

func (c *Compilation) synt () {
	var (
		
		com *Compiler
		p,
		li,
		co,
		tok int
		s string
		sUt,
		err bool
		
		recupere = func () {	// Errors recovery 
			if com.nbRecTerms == 0 || tok == lexFin {
				videPile(&c.pile)
			} else {
				loop:
				for {
					p, li, co, s, _, err, tok = c.lex()
					i := 0
					for i < com.nbRecTerms && com.recTerms[i].numTerm != tok {
						i++
					}
					if i < com.nbRecTerms {
						rT := com.recTerms[i]
						q := c.pile
						for q != nil {
							for j := 0; j < rT.nbEtats; j++ {
								rE := rT.recEtat[j]
								if q.etat == rE.etatDep {
									for c.pile != q {
										tireS(&c.pile)
									}
									gS := com.gotoSynt[rE.nTGoto]
									k := 0
									ll := int(gS.nbE) - 1
									for k < ll {
										m := (k + ll) / 2
										if gS.gotos[m].depart < c.pile.etat {
											k = m + 1
										} else {
											ll = m
										}
									}
									if gS.gotos[ll].depart != c.pile.etat {
										ll = int(gS.nbE) - 1
									}
									var a ObjectsList
									if gS.nbAtts == 0 {
										a = nil
									} else {
										a = initErr(gS.nbAtts)
									}
									pousseS(&c.pile, gS.gotos[ll].arrivee, a)
									break loop
								}
							}
							q = q.suivant
						}
					} else if tok == lexFin {
						videPile(&c.pile)
						break loop
					}
				}
			}
		}
	)

	c.initSynt()
	com = c.compil
	p, li, co, s, sUt, err, tok = c.lex()
	loop:
	for {
		aSy := com.actionSynt[c.pile.etat]
		i := 0
		j := int(aSy.nbT) - 1
		for i < j {
			k := (i + j) / 2
			if aSy.actions[k].derTerm < tok {
				i = k + 1
			} else {
				j = k
			}
		}
		if aSy.actions[j].premTerm > tok {
			j = int(aSy.nbT) - 1
		}
		aS := aSy.actions[j]
		switch aS.quoi {
			case deplaceS:
				var
					a ObjectsList
				if !sUt {
					a = nil
				} else {
					a = initT(p, li, co, s, err)
				}
				pousseS(&c.pile, aS.aux, a)
				p, li, co, s, sUt, err, tok = c.lex()
			case reduitS:
				rS := com.regleSynt[aS.aux]
				gS := com.gotoSynt[rS.nonTerm]
				var
					a ObjectsList
				if gS.nbAtts == 0 {
					a = nil
				} else {
					a = initNT(gS.nbAtts, gS.typsAt)
				}
				if !(c.yAErreurI && c.bSE || c.yAErreurE) {
					c.execute(aS.aux, a)
				}
				for i := 1; i <= rS.longueur; i++ {
					tireS(&c.pile)
				}
				i = 0
				m := int(gS.nbE) - 1
				for i < m {
					k := (i + m) / 2
					if gS.gotos[k].depart < c.pile.etat {
						i = k + 1
					} else {
						m = k
					}
				}
				if gS.gotos[m].depart != c.pile.etat {
					m = int(gS.nbE) - 1
				}
				pousseS(&c.pile, gS.gotos[m].arrivee, a)
			case accepteS:
				videPile(&c.pile)
				break loop
			case erreurS:
				c.erreurSynt(li, co, p, s)
				if !c.stopCompil {
					recupere()
					if c.pile == nil {
						break loop
					}
				}
		}
		if c.stopCompil {
			videPile(&c.pile)
			break loop
		}
	}
}

// End of syntaxic part

func NewDirectory (d Directorier) *Directory {
	dir := new(Directory)
	dir.Directorier = d
	return dir
}

func (d *Directory) readInt () int {
	return int(d.ReadInt())
}

func (d *Directory) readByte () int8 {
	return int8(d.ReadInt())
}

func (d *Directory) readBool () bool {
	i := d.ReadInt()
	if !(i == 0 || i == 1) {panic(0)}
	return i == 1
}

func (d *Directory) readChar () rune {
	return rune(d.ReadInt())
}

func (d *Directory) readString () (s string) {
	s = ""
	c := d.readChar()
	for c != eOS {
		s += string(c)
		c = d.readChar()
	}
	return
}
	
// Reads a compiler, i.e. the binary file built by module BabelBabel. 

func (d *Directory) ReadCompiler () *Compiler {
	
	lisEtatsRedLex := func  (nE int) (eL etatsLexT) {
		eL = make(etatsLexT, nE)
		for i := 0; i < nE; i++ {
			eL[i].recon = d.readInt()
			eL[i].nbTrans = d.readInt()
			eL[i].nbEps = d.readInt()
		}
		for i := 0; i < nE; i++ {
			if eL[i].nbTrans > 0 {
				t := make(transLex, eL[i].nbTrans)
				for j := 0; j < eL[i].nbEps; j++ {
					gT := new(gotoLexT)
					gT.goTo = d.readInt()
					gT.transit = d.readInt()
					t[j] = gT
				}
				for j := eL[i].nbEps; j < eL[i].nbTrans; j++ {
					gC := new(gotoLexC)
					gC.goTo = d.readInt()
					gC.premCar = d.readChar()
					gC.derCar = d.readChar()
					if !(gC.premCar <= gC.derCar) {panic(2)}
					t[j] = gC
				}
				eL[i].transL = t
			}
		}
		return
	}
	
	c := new(Compiler)
	c.nbToksLex = d.readInt()
	c.nbEtatsLex = d.readInt()
	c.profEtatsL = d.readInt()
	c.nbEtatsCom = d.readInt()
	c.profEtatsC = d.readInt()
	c.nbEtatsSynt = d.readInt()
	c.nbNonTSynt = d.readInt()
	c.nbRegleSynt = d.readInt()
	c.nbRecTerms = d.readInt()
	t := make(toksLexT, c.nbToksLex)
	for i := 0; i < c.nbToksLex; i++ {
		t[i].utile = d.readBool()
		t[i].valUt = d.readBool()
	}
	for i := 0; i < c.nbToksLex; i++ {
		k := d.readInt()
		if k != 0 {
			t[i].nom = d.readString()
		} else {
			t[i].nom = ""
		}
	}
	c.toksLex = t
	c.etatsLex = lisEtatsRedLex(c.nbEtatsLex)
	c.etatsCom = lisEtatsRedLex(c.nbEtatsCom)
	c.actionSynt = make(actionsSynt, c.nbEtatsSynt)
	for i := 0; i < c.nbEtatsSynt; i++ {
		c.actionSynt[i].nbT = d.readInt()
	}
	for i := 0; i < c.nbEtatsSynt; i++ {
		a := make(actSynt, c.actionSynt[i].nbT)
		for j := 0; j < c.actionSynt[i].nbT; j++ {
			a[j].quoi = d.readByte()
			a[j].premTerm = d.readInt()
			a[j].derTerm = d.readInt()
			a[j].aux = d.readInt()
		}
		c.actionSynt[i].actions = a
	}
	c.gotoSynt = make(gotosSyntT, c.nbNonTSynt)
	for i := 0; i < c.nbNonTSynt; i++ {
		c.gotoSynt[i].nbAtts = d.readInt()
		c.gotoSynt[i].nbE = d.readInt()
	}
	for i := 0; i < c.nbNonTSynt; i++ {
		if c.gotoSynt[i].nbAtts > 0 {
			ca := make(card, c.gotoSynt[i].nbAtts)
			for j := 0; j < c.gotoSynt[i].nbAtts; j++ {
				ca[j] = d.readInt()
			}
			c.gotoSynt[i].typsAt = ca
		}
		if c.gotoSynt[i].nbE > 0 {
			g := make(gotoS, c.gotoSynt[i].nbE)
			for j := 0; j < c.gotoSynt[i].nbE; j++ {
				g[j].depart = d.readInt()
				g[j].arrivee = d.readInt()
			}
			c.gotoSynt[i].gotos = g
		}
	}
	c.regleSynt = make(reglesSynt, c.nbRegleSynt)
	for i := 0; i < c.nbRegleSynt; i++ {
		c.regleSynt[i].longueur = d.readInt()
		c.regleSynt[i].nonTerm = d.readInt()
		c.regleSynt[i].nbAct = d.readInt()
	}
	for i := 0; i < c.nbRegleSynt; i++ {
		if c.regleSynt[i].nbAct > 0 {
			aS := make(actionsSem, c.regleSynt[i].nbAct)
			for j := 0; j < c.regleSynt[i].nbAct; j++ {
				aS[j].sOrH = d.readBool()
				aS[j].profG = d.readInt()
				aS[j].attG = d.readInt()
				aS[j].fonc = d.readInt()
				aS[j].nbPars = d.readInt()
			}
			for j := 0; j < c.regleSynt[i].nbAct; j++ {
				if aS[j].nbPars > 0 {
					p := make(params, aS[j].nbPars)
					for m := 0; m < aS[j].nbPars; m++ {
						p[m].profD = d.readInt()
						p[m].attD = d.readInt()
					}
					aS[j].pars = p
				}
			}
			c.regleSynt[i].act = aS
		}
	}
	if c.nbRecTerms > 0 {
		c.recTerms = make(recTermsT, c.nbRecTerms)
		for i := 0; i < c.nbRecTerms; i++ {
			c.recTerms[i].numTerm = d.readInt()
			c.recTerms[i].nbEtats = d.readInt()
		}
		for i := 0; i < c.nbRecTerms; i++ {
			r := make(recEtats, c.recTerms[i].nbEtats)
			for j := 0; j < c.recTerms[i].nbEtats; j++ {
				r[j].etatDep = d.readInt()
				r[j].nTGoto = d.readInt()
			}
			c.recTerms[i].recEtat = r
		}
	}
	
	return c
}

func (c *Compilation) Compile (co *Compiler, lockIfError bool) bool {
	
	// Start of a compilation co is the compiler used (must be loaded before) lockIfError: if true, inhibits the later calls of the Execution method when a lexical or syntactic error has occurred the boolean result indicates whether all is ok or not. 
	
	initEns := func (nbEtats, prof int) (e ensEtatsTab) {
		nbBS := (nbEtats - 1) / tBS + 1
		e = make(ensEtatsTab, prof)
		for j := 0; j < prof; j++ {
			for b := 0; b <= 1; b++ {
				if j == prof - 1 {
					e[j][b].pEtat = make(card, 1)
				} else {
					e[j][b].pEtat = make(card, nbEtats)
				}
				e[j][b].sommet = -1
				e[j][b].ensemble = make(ensBS, nbBS)
				for i := 0; i < nbBS; i++ {
					e[j][b].ensemble[i] = 0
				}
			}
		}
		return
	}

	if !(co != nil) {panic(3)}
	c.compil = co
	c.ensEtatLex = initEns(co.nbEtatsLex, co.profEtatsL)
	c.ensEtatCom = initEns(co.nbEtatsCom, co.profEtatsC)
	c.bSE = lockIfError
	c.yAErreurI = false
	c.yAErreurE = false
	c.yAEuErreur = false
	c.stopCompil = false
	c.initLex()
	c.synt()
	return !(c.yAErreurI || c.yAEuErreur)
}
	
// Extracts an Object of an ObjectsList: l: the list num: the index in the list (1 first). num must be strictly positive and not greater than the number of Objects in l. 

func Parameter (l ObjectsList, num int) *Object {
	if !(l != nil) {panic(4)}
	if !(num > 0) {panic(5)}
	if !(num <= len(l)) {panic(6)}
	return l[num - 1].subst()
}

// Gives the type of content of an Object: stringObj, userObj, termObj or nulObj. 

func (o *Object) ObjType () (t int8) {
	switch o.o.(type) {
		case *objetCOChaineObj:
			return StringObj
		case *objetCOUtilObj:
			return UserObj
		case *objetCOTermeObj:
			return TermObj
		case *objetCONulObj:
			return NulObj
	}
	return
}

// Gives the number of an userObj or termObj Object. 

func (o *Object) ObjNum () int {
	switch oo := o.o.(type) {
		case *objetCOUtilObj:
			return oo.numU
		case *objetCOTermeObj:
			return oo.numT
		default:
			panic(7)
	}
}

// Gives the number of the hard or soft function which created the userObj or termObj Object.

func (o *Object) ObjFunc () int {
	switch oo := o.o.(type) {
		case *objetCOUtilObj:
			return oo.foncU
		case *objetCOTermeObj:
			return oo.foncT
		default:
			panic(8)
	}
}

// Gives the length of the string contained in a stringObj Object. 

func (o *Object) ObjStringLen () int {
	oo, str := o.o.(*objetCOChaineObj)
	if !str {panic(9)}
	return len(oo.valC)
}

// Returns, in value, the string contained in a stringObj Object or, else, the value parameter is not changed. 

func (o *Object) ObjString () string {
	oo, str := o.o.(*objetCOChaineObj)
	if !str {panic(10)}
	return oo.valC
}

// Gives the data contained in an userObj Object. 

func (o *Object) ObjUser () Anyptr {
	oo, user := o.o.(*objetCOUtilObj)
	if !user {panic(11)}
	return oo.valU
}

// Gives the number of subtrees of a termObj Object. 

func (o *Object) ObjTermSonsNb () int {
	oo, term := o.o.(*objetCOTermeObj)
	if !term {panic(12)}
	return len(oo.filsT)
}

// Gives a subtree of a termObj Object: sonNum is the index of the subtree (1 first). sonNum must be strictly positive and not greater than the number of subtrees of o. 

func (o *Object) ObjTermSon (sonNum int) *Object {
	oo, term := o.o.(*objetCOTermeObj)
	if !term {panic(13)}
	if !(sonNum > 0) {panic(14)}
	if !(sonNum <= len(oo.filsT)) {panic(15)}
	return oo.filsT[sonNum - 1].subst()
}

// During a compilation, modifies the value of the lockIfError parameter of the Compile method. 

func (c *Compilation) LockIfError (lock bool) {
	c.bSE = lock
}

// Activates a serious semantic error: hard functions are no more called.

func (c *Compilation) Lock () {
	c.yAErreurE = true
	c.yAEuErreur = true
}

// Stops compilation at once. 

func (c *Compilation) StopCompil () {
	c.Lock()
	c.stopCompil = true
}

// Tests whether o has the type nulObj, or the type userObj with error, or the type termObj with error or with a descendant in error. 

func (o *Object) ErrorIn () bool {
	switch oo := o.o.(type) {
		case *objetCONulObj:
			return true
		case objetCOer:
			return oo.err()
		default:
			panic(16)
	}
}

// Position, numbered from 1, in the parsed text, of a stringObj or userObj Object. 

func (o *Object) Position () int {
	switch oo := o.o.(type) {
		case *objetCOChaineObj:
			return oo.posO
		case *objetCOUtilObj:
			return oo.posO
		default:
			panic(17)
	}
}

// Line number, numbered from 1, in the parsed text, of a stringObj or userObj Object. 

func (o *Object) Line () int {
	switch oo := o.o.(type) {
		case *objetCOChaineObj:
			return oo.ligO
		case *objetCOUtilObj:
			return oo.ligO
		default:
			panic(18)
	}
}

// Column number, numbered from 1, in the parsed text, of a stringObj or userObj Object. 

func (o *Object) Column () int {
	switch oo := o.o.(type) {
		case *objetCOChaineObj:
			return oo.colO
		case *objetCOUtilObj:
			return oo.colO
		default:
			panic(19)
	}
}

