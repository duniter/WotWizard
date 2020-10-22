/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelSyntaxe is part of the Babel subsystem, a compiler compiler. BabelSyntaxe builds the parser tables.

import (
	
	M	"util/misc"
	S	"util/sets2"

)

type (
	
	listeNN struct {
		 suivant *listeNN
		 nom string
		 num int
	 }
	
	gram struct {
		suivant *gram
		num int
		tOuNT bool
	}
	
	listeAttribs struct {
		suivant *listeAttribs
		numTNT,
		numAttrib int
	}
	
	action struct {
		numNT,
		numAtt,
		numFct int
		softOrHard bool
		params *listeAttribs
	}
	
	listeAction struct {
		 suivant *listeAction
		 actionF *action
	 }

)

const (
	
	dumping = false

)

type (
	
	ensTok = S.Set
	
	listeNT struct {
		 suivant *listeNT
		 mark bool
		 nomNT string
		 lAtt *listeNN
	 }
	
	regle struct {
		suivant *regle
		gNT int
		dGram *gram
		actions *listeAction
		numR,
		nouvNumR,
		precR int
		assocR bool
	}
	
	nTDesc struct {
		desc *listeNT
		numerNT int
		regles *regle
		first ensTok
		epsFirst,
		utilNT bool
	}
	
	nTTab []nTDesc
	
	lRItem struct {
		suivant *lRItem
		regleP *regle
		posP *gram
		posN int
		look ensTok
		propag *lRPropag
	}
	
	lRPropag struct {
		suivant *lRPropag
		to *lRItem
	}
	
	lRSet struct {
		 suivant *lRSet
		 set *lRItem
		 goTo *lRGoto
		 numSet int
	 }
	
	lRGoto struct {
		suivant *lRGoto
		trans *lRSet
		tNT int
		tOuNT bool
	}
	
	syntaxical struct {
		lx *lexical
		listeNTF *listeNT
		listeFH,
		listeFS *listeNN
		dNT nTTab
		termine,
		nbNT,
		derNT,
		nbRegles int
		lRS *lRSet
		yAAffichage,
		yAAttention bool
	}

)

func (ls *syntaxical) initEnsTok () (eT ensTok) {
	eT = S.NewSet()
	return
} // initEnsTok

func (ls *syntaxical) fixeTerminaison (numTerm int) {
	ls.termine = numTerm
} // fixeTerminaison

func (ls *syntaxical) insereNonTerm (nT string, marque bool, lA *listeNN) bool {
	var m *listeNT = nil
	l := ls.listeNTF
	loop:
	for l != nil {
		switch {
		case l.nomNT < nT:
			 m = l
			 l = l.suivant
		case l.nomNT > nT:
			 break loop
		default:
			 return false
		}
	}
	l = &listeNT{nomNT: nT, mark: marque, lAtt: lA}
	if m == nil {
		l.suivant = ls.listeNTF
		ls.listeNTF = l
	} else {
		l.suivant = m.suivant
		m.suivant = l
	}
	ls.nbNT++
	return true
} // insereNonTerm

func insereNN (l **listeNN, nomN string, numN int) bool {
	var q *listeNN = nil
	p := *l
	loop:
	for p != nil {
		switch {
		case p.nom < nomN:
			 q = p
			 p = p.suivant
		case p.nom > nomN:
			 break loop
		default:
			 return false
		}
	}
	p = &listeNN{nom: nomN, num: numN}
	if q == nil {
		p.suivant = *l
		*l = p
	} else {
		p.suivant = q.suivant
		q.suivant = p
	}
	return true
} // insereNN

func chercheNN (l *listeNN, nomN string, posN, numN *int) bool {
	*posN = 1
	for l != nil {
		switch {
		case l.nom < nomN:
			 *posN++
			 l = l.suivant
		case l.nom > nomN:
			 return false
		default:
			 *numN = l.num
			 return true
		}
	}
	return false
} // chercheNN

func (ls *syntaxical) insereFonctionHard (nomFct string, numFct int) bool {
	var n, m int
	if !chercheNN(ls.listeFS, nomFct, &n, &m) {
		return insereNN(&ls.listeFH, nomFct, numFct)
	}
	return false
} // insereFonctionHard

func (ls *syntaxical) insereFonctionSoft (nomFct string, numFct int) bool {
	var n, m int
	if !chercheNN(ls.listeFH, nomFct, &n, &m) {
		return insereNN(&ls.listeFS, nomFct, numFct)
	}
	return false
} // insereFonctionSoft

func (ls *syntaxical) chercheNT (nT string, numNT *int) bool {
	*numNT = 0
	l := ls.listeNTF
	for l != nil {
		switch {
		case l.nomNT < nT:
			 *numNT++
			 l = l.suivant
		case l.nomNT > nT:
			 return false
		default:
			 return true
		}
	}
	return false
} // chercheNT

func (ls *syntaxical) insereRegle (gauche int, droite *gram) {
	ls.nbRegles++
	r := &regle{suivant: ls.dNT[gauche].regles, gNT: gauche, dGram: droite, numR: ls.nbRegles, precR: 0, assocR: true}
	for droite != nil {
		if droite.tOuNT {
			r.precR, r.assocR = ls.lx.precedence(droite.num)
		}
		droite = droite.suivant
	}
	ls.dNT[gauche].regles = r
	ls.derNT = gauche
} // insereRegle

func (ls *syntaxical) insereSem (lA *listeAction) {
	ls.dNT[ls.derNT].regles.actions = lA
} // insereSem

func (ls *syntaxical) chercheGramCour (pos int, numTNT *int, termOuNT *bool) bool {
	if ls.derNT == 0 {
		return false
	}
	if pos == 0 {
		*numTNT = ls.derNT
		*termOuNT = false
		return true
	}
	l := ls.dNT[ls.derNT].regles.dGram
	for (l != nil) && (pos > 1) {
		pos--
		l = l.suivant
	}
	if l == nil {
		return false
	}
	*numTNT = l.num
	*termOuNT = l.tOuNT
	return true
} // chercheGramCour

func (ls *syntaxical) chercheAtt (numNT int, nomAttrib string, nAtt *int) bool {
	if numNT >= ls.nbNT {
		return false
	}
	var n int
	return chercheNN(ls.dNT[numNT].desc.lAtt, nomAttrib, nAtt, &n)
} // chercheAtt

func (ls *syntaxical) chercheFonction (nomFct string, numFct *int, sH *bool) bool {
	var n int
	if chercheNN(ls.listeFH, nomFct, &n, numFct) {
		*sH = false
		return true
	}
	if chercheNN(ls.listeFS, nomFct, &n, numFct) {
		*sH = true
		return true
	}
	return false
} // chercheFonction

func (ls *syntaxical) fixeRegleCourPrec (pr int, ass bool) {
	if ls.derNT > 0 {
		r := ls.dNT[ls.derNT].regles
		r.precR = pr
		r.assocR = ass
	}
} // fixeRegleCourPrec

func (ls *syntaxical) initSynt () {
	ls.termine = 0
	ls.nbNT = 0
	ls.nbRegles = 0
	ls.derNT = 0
} // initSynt

func (ls *syntaxical) debutSynt () {
} // debutSynt

