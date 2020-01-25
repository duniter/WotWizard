/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package main

// The module BabelBabel is part of the Babel subsystem, a compiler compiler. BabelBabel is the user interface which calls the programming interface module BabelInterface to build a tbl file from a definition document

import (
	
	A	"babel/api"
	B	"bytes"
	C	"babel/compil"
	F	"path/filepath"
	M	"util/misc"
	R	"util/resources"
	SM	"util/strMapping"
		"fmt"
		"io"
		"strings"
		"os"
		
	_	"babel/static"

)

const (
	
	sourcesPath = "bab"
	sourcesExt = ".txt"
	tablesPath = ""
	tablesExt = ".tbl"
	listExt = "List.txt"

)

type (
	
	Facer struct { //A.Facer
		f io.StringWriter
		w io.Writer
		r []rune
		size,
		pos int
	}

)

var (
	
	rsrcDir = R.FindDir()

)

func (f *Facer) Ln () {
	f.f.WriteString("\n")
} // Ln

func (f *Facer) String (s string) {
	f.f.WriteString(s)
} // String

func (f *Facer) BinInt (i int32) {
	b := make([]byte, 4)
	ui := uint32(i)
	for i := 0; i < 4; i++ {
		b[i] = byte(ui & 0xff)
		ui >>= 8
	}
	_, err := f.w.Write(b); M.Assert(err == nil, err, 100)
} // BinInt

func (f *Facer) Read () (ch rune, cLen int) {
	if f.pos >= f.size {
		ch = C.EOF1
	} else {
		ch = f.r[f.pos]
		f.pos++
	}
	cLen = 1
	return
} // Read

func (f *Facer) Pos () int {
	return f.pos
} // Pos

func (f *Facer) SetPos (pos int) {
	f.pos = M.Min(pos, f.size)
} // SetPos

func (f *Facer) Map (index string, p ... string) string {
	return SM.Map("#babel:" + index, p ...)
} // Map

func decompNom (nom, midPath string) (path, name string) {
	path, name = F.Split(nom)
	path = F.Join(path, midPath)
	name = string([]byte(name)[:len(name) - len(F.Ext(name))])
	return
} // decompNom

func compile1Text (input io.ReadSeeker, output io.StringWriter, binOutput io.Writer) (path, name string, ok bool) {
	n, err := input.Seek(0, io.SeekEnd); M.Assert(err == nil, err, 100)
	_, err = input.Seek(0, io.SeekStart); M.Assert(err == nil, err, 101)
	b := make([]byte, n)
	_, err = io.ReadFull(input, b); M.Assert(err == nil, err, 102)
	r := []rune(string(b))
	fr := &Facer{f: output, w: binOutput, r: r, size: len(r), pos: 0}
	f := A.NewFace(fr)
	nom, res := f.CompComp()
	path, name = decompNom(nom, tablesPath)
	if res != A.Errors {
		f.OutComp()
	}
	var mes string
	switch res {
	case A.WithoutDisp:
		mes = SM.Map("#babel:BOk")
	case A.Remarks:
		mes = SM.Map("#babel:BRem")
	case A.Warnings:
		mes = SM.Map("#babel:BWarning")
	case A.Errors:
		mes = SM.Map("#babel:BError")
	}
	fmt.Fprintln(os.Stderr, mes)
	ok = res != A.Errors
	if !ok {
		fmt.Fprintln(os.Stderr)
	}
	return
} // compile1Text

func compile1Name (nom string) bool {
	path, name := decompNom(nom, sourcesPath)
	fmt.Fprint(os.Stderr, "\t", nom, ": ")
	target := F.Join(rsrcDir, path, name + sourcesExt)
	input, err := os.Open(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, target, SM.Map("#babel:BNotFound"))
		return false
	}
	defer input.Close()
	output := new(strings.Builder)
	binOutput := new(B.Buffer)
	path, name, ok := compile1Text(input, output, binOutput)
	if ok {
		target := F.Join(rsrcDir, path)
		os.MkdirAll(target, 0777)
		f, err := os.Create(F.Join(target, name + tablesExt))
		if err == nil {
			defer f.Close()
			_, err = binOutput.WriteTo(f)
		}
		if err != nil {
			fmt.Fprintln(output)
			fmt.Fprint(output, SM.Map("#babel:BNoCreate"))
			ok = false
		}
		if output.Len() > 0 {
			f, err := os.Create(F.Join(target, name + listExt))
			if err == nil {
				defer f.Close()
				fmt.Fprint(f, output.String())
			}
		}
	} else {
		fmt.Fprint(os.Stderr, output.String())
	}
	return ok
} // compile1Name

func compileList (list []string) {
	for _, name := range list {
		compile1Name(name)
	}
} // compileList

func main () {
	compileList(os.Args[1:])
}
