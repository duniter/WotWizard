/*
Babel: a compiler compiler.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package api

// The module BabelInit is part of the Babel subsystem, a compiler compiler. BabelInit loads a hard coded version of the tbl file corresponding to the definition document BabelSrc.odc.

import (
	
	C	"babel/compil"

)

type (
	
	pInt = []int32
	
	directory struct {
		a pInt
		pos int
	}

)

func (d *directory) ReadInt () int32 {
	i := d.a[d.pos]
	d.pos++
	return i
} // ReadInt

func initBabel () *C.Compiler {
	compiler := fillArray()
	c := C.NewDirectory(&directory{a: compiler[:], pos: 0}).ReadCompiler()
	return c
} // initBabel