func (ls *syntaxical) finDeclaration () {
	ls.insereNonTerm("", false, nil)
	ls.dNT = make(nTTab, ls.nbNT)
	l := ls.listeNTF
	for i := 0; i < ls.nbNT; i++ {
		ls.dNT[i].desc = l
		ls.dNT[i].first = ls.initEnsTok()
		ls.dNT[i].epsFirst = false
		ls.dNT[i].utilNT = false
		l = l.suivant
	}
	ls.dNT[0].utilNT = true
	ls.dNT[0].regles = &regle{gNT: 0, dGram: &gram{tOuNT: false}, numR: 0, precR: 0, assocR: true}
} // finDeclaration

func (ls *syntaxical) fixeAxiome (numNT int) {
	if ls.dNT[0].regles.dGram != nil {
		ls.dNT[0].regles.dGram.num = numNT
	}
} // fixeAxiome

func (ls *syntaxical) ecrisNonTerm (nTerm int) string {
	l := ls.listeNTF
	for i := 1; i <= nTerm; i++ {
		l = l.suivant
	}
	return "_" + l.nomNT
} // ecrisNonTerm

func (ls *syntaxical) ecrisRegle (nR int) {
	var r *regle
	loop:
	for i := 0; i < ls.nbNT; i++ {
		r = ls.dNT[i].regles
		for r != nil {
			if r.numR == nR {
				break loop
			}
			r = r.suivant
		}
	}
	M.Assert(r != nil, 100)
	ls.lx.so.sString(ls.ecrisNonTerm(r.gNT))
	ls.lx.so.sString(" = ")
	g := r.dGram
	for g != nil {
		if g.tOuNT {
			ls.lx.so.sString(ls.lx.ecrisTerm(g.num))
		} else {
			ls.lx.so.sString(ls.ecrisNonTerm(g.num))
		}
		ls.lx.so.sString(" ")
		g = g.suivant
	}
	ls.lx.so.sString(";")
	ls.lx.so.sLn()
} // ecrisRegle

