/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelAnaSem is part of the Babel subsystem, a compiler compiler. BabelAnaSem interprets the definition document and uses modules BabelLexique and BabelSyntaxe to build the tbl file

import (
	
	C	"babel/compil"
	M	"util/misc"
	S	"util/sets"

)

const (
	
	trueB = 3
	consB = 13

)

type (
	
	pCard struct {
		i int
	}
	
	pCarEns struct {
		c *S.Set
	}
	
	pCardBool struct {
		c int
		b bool
	 }
	
	pChaine struct {
		s string
	}
	
	sem struct {
		*C.Compilation
		ls *syntaxical
		nom string
		yAPasErreur bool
	}

)

func (s *sem) erreurSem (numErr int, o *C.Object) {
	p := o.Position()
	li := o.Line()
	co := o.Column()
	var ind string
	switch numErr {
	case 0:
		ind = "AAlreadyDef"
	case 1:
		ind = "AIncreasing"
	case 2:
		ind = "AUnknownIdent"
	case 3:
		ind = "AOneOnly"
	case 4:
		ind = "ANonTTwiceDef"
	case 5:
		ind = "AUnknownTerm"
	case 6:
		ind = "AWrong"
	case 7:
		ind = "AUnknownAtt"
	case 8:
		ind = "ANoAttrib"
	case 9:
		ind = "AMandatoryAtt"
	case 10:
		ind = "AUnknownNonT"
	case 11:
		ind = "AHFuncTwiceDef"
	case 12:
		ind = "ASFuncTwiceDef"
	case 13:
		ind = "AUnknownFunc"
	case 14:
		ind = "ANullFuncNumber"
	case 15:
		ind = "AAttTwiceDef"
	case 16:
		ind = "ATwoBig"
	default:
		ind = "AUnknownErr"
	}
	s.Error(p, li, co, s.ls.lx.so.sMap(ind))
	s.yAPasErreur = false
} // erreurSem

func (se *sem) creeChaine (o *C.Object) (res string, ok bool) {
	ok = !o.ErrorIn()
	if ok {
		M.Assert(o.ObjType() == C.StringObj, 100)
		res = o.ObjString()
	} else {
		se.yAPasErreur = false
		res = ""
	}
	return
} // creeChaine

func (se *sem) creeNTChaine (o *C.Object) (res string, ok bool) {
	res, ok = se.creeChaine(o)
	if ok {
		r := []rune(res)
		res = string(slice(r, 1, len(r) - 1))
	}
	return
} // creeNTChaine

func (se *sem) extraitChaine (o *C.Object) (res string, ok bool) {
	res, ok = se.creeChaine(o)
	if ok {
		r := []rune(res)
		res = string(slice(r, 1, len(r) - 2))
	}
	return
} // extraitChaine

func (se *sem) valeur (o *C.Object) *pCard {
	s, ok := se.creeChaine(o)
	if !ok {
		return nil
	}
	i, ok := strToCard(s)
	if !ok {
		se.erreurSem(16, o)
	}
	return &pCard{i: i}
} // valeur

func (se *sem) creeLAtt (o *C.Object) *listeNN {
	if o.ErrorIn() {
		se.yAPasErreur = false
		return nil
	}
	var l *listeNN = nil
	for o.ObjFunc() == consB {
		obj := o.ObjTermSon(1)
		s, ok := se.creeChaine(obj); M.Assert(ok, 100)
		n := o.ObjTermSon(2).ObjUser().(*pCard)
		if !insereNN(&l, s, n.i) {
			se.erreurSem(15, obj)
		}
		o = o.ObjTermSon(3)
	}
	return l
} // creeLAtt

func (se *sem) creeGram (o *C.Object) *gram {
	if o.ErrorIn() {
		se.yAPasErreur = false
		return nil
	}
	var g *gram = nil
	for o.ObjFunc() == consB {
		n := o.ObjTermSon(1).ObjUser().(*pCard)
		g = &gram{suivant: g, tOuNT: o.ObjTermSon(2).ObjFunc() == trueB, num: n.i}
		o = o.ObjTermSon(3)
	}
	return g
} // creeGram

