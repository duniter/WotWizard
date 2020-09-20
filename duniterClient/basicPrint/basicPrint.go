/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package basicPrint

import (
	
	A	"util/avl"
	M	"util/misc"
	SM	"util/strMapping"
		"time"
		"unicode"

)

const (
	
	timeFormat = "02/01/2006 15:04:05"
	
	OldIcon = "×" // Icon for old items (old or leaving members)
	
	never = M.MaxInt64 // In WotWizard window
	Revoked = M.MinInt64 // Limit date for revoked members
	Already = M.MinInt64 + 1 // Already available certification date
	
	Lt = A.Lt
	Eq = A.Eq
	Gt = A.Gt

)

type (
	
	Comp = A.Comp

)

// Convert the date t to the string dt
func Ts2s (t int64) string {
	switch t {
	case never:
		return SM.Map("#duniterClient:Never")
	case Revoked:
		return SM.Map("#duniterClient:Revoked")
	case Already:
		return "**/**/**** **:**:**"
	default:
		dt := time.Unix(M.Abs64(t), 0).Format(timeFormat)
		if t < 0 {
			dt = OldIcon + dt // leaving member
		}
		return dt
	}
} //Ts2s

// Extract the significant character at the position i in s, or further; only alphanumeric characters are significant, and their case of lowest rank is returned
func downC (r []rune, i *int) rune {

	LetterOrDigit := func (c rune) bool {
		return c >= '0' && c <= '9' || unicode.IsLetter(c)
	} //LetterOrDigit

	//downC
	c := r[*i]
	for c != 0 {
		*i++
		if LetterOrDigit(c) {
			c = M.Min32(unicode.ToLower(c), unicode.ToUpper(c))
			break
		}
		c = r[*i]
	}
	return c
} //downC

// Standard comparison procedure for identifiers; they are first compared with only significant characters and case ignored, and if still equal, with all characters and case taken into account
func CompP (s1, s2 string) Comp {
	r1 := []rune(s1); r2 := []rune(s2)
	r1 = append(r1, 0); r2 = append(r2, 0)
	i1 := 0; i2 := 0
	var c1, c2 rune
	for {
		c1 = downC(r1, &i1)
		c2 = downC(r2, &i2)
		if c1 != c2 || c1 == 0 {break}
	}
	if c1 < c2 {
		return Lt
	}
	if c1 > c2 {
		return Gt
	}
	i := 0
	for r1[i] == r2[i] && r1[i] != 0 {
		i++
	}
	if r1[i] < r2[i] {
		return Lt
	}
	if r1[i] > r2[i] {
		return Gt
	}
	return Eq
} //CompP