func (ls *syntaxical) verifSynt () bool {
	b := true
	for i := 0; i < ls.nbNT; i++ {
		if ls.dNT[i].regles == nil {
			b = false
			ls.lx.so.sString(ls.lx.so.sMap("AnError"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.lx.so.sString(ls.lx.so.sMap("SNoRule", ls.ecrisNonTerm(i)))
			ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
		}
	}
	return b
} // verifSynt

type (
	
	lien struct {
		suivant *lien
		fleche int
	}
	
	nTGraph []*lien
	
	nTGraphs struct {
		suivant *nTGraphs
		graf nTGraph
	}

)

const (
	
	inconnu = iota
	syntLoc
	syntGlob
	herLoc
	herGlob
	faux

)

type (
	
	genresAtt []int8
	
	nTAtt struct {
		tailGrafs int
		grafs *nTGraphs
		genres genresAtt
	}
	
	nTAtts []nTAtt

)

const (
	
	invisible = iota
	visible
	vu

)

type (
	
	lienDouble struct {
		suivant *lienDouble
		versNT,
		versAtt int
	}
	
	attSem struct {
		entraine *lienDouble
		vis int8
	 }
	
	attsSem []attSem
	
	nTSem struct {
		numNTSem int
		attsSem attsSem
	}
	
	nTSems []nTSem
	
	indsSem []*nTGraphs
	
	regSem struct {
		 nbNTSem int
		 nTSems nTSems
		 indsSem indsSem
	 }
	
	regSems []regSem

)

func (ls *syntaxical) ecrisAtt (nNT, nAtt int) string {
	l := ls.dNT[nNT].desc.lAtt
	for i := 2; i <= nAtt; i++ {
		l = l.suivant
	}
	return l.nom
} // ecrisAtt

func parcours1 (ls *syntaxical, d nTSems, nNT, nAtt int, debNT, debAtt *int, suite, aff *bool) bool {
	s := d[nNT].attsSem
	switch s[nAtt - 1].vis {
	case invisible:
		s[nAtt - 1].vis = visible
		l := s[nAtt - 1].entraine
		for l != nil {
			if parcours1(ls, d, l.versNT, l.versAtt, debNT, debAtt, suite, aff) {
				if *aff {
					if *suite {
						ls.lx.so.sString(",")
						ls.lx.so.sLn()
						ls.lx.so.sString(ls.lx.so.sMap("Swhich"))
						ls.lx.so.sString(" ")
					} else {
						ls.lx.so.sLn()
						*suite = true
					}
					ls.lx.so.sString(ls.lx.so.sMap("SDerives", ls.ecrisAtt(d[nNT].numNTSem, nAtt), ls.ecrisNonTerm(d[nNT].numNTSem)))
					*aff = !((nNT == *debNT) && (nAtt == *debAtt))
				}
				return true
			}
			l = l.suivant
		}
		s[nAtt - 1].vis = vu
	case visible:
		ls.lx.so.sString(ls.lx.so.sMap("AnError"))
		ls.lx.so.sLn(); ls.lx.so.sLn()
		ls.lx.so.sString(ls.lx.so.sMap("SCyclicAtt"))
		ls.lx.so.sLn(); ls.lx.so.sLn()
		ls.lx.so.sString(ls.lx.so.sMap("SAttNonT", ls.ecrisAtt(d[nNT].numNTSem, nAtt), ls.ecrisNonTerm(d[nNT].numNTSem)))
		*debNT = nNT; *debAtt = nAtt
		*suite = false
		*aff = true
		return true
	case vu:
	}
	return false
} // parcours1

func parcours2 (d nTSems, g nTGraph, nNT, nAtt, src int) {
	
	insereLien := func (l **lien, nAtt int) {
		var q *lien = nil
		p := *l
		for p != nil && p.fleche < nAtt {
			q = p
			p = p.suivant
		}
		if p == nil || p.fleche > nAtt {
			r := &lien{suivant: p, fleche: nAtt}
			if q == nil {
				*l = r
			} else {
				q.suivant = r
			}
		}
	} // insereLien

	// parcours2
	if nNT == 0 && nAtt != src {
		insereLien(&g[src - 1], nAtt)
	}
	l := d[nNT].attsSem[nAtt - 1].entraine
	for l != nil {
		parcours2(d, g, l.versNT, l.versAtt, src)
		l = l.suivant
	}
} // parcours2

func (ls *syntaxical) verifSem () bool {
	
	creeNTAtts := func () nTAtts {
		nTA := make(nTAtts, ls.nbNT)
		l := ls.listeNTF
		for i := 0; i < ls.nbNT; i++ {
			nTA[i].tailGrafs = 0
			a := l.lAtt
			for a != nil {
				nTA[i].tailGrafs++
				a = a.suivant
			}
			nTA[i].grafs = new(nTGraphs)
			nTA[i].grafs.graf = make(nTGraph, nTA[i].tailGrafs)
			nTA[i].genres = make(genresAtt, nTA[i].tailGrafs)
			for j := 0; j < nTA[i].tailGrafs; j++ {
				nTA[i].genres[j] = inconnu
			}
			l = l.suivant
		}
		return nTA
	} // creeNTAtts
	
	insereLD := func (l **lienDouble, nNT, nAtt int) {
		p := *l
		for p != nil && (p.versNT != nNT || p.versAtt != nAtt) {
			p = p.suivant
		}
		if p == nil {
			*l = &lienDouble{suivant: *l, versNT: nNT, versAtt: nAtt}
		}
	} // insereLD
	
	creeRegSems := func (nTAtts nTAtts) regSems {

		estNT := func (g *gram, pos int) bool {
			if pos == 0 {
				return true
			}
			for i := 2; i <= pos; i++ {
				g = g.suivant
			}
			return !g.tOuNT
		} // estNT

		conv := func (g *gram, pos int) int {
			n := 0
			for i := 1; i <= pos; i++ {
				if !g.tOuNT {
					n++
				}
				g = g.suivant
			}
			return n
		} // conv
		
		regS := make(regSems, ls.nbRegles + 1)
		for i := 0; i < ls.nbNT; i++ {
			r := ls.dNT[i].regles
			for r != nil {
				regS[r.numR].nbNTSem = 1
				g := r.dGram
				for g != nil {
					if !g.tOuNT {
						regS[r.numR].nbNTSem++
					}
					g = g.suivant
				}
				regS[r.numR].nTSems = make(nTSems, regS[r.numR].nbNTSem)
				regS[r.numR].indsSem = make(indsSem, regS[r.numR].nbNTSem - 1)
				regS[r.numR].nTSems[0].numNTSem = i
				regS[r.numR].nTSems[0].attsSem = make(attsSem, nTAtts[i].tailGrafs)
				for k := 0; k < nTAtts[i].tailGrafs; k++ {
					regS[r.numR].nTSems[0].attsSem[k].vis = invisible
				}
				g = r.dGram
				j := 1
				for g != nil {
					if !g.tOuNT {
						regS[r.numR].nTSems[j].numNTSem = g.num
						regS[r.numR].nTSems[j].attsSem = make(attsSem, nTAtts[g.num].tailGrafs)
						for k := 0; k < nTAtts[g.num].tailGrafs; k++ {
							regS[r.numR].nTSems[j].attsSem[k].vis = invisible
						}
						j++
					}
					g = g.suivant
				}
				a := r.actions
				for a != nil {
					ac := a.actionF
					p := ac.params
					for p != nil {
						if estNT(r.dGram, p.numTNT) {
							insereLD(&regS[r.numR].nTSems[conv(r.dGram, p.numTNT)].attsSem[p.numAttrib - 1].entraine, conv(r.dGram, ac.numNT), ac.numAtt)
						}
						p = p.suivant
					}
					a = a.suivant
				}
				r = r.suivant
			}
		}
		return regS
	} // creeRegSems

	detGenre := func (nTA nTAtts) bool {
		erreurSH := func (nNT, nAtt int) {
			ls.lx.so.sString(ls.lx.so.sMap("AnError"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.lx.so.sString(ls.lx.so.sMap("SSyntAndHer", ls.ecrisAtt(nNT, nAtt), ls.ecrisNonTerm(nNT)))
			ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
		} // erreurSH

		//detGenre
		b := true
		for i := 0; i < ls.nbNT; i++ {
			r := ls.dNT[i].regles
			for r != nil {
				lA := r.actions
				for lA != nil {
					a := lA.actionF
					if a.numNT == 0 {
						switch nTA[i].genres[a.numAtt - 1] {
						case inconnu:
							 nTA[i].genres[a.numAtt - 1] = syntLoc
						case herLoc:
							 b = false
							 nTA[i].genres[a.numAtt - 1] = faux
							 erreurSH(i, a.numAtt)
						default:
						}
					} else {
						g := r.dGram
						for j := 2; j <= a.numNT; j++ {
							g = g.suivant
						}
						switch nTA[g.num].genres[a.numAtt - 1] {
						case inconnu:
							 nTA[g.num].genres[a.numAtt - 1] = herLoc
						case syntLoc:
							 b = false
							 nTA[g.num].genres[a.numAtt - 1] = faux
							 erreurSH(g.num, a.numAtt)
						default:
						}
					}
					lA = lA.suivant
				}
				r = r.suivant
			}
		}
		return b
	} // detGenre

	detGlob := func (nTA nTAtts) bool {
		erreurI := func (nNT, nAtt int, err bool) {
			if err {
				ls.lx.so.sString(ls.lx.so.sMap("AnError"))
			} else {
				ls.yAAttention = true
				ls.lx.so.sString(ls.lx.so.sMap("AnWarning"))
			}
			ls.lx.so.sLn(); ls.lx.so.sLn()
			if err {
				ls.lx.so.sString(ls.lx.so.sMap("SNeverCalc", ls.ecrisAtt(nNT, nAtt), ls.ecrisNonTerm(nNT)))
			} else {
				ls.lx.so.sString(ls.lx.so.sMap("SNeverUsedAtt", ls.ecrisAtt(nNT, nAtt), ls.ecrisNonTerm(nNT)))
			}
			ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
		} // erreurI

		// detGlob
		b := true
		for i := 0; i < ls.nbNT; i++ {
			r := ls.dNT[i].regles
			for r != nil {
				lA := r.actions
				for lA != nil {
					l := lA.actionF.params
					for l != nil {
						if l.numTNT == 0 {
							switch nTA[i].genres[l.numAttrib - 1] {
							case inconnu:
								 b = false
								 erreurI(i, l.numAttrib, true)
								 nTA[i].genres[l.numAttrib - 1] = faux
							case herLoc:
								 nTA[i].genres[l.numAttrib - 1] = herGlob
							default:
							}
						} else {
							g := r.dGram
							for j := 2; j <= l.numTNT; j++ {
								g = g.suivant
							}
							if !g.tOuNT {
								switch nTA[g.num].genres[l.numAttrib - 1] {
								case inconnu:
									 b = false
									 erreurI(g.num, l.numAttrib, true)
									 nTA[g.num].genres[l.numAttrib - 1] = faux
								case syntLoc:
									 nTA[g.num].genres[l.numAttrib - 1] = syntGlob
								default:
								}
							}
						}
						l = l.suivant
					}
					lA = lA.suivant
				}
				r = r.suivant
			}
		}
		for i := 0; i < ls.nbNT; i++ {
			for j := 1; j <= nTA[i].tailGrafs; j++ {
				if nTA[i].genres[j - 1] == inconnu {
					erreurI(i, j, false)
				}
			}
		}
		return b
	} // detGlob

	verifGenre := func (nTA nTAtts) bool {
		erreurPasCalc := func (nR, nNT, nAtt int, g int8) {
			ls.lx.so.sString(ls.lx.so.sMap("AnError"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			var s string
			switch g {
			case syntLoc:
				s = ls.lx.so.sMap("SLocSynt")
			case syntGlob:
				s = ls.lx.so.sMap("SGlobSynt")
			case herLoc:
				s = ls.lx.so.sMap("SLocHer")
			case herGlob:
				s = ls.lx.so.sMap("SGlobHer")
			}
			ls.lx.so.sString(ls.lx.so.sMap("SNoCalc", s, ls.ecrisAtt(nNT, nAtt), ls.ecrisNonTerm(nNT)))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.ecrisRegle(nR)
			ls.lx.so.sLn(); ls.lx.so.sLn()
		} // erreurPasCalc
		
		erreurTropCalc := func (nR, nNT, nAtt int) {
			ls.lx.so.sString(ls.lx.so.sMap("AnError"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.lx.so.sString(ls.lx.so.sMap("SSeveralCalc", ls.ecrisAtt(nNT, nAtt), ls.ecrisNonTerm(nNT)))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.ecrisRegle(nR)
			ls.lx.so.sLn(); ls.lx.so.sLn()
		} // erreurTropCalc

		vGenreGlob := func (posNT, nNT, tG, nR int, g int8, gs genresAtt, a *listeAction) bool {
			b := true
			for i := 1; i <= tG; i++ {
				if gs[i - 1] == g {
					n := 0
					l := a
					for l != nil {
						ac := l.actionF
						if ac.numNT == posNT && ac.numAtt == i {
							n++
						}
						l = l.suivant
					}
					if n == 0 {
						b = false
						erreurPasCalc(nR, nNT, i, g)
					} else if n > 1 {
						b = false
						erreurTropCalc(nR, nNT, i)
					}
				}
			}
			return b
		} // vGenreGlob

		vGenreLoc := func (posNT, nNT, tG, nR int, g int8, gs genresAtt, a *listeAction) bool {
			b := true
			for i := 1; i <= tG; i++ {
				if gs[i - 1] == g {
					n := 0
					calcule := false
					l := a
					for l != nil {
						ac := l.actionF
						if ac.numNT == posNT && ac.numAtt == i {
							n++
						}
						lA := ac.params
						for lA != nil {
							if lA.numTNT == posNT && lA.numAttrib == i {
								calcule = true
							}
							lA = lA.suivant
						}
						l = l.suivant
					}
					if calcule && n == 0 {
						b = false
						erreurPasCalc(nR, nNT, i, g)
					} else if n > 1 {
						b = false
						erreurTropCalc(nR, nNT, i)
					}
				}
			}
			return b
		} // vGenreLoc

		//verifGenre
		b := true
		for i := 0; i < ls.nbNT; i++ {
			r := ls.dNT[i].regles
			for r != nil {
				b = vGenreGlob(0, i, nTA[i].tailGrafs, r.numR, syntGlob, nTA[i].genres, r.actions) && b
				b = vGenreLoc(0, i, nTA[i].tailGrafs, r.numR, syntLoc, nTA[i].genres, r.actions) && b
				g := r.dGram
				j := 1
				for g != nil {
					if !g.tOuNT {
						b = vGenreGlob(j, g.num, nTA[g.num].tailGrafs, r.numR, herGlob, nTA[g.num].genres, r.actions) && b
						b = vGenreLoc(j, g.num, nTA[g.num].tailGrafs, r.numR, herLoc, nTA[g.num].genres, r.actions) && b
					}
					j++
					g = g.suivant
				}
				r = r.suivant
			}
		}
		return b
	} // verifGenre
	
	verifCycle := func (nTA nTAtts, regS regSems) bool {
		initInd := func (nTA nTAtts, nTSems nTSems, nbInd int, ind indsSem) {
			for i := 1; i <= nbInd; i++ {
				ind[i - 1] = nTA[nTSems[i].numNTSem].grafs
			}
		} // initInd
		
		copieDagR := func (nTA nTAtts, nbNTSem int, src nTSems) nTSems {
			
			copieDLiens := func (src *lienDouble) *lienDouble {
				but := new(lienDouble)
				b := but
				for src != nil {
					b.suivant = new(lienDouble)
					b = b.suivant
					*b = *src
					src = src.suivant
				}
				b.suivant = nil
				return but.suivant
			} // copieDLiens
		
			//copieDagR
			but := make(nTSems, nbNTSem)
			for i := 0; i < nbNTSem; i++ {
				but[i].numNTSem = src[i].numNTSem
				but[i].attsSem = make(attsSem, nTA[but[i].numNTSem].tailGrafs)
				for j := 0; j < nTA[but[i].numNTSem].tailGrafs; j++ {
					but[i].attsSem[j].entraine = copieDLiens(src[i].attsSem[j].entraine)
					but[i].attsSem[j].vis = invisible
				}
			}
			return but
		} // copieDagR

		ajouteDLiens := func (nTA nTAtts, nbNTSem int, ind indsSem, d nTSems) {
			for i := 1; i < nbNTSem; i++ {
				g := ind[i - 1]
				for j := 0; j < nTA[d[i].numNTSem].tailGrafs; j++ {
					l := g.graf[j]
					for l != nil {
						insereLD(&d[i].attsSem[j].entraine, i, l.fleche)
						l = l.suivant
					}
				}
			}
		} // ajouteDLiens

		cyclique := func (nTA nTAtts, nbNTSem int, d nTSems) bool {
			var (
				debNT, debAtt int
				suite, aff bool
			)
			for i := 0; i < nbNTSem; i++ {
				for j := 1; j <= nTA[d[i].numNTSem].tailGrafs; j++ {
					if (d[i].attsSem[j - 1].vis == invisible) && parcours1(ls, d, i, j, &debNT, &debAtt, &suite, &aff) {
						ls.lx.so.sString(".")
						ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
						return true
					}
				}
			}
			return false
		} // cyclique

		ajouteLiens := func (tail int, d nTSems, g nTGraph) {
			for src := 1; src <= tail; src++ {
				parcours2(d, g, 0, src, src)
			}
		} // ajouteLiens

		insereGraphNT := func (tail int, lG *nTGraphs, g nTGraph) bool {
			
			egGraph := func (tail int, g1, g2 nTGraph) bool {
				for i := 0; i < tail; i++ {
					l1 := g1[i]
					l2 := g2[i]
					for {
						if l1 == nil {
							if l2 == nil {
								break
							}
							return false
						}
						if l2 == nil {
							return false
						}
						if l1.fleche != l2.fleche {
							return false
						}
						l1 = l1.suivant
						l2 = l2.suivant
					}
				}
				return true
			} // egGraph
		
			//insereGraphNT
			var q *nTGraphs = nil
			p := lG
			for p != nil && !egGraph(tail, p.graf, g) {
				q = p
				p = p.suivant
			}
			if p == nil {
				q.suivant = &nTGraphs{graf: g}
				return true
			}
			return false
		} // insereGraphNT

		avance := func (nTA nTAtts, nTS nTSems, nbInd int, ind indsSem) bool {
			i := nbInd
			for i > 0 {
				ind[i - 1] = ind[i - 1].suivant
				if ind[i - 1] != nil {
					return true
				}
				ind[i - 1] = nTA[nTS[i].numNTSem].grafs
				i--
			}
			return false
		} // avance

		//verifCycle
		for {
			stop := true
			for i := 0; i <= ls.nbRegles; i++ {
				initInd(nTA, regS[i].nTSems, regS[i].nbNTSem - 1, regS[i].indsSem)
				for {
					d := copieDagR(nTA, regS[i].nbNTSem, regS[i].nTSems)
					ajouteDLiens(nTA, regS[i].nbNTSem, regS[i].indsSem, d)
					if cyclique(nTA, regS[i].nbNTSem, d) {
						return false
					}
					g := make(nTGraph, nTA[regS[i].nTSems[0].numNTSem].tailGrafs)
					ajouteLiens(nTA[regS[i].nTSems[0].numNTSem].tailGrafs, d, g)
					stop = !insereGraphNT(nTA[regS[i].nTSems[0].numNTSem].tailGrafs, nTA[regS[i].nTSems[0].numNTSem].grafs, g) && stop
					if !avance(nTA, regS[i].nTSems, regS[i].nbNTSem - 1, regS[i].indsSem) {
						break
					}
				}
			}
			if stop {
				break
			}
		}
		return true
	} // verifCycle

	nTA := creeNTAtts()
	if detGenre(nTA) && detGlob(nTA) && verifGenre(nTA) {
		regS := creeRegSems(nTA)
		return verifCycle(nTA, regS)
	}
	return false
} // verifSem

func compLRItem (l *lRItem, r *regle, pos int) int8 {
	if l.regleP.numR < r.numR {
		return inf
	}
	if l.regleP.numR > r.numR {
		return sup
	}
	if l.posN < pos {
		return inf
	}
	if l.posN > pos {
		return sup
	}
	return ega
} // compLRItem

func (ls *syntaxical) insereLRItem (l **lRItem, r *regle, pP *gram, pN int) {
	var q *lRItem = nil
	p := *l
	loop:
	for p != nil {
		switch compLRItem(p, r, pN) {
		case inf:
			 q = p
			 p = p.suivant
		case sup:
			 break loop
		default:
			 return
		}
	}
	p = &lRItem{regleP: r, posP: pP, posN: pN, look: ls.initEnsTok()}
	if q == nil {
		p.suivant = *l
		*l = p
	} else {
		p.suivant = q.suivant
		q.suivant = p
	}
} // insereLRItem

func egLRSet (l1, l2 *lRItem) bool {
	for l1 != nil && l2 != nil && l1.regleP == l2.regleP && l1.posN == l2.posN {
		l1 = l1.suivant
		l2 = l2.suivant
	}
	return l1 == l2
} // egLRSet

func (ls *syntaxical) insereLRSet (nSet *lRItem, org *lRSet, gram int, termOuNT bool) {
	var q *lRSet = nil
	p := ls.lRS
	for p != nil && !egLRSet(p.set, nSet) {
		q = p
		p = p.suivant
	}
	if p == nil {
		p = &lRSet{set: nSet}
		if q == nil {
			p.suivant = ls.lRS
			ls.lRS = p
		} else {
			p.suivant = q.suivant
			q.suivant = p
		}
	}
	org.goTo = &lRGoto{suivant: org.goTo, trans: p, tNT: gram, tOuNT: termOuNT}
} // insereLRSet

func (ls *syntaxical) copieLRItem (s *lRItem) *lRItem {
	p := new(lRItem)
	q := p
	for s != nil {
		q.suivant = new(lRItem)
		q = q.suivant
		*q = *s
		q.look = s.look.Copy()
		s = s.suivant
	}
	q.suivant = nil
	return p.suivant
} // copieLRItem

func (ls *syntaxical) closure0 (s *lRItem) *lRItem {
	s = ls.copieLRItem(s)
	t := s
	for t != nil {
		if !(t.posP == nil || t.posP.tOuNT) {
			l := ls.dNT[t.posP.num].regles
			for l != nil {
				var q *lRItem = nil
				p := s
				for p != nil && !(p.posN == 1 && p.regleP == l) {
					q = p
					p = p.suivant
				}
				if p == nil {
					q.suivant = &lRItem{suivant: q.suivant, regleP: l, posP: l.dGram, posN: 1, look: ls.initEnsTok()}
				}
				l = l.suivant
			}
		}
		t = t.suivant
	}
	return s
} // closure0

func (ls *syntaxical) goTo (s *lRItem, tNT int, termOuNT bool) *lRItem {
	var t *lRItem = nil
	for s != nil {
		if s.posP != nil {
			g := s.posP
			if g.num == tNT && g.tOuNT == termOuNT {
				ls.insereLRItem(&t, s.regleP, g.suivant, s.posN + 1)
			}
		}
		s = s.suivant
	}
	return t
} // goTo

func (ls *syntaxical) faitLR0 () {
	var l *lRItem = nil
	ls.insereLRItem(&l, ls.dNT[0].regles, ls.dNT[0].regles.dGram, 1)
	ls.lRS = &lRSet{set: l}
	m := ls.lRS
	for m != nil {
		l := ls.closure0(m.set)
		for i := 0; i < ls.lx.numeroT; i++ {
			p := ls.goTo(l, i, true)
			if p != nil {
				ls.insereLRSet(p, m, i, true)
			}
		}
		for i := 0; i < ls.nbNT; i++ {
			p := ls.goTo(l, i, false)
			if p != nil {
				ls.insereLRSet(p, m, i, false)
				ls.dNT[i].utilNT = true
			}
		}
		m = m.suivant
	}
} // faitLR0

func (ls *syntaxical) firstX () {
	for {
		stop := true
		for i := 0; i < ls.nbNT; i++ {
			r := ls.dNT[i].regles
			for r != nil {
				g := r.dGram
				for {
					if g == nil {
						if !ls.dNT[i].epsFirst {
							stop = false
							ls.dNT[i].epsFirst = true
						}
						break
					}
					if g.tOuNT {
						if !ls.dNT[i].first.In(g.num) {
							stop = false
							ls.dNT[i].first.Incl(g.num)
						}
						break
					}
					if !ls.dNT[g.num].first.Diff(ls.dNT[i].first).IsEmpty() {
						stop = false
						ls.dNT[i].first.Add(ls.dNT[g.num].first)
					}
					if !ls.dNT[g.num].epsFirst {
						break
					}
					g = g.suivant
				}
				r = r.suivant
			}
		}
		if stop {
			break
		}
	}
} // firstX

func (ls *syntaxical) first (g *gram, derT ensTok) ensTok {
	f := ls.initEnsTok()
	eps := true
	for eps && g != nil {
		if g.tOuNT {
			f.Incl(g.num)
			eps = false
		} else {
			f.Add(ls.dNT[g.num].first)
			eps = ls.dNT[g.num].epsFirst
		}
		g = g.suivant
	}
	if eps {
		f.Add(derT)
	}
	return f
} // first

func (ls *syntaxical) closure1 (s *lRItem) *lRItem {
	s = ls.copieLRItem(s)
	for {
		stop := true
		t := s
		for t != nil {
			if !(t.posP == nil || t.posP.tOuNT) {
				f := ls.first(t.posP.suivant, t.look)
				l := ls.dNT[t.posP.num].regles
				for l != nil {
					var q *lRItem = nil
					p := s
					for p != nil && !(p.posN == 1 && p.regleP == l) {
						q = p
						p = p.suivant
					}
					if p == nil {
						p = &lRItem{regleP: l, posP: l.dGram, posN: 1, look: ls.initEnsTok()}
						q.suivant = p
					}
					if !f.Diff(p.look).IsEmpty() {
						stop = false
						p.look.Add(f)
					}
					l = l.suivant
				}
			}
			t = t.suivant
		}
		if stop {
			break
		}
	}
	return s
} // closure1

func (ls *syntaxical) tabuleLook () {
	s := ls.lRS
	for s != nil {
		i := s.set
		for i != nil {
			f := i.look
			i.look = ls.initEnsTok()
			i.look.Incl(1)
			p := i.suivant
			i.suivant = nil
			j := ls.closure1(i)
			i.suivant = p
			i.look = f
			k := j
			for k != nil {
				if k.posP != nil {
					t := s.goTo
					for !(t.tOuNT == k.posP.tOuNT && t.tNT == k.posP.num) {
						t = t.suivant
					}
					l := t.trans.set
					for !(l.regleP == k.regleP && l.posP == k.posP.suivant) {
						l = l.suivant
					}
					l.look.Add(k.look)
					if k.look.In(1) {
						l.look.Excl(1)
						i.propag = &lRPropag{suivant: i.propag, to: l}
					}
				}
				k = k.suivant
			}
			i = i.suivant
		}
		s = s.suivant
	}
} // tabuleLook

func (ls *syntaxical) faitLR1 () {
	ls.lRS.set.look.Incl(ls.termine)
	ls.tabuleLook()
	for {
		stop := true
		s := ls.lRS
		for s != nil {
			i := s.set
			for i != nil {
				p := i.propag
				for p != nil {
					if !i.look.Diff(p.to.look).IsEmpty() {
						stop = false
						p.to.look.Add(i.look)
					}
					p = p.suivant
				}
				i = i.suivant
			}
			s = s.suivant
		}
		if stop {
			break
		}
	}
} // faitLR1

func (ls *syntaxical) fabriqueSynt () {
	ls.faitLR0()
	ls.firstX()
	ls.faitLR1()
} // fabriqueSynt

func (ls *syntaxical) creeSynt (c *compiler) {

	var dump bool

	type (
		
		actionT struct {
			quoiT int8
			auxT,
			numRT,
			precT int
			assocT bool
		}
		
		actionTab []actionT
		
		recupT struct {
			suivant *recupT
			etatDep,
			nTGoto int
		}
		
		recupTab = []*recupT
	)
	
	ecrisItemPtr := func (i *lRItem) {
		for i != nil {
			r := i.regleP
			ls.lx.so.sString(ls.ecrisNonTerm(r.gNT))
			ls.lx.so.sString(" = ")
			g := r.dGram
			for j := 2; j <= i.posN; j++ {
				if g.tOuNT {
					ls.lx.so.sString(ls.lx.ecrisTerm(g.num))
				} else {
					ls.lx.so.sString(ls.ecrisNonTerm(g.num))
				}
				ls.lx.so.sString(" ")
				g = g.suivant
			}
			ls.lx.so.sString("... ")
			for g != nil {
				if g.tOuNT {
					ls.lx.so.sString(ls.lx.ecrisTerm(g.num))
				} else {
					ls.lx.so.sString(ls.ecrisNonTerm(g.num))
				}
				ls.lx.so.sString(" ")
				g = g.suivant
			}
			ls.lx.so.sString(";")
			ls.lx.so.sLn()
			i = i.suivant
			if i != nil {
				ls.lx.so.sString(ls.lx.so.sMap("Sor"))
				ls.lx.so.sLn()
			}
		}
	} // ecrisItemPtr
	
	conflit := func (nS, term int, act1, act2 int8, aux1, aux2 int) {
	
		ecrisItem := func (nS int) {
			l := ls.lRS
			for l.numSet != nS {
				l = l.suivant
			}
			ecrisItemPtr(l.set)
		} // ecrisItem
	
		// conflit
		if dumping {
			dump = true
		}
		ls.yAAttention = true
		ls.lx.so.sString(ls.lx.so.sMap("AnWarning"))
		ls.lx.so.sLn(); ls.lx.so.sLn()
		if dumping {
			ls.lx.so.sString("Dans l'état ")
			ls.lx.so.sString(cardToStr(nS))
			ls.lx.so.sString(", lorsqu'on a lu :")
		} else {
			ls.lx.so.sString(ls.lx.so.sMap("SHaveRead"))
		}
		ls.lx.so.sLn(); ls.lx.so.sLn()
		ecrisItem(nS)
		ls.lx.so.sLn()
		ls.lx.so.sString(ls.lx.so.sMap("SIfRead"))
		ls.lx.so.sString(" ")
		ls.lx.so.sString(ls.lx.ecrisTerm(term))
		ls.lx.so.sString(" ;")
		ls.lx.so.sLn(); ls.lx.so.sLn()
		if act1 == deplaceS {
			ls.lx.so.sString(ls.lx.so.sMap("SMustRead"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ecrisItem(aux1)
		} else {
			ls.lx.so.sString(ls.lx.so.sMap("SMustUnderstand"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.ecrisRegle(aux1)
		}
		ls.lx.so.sLn()
		ls.lx.so.sString(ls.lx.so.sMap("Sor"))
		ls.lx.so.sString(" ")
		if act2 == deplaceS {
			ls.lx.so.sString(ls.lx.so.sMap("SMustRead"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ecrisItem(aux2)
		} else {
			ls.lx.so.sString(ls.lx.so.sMap("SMustUnderstand"))
			ls.lx.so.sLn(); ls.lx.so.sLn()
			ls.ecrisRegle(aux2)
		}
		ls.lx.so.sLn()
		ls.lx.so.sString(ls.lx.so.sMap("SFirst"))
		ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
	} // conflit

	numeroteEtats := func () {
		c.nbEtatsSynt = 0
		l := ls.lRS
		for l != nil {
			l.numSet = c.nbEtatsSynt
			c.nbEtatsSynt++
			l = l.suivant
		}
	} // numeroteEtats

	creeTables := func () {
		c.actionSynt = make(actionsSynt, c.nbEtatsSynt)
		c.nbNonTSynt = 0
		c.nbRegleSynt = 0
		for i := 0; i < ls.nbNT; i++ {
			if ls.dNT[i].utilNT {
				ls.dNT[i].numerNT = c.nbNonTSynt
				c.nbNonTSynt++
				r := ls.dNT[i].regles
				for r != nil {
					r.nouvNumR = c.nbRegleSynt
					c.nbRegleSynt++
					l := r.actions
					for l != nil {
						p := l.actionF.params
						for p != nil {
							if p.numTNT > 0 {
								g := r.dGram
								for j := 2; j <= p.numTNT; j++ {
									g = g.suivant
								}
								if g.tOuNT {
									c.toksLex[g.num].valUt = true
								}
							}
							p = p.suivant
						}
						l = l.suivant
					}
					r = r.suivant
				}
			} else {
				ls.yAAttention = true
				ls.lx.so.sString(ls.lx.so.sMap("AnWarning"))
				ls.lx.so.sLn(); ls.lx.so.sLn()
				ls.lx.so.sString(ls.lx.so.sMap("SNeverUsedNonT", ls.ecrisNonTerm(i)))
				ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
				r := ls.dNT[i].regles
				for r != nil {
					r.nouvNumR = 0
					r = r.suivant
				}
			}
		}
		c.gotoSynt = make(gotosSynt, c.nbNonTSynt)
		c.regleSynt = make(reglesSynt, c.nbRegleSynt)
	} // creeTables

	creeAux := func () (t actionTab, test card, rec recupTab) {
		t = make(actionTab, c.nbToksLex)
		test = make(card, c.nbEtatsSynt + c.nbRegleSynt)
		rec = make(recupTab, c.nbToksLex)
		return
	} // creeAux

	parcoursEtats := func (t actionTab, test card, rec recupTab) {
		
		creeReduit := func (nS int, m *lRItem, t actionTab) {
			if m.posP == nil {
				for i := 0; i < c.nbToksLex; i++ {
					if m.look.In(i) {
						var a int8
						if m.regleP.numR == 0 {
							a = accepteS
						} else {
							a = reduitS
						}
						if t[i].quoiT == erreurS {
							t[i].quoiT = a
							r := m.regleP
							t[i].auxT = r.nouvNumR
							t[i].numRT = r.numR
							t[i].precT = r.precR
							t[i].assocT = r.assocR
						} else if t[i].numRT <= m.regleP.numR {
							conflit(nS, i, t[i].quoiT, a, t[i].numRT, m.regleP.numR)
						} else {
							conflit(nS, i, a, t[i].quoiT, m.regleP.numR, t[i].numRT)
							t[i].quoiT = a
							r := m.regleP
							t[i].auxT = r.nouvNumR
							t[i].numRT = r.numR
							t[i].precT = r.precR
							t[i].assocT = r.assocR
						}
					}
				}
			}
		} // creeReduit
	
		creeRec := func (nS int, m *lRItem, rec recupTab) {
			if m.posN == 1 && ls.dNT[m.regleP.gNT].desc.mark {
				for i := 0; i < c.nbToksLex; i++ {
					if m.look.In(i) {
						var q *recupT = nil
						s := rec[i]
						for s != nil && s.etatDep != nS {
							q = s
							s = s.suivant
						}
						if s == nil {
							s = &recupT{etatDep: nS, nTGoto: m.regleP.gNT}
							if q == nil {
								rec[i] = s
							} else {
								q.suivant = s
							}
						}
					}
				}
			}
		} // creeRec
	
		creeDeplace := func (nS int, g *lRGoto, t actionTab) {
			for g != nil {
				if g.tOuNT {
					if t[g.tNT].quoiT == erreurS {
						t[g.tNT].quoiT = deplaceS
						t[g.tNT].auxT = g.trans.numSet
					} else {
						prTerm, _ := ls.lx.precedence(g.tNT)
						if t[g.tNT].precT > prTerm || t[g.tNT].precT == prTerm && !t[g.tNT].assocT {
							conflit(nS, g.tNT, t[g.tNT].quoiT, deplaceS, t[g.tNT].numRT, g.trans.numSet)
						} else {
							conflit(nS, g.tNT, deplaceS, t[g.tNT].quoiT, g.trans.numSet, t[g.tNT].numRT)
							t[g.tNT].quoiT = deplaceS
							t[g.tNT].auxT = g.trans.numSet
						}
					}
				}
				g = g.suivant
			}
		} // creeDeplace
	
		tailleAction := func (t actionTab, test card) (int, int) {
			nbErr := 0
			for i := 0; i < c.nbEtatsSynt + c.nbRegleSynt; i++ {
				test[i] = 0
			}
			ancTest := c.nbEtatsSynt + c.nbRegleSynt + 1
			for i := 0; i < c.nbToksLex; i++ {
				switch t[i].quoiT {
				case deplaceS:
					if t[i].auxT + 1 != ancTest {
						ancTest = t[i].auxT + 1
						test[t[i].auxT]++
					}
				case reduitS:
					if c.nbEtatsSynt + t[i].auxT + 1 != ancTest {
						ancTest = c.nbEtatsSynt + t[i].auxT + 1
						test[c.nbEtatsSynt + t[i].auxT]++
					}
				case accepteS:
					if c.nbEtatsSynt + 1 != ancTest {
						ancTest = c.nbEtatsSynt + 1
						test[c.nbEtatsSynt]++
					}
				case erreurS:
					if ancTest != 0 {
						ancTest = 0
						nbErr++
					}
				}
			}
			ind := - 1; sup := nbErr
			for i := 0; i < c.nbEtatsSynt + c.nbRegleSynt; i++ {
				if test[i] > sup {
					ind = i
					sup = test[i]
				}
			}
			var n int
			if ind == - 1 {
				n = 1
			} else {
				n = nbErr
			}
			for i := 0; i < c.nbEtatsSynt + c.nbRegleSynt; i++ {
				if i == ind {
					n++
				} else {
					n += test[i]
				}
			}
			return n, ind
		} // tailleAction
	
		creeAction := func (actions actSynt, t actionTab, ind, nbT int) {
			n := - 1
			ancTest := c.nbEtatsSynt + c.nbRegleSynt
			for i := 0; i < c.nbToksLex; i++ {
				switch t[i].quoiT {
				case deplaceS:
					var x int
					if t[i].auxT == ind {
						ancTest = ind
						x = nbT - 1
					} else {
						if t[i].auxT != ancTest {
							ancTest = t[i].auxT
							n++
							actions[n].premTerm = i
						}
						x = n
					}
					actions[x].quoi = deplaceS
					actions[x].derTerm = i
					actions[x].aux = t[i].auxT
				case reduitS:
					var x int
					if c.nbEtatsSynt + t[i].auxT == ind {
						ancTest = ind
						x = nbT - 1
					} else {
						if c.nbEtatsSynt + t[i].auxT != ancTest {
							ancTest = c.nbEtatsSynt + t[i].auxT
							n++
							actions[n].premTerm = i
						}
						x = n
					}
					actions[x].quoi = reduitS
					actions[x].derTerm = i
					actions[x].aux = t[i].auxT
				case accepteS:
					var x int
					if c.nbEtatsSynt == ind {
						ancTest = ind
						x = nbT - 1
					} else {
						if c.nbEtatsSynt != ancTest {
							ancTest = c.nbEtatsSynt
							n++
							actions[n].premTerm = i
						}
						x = n
					}
					actions[x].quoi = accepteS
					actions[x].derTerm = i
				case erreurS:
					var x int
					if ind == - 1 {
						ancTest = ind
						x = nbT - 1
					} else {
						if ancTest != - 1 {
							ancTest = - 1
							n++
							actions[n].premTerm = i
						}
						x = n
					}
					actions[x].quoi = erreurS
					actions[x].derTerm = i
				}
			}
		} // creeAction
	
		//parcoursEtats
		l := ls.lRS
		for i := 0; i < c.nbEtatsSynt; i++ {
			for j := 0; j < c.nbToksLex; j++ {
				t[j].quoiT = erreurS
			}
			m := ls.closure1(l.set)
			p := m
			for p != nil {
				creeReduit(l.numSet, p, t)
				creeRec(l.numSet, p, rec)
				p = p.suivant
			}
			creeDeplace(l.numSet, l.goTo, t)
			a := c.actionSynt
			var j int
			a[l.numSet].nbT, j = tailleAction(t, test)
			a[l.numSet].actions = make(actSynt, a[l.numSet].nbT)
			creeAction(a[l.numSet].actions, t, j, a[l.numSet].nbT)
			l = l.suivant
		}
	} // parcoursEtats
	
	parcoursGoto := func (test card) {
		
		creeAtts := func (l *listeNT, gto *gotoSyntT) {
			gto.nbAtts = 0
			lNN := l.lAtt
			for lNN != nil {
				gto.nbAtts++
				lNN = lNN.suivant
			}
			gto.typsAt = make(card, gto.nbAtts)
			lNN = l.lAtt
			for i := 0; i < gto.nbAtts; i++ {
				gto.typsAt[i] = lNN.num
				lNN = lNN.suivant
			}
		} // creeAtts
	
		creeGoto := func (nNT int, gto *gotoSyntT, test card) {
			for i := 0; i < c.nbEtatsSynt; i++ {
				test[i] = 0
			}
			l := ls.lRS
			for l != nil {
				g := l.goTo
				for g != nil {
					if !g.tOuNT && g.tNT == nNT {
						test[g.trans.numSet]++
					}
					g = g.suivant
				}
				l = l.suivant
			}
			ind := - 1; n := 0
			for i := 0; i < c.nbEtatsSynt; i++ {
				if test[i] > n {
					ind = i
					n = test[i]
				}
			}
			gto.nbE = 0
			for i := 0; i < c.nbEtatsSynt; i++ {
				if i == ind {
					gto.nbE++
				} else {
					gto.nbE += test[i]
				}
			}
			gto.gotos = make(gotoS, gto.nbE)
			if gto.nbE != 0 {
				l := ls.lRS; n := 0
				for l != nil {
					g := l.goTo
					for g != nil {
						if !g.tOuNT && g.tNT == nNT {
							var x int
							if g.trans.numSet == ind {
								x = gto.nbE - 1
							} else {
								x = n
								n++
							}
							gto.gotos[x].depart = l.numSet
							gto.gotos[x].arrivee = g.trans.numSet
						}
						g = g.suivant
					}
					l = l.suivant
				}
			}
		} // creeGoto
	
		parcoursRegles := func (nNT int) {
			
			creeRegle := func (nNT int, r *regle) {
				c.regleSynt[r.nouvNumR].longueur = 0
				d := r.dGram
				for d != nil {
					c.regleSynt[r.nouvNumR].longueur++
					d = d.suivant
				}
				c.regleSynt[r.nouvNumR].nonTerm = nNT
			} // creeRegle
		
			creeSem := func (r *regle) {
				c.regleSynt[r.nouvNumR].nbAct = 0
				lA := r.actions
				for lA != nil {
					c.regleSynt[r.nouvNumR].nbAct++
					lA = lA.suivant
				}
				c.regleSynt[r.nouvNumR].act = make(actionsSem, c.regleSynt[r.nouvNumR].nbAct)
				lA = r.actions
				s := c.regleSynt[r.nouvNumR].act
				for i := 0; i < c.regleSynt[r.nouvNumR].nbAct; i++ {
					a := lA.actionF
					s[i].sOrH = a.softOrHard
					if a.numNT == 0 {
						s[i].profG = 0
					} else {
						s[i].profG = c.regleSynt[r.nouvNumR].longueur + 1 - a.numNT
					}
					s[i].attG = a.numAtt
					s[i].fonc = a.numFct
					s[i].nbPars = 0
					lAt := a.params
					for lAt != nil {
						s[i].nbPars++
						lAt = lAt.suivant
					}
					s[i].pars = make(params, s[i].nbPars)
					lAt = a.params
					for j := 0; j < s[i].nbPars; j++ {
						if lAt.numTNT == 0 {
							s[i].pars[j].profD = 0
						} else {
							s[i].pars[j].profD = c.regleSynt[r.nouvNumR].longueur + 1 - lAt.numTNT
						}
						s[i].pars[j].attD = lAt.numAttrib
						lAt = lAt.suivant
					}
					lA = lA.suivant
				}
			} // creeSem
		
			//parcoursRegles
			r := ls.dNT[nNT].regles
			for r != nil {
				creeRegle(ls.dNT[nNT].numerNT, r)
				creeSem(r)
				r = r.suivant
			}
		} // parcoursRegles
	
		//parcoursGoto
		for i := 0; i < ls.nbNT; i++ {
			if ls.dNT[i].utilNT {
				creeAtts(ls.dNT[i].desc, &c.gotoSynt[ls.dNT[i].numerNT])
				creeGoto(i, &c.gotoSynt[ls.dNT[i].numerNT], test)
				parcoursRegles(i)
			}
		}
	} // parcoursGoto
	
	creeRec := func (rec recupTab) {
		
		affRec := func () {
			for i := 0; i < ls.nbNT; i++ {
				b := true
				for j := 0; j < c.nbRecTerms; j++ {
					k := 0
					for {
						if k >= c.recTerms[j].nbEtats {
							break
						}
						if c.recTerms[j].recEtat[k].nTGoto == i {
							if b {
								b = false
								ls.lx.so.sString(ls.lx.so.sMap("AnRemark"))
								ls.lx.so.sLn(); ls.lx.so.sLn()
								ls.lx.so.sString(ls.lx.so.sMap("SResumeAfter", ls.ecrisNonTerm(i), ls.lx.ecrisTerm(c.recTerms[j].numTerm)))
							} else {
								ls.lx.so.sString(ls.lx.so.sMap("SorAfter", ls.lx.ecrisTerm(c.recTerms[j].numTerm)))
							}
							break
						}
						k++
					}
				}
				if !b {
					ls.yAAffichage = true
					ls.lx.so.sString(".")
					ls.lx.so.sLn(); ls.lx.so.sLn(); ls.lx.so.sLn()
				}
			}
		} // affRec
	
		//creeRec
		c.nbRecTerms = 0
		for i := 0; i < c.nbToksLex; i++ {
			if rec[i] != nil {
				c.nbRecTerms++
			}
		}
		c.recTerms = make(recTerms, c.nbRecTerms)
		if c.nbRecTerms != 0 {
			j := - 1
			for i := 0; i < c.nbToksLex; i++ {
				if rec[i] != nil {
					j++
					s := rec[i]
					c.recTerms[j].numTerm = i
					c.recTerms[j].nbEtats = 0
					for s != nil {
						c.recTerms[j].nbEtats++
						s = s.suivant
					}
					c.recTerms[j].recEtat = make(recEtats, c.recTerms[j].nbEtats)
					s = rec[i]
					for k := 0; k < c.recTerms[j].nbEtats; k++ {
						c.recTerms[j].recEtat[k].etatDep = s.etatDep
						c.recTerms[j].recEtat[k].nTGoto = s.nTGoto
						s = s.suivant
					}
				}
			}
			affRec()
			for j := 0; j < c.nbRecTerms; j++ {
				for k := 0; k < c.recTerms[j].nbEtats; k++ {
					c.recTerms[j].recEtat[k].nTGoto = ls.dNT[c.recTerms[j].recEtat[k].nTGoto].numerNT
				}
			}
		}
	} // creeRec

	dumpF := func () {
		ls.lx.so.sString("Automate syntaxique :")
		ls.lx.so.sLn(); ls.lx.so.sLn()
		l := ls.lRS
		for l != nil {
			ls.lx.so.sString("Etat ")
			ls.lx.so.sString(cardToStr(l.numSet))
			ls.lx.so.sLn()
			ecrisItemPtr(l.set)
			ls.lx.so.sLn()
			if l.goTo != nil {
				g := l.goTo
				for{
					if g.tOuNT {
						ls.lx.so.sString(ls.lx.ecrisTerm(g.tNT))
					} else {
						ls.lx.so.sString(ls.ecrisNonTerm(g.tNT))
					}
					ls.lx.so.sString(" : ")
					ls.lx.so.sString(cardToStr(g.trans.numSet))
					ls.lx.so.sLn()
					g = g.suivant
					if g == nil {
						break
					}
				}
				ls.lx.so.sLn()
			}
			i := l.set
			n := 0; b2 := false
			for i != nil {
				n++
				if i.posP == nil {
					b2 = true
					ls.lx.so.sString(cardToStr(n))
					ls.lx.so.sString(" : ")
					b1 := false
					for j := 0; j < c.nbToksLex; j++ {
						if i.look.In(j) {
							if b1 {
								ls.lx.so.sString(", ")
							} else {
								b1 = true
							}
							ls.lx.so.sString(ls.lx.ecrisTerm(j))
						}
					}
					ls.lx.so.sString("."); ls.lx.so.sLn()
				}
				i = i.suivant
			}
			if b2 {
				ls.lx.so.sLn()
			}
			l = l.suivant
		}
	} // dumpF

	//creeSynt
	if dumping {
		dump = false
	}
	numeroteEtats()
	creeTables()
	t, test, rec := creeAux()
	parcoursEtats(t, test, rec)
	parcoursGoto(test)
	creeRec(rec)
	if dumping && dump {
		dumpF()
	}
} // creeSynt