func (se *sem) creeLAct (o *C.Object) *listeAction {
	
	creeAct := func (o *C.Object) *action {
		
		creeFonct := func (o *C.Object) *action {
			
			creeParams := func (o *C.Object) *listeAttribs {
				
				creeAtt := func (o *C.Object) *listeAttribs {
					obj := o.ObjTermSon(1)
					i := obj.ObjUser().(*pCard)
					var (j int; tOuNT bool; lPar *listeAttribs = nil)
					if se.ls.chercheGramCour(i.i, &j, &tOuNT) {
						if o.ObjTermSonsNb() == 1 {
							if tOuNT {
								lPar = &listeAttribs{numTNT: i.i, numAttrib: 1}
							} else {
								se.erreurSem(9, obj)
							}
						} else {
							if tOuNT {
								se.erreurSem(8, obj)
							} else {
								obj := o.ObjTermSon(2)
								s, ok := se.creeChaine(obj)
								var k int
								if ok && se.ls.chercheAtt(j, s, &k) {
									lPar = &listeAttribs{numTNT: i.i, numAttrib: k}
								} else {
									se.erreurSem(7, obj)
								}
							}
						}
					} else {
						se.erreurSem(6, obj)
					}
					return lPar
				} // creeAtt
				
				// creeParams
				var lPar *listeAttribs = nil
				for o.ObjFunc() == consB {
					l := creeAtt(o.ObjTermSon(1))
					if l != nil {
						l.suivant = lPar
						lPar = l
					}
					o = o.ObjTermSon(2)
				}
				return lPar
			} // creeParams
			
			// creeFonct
			cB := o.ObjTermSon(1).ObjUser().(*pCardBool)
			return &action{numFct: cB.c, softOrHard: cB.b, params: creeParams(o.ObjTermSon(2))}
		} // creeFonct
		
		// creeAct
		if o.ErrorIn() {
			se.yAPasErreur = false
			return nil
		} else {
			obj := o.ObjTermSon(1)
			i := obj.ObjUser().(*pCard)
			var (j int; tOuNT bool; act *action = nil)
			if se.ls.chercheGramCour(i.i, &j, &tOuNT) {
				if tOuNT {
					se.erreurSem(8, obj)
				} else {
					obj := o.ObjTermSon(2)
					s, ok := se.creeChaine(obj)
					var k int
					if ok && se.ls.chercheAtt(j, s, &k) {
						act = creeFonct(o.ObjTermSon(3))
						act.numNT = i.i
						act.numAtt = k
					} else {
						se.erreurSem(7, obj)
					}
				}
			} else {
				se.erreurSem(6, obj)
			}
			return act
		}
	} // creeAct
	
	// creeLAct
	var lAct *listeAction = nil
	for o.ObjFunc() == consB {
		lAct = &listeAction{suivant: lAct, actionF: creeAct(o.ObjTermSon(1))}
		o = o.ObjTermSon(2)
	}
	return lAct
} // creeLAct

func (se *sem) creeLExp (o *C.Object) lexExp {

	const (
		
		slash = 4 + iota
		ou
		cat
		star
		eps
		mult
		catCha
		eOL
		creeCar
	
	)
	
	creeMultL := func (lExp lexExp, x, y *pCard, o *C.Object) lexExp {
		
		mult := func (e lexExp, n int) lexExp {
			var f lexExp
			if n == 0 {
				f = creeEpsL()
			} else {
				f = copieLexExp(e)
				for i := 2; i <= n; i++ {
					f = creeCatL(f, copieLexExp(e))
				}
			}
			return f
		} // mult
		
		// creeMultL
		if x.i > y.i {
			se.erreurSem(1, o)
			return nil
		}
		e := mult(lExp, x.i)
		for i := x.i + 1; i <= y.i; i++ {
			e = creeCatL(e, creeOrL(creeEpsL(), copieLexExp(lExp)))
		}
		return e
	} // creeMultL
	
	creeCatChaL := func (s string) lexExp {
		var e lexExp
		if s == "" {
			e = creeEpsL()
		} else {
			r := []rune(s)
			c := S.NewSet()
			c.Incl(int(r[0]))
			e = creeCarL(c)
			for i := 1; i < len(r); i++ {
				c := S.NewSet()
				c.Incl(int(r[i]))
				e = creeCatL(e, creeCarL(c))
			}
		}
		return e
	} // creeCatChaL

	creeCaracL := func (cE *pCarEns) lexExp {
		return creeCarL(cE.c)
	} // creeCaracL

	// creeLExp
	if o.ErrorIn() {
		se.yAPasErreur = false
		return nil
	}
	if o.ObjType() == C.UserObj {
		p := o.ObjUser()
		if p == nil {
			return nil
		}
		return copieLexExp(p.(lexExp))
	}
	switch o.ObjFunc() {
	case slash:
		return creeSuiviL(se.creeLExp(o.ObjTermSon(1)), se.creeLExp(o.ObjTermSon(2)))
	case ou:
		return creeOrL(se.creeLExp(o.ObjTermSon(1)), se.creeLExp(o.ObjTermSon(2)))
	case cat:
		return creeCatL(se.creeLExp(o.ObjTermSon(1)), se.creeLExp(o.ObjTermSon(2)))
	case star:
		return creeStarL(se.creeLExp(o.ObjTermSon(1)))
	case eps:
		return creeEpsL()
	case mult:
		return creeMultL(se.creeLExp(o.ObjTermSon(1)), o.ObjTermSon(2).ObjUser().(*pCard), o.ObjTermSon(3).ObjUser().(*pCard), o.ObjTermSon(3))
	case catCha:
		return creeCatChaL(o.ObjTermSon(1).ObjUser().(*pChaine).s)
	case eOL:
		return creeEOLL()
	case creeCar:
		return creeCaracL(o.ObjTermSon(1).ObjUser().(*pCarEns))
	default:
		M.Assert(false, 100)
		return nil
	}
} // creeLExp

