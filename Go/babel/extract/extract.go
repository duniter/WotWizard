/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package main

import (
	"fmt"
	"os"
	"bufio"
)
	
const (
	deplaceS = iota	// Shift action 
	reduitS	// Reduce action 
	accepteS	// Accepts action 
	erreurS	// Error 
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
	
	gotoLex struct {
		goTo int
	}
	
	gotoLexC struct {
		gotoLex
		premCar,
		derCar rune
	}
	
	gotoLexT struct {
		gotoLex
		transit int
	}
	
	transLex []Anyptr	// *gotoLex, *gotoLexC or *gotoLexT

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
	
// Reads a compiler, i.e. the binary file built by module BabelBabel. 

func readCompiler (readInt func () int) *Compiler {
	
	const
		eOS = 0x0

	readByte := func () int8 {
		return int8(readInt())
	}

	readBool := func () bool {
		i := readInt()
		if !(i == 0 || i == 1) {panic(0)}
		return i == 1
	}

	readChar := func () rune {
		return rune(readInt())
	}

	readString := func () (s string) {
		s = ""
		c := readChar()
		for c != eOS {
			s += string(c)
			c = readChar()
		}
		return
	}
	
	lisEtatsRedLex := func  (nE int) (eL etatsLexT) {
		eL = make(etatsLexT, nE)
		for i := 0; i < nE; i++ {
			eL[i].recon = readInt()
			eL[i].nbTrans = readInt()
			eL[i].nbEps = readInt()
		}
		for i := 0; i < nE; i++ {
			if eL[i].nbTrans > 0 {
				t := make(transLex, eL[i].nbTrans)
				for j := 0; j < eL[i].nbEps; j++ {
					gT := new(gotoLexT)
					gT.goTo = readInt()
					gT.transit = readInt()
					t[j] = gT
				}
				for j := eL[i].nbEps; j < eL[i].nbTrans; j++ {
					gC := new(gotoLexC)
					gC.goTo = readInt()
					gC.premCar = readChar()
					gC.derCar = readChar()
					if !(gC.premCar <= gC.derCar) {panic(1)}
					t[j] = gC
				}
				eL[i].transL = t
			}
		}
		return
	}
	
	c := new(Compiler)
	c.nbToksLex = readInt()
	c.nbEtatsLex = readInt()
	c.profEtatsL = readInt()
	c.nbEtatsCom = readInt()
	c.profEtatsC = readInt()
	c.nbEtatsSynt = readInt()
	c.nbNonTSynt = readInt()
	c.nbRegleSynt = readInt()
	c.nbRecTerms = readInt()
	t := make(toksLexT, c.nbToksLex)
	for i := 0; i < c.nbToksLex; i++ {
		t[i].utile = readBool()
		t[i].valUt = readBool()
	}
	for i := 0; i < c.nbToksLex; i++ {
		k := readInt()
		if k != 0 {
			t[i].nom = readString()
		} else {
			t[i].nom = ""
		}
	}
	c.toksLex = t
	c.etatsLex = lisEtatsRedLex(c.nbEtatsLex)
	c.etatsCom = lisEtatsRedLex(c.nbEtatsCom)
	c.actionSynt = make(actionsSynt, c.nbEtatsSynt)
	for i := 0; i < c.nbEtatsSynt; i++ {
		c.actionSynt[i].nbT = readInt()
	}
	for i := 0; i < c.nbEtatsSynt; i++ {
		a := make(actSynt, c.actionSynt[i].nbT)
		for j := 0; j < c.actionSynt[i].nbT; j++ {
			a[j].quoi = readByte()
			a[j].premTerm = readInt()
			a[j].derTerm = readInt()
			a[j].aux = readInt()
		}
		c.actionSynt[i].actions = a
	}
	c.gotoSynt = make(gotosSyntT, c.nbNonTSynt)
	for i := 0; i < c.nbNonTSynt; i++ {
		c.gotoSynt[i].nbAtts = readInt()
		c.gotoSynt[i].nbE = readInt()
	}
	for i := 0; i < c.nbNonTSynt; i++ {
		if c.gotoSynt[i].nbAtts > 0 {
			ca := make(card, c.gotoSynt[i].nbAtts)
			for j := 0; j < c.gotoSynt[i].nbAtts; j++ {
				ca[j] = readInt()
			}
			c.gotoSynt[i].typsAt = ca
		}
		if c.gotoSynt[i].nbE > 0 {
			g := make(gotoS, c.gotoSynt[i].nbE)
			for j := 0; j < c.gotoSynt[i].nbE; j++ {
				g[j].depart = readInt()
				g[j].arrivee = readInt()
			}
			c.gotoSynt[i].gotos = g
		}
	}
	c.regleSynt = make(reglesSynt, c.nbRegleSynt)
	for i := 0; i < c.nbRegleSynt; i++ {
		c.regleSynt[i].longueur = readInt()
		c.regleSynt[i].nonTerm = readInt()
		c.regleSynt[i].nbAct = readInt()
	}
	for i := 0; i < c.nbRegleSynt; i++ {
		if c.regleSynt[i].nbAct > 0 {
			aS := make(actionsSem, c.regleSynt[i].nbAct)
			for j := 0; j < c.regleSynt[i].nbAct; j++ {
				aS[j].sOrH = readBool()
				aS[j].profG = readInt()
				aS[j].attG = readInt()
				aS[j].fonc = readInt()
				aS[j].nbPars = readInt()
			}
			for j := 0; j < c.regleSynt[i].nbAct; j++ {
				if aS[j].nbPars > 0 {
					p := make(params, aS[j].nbPars)
					for m := 0; m < aS[j].nbPars; m++ {
						p[m].profD = readInt()
						p[m].attD = readInt()
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
			c.recTerms[i].numTerm = readInt()
			c.recTerms[i].nbEtats = readInt()
		}
		for i := 0; i < c.nbRecTerms; i++ {
			r := make(recEtats, c.recTerms[i].nbEtats)
			for j := 0; j < c.recTerms[i].nbEtats; j++ {
				r[j].etatDep = readInt()
				r[j].nTGoto = readInt()
			}
			c.recTerms[i].recEtat = r
		}
	}
	return c
}

func writeCompiler (c *Compiler) {

	var (
		rien = "nothing"
		sansNom = "unnamed"
		utile = "used"
		inutile = "unused"
		valeurU = "used value"
		valeurNonU = "unused value"
		copie = "copy function"
		douce = "soft function"
		dure = "hard function"
	)
	
	writeChar := func (c rune) {
		if c < 0x20 {
			fmt.Print("chr(", c, ")")
		} else {
			fmt.Print(string(c));
		}
	}

	ecrisEtatsRedLex := func (nT, nE, pE int, eL etatsLexT, nN, pN, lN string) {
		fmt.Println(nN, " = ", nE)
		fmt.Println(pN, " = ", pE)
		for i := 0; i < nE; i++ {
			fmt.Println("\t", lN, " ", i)
			fmt.Print("\t\trecon = ")
			if eL[i].recon < nT {
				fmt.Println(eL[i].recon)
			} else {
				fmt.Println(rien)
			}
			fmt.Println("\t\tnbTrans = ", eL[i].nbTrans)
			fmt.Println("\t\tnbEps = ", eL[i].nbEps)
			for j := 1; j <= eL[i].nbEps; j++ {
				fmt.Println("\t\t\ttransL ", j - 1)
				fmt.Println("\t\t\t\tgoto = ", eL[i].transL[j - 1].(*gotoLexT).goTo)
				fmt.Println("\t\t\t\ttransit = ", eL[i].transL[j - 1].(*gotoLexT).transit)
			}
			for j := eL[i].nbEps + 1; j <= eL[i].nbTrans; j++ {
				fmt.Println("\t\t\ttransL ", j - 1)
				fmt.Println("\t\t\t\tgoto = ", eL[i].transL[j - 1].(*gotoLexC).goTo)
				fmt.Print("\t\t\t\tpremCar = "); writeChar(eL[i].transL[j - 1].(*gotoLexC).premCar); fmt.Println()
				fmt.Print("\t\t\t\tderCar = "); writeChar(eL[i].transL[j - 1].(*gotoLexC).derCar); fmt.Println()
			}
		}
	}

	fmt.Println("nbToksLex = ", c.nbToksLex)
	for i := 0; i < c.nbToksLex; i++ {
		fmt.Println("\ttoksLex ", i)
		fmt.Print("\t\t") 
		if c.toksLex[i].nom == "" {
			fmt.Println(sansNom)
		} else {
			fmt.Println(c.toksLex[i].nom)
		}
		fmt.Print("\t\t") 
		if c.toksLex[i].utile {
			fmt.Println(utile)
		} else {
			fmt.Println(inutile)
		}
		fmt.Print("\t\t")
		if c.toksLex[i].valUt {
			fmt.Println(valeurU)
		} else {
			fmt.Println(valeurNonU)
		}
	}
	ecrisEtatsRedLex(c.nbToksLex, c.nbEtatsLex, c.profEtatsL, c.etatsLex, "nbEtatsLex", "profEtatsL", "etatsLex")
	ecrisEtatsRedLex(3, c.nbEtatsCom, c.profEtatsC, c.etatsCom, "nbEtatsCom", "profEtatsC", "etatsCom")
	fmt.Println("nbEtatsSynt = ", c.nbEtatsSynt)
	for i := 0; i < c.nbEtatsSynt; i++ {
		fmt.Println("\tactionSynt ", i)
		fmt.Println("\t\tnbT = ", c.actionSynt[i].nbT)
		for j := 0; j < c.actionSynt[i].nbT; j++ {
			fmt.Println("\t\t\tactions ", j + 1)
			fmt.Print("\t\t\t\tquoi = ")
			switch c.actionSynt[i].actions[j].quoi {
				case deplaceS:
					 fmt.Println("deplaceS")
					 fmt.Println("\t\t\t\taux = ", c.actionSynt[i].actions[j].aux)
				case reduitS:
					 fmt.Println("reduitS")
					 fmt.Println("\t\t\t\taux = ", c.actionSynt[i].actions[j].aux)
				case accepteS:
					 fmt.Println("accepteS")
				case erreurS:
					 fmt.Println("erreurS")
			}
			if j < c.actionSynt[i].nbT - 1 {
				fmt.Println("\t\t\t\tpremTerm = ", c.actionSynt[i].actions[j].premTerm)
				fmt.Println("\t\t\t\tderTerm = ", c.actionSynt[i].actions[j].derTerm)
			}
		}
	}
	fmt.Println("nbNonTSynt = ", c.nbNonTSynt)
	for i := 0; i < c.nbNonTSynt; i++ {
		fmt.Println("\tgotoSynt ", i)
		fmt.Println("\t\tnbAtts = ", c.gotoSynt[i].nbAtts)
		for j := 0; j < c.gotoSynt[i].nbAtts; j++ {
			fmt.Println("\t\t\ttypsAt ", j + 1, " = ", c.gotoSynt[i].typsAt[j])
		}
		fmt.Println("\t\tnbE = ", c.gotoSynt[i].nbE)
		for j := 0; j < c.gotoSynt[i].nbE; j++ {
			fmt.Println("\t\t\tgotos ", j + 1)
			fmt.Println("\t\t\t\tdepart = ", c.gotoSynt[i].gotos[j].depart)
			fmt.Println("\t\t\t\tarrivee = ", c.gotoSynt[i].gotos[j].arrivee)
		}
	}
	fmt.Println("nbRegleSynt = ", c.nbRegleSynt)
	for i := 0; i < c.nbRegleSynt; i++ {
		fmt.Println("\tregleSynt ", i)
		fmt.Println("\t\tlongueur = ", c.regleSynt[i].longueur)
		fmt.Println("\t\tnonTerm = ", c.regleSynt[i].nonTerm)
		fmt.Println("\t\tnbAct = ", c.regleSynt[i].nbAct)
		for j := 0; j < c.regleSynt[i].nbAct; j++ {
			fmt.Println("\t\t\tact ", j + 1)
			fmt.Print("\t\t\t\t")
			if c.regleSynt[i].act[j].sOrH {
				if c.regleSynt[i].act[j].fonc == 0 {
					fmt.Println(copie)
				} else {
					fmt.Println(douce)
				}
			} else {
				fmt.Println(dure)
			}
			fmt.Println("\t\t\t\tprofG = ", c.regleSynt[i].act[j].profG)
			fmt.Println("\t\t\t\tattG = ", c.regleSynt[i].act[j].attG)
			if c.regleSynt[i].act[j].fonc > 0 {
				fmt.Println("\t\t\t\tfonc = ", c.regleSynt[i].act[j].fonc)
			}
			fmt.Println("\t\t\t\tnbPars = ", c.regleSynt[i].act[j].nbPars)
			for k := 0; k < c.regleSynt[i].act[j].nbPars; k++ {
				fmt.Println("\t\t\t\t\tpars ", k + 1)
				fmt.Println("\t\t\t\t\t\tprofD = ", c.regleSynt[i].act[j].pars[k].profD)
				fmt.Println("\t\t\t\t\t\tattD = ", c.regleSynt[i].act[j].pars[k].attD)
			}
		}
	}
	fmt.Println("nbRecTerms = ", c.nbRecTerms)
	for i := 0; i < c.nbRecTerms; i++ {
		fmt.Println("\trecTerms ", i + 1)
		fmt.Println("\t\tnumTerm = ", c.recTerms[i].numTerm)
		fmt.Println("\t\tnbEtats = ", c.recTerms[i].nbEtats)
		for j := 0; j < c.recTerms[i].nbEtats; j++ {
			fmt.Println("\t\t\trecEtat ", j + 1)
			fmt.Println("\t\t\t\tetatDep = ", c.recTerms[i].recEtat[j].etatDep)
			fmt.Println("\t\t\t\tnTGoto = ", c.recTerms[i].recEtat[j].nTGoto)
		}
	}
}


func main () {
	
	var
		r *bufio.Reader
	
	readInt := func () int {
		res := int32(0)
		p := uint(0)
		for i := 0; i < 4; i++ {
			n, err := r.ReadByte();
			if err != nil {panic(0)}
			res = res + int32(n) << p
			p += 8
		}
		return int(res)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {panic(0)}
	r = bufio.NewReader(f)
	c := readCompiler(readInt)
	writeCompiler(c)
}
