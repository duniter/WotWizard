/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelInterface is part of the Babel subsystem, a compiler compiler. BabelInterface is the programming interface corresponding (and called by) the user interface BabelBabel. It builds a tbl file from a definition document.

import (
	
	C	"babel/compil"
	M	"util/misc"

)

const (
	
	// Indicates that compilation was done without display of comments.
	WithoutDisp = iota
	
	// Indicates the production of remarks during compilation ...
	Remarks
	
	// ... or of warnings ...
	Warnings
	
	// ... or of error messages.
	Errors

)

type (
	
	Facer interface {
		
		// Method for writing a string in the text of comments.
		String (s string)
		
		// Method for writing an end of line in the text of comments.
		Ln ()
		
		// Maps an index text to a more lengthy one, replacing ^0, ^1, etc... occurrences by p's values, without any modification of the latter strings. Used to make explicit or to translate error messages.
		Map (index string, p ... string) string
		
		// Method for writing in the binary file; i: next integer to write.
		BinInt (i int32)
		
		// Method for reading the parsed text; ch is the next parsed character.
		Read () (ch rune, cLen int)
		
		// Returns the running position in the parsed text (any origin is allowed).
		Pos () int
		
		// Moves the running position in the parsed text to pos (the origin must be the same than for Pos).
		SetPos (pos int)
	
	}
	
	Face struct {
		Facer
		c *compiler
		deb int
	}
	
	sorties struct {
		f *Face
	}
	
	semI struct {
		*sem
		*Face
		so *sorties
	}

)

var (
	
	comp *C.Compiler

)

func (f *Face) binInt (i int) {
	f.BinInt(int32(i))
}

func (f *Face) binByte (b int8) {
	f.BinInt(int32(b))
} // binByte

func (f *Face) binBool (b bool) {
	if b {
		f.BinInt(1)
	} else {
		f.BinInt(0)
	}
} // binBool

func (f *Face) binChar (c rune) {
	f.BinInt(int32(c))
} // binChar

func (f *Face) binStr (s string) {
	for _, c := range []rune(s) {
		f.binChar(c)
	}
	f.binChar('\x00')
} // binStr

func (f *Face) Int (n int) {
	f.String(cardToStr(n))
}

func (so *sorties) sString (s string) {
	so.f.String(s)
} // sString

func (so *sorties) sLn () {
	so.f.Ln()
} // sLn

func (so *sorties) sMap (index string, p ... string) string {
	return so.f.Map(index, p ...)
} // sMap

func (se *semI) Map (index string) string {
	return se.so.sMap(index)
} // Map

// Reports an error which takes place at position p, line l and column c in the parsed text. mes is the error message. The implemented procedure writes these informations with f.String, f.Ln and f.Map.
func (f *Face) Error (p, l, c int, mes string) {
	
	const (
		
		errorMark = "@"
	
	)
	
	f.String(f.Map("AnError"))
	f.Ln(); f.Ln()
	f.String(f.Map("ILine")); f.String(" ")
	f.Int(l)
	f.String(", ")
	f.String(f.Map("ICol")); f.String(" ")
	f.Int(c)
	f.String(".")
	f.Ln(); f.Ln()
	pos := f.Pos()
	f.SetPos(p - c + f.deb)
	ch, _ := f.Read()
	n := 0
	for (ch != C.EOF1) && (ch != C.EOF2) && (ch != C.EOL1) && (ch != C.EOL2) {
		n++
		ch, _ = f.Read()
	}
	if n > 0 {
		f.SetPos(p - c + f.deb)
		for i := 1; i < c; i++ {
			ch, _ := f.Read()
			f.String(string(ch))
		}
		f.String(errorMark)
		for i := c; i <= n; i++ {
			ch, _ := f.Read()
			f.String(string(ch))
		}
		f.Ln()
	}
	f.SetPos(pos)
	f.Ln()
	f.String(mes)
	f.Ln(); f.Ln(); f.Ln()
} // Error

