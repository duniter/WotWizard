/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelLexique is part of the Babel subsystem, a compiler compiler. BabelLexique translates regular expressions given in a definition document into automata.

import (
	
	C	"util/sets"
	M	"util/misc"
	U	"unicode"

)

const (
	
	// "end of file" name
	
	nomEOFC = "LEOF"

)

type (
	
	lexExp interface {
		lexExpCom () *lexExpC
	}
	
	lexExpC struct {
		nullable bool
		firstPos,
		lastPos,
		followPos *listLexExp
		numPos int
	}
		
	// A lexical regular expression.
	
	lexExpOrL struct {
		lexExpC
		exp1,
		exp2 lexExp
	}
	
	lexExpCatL struct {
		lexExpC
		exp1,
		exp2 lexExp
	}
	
	lexExpStarL struct {
		lexExpC
		exp lexExp
	}
	
	lexExpSlashL struct {
		lexExpC
		exp lexExp
	}
	
	lexExpEpsL struct {
		lexExpC
	}
	
	lexExpCarL struct {
		lexExpC
		cars *C.Set
	}
	
	lexExpEndL struct {
		lexExpC
		numTok int
	}
	
	listLexExp struct {
		suivant *listLexExp
		lExp lexExp
	}
	
	lDef struct {
		suivant *lDef
		lNom string
		lExp lexExp
		prec int
		assoc,
		utilise bool
	}
	
	listLLE struct {
		 suivant *listLLE
		 num,
		 numA int
		 etat *listLexExp
		 trans listTrans
	 }
	
	listTrans interface {
		listTransCom () *listTransC
	 }
	
	listTransC struct {
		 suivant listTrans
		 etatSuiv *listLLE
	 }
	
	listTransReg struct {
		listTransC
		eC *C.Set
	 }
	
	listTransEps struct {
		listTransC
		etatEps *listLLE
	 }
	
	lexical struct {
		so sortieser
		listeDef,
		listeTerm *lDef
		listeTok,
		listeCom lexExp
		listeEtats,
		listeECom *listLLE
		numeroT,
		profLex,
		profCom int
	}
	
	auto struct {
		suivant,
		precedent *auto
		lllE *listLLE
	}
	
	ensAuto struct {
		suivant *ensAuto
		auto *auto
		rec int
	 }
	
	autoTab []*ensAuto
	
	transAuto interface {
		transAutoCom () *transAutoC
	}
	
	transAutoC struct {
		suivant transAuto
		eS int
	}
	
	transAutoReg struct {
		transAutoC
		ensC *C.Set
	}
	
	transAutoEps struct {
		transAutoC
		eps int
	}
	
	listePos struct {
		suivant *listePos
		pos *listLexExp
		eC *C.Set
	}

)

var (
	
	nomEOF string

)

