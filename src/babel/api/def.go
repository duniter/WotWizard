/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelDef is part of the Babel subsystem, a compiler compiler. BabelDef holds general definitions and methods.

import (
	
	M	"util/misc"
	S	"strconv"

)

const (

	eOF1 = '\x00'
	eOF2 = '\x1A' // End of file = eOF1 or eOF2

	eOL1 = '\x0D'
	eOL2 = '\x0A'
	
	// less than, equal, more than
	
	inf = -1; ega = 0; sup = + 1

)

type (
	
	// The following definitions are the same than those found in BabelCompil. See BabelCompil for further explanations
	
	tokLex struct { // A token
		nom string
		utile,
		valUt bool
	}
	
	toksLexT []tokLex
	
	/*
	gotoLex interface { // *gotoLexC or *gotoLexT
	}
	
	gotoLexC = struct {
		gotoF int
		premCar,
		derCar rune
	}
	
	gotoLexT struct {
		gotoF,
		transit int
	}
	*/
	
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
	
	card []int
	
	etatLex struct {
		recon,
		nbTrans,
		nbEps int
		transL transLex
	}
	
	etatsLex []etatLex

)

const (
	
	deplaceS = iota
	reduitS
	accepteS
	erreurS

)

type (
	
	actionS struct {
		quoi int8 // deplaceS, reduitS, accepteS ou erreurS
		premTerm,
		derTerm,
		aux int
	}
	
	actSynt []actionS
	
	actionSynt struct {
		 nbT int
		 actions actSynt
	 }
	
	actionsSynt []actionSynt
	
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
	
	gotosSynt []gotoSyntT
	
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
	
	regleSyntT struct {
		longueur,
		nonTerm,
		nbAct int
		act actionsSem
	}
	
	reglesSynt []regleSyntT
	
	recEtat struct {
		etatDep,
		nTGoto int
	}
	
	recEtats []recEtat
	
	recTerm struct {
		numTerm,
		nbEtats int
		recEtat recEtats
	}
	
	recTerms []recTerm
	
	compiler struct {
		nbToksLex int
		toksLex toksLexT
		nbEtatsLex,
		profEtatsL int
		etatsLex etatsLex
		nbEtatsCom,
		profEtatsC int
		etatsCom etatsLex
		nbEtatsSynt int
		actionSynt actionsSynt
		nbNonTSynt int
		gotoSynt gotosSynt
		nbRegleSynt int
		regleSynt reglesSynt
		nbRecTerms int
		recTerms recTerms
	}
	
	// A collection of inout user defined methods.
	
	sortieser interface {
		
		// Writes a string
		
		sString (s string)
		
		// Writes a new line
		
		sLn ()
		
		// Maps an index text to a more lengthy one. p... replaces, without mapping, instances of , respectively, ^0, ^1 and ^2 appearing in the replacement text.
		
		sMap (index string, p ... string) string
	
	}

)

// Return a slice of sS, beginning at p and of lenth l.

func slice (sS []rune, p, l int) []rune {
	M.Assert(p >= 0 && l >= 0 && p + l <= len(sS), 20)
	return sS[p:p + l]
} // slice

// Writes the integer n into the string s.

func cardToStr (n int) (s string) {
	M.Assert(n >= 0, 20)
	return S.Itoa(n)
} // cardToStr

// Return the value of the integer read from s and true if all is ok.

func strToCard (s string) (int, bool) {
	j := len(s) - 1
	hexa := s[j] == 'H'
	if hexa {
		s = s[:j]
	}
	for s[0] == 0 && len(s) > 1 {
		s = s[1:]
	}
	if hexa {
		s = "0x" + s
	}
	n, err := S.ParseInt(s, 0, 0)
	return int(n), err == nil && n >= 0
} // strToCard
