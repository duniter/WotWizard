/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package sort
	
	import
		"util/alea"
	
	type (
		
		Finder interface {
			Less (p1, p2 int) bool
		}
		
		Sorter interface {
			Finder
			Swap (p1, p2 int)
		}
		
		TF struct {
			Finder
		}
		
		TS struct {
			Sorter
		}
	
	)
	
	var
		g *alea.Generator
	
	func (f TF) BinSearch (min, max int, target *int) {
		i := min; j := max + 1
		for i < j {
			k := (i + j) / 2
			if f.Less(k, *target) {
				i = k + 1
			} else {
				j = k
			}
		}
		if j <= max && !f.Less(j, *target) && !f.Less(*target, j) {
			*target = j
		}
	}
	
	func (f TF) BinSearchNext (min, max int, target *int) {
		i := min; j := max + 1
		for i < j {
			k := (i + j) / 2
			if f.Less(k, *target) {
				i = k + 1
			} else {
				j = k
			}
		}
		if j <= max {
			if !f.Less(j, *target) {
				*target = j
			} else {
				*target = j + 1
			}
		}
	}
	
	func (s TS) QuickSort (min, max int) {
		
		insertion := func (l, r int) {
			for i := l + 1; i <= r; i++ {
				for j := i; (j > l) && s.Less(j, j - 1); j-- {
					s.Swap(j, j - 1)
				}
			}
		}
		
		const maxIns = 24
		
		for {
			if max - min < maxIns {
				insertion(min, max)
				break
			}
			p0 := int(g.IntRand(int64(min), int64(max) + 1))
			p := int(g.IntRand(int64(min), int64(max) + 1))
			p2 := int(g.IntRand(int64(min), int64(max) + 1))
			if s.Less(p, p0) {
				p, p0 = p0, p
			}
			if s.Less(p2, p0) {
				p = p0;
			} else if s.Less(p2, p) {
				p = p2;
			}
			i := min; j := max
			for {
				for s.Less(i, p) {
					i++
				}
				for s.Less(p, j) {
					j--
				}
				if i <= j {
					s.Swap(i, j)
					if p == i {
						p = j
					} else if p == j {
						p = i
					}
					i++; j--
				}
				if i > j {break}
			}
			if j - min < max - i {
				s.QuickSort(min, j)
				min = i
			} else {
				s.QuickSort(i, max)
				max = j
			}
		}
	}
	
	func init () {
		g = alea.New();
	}