func (l *lexExpOrL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpCatL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpStarL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpSlashL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpEpsL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpCarL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func (l *lexExpEndL) lexExpCom () *lexExpC {
	return &l.lexExpC
} // lexExpCom

func creeOrL (e1, e2 lexExp) lexExp {
	return &lexExpOrL{exp1: e1, exp2: e2}
} // creeOrL

func creeCatL (e1, e2 lexExp) lexExp {
	return &lexExpCatL{exp1: e1, exp2: e2}
} // creeCatL

func creeStarL (e lexExp) lexExp {
	return &lexExpStarL{exp: e}
} // creeStarL

func creeEpsL () lexExp {
	return new(lexExpEpsL)
} // creeEpsL

func creeCarL (c *C.Set) lexExp {
	return &lexExpCarL{cars: c}
} // creeCarL

func creeSlashL (e lexExp) lexExp {
	return &lexExpSlashL{exp: e}
} // creeSlashL

func creeEndL (n int) lexExp {
	return &lexExpEndL{numTok: n}
} // creeEndL

func creeEOLL () lexExp {
	c1 := C.NewSet()
	c1.Incl(int(eOL1))
	c2 := C.NewSet()
	c2.Incl(int(eOL2))
	return creeOrL(creeCatL(creeCarL(c1), creeOrL(creeCarL(c2), creeEpsL())), creeCatL(creeOrL(creeCarL(c1), creeEpsL()), creeCarL(c2)))
} // creeEOLL

func creeSuiviL (e1, e2 lexExp) lexExp {
	return creeCatL(e1, creeSlashL(creeCatL(e2, creeEndL(0))))
} // creeSuiviL

func copieListLexExp (l *listLexExp) *listLexExp {
	m := new(listLexExp)
	p := m
	for l != nil {
		p.suivant = new(listLexExp)
		p = p.suivant
		*p = *l
		l = l.suivant
	}
	p.suivant = nil
	return m.suivant
} // copieListLexExp

func copieLexExp (e lexExp) lexExp {
	if e == nil {
		return nil
	}
	var f lexExp
	switch  ee := e.(type) {
	case *lexExpOrL:
		f = creeOrL(copieLexExp(ee.exp1), copieLexExp(ee.exp2))
	case *lexExpCatL:
		f = creeCatL(copieLexExp(ee.exp1), copieLexExp(ee.exp2))
	case *lexExpStarL:
		f = creeStarL(copieLexExp(ee.exp))
	case *lexExpSlashL:
		f = creeSlashL(copieLexExp(ee.exp))
	case *lexExpEpsL:
		f = creeEpsL()
	case *lexExpCarL:
		f = creeCarL(ee.cars)
	case *lexExpEndL:
		f = creeEndL(ee.numTok)
	}
	*f.lexExpCom() = *e.lexExpCom()
	return f
} // copieLexExp

func insereListL (liste **lDef, nom string, pr int, ass bool, expL lexExp) bool {
	l := *liste
	m := (*lDef)(nil)
	loop:
	for l != nil {
		switch {
		case l.lNom < nom:
			m = l
			l = l.suivant
		case l.lNom == nom:
			return false
		case l.lNom > nom:
			break loop
		}
	}
	l = &lDef{lNom: nom, lExp: expL, prec: pr, assoc: ass, utilise: false}
	if m == nil {
		l.suivant = *liste
		*liste = l
	} else {
		l.suivant = m.suivant
		m.suivant = l
	}
	return true
} // insereListL

func (lx *lexical) insereDefL (nom string, expL lexExp) bool {
	return insereListL(&lx.listeDef, nom, 0, false, expL)
} // insereDefL

func (lx *lexical) defComment (debCom, finCom lexExp) {
	e := creeOrL(creeCatL(debCom, creeEndL(1)), creeCatL(finCom, creeEndL(2)))
	lx.listeTok = creeOrL(lx.listeTok, copieLexExp(e))
	lx.listeCom = creeOrL(lx.listeCom, e)
} // defComment

func chercheListL (liste *lDef, nom string, lExpL **lDef) bool {
	l := liste
	for {
		if l == nil {
			return false
		}
		switch {
		case l.lNom < nom:
			l = l.suivant
		case l.lNom == nom:
			*lExpL = l
			return true
		case l.lNom > nom:
			return false
		}
	}
	return false
} // chercheListL

func (lx *lexical) insereTokL (nom string, pr int, ass bool, expL lexExp) bool {
	expL = creeCatL(expL, creeEndL(lx.numeroT))
	lx.numeroT++
	var l *lDef
	if !chercheListL(lx.listeDef, nom, &l) && insereListL(&lx.listeTerm, nom, pr, ass, expL) {
		if lx.listeTok == nil {
			lx.listeTok = copieLexExp(expL)
		} else {
			lx.listeTok = creeOrL(lx.listeTok, copieLexExp(expL))
		}
		return true
	}
	return false
} // insereTokL

func (lx *lexical) initTokL () {
	nomEOF = lx.so.sMap(nomEOFC)
	lx.numeroT = 0
	c := C.NewSet()
	c.Incl(int(eOF1))
	c.Incl(int(eOF2))
	lx.insereTokL("", 0, false, creeCarL(c))
	lx.listeCom = creeCatL(creeCarL(c), creeEndL(0))
	lx.listeTerm.utilise = true
	lx.numeroT = 3
} // initTokL

func (lx *lexical) chercheTokL (nom string, expL *lexExp) bool {
	var l *lDef
	if chercheListL(lx.listeDef, nom, &l) {
		*expL = l.lExp
		return true
	}
	if chercheListL(lx.listeTerm, nom, &l) {
		*expL = l.lExp.(*lexExpCatL).exp1
		return true
	}
	return false
} // chercheTokL

func insereLLE (l **listLexExp, e lexExp) {
	p := *l
	m := (*listLexExp)(nil)
	n := e.lexExpCom().numPos
	for p != nil && p.lExp.lexExpCom().numPos < n {
		m = p
		p = p.suivant
	}
	p = &listLexExp{lExp: e}
	if m == nil {
		p.suivant = *l
		*l = p
	} else {
		p.suivant = m.suivant
		m.suivant = p
	}
} // insereLLE

func unionLLE (l1, l2 *listLexExp) *listLexExp {
	l := new(listLexExp)
	m := l
	for l1 != nil || l2 != nil {
		l.suivant = new(listLexExp)
		l = l.suivant
		if l1 == nil {
			l.lExp = l2.lExp
			l2 = l2.suivant
		} else if l2 == nil {
			l.lExp = l1.lExp
			l1 = l1.suivant
		} else if l1.lExp.lexExpCom().numPos < l2.lExp.lexExpCom().numPos {
			l.lExp = l1.lExp
			l1 = l1.suivant
		} else if l1.lExp.lexExpCom().numPos > l2.lExp.lexExpCom().numPos {
			l.lExp = l2.lExp
			l2 = l2.suivant
		} else {
			l.lExp = l1.lExp
			l1 = l1.suivant
			l2 = l2.suivant
		}
	}
	l.suivant = nil
	return m.suivant
} // unionLLE

func egLLE (l1, l2 *listLexExp) bool {
	for l1 != nil && l2 != nil && l1.lExp == l2.lExp {
		l1 = l1.suivant
		l2 = l2.suivant
	}
	return l1 == l2
} // egLLE

func cherchInsListLLE (l **listLLE, e *listLexExp, n *int, nA int) *listLLE {
	p := *l
	m := (*listLLE)(nil)
	for p != nil && !egLLE(p.etat, e) {
		m = p
		p = p.suivant
	}
	if p == nil {
		p = &listLLE{num: *n, numA: nA, etat: e}
		*n++
		if m == nil {
			p.suivant = *l
			*l = p
		} else {
			p.suivant = m.suivant
			m.suivant = p
		}
	}
	return p
} // cherchInsListLLE

func (l *listTransReg) listTransCom () *listTransC {
	return &l.listTransC
} // listTransCom

func (l *listTransEps) listTransCom () *listTransC {
	return &l.listTransC
} // listTransCom

func insereTrans (l *listTrans, eC *C.Set, e *listLLE) {
	p := *l
	m := listTrans(nil)
	for p != nil && p.listTransCom().etatSuiv != e {
		m = p
		p = p.listTransCom().suivant
	}
	var q *listTransReg
	if p == nil {
		q = &listTransReg{eC: C.NewSet(), listTransC: listTransC{etatSuiv: e}}
		if m == nil {
			q.suivant = *l
			*l = q
		} else {
			q.suivant = m.listTransCom().suivant
			m.listTransCom().suivant = q
		}
	} else {
		q = p.(*listTransReg)
	}
	q.eC = q.eC.Union(eC)
} // insereTrans

func insereTransEps (l *listTrans, epsT, e *listLLE) {
	*l = &listTransEps{etatEps: epsT, listTransC: listTransC{etatSuiv: e, suivant: *l}}
} // insereTransEps

func (lx *lexical) effaceDef () {
	lx.listeDef = nil
} // effaceDef

func calcul (e lexExp, n *int, prof int, profMax *int) {
	
	faitFollow := func (l1, l2 *listLexExp) {
		for l1 != nil {
			e := l1.lExp.lexExpCom()
			e.followPos = unionLLE(e.followPos, l2)
			l1 = l1.suivant
		}
	} // faitFollow
	
	//calcul
	*profMax = M.Max(prof, *profMax)
	switch ee := e.(type) {
	case *lexExpOrL:
		calcul(ee.exp1, n, prof, profMax); ec1 := ee.exp1.lexExpCom()
		calcul(ee.exp2, n, prof, profMax); ec2 := ee.exp2.lexExpCom()
		ee.nullable = ec1.nullable || ec2.nullable
		ee.firstPos = unionLLE(ec1.firstPos, ec2.firstPos)
		ee.lastPos = unionLLE(ec1.lastPos, ec2.lastPos)
	case *lexExpCatL:
		calcul(ee.exp1, n, prof, profMax); ec1 := ee.exp1.lexExpCom()
		calcul(ee.exp2, n, prof, profMax); ec2 := ee.exp2.lexExpCom()
		ee.nullable = ec1.nullable && ec2.nullable
		if ec1.nullable {
			ee.firstPos = unionLLE(ec1.firstPos, ec2.firstPos)
		} else {
			ee.firstPos = copieListLexExp(ec1.firstPos)
		}
		if ec2.nullable {
			ee.lastPos = unionLLE(ec1.lastPos, ec2.lastPos)
		} else {
			ee.lastPos = copieListLexExp(ec2.lastPos)
		}
		faitFollow(ec1.lastPos, ec2.firstPos)
	case *lexExpStarL:
		calcul(ee.exp, n, prof, profMax); ec := ee.exp.lexExpCom()
		ee.nullable = true
		ee.firstPos = copieListLexExp(ec.firstPos)
		ee.lastPos = copieListLexExp(ec.lastPos)
		faitFollow(ee.lastPos, ee.firstPos)
	case *lexExpEpsL:
		ee.nullable = true
	case *lexExpSlashL:
		calcul(ee.exp, n, prof + 1, profMax)
		ee.numPos = *n
		*n++
		ee.nullable = false
		insereLLE(&ee.firstPos, ee)
		insereLLE(&ee.lastPos, ee)
	case *lexExpCarL:
		ee.numPos = *n
		*n++
		ee.nullable = ee.cars.IsEmpty()
		insereLLE(&ee.firstPos, ee)
		insereLLE(&ee.lastPos, ee)
	case *lexExpEndL:
		ee.numPos = *n
		*n++
		ee.nullable = false
		insereLLE(&ee.firstPos, ee)
		insereLLE(&ee.lastPos, ee)
	}
} // calcul

func (t *transAutoReg) transAutoCom () *transAutoC {
	return &t.transAutoC
} // transAutoCom

func (t *transAutoEps) transAutoCom () *transAutoC {
	return &t.transAutoC
} // transAutoCom

func (lx *lexical) fabriqueLex () {
	
	etats := func (lT lexExp, lE **listLLE) {
		
		reduit := func (lE *listLLE, nbA int) {
			
			var aT autoTab
			
			red := func (lE *listLLE, nA int) {
				
				initEns := func (lE *listLLE, nA int) (eA*ensAuto) {
					a := (*auto)(nil)
					for lE != nil {
						if lE.numA == nA {
							a = &auto{suivant: a, lllE: lE}
						}
						lE = lE.suivant
					}
					eA = nil
					for a != nil {
						b := a.suivant
						n := M.MaxInt32
						p := a.lllE.etat
						for p != nil {
							le := p.lExp
							if l, ok := le.(*lexExpEndL); ok && l.numTok < n {
								n = l.numTok
							}
							p = p.suivant
						}
						e := eA
						for (e != nil) && (e.rec != n) {
							e = e.suivant
						}
						if e == nil {
							c := new(auto)
							c.suivant = c
							c.precedent = c
							e = &ensAuto{suivant: eA, auto: c, rec: n}
							eA = e
						}
						a.precedent = e.auto
						a.suivant = e.auto.suivant
						aT[a.lllE.num] = e
						e.auto.suivant.precedent = a
						e.auto.suivant = a
						a = b
					}
					return
				} // initEns
				
				partitionne := func (eA *ensAuto) {
					
					consTransA := func (l listTrans) transAuto {
						
						compTransA := func (eS1, eS2, eE1, eE2 int, tE1, tE2 bool) int8 {
							if tE1 {
								if tE2 {
									if eS1 < eS2 {
										return inf
									}
									if eS1 > eS2 {
										return sup
									}
									if eE1 < eE2 {
										return inf
									}
									if eE1 > eE2 {
										return sup
									}
									return ega
								}
								return inf
							}
							if tE2 {
								return sup
							}
							if eS1 < eS2 {
								return inf
							}
							if eS1 > eS2 {
								return sup
							}
							return ega
						} // compTransA
						
						//consTransA
						var t transAuto = nil
						for l != nil {
							nS := aT[l.listTransCom().etatSuiv.num].rec
							var nE int
							ll, lE := l.(*listTransEps)
							if lE {
								nE = ll.etatEps.num
							}
							var (n int; b bool)
							tt := t; ttt := transAuto(nil)
							for tt != nil {
								var tte *transAutoEps
								tte, b = tt.(*transAutoEps)
								if b {
									n = tte.eps
								}
								if compTransA(tt.transAutoCom().eS, nS, n, nE, b, lE) != inf {
									break
								}
								ttt = tt
								tt = tt.transAutoCom().suivant
							}
							if tt == nil || compTransA(tt.transAutoCom().eS, nS, n, nE, b, lE) == sup {
								if lE {
									tt = &transAutoEps{eps: nE}
								} else {
									tt = &transAutoReg{ensC: C.NewSet()}
								}
								tt.transAutoCom().eS = nS
								if ttt == nil {
									tt.transAutoCom().suivant = t
									t = tt
								} else {
									tt.transAutoCom().suivant = ttt.transAutoCom().suivant
									ttt.transAutoCom().suivant = tt
								}
							}
							if !lE {
								tt.(*transAutoReg).ensC = tt.(*transAutoReg).ensC.Union(l.(*listTransReg).eC)
							}
							l = l.listTransCom().suivant
						}
						return t
					} //consTransA
					
					egalTransA := func (t1, t2 transAuto) bool {
						
						egal1TransA := func (t1, t2 transAuto) bool {
							if t1.transAutoCom().eS != t2.transAutoCom().eS {
								return false
							}
							tE1, b1 := t1.(*transAutoEps)
							tE2, b2 := t2.(*transAutoEps)
							if b1 != b2 {
								return false
							}
							if b1 {
								return tE1.eps == tE2.eps
							}
							return t1.(*transAutoReg).ensC.Equal(t2.(*transAutoReg).ensC)
						} // egal1TransA
						
						// egalTransA
						for t1 != nil && t2 != nil && egal1TransA(t1, t2) {
							t1 = t1.transAutoCom().suivant; t2 = t2.transAutoCom().suivant
						}
						return t1 == t2
					} // egalTransA
				
					//partitionne
					var f *ensAuto = nil
					n := 0; e := eA
					for e != nil {
						e.rec = n
						n++
						f = e
						e = e.suivant
					}
					for {
						eF := f
						e := eA
						for {
							b := e.auto.suivant
							c := b.suivant
							if c != e.auto {
								a := new(auto)
								a.suivant = a
								a.precedent = a
								tB := consTransA(b.lllE.trans)
								for {
									d := c.suivant
									tC := consTransA(c.lllE.trans)
									if !egalTransA(tB, tC) {
										c.suivant.precedent = c.precedent
										c.precedent.suivant = c.suivant
										c.suivant = a
										c.precedent = a.precedent
										a.precedent.suivant = c
										a.precedent = c
									}
									c = d
									if c == e.auto {
										break
									}
								}
								if a.suivant != a {
									f.suivant = &ensAuto{auto: a, rec: n}
									f = f.suivant
									n++
								}
							}
							if e == f {
								break
							}
							e = e.suivant
						}
						if f == eF {
							break
						}
						e = eF
						for {
							e = e.suivant
							a := e.auto.suivant
							for {
								aT[a.lllE.num] = e
								a = a.suivant
								if a == e.auto {
									break
								}
							}
							if e == f {
								break
							}
						}
					}
				} //partitionne
				
				elimine := func (eA **ensAuto) {
					
					isTransEps := func (l listTrans) bool {
						_, b := l.(*listTransEps)
						return b
					} // isTransEps
					
					//elimine
					e := *eA
					for e != nil {
						a := e.auto.suivant
						l := a.lllE
						tr := l.trans
						t1 := tr; t2 := listTrans(nil)
						for t1 != nil {
							tc := t1.listTransCom()
							tc.etatSuiv = aT[tc.etatSuiv.num].auto.suivant.lllE
							if isTransEps(t1) && tc.etatSuiv == l {
								t1 = tc.suivant
								if t2 == nil {
									tr = t1; l.trans = t1
								} else {
									t2.listTransCom().suivant = t1
								}
							} else {
								t3 := tr
								for t3 != t1 && (t3.listTransCom().etatSuiv != tc.etatSuiv || (isTransEps(t3) || isTransEps(t1)) && (!(isTransEps(t3) && isTransEps(t1)) || t3.(*listTransEps).etatEps != t1.(*listTransEps).etatEps)) {
									t3 = t3.listTransCom().suivant
								}
								if t3 == t1 {
									t2 = t1
									t1 = tc.suivant
								} else {
									if !isTransEps(t3) {
										t3.(*listTransReg).eC = t3.(*listTransReg).eC.Union(t1.(*listTransReg).eC)
									}
									t1 = tc.suivant
									t2.listTransCom().suivant = t1
								}
							}
						}
						e = e.suivant
					}
					for *eA != nil {
						a := (*eA).auto.suivant
						l := a.lllE
						var ll *listLLE = nil
						(*eA).auto.precedent.suivant = nil
						a = a.suivant
						*eA = (*eA).suivant
						for a != nil {
							for l != a.lllE {
								ll = l
								l = l.suivant
							}
							ll.suivant = l.suivant
							l = ll.suivant
							a = a.suivant
						}
					}
				} // elimine
				
				// red
				eA := initEns(lE, nA)
				partitionne(eA)
				elimine(&eA)
			} //red
			
			numerote := func (lE *listLLE) {
				n := 0
				for lE != nil {
					lE.num = n
					n++
					lE = lE.suivant
				}
			} // numerote
		
			//reduit
			n := 0
			l := lE
			for l != nil {
				n++
				l = l.suivant
			}
			aT = make(autoTab, n)
			for i := 0; i < nbA; i++ {
				red(lE, i)
			}
			numerote(lE)
		} // reduit
		
		ajouteListePos := func (lP **listePos, e *lexExpCarL) {
			eC1 := e.cars.Copy()
			l := *lP
			for (l != nil) && !eC1.IsEmpty() {
				eC2 := l.eC.Inter(eC1)
				if !eC2.IsEmpty() {
					m := &listLexExp{suivant: l.pos, lExp: e}
					if eC2.Equal(l.eC) {
						l.pos = m
					} else {
						l.eC = l.eC.Diff(eC2)
						*lP = &listePos{suivant: *lP, pos: m, eC: eC2.Copy()}
					}
					eC1 = eC1.Diff(eC2)
				}
				l = l.suivant
			}
			if !eC1.IsEmpty() {
				*lP = &listePos{suivant: *lP, pos: &listLexExp{lExp: e}, eC: eC1}
			}
		} // ajouteListePos
	
		// etats
		n := 0
		var u *listLexExp = nil
		u = unionLLE(u, lT.lexExpCom().firstPos)
		l := cherchInsListLLE(lE, u, &n, 0)
		nA := 1
		for l != nil {
			var lP *listePos = nil
			m := l.etat
			for m != nil {
				e := m.lExp
				if ee, b := e.(*lexExpCarL); b {
					ajouteListePos(&lP, ee)
				}
				m = m.suivant
			}
			for lP != nil {
				var u *listLexExp = nil
				m := lP.pos
				for m != nil {
					u = unionLLE(u, m.lExp.lexExpCom().followPos)
					m = m.suivant
				}
				if u != nil {
					insereTrans(&l.trans, lP.eC, cherchInsListLLE(lE, u, &n, l.numA))
				}
				lP = lP.suivant
			}
			m = l.etat
			for m != nil {
				e := m.lExp
				if ee, b := e.(*lexExpSlashL); b {
					var u *listLexExp = nil
					u = unionLLE(u, ee.followPos)
					p := cherchInsListLLE(lE, u, &n, l.numA)
					if p != l {
						u = nil
						u = unionLLE(u, e.(*lexExpSlashL).exp.lexExpCom().firstPos)
						q := n
						insereTransEps(&l.trans, cherchInsListLLE(lE, u, &n, nA), p)
						if n > q {
							nA++
						}
					}
				}
				m = m.suivant
			}
			l = l.suivant
		}
		reduit(*lE, nA)
	} // etats
	
	//fabriqueLex
	n := 0
	lx.profLex = 1
	calcul(lx.listeTok, &n, 1, &lx.profLex)
	etats(lx.listeTok, &lx.listeEtats)
	n = 0
	lx.profCom = 1
	calcul(lx.listeCom, &n, 1, &lx.profCom)
	etats(lx.listeCom, &lx.listeECom)
} // fabriqueLex

func (lx *lexical) chercheTerm (term string, numT *int) bool {
	var l *lDef
	if chercheListL(lx.listeTerm, term, &l) {
		*numT = l.lExp.(*lexExpCatL).exp2.(*lexExpEndL).numTok
		l.utilise = true
		return true
	}
	return false
} // chercheTerm

func (lx *lexical) cherchePrec (term string, pr *int, ass *bool) bool {
	var l *lDef
	if chercheListL(lx.listeTerm, term, &l) {
		*pr = l.prec
		*ass = l.assoc
		return true
	}
	return false
} // cherchePrec

func (lx *lexical) precedence (numT int) (pr int, ass bool) {
	l := lx.listeTerm
	for l.lExp.(*lexExpCatL).exp2.(*lexExpEndL).numTok != numT {
		l = l.suivant
	}
	pr = l.prec
	ass = l.assoc
	return
} // precedence

func unique (e lexExp, n *int) bool {
	switch ee := e.(type) {
	case *lexExpCatL:
		return unique(ee.exp1, n) && unique(ee.exp2, n)
	case *lexExpCarL:
		*n++
		if ee.cars.NbElems() != 1 {
			return false
		}
		c, b := ee.cars.Attach().FirstE(); M.Assert(b, 100)
		return U.IsPrint(rune(c))
	default:
		return false
	}
} // unique

func copieNom (e lexExp, n *int, nom []rune) {
	switch ee := e.(type) {
	case *lexExpCatL:
		copieNom(ee.exp1, n, nom)
		copieNom(ee.exp2, n, nom)
	case *lexExpCarL:
		c, b := ee.cars.Attach().FirstE(); M.Assert(b, 100)
		nom[*n] = rune(c)
		*n++
	}
} // copieNom

func (lx *lexical) creeLex () *compiler {

	const (
		
		nbToksCom = 3

	)
	
	faitNom := func (e lexExp, nom *string) bool {
		n := 2
		if unique(e, &n) {
			rs := make([]rune, n)
			rs[0] = '"'
			rs[n - 1] = '"'
			n = 1
			copieNom(e, &n, rs)
			*nom = string(rs)
			return true
		}
		return false
	} // faitNom
	
	creeAuto := func (maxT int, lE *listLLE) (nE int, eL etatsLex) {
		nE = 0
		m := lE
		for m != nil {
			nE++
			m = m.suivant
		}
		eL = make(etatsLex, nE)
		m = lE
		for i := 0; i < nE; i++ {
			p := m.etat
			eL[m.num].recon = maxT
			for p != nil {
				switch e := p.lExp.(type) {
				case *lexExpEndL:
					if e.numTok < eL[m.num].recon {
						eL[m.num].recon = e.numTok
					}
				default:
				}
				p = p.suivant
			}
			eL[m.num].nbTrans = 0
			eL[m.num].nbEps = 0
			q := m.trans
			loop:
			for q != nil {
				switch q.(type) {
				case *listTransEps:
					eL[m.num].nbTrans++
					eL[m.num].nbEps++
				default:
					break loop
				}
				q = q.listTransCom().suivant
			}
			for q != nil {
				it := q.(*listTransReg).eC.Attach()
				_, _, ok := it.First()
				for ok {
					eL[m.num].nbTrans++
					_, _, ok = it.Next()
				}
				q = q.listTransCom().suivant
			}
			if eL[m.num].nbTrans != 0 {
				eL[m.num].transL = make(transLex, eL[m.num].nbTrans)
				j := 0
				q := m.trans
				loop2:
				for q != nil {
					switch qq := q.(type) {
					case *listTransEps:
						eL[m.num].transL[j] = &gotoLexT{goTo: qq.etatSuiv.num, transit: qq.etatEps.num}
						j++
					default:
						break loop2
					}
					q = q.listTransCom().suivant
				}
				for q != nil {
					it := q.(*listTransReg).eC.Attach()
					k, l, ok := it.First()
					for ok {
						eL[m.num].transL[j] = &gotoLexC{premCar: rune(k), derCar: rune(l), goTo: q.listTransCom().etatSuiv.num}
						j++
						k, l, ok = it.Next()
					}
					q = q.listTransCom().suivant
				}
				for j := eL[m.num].nbEps + 1; j < eL[m.num].nbTrans; j++ {
					g := eL[m.num].transL[j].(*gotoLexC)
					k := j
					for k > eL[m.num].nbEps && eL[m.num].transL[k - 1].(*gotoLexC).premCar > g.derCar {
						eL[m.num].transL[k] = eL[m.num].transL[k - 1]
						k--
					}
					eL[m.num].transL[k] = g
				}
			}
			m = m.suivant
		}
		return
	} // creeAuto

	//creeLex
	c := &compiler{nbToksLex: lx.numeroT, toksLex: make(toksLexT, lx.numeroT), profEtatsL: lx.profLex, profEtatsC: lx.profCom}
	for i := 1; i<= 2; i++ {
		c.toksLex[i].utile = false
		c.toksLex[i].valUt = false
	}
	l := lx.listeTerm
	for l != nil {
		e := l.lExp.(*lexExpCatL)
		ee := e.exp2.(*lexExpEndL)
		if ee.numTok == 0 {
			c.toksLex[ee.numTok].nom = nomEOF
		} else if l.utilise {
			if !faitNom(e.exp1, &c.toksLex[ee.numTok].nom) {
				c.toksLex[ee.numTok].nom = l.lNom
			}
		}
		c.toksLex[ee.numTok].utile = l.utilise
		c.toksLex[ee.numTok].valUt = false
		l = l.suivant
	}
	c.nbEtatsLex, c.etatsLex = creeAuto(lx.numeroT, lx.listeEtats)
	c.nbEtatsCom, c.etatsCom = creeAuto(nbToksCom, lx.listeECom)
	return c
} // creeLex

func (lx *lexical) ecrisTerm (term int) string {
	if term == 0 {
		return nomEOF
	} else {
		l := lx.listeTerm
		for l.lExp.(*lexExpCatL).exp2.(*lexExpEndL).numTok != term {
			l = l.suivant
		}
		return l.lNom
	}
} // ecrisTerm