func (se *sem) Execution (numFct, nbPars int, params C.ObjectsList) (o *C.Object, a C.Anyptr, ok bool) {
	
	creeTokL := func (o *C.Object) lexExp {
		s, ok := se.creeChaine(o)
		if !ok {
			return nil
		}
		var e lexExp = nil
		if se.ls.lx.chercheTokL(s, &e) {
			e = copieLexExp(e)
		} else {
			se.erreurSem(2, o)
		}
		return e
	} // creeTokL
	
	// Execution
	const (
		maxChar = 0xE01EF
	)
	se.yAPasErreur = true
	switch numFct {
	case 1: //InsereDefL
		o = C.Parameter(params, 1)
		s, ok := se.creeChaine(o)
		lExp := se.creeLExp(C.Parameter(params, 2))
		if ok && !se.ls.lx.insereDefL(s, lExp) {
			se.erreurSem(0, o)
		}
	case 2: //DefComment
		o = C.Parameter(params, 1)
		se.ls.lx.defComment(se.creeLExp(o), se.creeLExp(C.Parameter(params, 2)))
	case 3: //InitSynt
		se.ls.lx.effaceDef()
		se.ls.debutSynt()
	case 4: //InsereTokL
		o = C.Parameter(params, 1)
		s, ok := se.creeChaine(o)
		lExp := se.creeLExp(C.Parameter(params, 4))
		if ok {
			o = C.Parameter(params, 2)
			var n int
			if o.ErrorIn() {
				n = 0
			} else {
				n = o.ObjUser().(*pCard).i
			}
			if !se.ls.lx.insereTokL(s, n, C.Parameter(params, 3).ObjFunc() == trueB, lExp) {
				se.erreurSem(0, C.Parameter(params, 1))
			}
		}
	case 5: //ZERO
		a = &pCard{i: 0}
	case 6: //valeur
		o = C.Parameter(params, 1)
		a = se.valeur(o)
	case 7: //ToutCar
		a = &pCarEns{c: S.Interval(0, maxChar).Diff(S.Small(1 << uint(eOF1) | 1 << uint(eOF2) | 1 << uint(eOL1) | 1 << uint(eOL2)))}
	case 8: //Union
		o = C.Parameter(params, 1)
		p1 := o.ObjUser()
		p2 := C.Parameter(params, 2).ObjUser()
		if p1 == nil || p2 == nil {
			se.yAPasErreur = false
			a = nil
		} else {
			a = &pCarEns{c: p1.(*pCarEns).c.Union(p2.(*pCarEns).c)}
		}
	case 9: //Difference
		o = C.Parameter(params, 1)
		p1 := o.ObjUser()
		p2 := C.Parameter(params, 2).ObjUser()
		if p1 == nil || p2 == nil {
			se.yAPasErreur = false
			a = nil
		} else {
			a = &pCarEns{c: p1.(*pCarEns).c.Diff(p2.(*pCarEns).c)}
		}
	case 10: //Vide
		a = &pCarEns{c: S.NewSet()}
	case 11: //Intervalle
		o = C.Parameter(params, 1)
		if o.ErrorIn() {
			se.yAPasErreur = false
			a = nil
		} else {
			r1 := []rune(o.ObjUser().(*pChaine).s)
			if len(r1) == 1 {
				o = C.Parameter(params, 2)
				if o.ErrorIn() {
					se.yAPasErreur = false
					a = nil
				} else {
					r2 := []rune(o.ObjUser().(*pChaine).s)
					if len(r2) == 1 {
						a = &pCarEns{c: S.Interval(int(r1[0]), int(r2[0]))}
					} else {
						se.erreurSem(3, o)
						a = nil
					}
				}
			} else {
				se.erreurSem(3, o)
				a = nil
			}
		}
	case 12: //extraitChaine
		o = C.Parameter(params, 1)
		s, _ := se.extraitChaine(o)
		a = &pChaine{s: s}
	case 13: //NombreChaine
		o = C.Parameter(params, 1)
		if o.ErrorIn() {
			se.yAPasErreur = false
			a = nil
		} else {
			card := o.ObjUser().(*pCard)
			if card.i > S.SMax {
				se.erreurSem(16, o)
				a = nil
			} else {
				r := make([]rune, 1)
				r[0] = rune(card.i)
				a = &pChaine{s: string(r)}
			}
		}
	case 14: //FinDeclaration
		se.ls.finDeclaration()
	case 15: //InsereNonTerm
		o = C.Parameter(params, 1)
		s, ok := se.creeNTChaine(o)
		lAtt := se.creeLAtt(C.Parameter(params, 3))
		if ok && !se.ls.insereNonTerm(s, C.Parameter(params, 2).ObjFunc() == trueB, lAtt) {
			se.erreurSem(4, o)
		}
	case 16: //InsereFonctionHard
		s, ok := se.creeChaine(C.Parameter(params, 1))
		o = C.Parameter(params, 2)
		var n int
		if o.ErrorIn() {
			se.yAPasErreur = false
			n = 1
		} else {
			n = o.ObjUser().(*pCard).i
			if n == 0 {
				se.erreurSem(14, o)
			}
		}
		if ok && !se.ls.insereFonctionHard(s, n) {
			se.erreurSem(11, o)
		}
	case 17: //InsereFonctionSoft
		s, ok := se.creeChaine(C.Parameter(params, 1))
		o = C.Parameter(params, 2)
		var n int
		if o.ErrorIn() {
			se.yAPasErreur = false
			n = 1
		} else {
			n = o.ObjUser().(*pCard).i
			if n == 0 {
				se.erreurSem(14, o)
			}
		}
		if ok && !se.ls.insereFonctionSoft(s, n) {
			se.erreurSem(12, o)
		}
	case 18: //InsereSem
		o = C.Parameter(params, 1)
		se.ls.insereSem(se.creeLAct(o))
	case 19: //ChercheNT
		o = C.Parameter(params, 1)
		s, _ := se.creeNTChaine(o)
		card := new(pCard)
		if se.yAPasErreur && !se.ls.chercheNT(s, &card.i) {
			se.erreurSem(10, o)
			a = nil
		}
		a = card
	case 20: //InsereRegle
		o = C.Parameter(params, 1)
		var n int
		if o.ErrorIn() {
			se.yAPasErreur = false
			n = 0
		} else {
			n = o.ObjUser().(*pCard).i
		}
		o = C.Parameter(params, 2)
		se.ls.insereRegle(n, se.creeGram(o))
	case 21: //ChercheTerm
		o = C.Parameter(params, 1)
		s, _ := se.creeChaine(o)
		card := new(pCard)
		if se.yAPasErreur && !se.ls.lx.chercheTerm(s, &card.i) {
			se.erreurSem(5, o)
			a = nil
		}
		a = card
	case 22: //ChercheFixePrec
		o = C.Parameter(params, 1)
		s, _ := se.creeChaine(o)
		if se.yAPasErreur {
			var (n int; b bool)
			if se.ls.lx.cherchePrec(s, &n, &b) {
				se.ls.fixeRegleCourPrec(n, b)
			} else {
				se.erreurSem(5, o)
			}
		}
	case 23: //ChercheFonction
		o = C.Parameter(params, 1)
		s, _ := se.creeChaine(o)
		cB := new(pCardBool)
		if se.yAPasErreur && !se.ls.chercheFonction(s, &cB.c, &cB.b) {
			se.erreurSem(13, o)
			a = nil
		}
		a = cB
	case 24: //COPIE
		a = &pCardBool{c: 0, b: true}
	case 25: //FixeTerminaison
		o = C.Parameter(params, 1)
		if o.ErrorIn() {
			se.yAPasErreur = false
		} else {
			se.ls.fixeTerminaison(o.ObjUser().(*pCard).i)
		}
	case 26: //FixeAxiome
		o = C.Parameter(params, 1)
		if o.ErrorIn() {
			se.yAPasErreur = false
		} else {
			se.ls.fixeAxiome(o.ObjUser().(*pCard).i)
		}
	case 27: //ChercheTokL
		o = C.Parameter(params, 1)
		a = creeTokL(o)
	case 28: //Nomme
		o = C.Parameter(params, 1)
		se.nom, _ = se.creeChaine(o)
		se.nom = string([]rune(se.nom)[1:len(se.nom) - 1])
	}
	ok = se.yAPasErreur
	return
} // Execution
