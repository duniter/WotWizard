/* 
Util: Utility tools.

Copyright (C) 2001…2019 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package extStrings

// Defines an abstract type of  extensible strings and implements it.

import (
	
	Avl	"util/avl"
	M	"util/misc"
		"bytes"

)

const
	
	eOS = 0

type (
	
	// String of rune(s)
	String interface { // String of characters.
		// Creates and returns a new reader of the String, positionned at 0.
		NewReader () Reader
		// Creates and returns a new writer of the String, positionned at its end.
		NewWriter () Writer
		// Returns the length of the String.
		Length () int
		// Returns the position of the first occurrence of target in the String, after the position beg. Returns -1 if not found.
		Pos (beg int, target string) int
		// Sets the value of the String to st.
		Set (st string)
		// Returns the concatenation of the String and of s.
		Cat (s String) String
		// Returns the String with src inserted at position pos. The String is not modified, nor src. If pos < 0, pos == 0 is assumed; if pos > String.Length(), pos == String.Length() is assumed.
		Insert (pos int, src String) String
		// Extracts and returns the substring of the String beginning at position pos and of length length.
		Extract (pos, length int) String
		// Copies and returns the String.
		Copy () String
		// Converts the String into a string.
		Convert () string
	}
	
	Reader interface { // Reader of string.
		// Indicates when end of string has been read.
		Eos () bool
		// Last character read.
		Char () rune
		// Returns the String Reader is reading.
		Base () String
		// Returns the position of Reader in the String.
		Pos () int
		// Sets the position of Reader at pos. pos must verify 0 <= pos <= Reader.Base().Length().
		SetPos (pos int)
		// Reads a character at the position of Reader and advances Reader to the next position. If  the end of the string has been reached, Eos() returns true and Char() returns eOS.
		Read ()
		// Does just as Reader.Read but returns the character read.
		ReadChar () rune
		// Reads at most length characters by repetition of Read().
		ReadString (length int) string
	}
	
	Writer interface { // Writer of string.
		// Returns the string Writer is writing.
		Base () String
		// Returns the position of Writer.
		Pos () int
		// Sets the position of Writer at pos. pos must verify 0 <= pos <= Writer.Base().Length().
		SetPos (pos int)
		// Writes the character c at the position of Writer and advances Writer to the next position. If  the end of the string has been reached, c is appended to the string.
		WriteChar (c rune)
		// Writes the string s by repetition of WriteChar().
		WriteString (s string)
	}
	
	Directory interface { // Factory for strings.
		// Creates a new empty String.	
		New () String
	}
	
	char rune
	
	stdString struct {
		chars *Avl.Tree
	}
	
	stdReader struct {
		eos bool
		char rune
		s *stdString
		p int
	}
	
	stdWriter struct {
		s *stdString
		p int
	}
	
	stdDirectory struct {
	}

)

var (
	
	stdDir Directory = &stdDirectory{}
	dir = stdDir

)

func (c char) Copy () Avl.Copier {
	return c
} //Copy

func (s *stdString) NewReader () Reader {
	return &stdReader{s: s, p: 0}
} //NewReader

func (s *stdString) NewWriter () Writer {
	return &stdWriter{s, s.Length()}
} //NewWriter

func (s *stdString) Length () int {
	return s.chars.NumberOfElems()
} //Length

func (s *stdString) Pos (beg int, target string) int {
	rs := []rune(target)
	m := len(rs)
	n := s.chars.NumberOfElems() - m + 1
	j := 0
	for (j < m) && (beg < n) {
		beg++
		e, _ := s.chars.Find(beg)
		j = 0
		for j < m && e.Val().(char) == char(rs[j]) {
			e = s.chars.Next(e)
			j++
		}
	}
	if j < m {
		beg = 0
	}
	return beg - 1
} //Pos

func (s *stdString) Set (st string) {
	s.chars = Avl.New()
	for _, r := range st {
		s.chars.Append(char(r))
	}
} //Set

func (s1 *stdString) Cat (s2 String) String {
	ss2, ok := s2.(*stdString); M.Assert(ok, 20)
	c1 := s1.chars.Copy()
	c2 := ss2.chars.Copy()
	c1.Cat(c2)
	return &stdString{c1}
} //Cat

func (s *stdString) Insert (pos int, src String) String {
	s2, ok := src.(*stdString); M.Assert(ok, 20)
	c1 := s.chars.Copy()
	c2 := s2.chars.Copy()
	d := c1.Split(pos)
	c1.Cat(c2)
	c1.Cat(d)
	return &stdString{c1}
} //Insert

func (s *stdString) Extract (pos, length int) String {
	t := s.chars.Copy()
	length = M.Max(0, length)
	if pos <= M.MaxInt32 - length {
		t.Split(pos + length)
	}
	u := t.Split(pos)
	return &stdString{u}
} //Extract

func (s *stdString) Copy () String {
	return &stdString{s.chars.Copy()}
} //Copy

func (s *stdString) Convert () string {
	buf := new(bytes.Buffer)
	for e := s.chars.Next(nil); e != nil; e = s.chars.Next(e) {
		buf.WriteRune(rune(e.Val().(char)))
	}
	return buf.String()
} //Convert

func (r *stdReader) Eos () bool {
	return r.eos
}

func (r *stdReader) Char () rune {
	return r.char
}

func (r *stdReader) Base () String {
	return r.s
} //Base

func (r *stdReader) Pos () int {
	r.p = M.Min(r.p, r.s.Length())
	return r.p
} //Pos

func (r *stdReader) SetPos (pos int) {
	M.Assert(pos >= 0, 20)
	M.Assert(pos <= r.s.Length(), 21)
	r.p = pos
} //SetPos

func (r *stdReader) Read () {
	l := r.s.Length()
	r.p = M.Min(r.p, l)
	r.eos = r.p == l
	if r.eos {
		r.char = eOS
	} else {
		e, ok := r.s.chars.Find(r.p + 1); M.Assert(ok)
		r.char = rune(e.Val().(char))
		r.p++
	}
} //Read

func (r *stdReader) ReadChar () rune {
	r.Read()
	return r.char
} //ReadChar

func (r *stdReader) ReadString (length int) string {
	buf := new(bytes.Buffer)
	if length > 0 {
		l := r.s.Length()
		r.p = M.Min(r.p, l)
		e, ok := r.s.chars.Find(r.p + 1)
		if !ok {
			e = nil
		}
		i := 0
		for e != nil && i < length {
			r.char = rune(e.Val().(char))
			buf.WriteRune(rune(r.char))
			e = r.s.chars.Next(e)
			r.p++
			i++
		}
		r.eos = i < length
		if r.eos {
			r.char = eOS
		}
	}
	return buf.String()
} //ReadString

func (w *stdWriter) Base () String {
	return w.s
} //Base

func (w *stdWriter) Pos () int {
	w.p = M.Min(w.p, w.s.Length())
	return w.p
} //Pos

func (w *stdWriter) SetPos (pos int) {
	M.Assert(pos >= 0, 20)
	M.Assert(pos <= w.s.Length(), 21)
	w.p = pos
} //SetPos

func (w *stdWriter) WriteChar (c rune) {
	l := w.s.Length()
	w.p = M.Min(M.Max(0, w.p), l)
	if w.p == l {
		w.s.chars.Append(char(c))
	} else {
		e, ok := w.s.chars.Find(w.p + 1); M.Assert(ok)
		e.SetVal(char(c))
	}
	w.p++
} //WriteChar

func (w *stdWriter) WriteString (s string) {
	for _, r := range s {
		w.WriteChar(r)
	}
} //WriteString

func (d *stdDirectory) New () String {
	return &stdString{Avl.New()}
} //New

// Current directory
func Dir () Directory {
	return dir
}

// Standard directory
func StdDir () Directory {
	return stdDir
}

// Sets the dir global variable.
func SetDir (d Directory) {
	dir = d
} //SetDir