// Compiles a definition document, builds from this document a compiler, writes possibly comments and returns in res a value (equal to WithoutDisp, Remarks, Warnings or Errors) indicating the state of compilation. If Errors is returned, the compiler was not created. If another value that WithoutDisp is returned, a text of comments were written by the Face.String and Face.Ln methods. The definition text is read by Face.Read, with the help of Face.Pos and Face.SetPos. name returns the name of the document, as written after the first keyword BABEL.
func (f *Face) CompComp () (name string, res int8) {
	f.deb = f.Pos()
	so := &sorties{f: f}
	se := new(sem)
	si := &semI{sem: se, Face: f, so: so}
	se.Compilation = C.NewCompilation(si)
	se.ls = &syntaxical{lx: &lexical{so: so}, yAAffichage: false, yAAttention: false}
	se.ls.lx.initTokL()
	se.ls.initSynt()
	if si.Compile(comp, false) && si.ls.verifSynt() && si.ls.verifSem() {
		si.ls.lx.fabriqueLex()
		f.c = si.ls.lx.creeLex()
		si.ls.fabriqueSynt()
		si.ls.creeSynt(f.c)
		name = si.nom
		if si.ls.yAAttention {
			res = Warnings
		} else if si.ls.yAAffichage {
			res = Remarks
		} else {
			res = WithoutDisp
		}
	} else {
		f.c = nil
		name = ""
		res = Errors
	}
	return
} // CompComp

// Writes the binary file, corresponding to the compiler created by Face.CompComp, using the Face.BinInt method.
func (f *Face) OutComp () {
	
	ecrisEtatsRedLex := func (eL etatsLex) {
		for _, e := range eL {
			f.binInt(e.recon)
			f.binInt(e.nbTrans)
			f.binInt(e.nbEps)
		}
		for _, e := range eL {
			t := e.transL
			for j := 0; j < e.nbEps; j++ {
				tt := t[j].(*gotoLexT)
				f.binInt(tt.goTo)
				f.binInt(tt.transit)
			}
			for j := e.nbEps; j < e.nbTrans; j++ {
				tt := t[j].(*gotoLexC)
				f.binInt(tt.goTo)
				f.binChar(tt.premCar)
				f.binChar(tt.derCar)
			}
		}
	} // ecrisEtatsRedLex

	//OutComp
	M.Assert(f.c != nil, 20)
	f.binInt(f.c.nbToksLex)
	f.binInt(f.c.nbEtatsLex)
	f.binInt(f.c.profEtatsL)
	f.binInt(f.c.nbEtatsCom)
	f.binInt(f.c.profEtatsC)
	f.binInt(f.c.nbEtatsSynt)
	f.binInt(f.c.nbNonTSynt)
	f.binInt(f.c.nbRegleSynt)
	f.binInt(f.c.nbRecTerms)
	tt := f.c.toksLex
	for _, t := range tt {
		f.binBool(t.utile)
		f.binBool(t.valUt)
	}
	for _, t := range tt {
		if t.nom == "" {
			f.binInt(0)
		} else {
			f.binInt(len(t.nom) + 1)
			f.binStr(t.nom)
		}
	}
	ecrisEtatsRedLex(f.c.etatsLex)
	ecrisEtatsRedLex(f.c.etatsCom)
	for _, as := range f.c.actionSynt {
		f.binInt(as.nbT)
	}
	for _, as := range f.c.actionSynt {
		for _, a := range as.actions {
			f.binByte(a.quoi)
			f.binInt(a.premTerm)
			f.binInt(a.derTerm)
			f.binInt(a.aux)
		}
	}
	for _, g := range f.c.gotoSynt {
		f.binInt(g.nbAtts)
		f.binInt(g.nbE)
	}
	for _, g := range f.c.gotoSynt {
		for _, ca := range g.typsAt {
			f.binInt(ca)
		}
		for _, gg := range g.gotos {
			f.binInt(gg.depart)
			f.binInt(gg.arrivee)
		}
	}
	for _, r := range f.c.regleSynt {
		f.binInt(r.longueur)
		f.binInt(r.nonTerm)
		f.binInt(r.nbAct)
	}
	for _, r := range f.c.regleSynt {
		for _, aS := range r.act {
			f.binBool(aS.sOrH)
			f.binInt(aS.profG)
			f.binInt(aS.attG)
			f.binInt(aS.fonc)
			f.binInt(aS.nbPars)
		}
		for _, aS := range r.act {
			for _, p := range aS.pars {
				f.binInt(p.profD)
				f.binInt(p.attD)
			}
		}
	}
	for _, r := range f.c.recTerms {
		f.binInt(r.numTerm)
		f.binInt(r.nbEtats)
	}
	for _, r := range f.c.recTerms {
		for _, e := range r.recEtat {
			f.binInt(e.etatDep)
			f.binInt(e.nTGoto)
		}
	}
} // OutComp

func NewFace (f Facer) *Face {
	return &Face{Facer: f}
} // NewFace

func init () {
	comp = initBabel()
} // init
