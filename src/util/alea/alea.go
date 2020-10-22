/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package alea

import (
	
	"math"
	"sync"
	"time"

)

const (
	
	rLong = 32
	m1 = 2147483563 
	ia1 = 40014
	m2 = 2147483399 
	ia2 = 40692
	
	rm1 float64 = 1. / float64(m1);
	nDiv = 1 + (m1 - 1) / rLong;

)

type (
	
	rT [rLong]int32 //R
	
	Generator struct {
		ix1, ix2, iy int32
		r rT
		gSet bool
		g float64
	}

)

var (
	
	gen *Generator
	
	randomMutex = new(sync.Mutex)
	gaussMutex = new(sync.Mutex)

)

// Generates a new pseudo-random real number x such that  (0 <= x < 1), with uniforme deviates.  (Period > 2 × 10^18). Ref.: Generator ran2, W. H. Press, S. A. Teukolsky, W. T. Vetterling, B. P. Flannery, Numerical Recipes, The Art of Scientific Computing, second edition, 1997.
func (g *Generator) Random () float64 {
	randomMutex.Lock()
	g.ix1 = int32(int64(g.ix1) * ia1 % m1)
	g.ix2 = int32(int64(g.ix2) * ia2 % m2)
	j := int(g.iy / nDiv)
	g.iy = g.r[j] - g.ix2
	if g.iy <= 0 {
		g.iy += m1 - 1
	}
	g.r[j] = g.ix1
	randomMutex.Unlock()
	res := rm1 * float64(g.iy)
	if !((res >= 0.) && (res < 1.)) {panic(60)}
	return res
}

// Generates a new integer number between min (included) and max (excluded)
func (g *Generator) IntRand (min, max int64) int64 {
	return min + int64(float64(max - min) * g.Random())
}

// Generates a new unsigned integer number between min (included) and max (excluded)
func (g *Generator) UintRand (min, max uint64) uint64 {
	if min > max {
		min, max = max, min
	}
	return min + uint64(float64(max - min) * g.Random())
}

// Generates a new pseudo-random real number x with gaussian (normal) deviates. Probability distribution p(x) = 1 / sqrt(2 * pi) * exp(-x^2 / 2). Box-Muller method.
func (g *Generator) GaussRand () float64 {
	var v1, v2, r float64
	gaussMutex.Lock()
	g.gSet = !g.gSet
	if g.gSet {
		for {
			v1 = 2. * g.Random() - 1.
			v2 = 2. * g.Random() - 1.
			r = v1 * v1 + v2 * v2
			if (0. < r) && (r < 1.) {break}
		}
		r = math.Sqrt( -2. * math.Log(r) / r)
		g.g = v1 * r
		r = v2 * r
	} else {
		r = g.g
	}
	gaussMutex.Unlock()
	return r
}

// Initializes the pseudo-random number generator with seed.
func (g *Generator) Randomize (seed int64) {
	const
		warmUp = 8
	
	randomMutex.Lock()
	if g.ix1 = int32(seed % m1); g.ix1 < 1 {g.ix1 = 1}
	g.ix2 = g.ix1
	for i := 1; i <= warmUp; i++ {
		g.ix1 = int32(int64(g.ix1) * ia1 % m1)
	}
	for i := 0; i < rLong; i++ {
		g.ix1 = int32(int64(g.ix1) * ia1 % m1)
		g.r[i] = g.ix1
	}
	g.iy = g.r[rLong - 1]
	randomMutex.Unlock()
}

// Initializes the generator with a random seed.
func New () *Generator {
	g := new(Generator)
	g.Randomize(time.Now().Unix())
	g.gSet = false
	return g
}

func Random () float64 {
	return gen.Random();
}

func IntRand (min, max int64) int64 {
	return gen.IntRand (min, max)
}

func UintRand (min, max uint64) uint64 {
	return gen.UintRand (min, max)
}

func GaussRand () float64 {
		return gen.GaussRand();
}

func Randomize (seed int64) {
		gen.Randomize(seed);
}

func init () {
	gen = New()
}
