/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

// The package json is an implementation of the JSON language
// This package must be imported for its side effects when the JSON compiler is stored in "util/json/static" (file "compVar.go", variable "compiler")
package static

import (
	
	J	"util/json"
		"errors"

)

type (
	
	compReader struct {
		compiler []int32
		pos int64
		n uint32
		i uint
	}

)

func (r *compReader) Read (p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		if r.i == 0 {
			if r.pos >= int64(len(r.compiler)) {
				return i, errors.New("Reading beyond the end of data")
			}
			r.n = uint32(r.compiler[r.pos])
			r.pos++
		}
		p[i] = byte(r.n % 0x100)
		r.n = r.n / 0x100
		r.i = (r.i + 1) % 4
	}
	return len(p), nil
}

func init () {
	compiler := fillArray()
	J.SetRComp(&compReader{compiler: compiler[:], pos: 0, i: 0})
}
