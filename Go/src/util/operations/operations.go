/* 
Hermes: Scientific Spreadsheet.

Copyright (C) 2005…2006 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package operations

import (
	
	M	"util/misc"
		"math"
		//"fmt"

)

const (
	
	PtsNbDef = 5 // Nombre de points de lissage maximal par defaut
	DegreeDef = 2 // Degré maximal du polynome de lissage par defaut (>= 1 && <= (PtsNbDef - 1) / 2 * 2)

)

type (
	
	Dot struct {
		x, y float64
	}
	
	Dots []Dot

)

var (
	
	ptsNb, degree int

)

func (d Dot) X () float64 {
	return d.x
}

func (d Dot) Y () float64 {
	return d.y
}

func (d *Dot) SetXY (x, y float64) {
	d.x = x
	d.y = y
}

func luDcmp (a [][]float64) (indx []int) {
	
	const tiny = 1.E-20
	
	n := len(a)
	M.Assert(n > 0 && len(a[0]) == n, 20)
	indx = make([]int, n)
	vv := make([]float64, n)
	for i := 0; i < n; i++ {
		aamax := 0.
		for j := 0; j < n; j++ {
			aamax = M.MaxF64(aamax, M.AbsF64(a[i][j]))
		}
		M.Assert(aamax > 0., 21)
		vv[i] = 1. / aamax
	}
	for j := 0; j < n; j++ {
		for i := 0; i < j; i++ {
			sum := a[i][j]
			for k := 0; k < i; k++ {
				sum -= a[i][k] * a[k][j]
			}
			a[i][j] = sum
		}
		aamax := 0.; imax := 0
		for i := j; i < n; i++ {
			sum := a[i][j]
			for k := 0; k < j; k++ {
				sum -= a[i][k] * a[k][j]
			}
			a[i][j] = sum
			dum := vv[i] * M.AbsF64(sum)
			if dum >= aamax {
				imax = i
				aamax = dum
			}
		}
		if j != imax {
			a[imax], a[j] = a[j], a[imax]
			vv[imax] = vv[j]
		}
		indx[j] = imax
		if a[j][j] == 0. {
			a[j][j] = tiny
		}
		if j != n - 1 {
			dum := 1. / a[j][j]
			for i := j + 1; i < n; i++ {
				a[i][j] *= dum
			}
		}
	}
	return
}

func luBkSb (a [][]float64, indx []int, b[]float64) {
	n := len(a)
	M.Assert(n > 0 && len(a[0]) == n && len(indx) == n && len(b) == n, 20)
	ii := - 1
	for i := 0; i < n; i++ {
		ll := indx[i]
		sum := b[ll]
		b[ll] = b[i]
		if ii >= 0 {
			for j := ii; j < i; j++ {
				sum -= a[i][j] * b[j]
			}
		} else if sum != 0. {
			ii = i
		}
		b[i] = sum
	}
	for i := n - 1; i >= 0; i-- {
		sum := b[i]
		for j := i + 1; j < n; j++ {
			sum -= a[i][j] * b[j]
		}
		b[i] = sum / a[i][i]
	}
}

func savGol (nl, nr, ld, m int) (c []float64) {
	
	intPower := func (x, y int) float64 {
		xx := float64(x)
		res := 1.
		for y > 0 {
			if M.Odd(y) {
				res *= xx
				y--
			} else {
				xx *= xx
				y /= 2
			}
		}
		return res
	}
	
	// savGol
	M.Assert((nl >= 0) && (nr >= 0), 20)
	np := nl + nr + 1
	M.Assert(nl + nr >= m, 21)
	M.Assert(ld <= m, 22)
	c = make([]float64, np)
	a := make([][]float64, m + 1)
	for i := 0; i <= m; i++ {
		a[i] = make([]float64, m + 1)
	}
	b := make([]float64, m + 1)
	for ipj := 0; ipj <= 2 * m; ipj++ {
		sum := 0.
		if ipj == 0 {
			sum = 1.
		}
		for k := 1; k <= nr; k++ {
			sum += intPower(k, ipj)
		}
		for k := 1; k <= nl; k++ {
			sum += intPower(-k, ipj)
		}
		mm := M.Min(ipj, 2 * m - ipj)
		for imj := -mm; imj <= mm; imj += 2 {
			a[(ipj + imj) / 2][(ipj - imj) / 2] = sum
		}
	}
	indx := luDcmp(a)
	for j := 0; j <= m; j++ {
		b[j] = 0.
	}
	b[ld] = 1.
	luBkSb(a, indx, b)
	for j := 0; j < np; j++ {
		c[j] = 0.
	}
	for k := - nl; k <= nr; k++ {
		sum := b[0]
		fac := 1.
		for mm := 1; mm <= m; mm++ {
			fac = fac * float64(k)
			sum += b[mm] * fac
		}
		c[k + nl] = sum
	}
	return
}

func Derive (p Dots, ptsNb, degree int) (dp Dots) {

	der := func (p Dots, deb, length, pos int, c []float64, dp Dots) {
		M.Assert(length <= len(c), 20)
		M.Assert((deb >= 0) && (deb + length <= len(p)), 21)
		M.Assert((pos >= 0) && (pos <= len(dp)), 22)
		u := 0.; v := 0.
		for i := 0; i < length; i++ {
			u += c[i] * p[deb + i].y
			v += c[i] * p[deb + i].x
		}
		if u != 0. {
			u = u / v
			if math.IsInf(u, 1) {
				u = math.MaxFloat64
			} else if math.IsInf(u, -1) {
				u = -math.MaxFloat64
			}
		}
		dp[pos].y = u
	}
	
	// Derive
	n := len(p)
	M.Assert(n > 0, 20)
	M.Assert(degree >= 1, 21)
	M.Assert((ptsNb - 1) / 2 * 2 >= degree, 22)
	dp = make(Dots, n)
	for i := 0; i < n; i++ {
		dp[i].x = p[i].x
	}
	if n == 1 {
		dp[0].y = 0.
	} else if n == 2 {
		dp[0].y = (p[1].y - p[0].y) / (p[1].x - p[0].x)
		dp[1].y = dp[0].y
	} else {
		q := (M.Min(n, ptsNb) - 1) / 2
		np := 2 * q + 1
		m := M.Min(np - 1, degree)
		M.Assert(m >= 1, 100)
		for i := 0; i < q; i++ {
			c := savGol(i, np - 1 - i, 1, m)
			der(p, 0, np, i, c, dp)
		}
		c := savGol(q, q, 1, m)
		for i := q; i < n - q; i++ {
			der(p, i - q, np, i, c, dp)
		}
		for i := n - q; i < n; i++ {
			c := savGol(i - n + np, n - 1 - i, 1, m)
			der(p, n - np, np, i, c, dp)
		}
	}
	return
}
