/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package wotWizard

/**************************************************************************************
Bug : Dans calcRec, si la date de passage d'un dossier devient supérieure à la date limite d'une de ses certifications, il faut en tenir compte : enlever la certification, recalculer PrincCertif, etc...
***************************************************************************************/

// For versions 1.4+ of Duniter
// This version suppose equiprobable all external certifications (toward a non-member identity) which are concurrent at the same date, and process concurrent internal certifications (toward an already-member identity) afterwards

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	M	"util/misc"
	S	"duniter/sandbox"
		"math"
		"time"
		"unsafe"

)

const (
	
	maxSizeDef = 430000000 // Default value for the greatest allowed allocated memory size

)

type (
	
	Uid *string
	Hash *B.Hash
	Pubkey *B.Pubkey
	
	// Internal certification, or dossier of external certifications toward the same identity
	CertOrDoss interface {
		Date () int64
		Limit () int64
	}
	
	File []CertOrDoss // Array of internal certifications and dossiers
	
	// Internal certification
	Certif struct {
		date, // Entry date
		limit int64 // Expiration date
		From, To Uid
		ToH Hash
		fromP Pubkey
	}
	
	// Dossier of external certifications
	Dossier struct {
		date, // Entry date
		MinDate, // Minimum date: last membership application + msPeriod
		limit int64 // Expiration date
		Id Uid // Certified identity
		Hash Hash
		pub Pubkey
		PrincCertif int // Rank of the certification whose entry date gives the entry date of the dossier (1 <= PrincCertif <= len(Certifs)
		ProportionOfSentries float64 // Proportion of sentries reachable through B.pars.stepMax steps
		Certifs File // Array of certifications
	}
)

func (c *Certif) Date () int64 {
	return c.date
}

func (c *Certif) Limit () int64 {
	return c.limit
}

func (d *Dossier) Date () int64 {
	return d.date
}

func (d *Dossier) Limit () int64 {
	return d.limit
}

type (
	
	// AVL trees of Propagation are used in the main CalcPermutations procedure; trees of Propagation are sorted by Id(s) first and then by Date(s)
	Propagation struct {
		Hash B.Hash
		Id string // uid of a candidate
		Date int64 // A possible date of her entry
		After bool // After = true if this entry can occur at the date Date or after (uncertainty due to incomplete computation)
		Proba float64 // Probability of this entry
	}
	
	// AVL trees of PropName and PropDate are the ouput formats of procedures CalcEntries and BuildEntries
	PropName Propagation

	PropDate Propagation
	
	// Element of sets of trees of Propagation, each tree corresponding to a possible permutation of the order of the entries
	Set struct {
		Proba float64 // Probability of the permutation
		T *A.Tree // AVL tree
	}
	
	// Chained list of Pubkey(s); used to test the distance rule in the procedure notTooFar
	pubList struct {
		next *pubList
		pub Pubkey
		date int64
	}
	
	// The virtual stack elements (stored in a queue) used in the main procedure CalcPermutations and representing the tree of all possible permutations of Dossier(s) in the final order of the File f
	node struct {
		next, // Next node in the queue (or brother in the possibilities tree)
		stack *node // Previous node in the stack (or father in the possibilities tree)
		f File // The File used at this point
		step, // Current position in f
		sons, // Number of sons in the possibilities tree
		seenSons int // Number of sons already seen
		sets *A.Tree; // Set of permutations of entries of Dossier(s) found in the descendants of this node
	}
	
	// Queue (or fifo list) of node(s)
	queue struct {
		end *node
	}
	
	// Set of Pubkeys, used in FillFile
	pubSet struct {
		p B.Pubkey
	}

)

var (
	
	// Maximum allowed memory size for the execution of CalcPermutations
	maxSize int64 = maxSizeDef

)

// Standard managing procedures for queue

func (q *queue) init () {
	q.end = nil
}

func (q *queue) isEmpty () bool {
	return q.end == nil
}

func (q *queue) put (n *node) {
	if q.end == nil {
		q.end = n
		n.next = n
	} else {
		n.next = q.end.next
		q.end.next = n
		q.end = n
	}
}

func (q *queue) get () (n *node) {
	M.Assert(q.end != nil, 20)
	n = q.end.next
	q.end.next = n.next
	if q.end == n {
		q.end = nil
	}
	return
}

// Comparison procedures for Propagation, PropName, PropDate and Set

func compareProp (e1, e2 *A.Elem) A.Comp {
	if e1 == nil {
		if e2 == nil {
			return A.Eq
		}
		return A.Lt
	}
	if e2 == nil {
		return A.Gt
	}
	p1 := e1.Val().(*Propagation)
	p2 := e2.Val().(*Propagation)
	if p1.Id < p2.Id {
		return A.Lt
	}
	if p1.Id > p2.Id {
		return A.Gt
	}
	if p1.Date < p2.Date {
		return A.Lt
	}
	if p1.Date > p2.Date {
		return A.Gt
	}
	if p1.After && !p2.After {
		return A.Lt
	}
	if !p1.After && p2.After {
		return A.Gt
	}
	return A.Eq
}

func (p1 *Propagation) Compare (p2 A.Comparer) A.Comp {
	pp2 := p2.(*Propagation)
	if p1.Id < pp2.Id {
		return A.Lt
	}
	if p1.Id > pp2.Id {
		return A.Gt
	}
	if p1.Date < pp2.Date {
		return A.Lt
	}
	if p1.Date > pp2.Date {
		return A.Gt
	}
	if p1.After && !pp2.After {
		return A.Lt
	}
	if !p1.After && pp2.After {
		return A.Gt
	}
	return A.Eq
}

func (p1 *PropName) Compare (p2 A.Comparer) A.Comp {
	pp2 := p2.(*PropName)
	b := BA.CompP(p1.Id, pp2.Id)
	if b != A.Eq {
		return b
	}
	if p1.Date < pp2.Date {
		return A.Lt
	}
	if p1.Date > pp2.Date {
		return A.Gt
	}
	return A.Eq
}

func (p1 *PropDate) Compare (p2 A.Comparer) A.Comp {
	pp2 := p2.(*PropDate)
	if p1.Date < pp2.Date {
		return A.Lt
	}
	if p1.Date > pp2.Date {
		return A.Gt
	}
	if !p1.After && pp2.After {
		return A.Lt
	}
	if p1.After && !pp2.After {
		return A.Gt
	}
	return BA.CompP(p1.Id, pp2.Id)
}

func (s1 *Set) Compare (s2 A.Comparer) A.Comp {
	ss2 := s2.(*Set)
	e1 := s1.T.Next(nil); e2 := ss2.T.Next(nil)
	b := compareProp(e1, e2)
	for b == A.Eq && e1 != nil {
		e1 = s1.T.Next(e1); e2 = ss2.T.Next(e2)
		b = compareProp(e1, e2)
	}
	return b
}

func (pub1 *pubSet) Compare (pub2 A.Comparer) A.Comp {
	ppub2 := pub2.(*pubSet)
	if pub1.p < ppub2.p {
		return A.Lt
	}
	if pub1.p > ppub2.p {
		return A.Gt
	}
	return A.Eq
}

// Copy procedures for Propagation and InvProp

func (p *PropDate) Copy () A.Copier {
	q := new(PropDate); *q = *p
	return q
}

func (p *PropName) Copy () A.Copier {
	q := new(PropName); *q = *p
	return q
}

// Copy f and return its copy; the copy is shallow for elements of ranks less than deepFrom, and deep afterwards; size is the size of the copy
func copyFile (f File, deepFrom int, size *int64) File {
	l := len(f)
	M.Assert(deepFrom <= l, 20)
	g := make(File, l)
	*size += int64(unsafe.Sizeof(g)) + int64(l) * int64(unsafe.Sizeof(g[0]))
	for i := 0; i < deepFrom; i++ {
		g[i] = f[i]
	}
	for i := deepFrom; i < l; i++ {
		switch cd := f[i].(type) {
		case *Certif:
			c := new(Certif)
			*c = *cd
			*size += int64(unsafe.Sizeof(*c))
			g[i] = c
		case *Dossier:
			d := new(Dossier)
			*d = *cd
			*size += int64(unsafe.Sizeof(*d))
			g[i] = d
			h := make(File, len(cd.Certifs))
			*size += int64(unsafe.Sizeof(h)) + int64(len(cd.Certifs)) * int64(unsafe.Sizeof(h[0]))
			for j := 0; j < len(cd.Certifs); j++ {
				c := new(Certif)
				*c = *cd.Certifs[j].(*Certif)
				*size += int64(unsafe.Sizeof(*c))
				h[j] = c
			}
			d.Certifs = h
		}
	}
	return g
}

func calcPrinc (certifs File) (princCertif int) {
	n := len(certifs)
	princCertif = int(B.Pars().SigQty) - 1
	var ok bool
	for {
		princCertif++
		certifiers := make([]B.Pubkey, princCertif)
		for j := 0; j < princCertif; j++ {
			certifiers[j] = *certifs[j].(*Certif).fromP
		}
		ok = B.DistanceRuleOk(certifiers)
		if ok || princCertif == n {break}
	}
	M.Assert(ok, 60)
	return
}

// Fix the next possible date d of a certification from p
func fixCertNextDate (p B.Pubkey) (d int64) {
	d = 0
	var pos B.CertPos
	if B.CertFrom(p, &pos) {
		from, to, ok := pos.CertNextPos()
		for ok {
			block_number, _, b := B.Cert(from, to); M.Assert(b, 100)
			tm, _, b := B.TimeOf(block_number); M.Assert(b, 101)
			d = M.Max64(d, tm)
			from, to, ok = pos.CertNextPos()
		}
		d += int64(B.Pars().SigPeriod)
	}
	return
}

// Fix the disponibility dates of the elements of f, always after the current date
func fileDates (f File) {
	for i := 0; i < len(f); i++ {
		switch cd := f[i].(type) {
		case *Dossier:
			if len(cd.Certifs) == 0 {
				cd.date = 0
			} else {
				cd.date = cd.Certifs[cd.PrincCertif - 1].(*Certif).date
			}
			cd.date = M.Max64(cd.date, cd.MinDate)
			if cd.date > cd.limit {
				cd.date = BA.Never
			}
		default:
		}
	}
} //fileDates

// Sort, by insertion, the end of f, starting at position i0; all the dates must have been fixed before
func sortAll (f File, i0, critical int) (modif bool) {

	// Dossiers with at least B.pars.sigQty certifications and internal certifications first; other dossiers last; dossiers are sorted by number of certifications first, then by dates; internal certifications are sorted by dates; if a dossier with at least B.pars.sigQty certifications and an internal certification have the same date, the dossier comes first

	less := func (cd1, cd2 CertOrDoss) bool {
		var ok bool
		switch cd01 := cd1.(type) {
		case *Dossier:
			switch cd02 := cd2.(type) {
			case *Dossier:
				ok = cd02.PrincCertif < int(B.Pars().SigQty) && (cd01.PrincCertif > cd02.PrincCertif || cd01.PrincCertif == cd02.PrincCertif && (cd01.date < cd02.date || cd01.date == cd02.date && *cd01.pub < *cd02.pub)) || cd01.PrincCertif >= int(B.Pars().SigQty) && cd02.PrincCertif >= int(B.Pars().SigQty) && (cd01.date < cd02.date || cd01.date == cd02.date && *cd01.pub < *cd02.pub)
			case *Certif:
				ok = cd01.PrincCertif >= int(B.Pars().SigQty) && cd01.date <= cd02.date
			}
		case *Certif:
			switch cd02 := cd2.(type) {
			case *Dossier:
				ok = cd02.PrincCertif < int(B.Pars().SigQty) || cd01.date < cd02.date
			case *Certif:
				ok = cd01.date < cd02.date || cd01.date == cd02.date && (*cd01.fromP < *cd02.fromP || *cd01.fromP == *cd02.fromP && *cd01.To < *cd02.To)
			}
		}
		return ok
	}
	
	modif = false
	for i := i0 + 1; i < len(f); i++ {
		j := i; cdx := f[j]
		for j > i0 && less(cdx, f[j - 1]) { // No sentinel; inefficient
			f[j] = f[j - 1]
			j--
			modif = modif || j == critical
		}
		f[j] = cdx
	}
	return
} //sortAll

// Sort the end of f, starting at position i0; the dates of all Certification(s) must have been fixed before, but not those of Dossier(s); if afterNow, the disponibility dates are put after the current date
func sortFile (f File, i0 int) {
	for i := i0; i < len(f); i++ {
		switch cd := f[i].(type) {
		case *Dossier:
			if sortAll(cd.Certifs, 0, cd.PrincCertif - 1) {
				cd.PrincCertif = calcPrinc(cd.Certifs)
			}
		default:
		}
	}
	fileDates(f)
	sortAll(f, i0, 0)
} //sortFile

// Update the list of sentries and the distance rule test at each step? Too long execution time!
// Simulate the entry of f[step] in the blockchain and update all f[i] with i > step
func propagate (f *File, step int) {
	
	incDates := func (uid string, newDate int64, i0 int) {
		for i := i0; i < len(*f); i++ {
			switch cd := (*f)[i].(type) {
			case *Certif:
				if *cd.From == uid {
					if newDate > cd.limit {
						cd.date = BA.Never
					} else {
						cd.date = newDate
					}
				}
			case *Dossier:
				j := len(cd.Certifs)
				for j > 0 {
					j--
					c := cd.Certifs[j].(*Certif)
					if *c.From == uid {
						if c.date != BA.Already {
							if newDate > M.Min64(cd.limit, c.limit) {
								c.date = BA.Never
							} else {
								c.date = newDate
							}
						}
						j = 0
					}
				}
			}
		}
	} //incDates
	
	// propagate
	// If two newcomers has the same id or the same pubkey, only the one that comes first can enter the WoT
	for {
		ok := true
		switch cd := (*f)[step].(type) {
		case *Dossier:
			j := 0
			for ok && j < step {
				switch cd2 := (*f)[j].(type) {
				case *Dossier:
					ok = *cd2.Id != *cd.Id && *cd2.pub != *cd.pub
				default:
				}
				j++
			}
			if !ok { // Small decrease in allocated size, not recorded
				g := make(File, len(*f) - 1)
				for j := 0; j < step; j++ {
					g[j] = (*f)[j]
				}
				for j := step; j < len(g); j++ {
					g[j] = (*f)[j + 1]
				}
				*f = g
			}
		default:
		}
		if ok || step == len(*f) {break}
	}
	if step < len(*f) && (*f)[step].Date() != BA.Never {
		switch cd := (*f)[step].(type) {
		case *Certif:
			incDates(*cd.From, cd.date + int64(B.Pars().SigPeriod), step + 1)
		case *Dossier:
			j := 0
			// All certifs with the same date as cd must lead to updates and certifs with a date less than cd.date + B.pars.avgGenTime have more than 50% probability to be inserted in the same block and, so, must lead to update too
			for j < cd.PrincCertif || j < len(cd.Certifs) && cd.Certifs[j].(*Certif).date < cd.date + int64(B.Pars().AvgGenTime) {
				incDates(*cd.Certifs[j].(*Certif).From, cd.date + int64(B.Pars().SigPeriod), step + 1);
				j++
			}
			nc := len(cd.Certifs) - j
			if nc > 0 { // No change in size; just transfers
				n1 := len(*f); n2 := n1 + nc
				g := make(File, n2)
				for m := 0; m < n1; m++ {
					g[m] = (*f)[m]
				}
				k := j
				for m := n1; m < n2; m++ {
					g[m] = cd.Certifs[k]
					k++
				}
				*f = g
				g = make(File, j)
				for m := 0; m < j; m++ {
					g[m] = cd.Certifs[m]
				}
				cd.Certifs = g
			}
		}
		sortFile(*f, step + 1)
	}
} //propagate

// Sort by selection
func sortPubList (c **pubList) {

	least := func (c *pubList) *pubList {
		cc := c; c = c.next
		for c.next != nil {
			if c.next.date < cc.next.date || c.next.date == cc.next.date && *c.next.pub < *cc.next.pub {
				cc = c
			}
			c = c.next
		}
		return cc
	}

	c0 := &pubList{next: *c}
	cc := c0
	for cc.next != nil {
		c1 := least(cc)
		c2 := c1.next; c1.next = c2.next
		c2.next = cc.next; cc.next = c2
		cc = c2
	}
	*c = c0.next
}

// Say whether the list of certifiers' Pubkey(s) c verifies the Duniter's distance rule and gives, in proportionOfSentries the proportion of sentries members reachable in less than B.pars.stepMax steps
func notTooFar (c **pubList, n int) (needed int, proportionOfSentries float64, ok bool) {
	if n == 0 {
		needed = 0
		proportionOfSentries = 0.
		ok = false
	} else if n < int(B.Pars().SigQty) {
		sortPubList(c)
		certifiers := make([]B.Pubkey, n)
		cc := *c
		for j := 0; j < n; j++ {
			certifiers[j] = *cc.pub
			cc = cc.next
		}
		needed = n
		proportionOfSentries = B.PercentOfSentries(certifiers)
		ok = proportionOfSentries >= B.Pars().Xpercent
	} else {
		sortPubList(c)
		needed = int(B.Pars().SigQty) - 1
		for {
			needed++
			certifiers := make([]B.Pubkey, needed)
			cc := *c;
			for j := 0; j < needed; j++ {
				certifiers[j] = *cc.pub
				cc = cc.next
			}
			proportionOfSentries = B.PercentOfSentries(certifiers)
			ok = proportionOfSentries >= B.Pars().Xpercent
			if ok || needed == n {break}
		}
	}
	return
}

// WotWizard main procedure; return in sets all the elements of type Set, i.e. all the possible permutations in the order of entries of the Dossier(s) in f, along with their probabilities
func CalcPermutations (f File) *A.Tree {

	// Put into n.sets the set containing, as unique element, the list of entries in n.f
	evaluate := func (n *node) {
		n.sets = A.New()
		set := new(Set); set.T = A.New()
		for i := 0; i < len(n.f); i++ {
			switch cd := n.f[i].(type) {
			case *Dossier:
				_, b, _ := set.T.SearchIns(&Propagation{Hash: *cd.Hash, Id: *cd.Id, Date: cd.date, After: i >= n.step && (cd.date != BA.Never)}); M.Assert(!b, 100)
			default:
			}
		}
		n.f = nil
		set.Proba = 1.
		_, b, _ := n.sets.SearchIns(set); M.Assert(!b, 101)
	}
	
	// Merge the set of possible permutations (of type Set) setsUp into setsDown, combining probabilities
	addProba := func (setsDown, setsUp *A.Tree, permNb int) {
		x := 1. / float64(permNb)
		e := setsUp.Next(nil)
		for e != nil {
			s := e.Val().(*Set)
			y := s.Proba * x
			s.Proba = 0.
			ee, _, _ := setsDown.SearchIns(s)
			s = ee.Val().(*Set)
			s.Proba += y
			e = setsUp.Next(e)
		}
	}
	
	var ok bool

	// Process f in the order of its elements by calling propagate as long as successive Dossier(s) have the same date, and call itself (by creating new sons' nodes) for every first possible entry in a set of Dossier(s) with the same date and with a number of Dossier(s) greater than one, merge all the sets of possible permutations returned in sons' nodes into their father's node. It's not a recursive procedure, but it uses queues (fifo lists) to process all possible events in breadth-first order
	calcRec := func (f File) (sets *A.Tree) {
		
		isCertif := func (cd CertOrDoss) bool {
			_, ok := cd.(*Certif)
			return ok
		}
		
		// calcRec
		stackSize := int64(0)
		var q1, q2 queue
		q1.init(); q2.init()
		root := &node{stack: nil, f: f, step: 0, seenSons: 0, sets: nil}
		stackSize += int64(unsafe.Sizeof(*root))
		q1.put(root)
		for !q1.isEmpty() {
			n := q1.get()
			for n.step < len(n.f) && n.f[n.step].Date() != BA.Never && (n.step >= len(n.f) - 1 ||  isCertif(n.f[n.step]) ||  isCertif(n.f[n.step + 1]) || n.f[n.step + 1].Date() - n.f[n.step].Date() >= int64(B.Pars().AvgGenTime)) {
				propagate(&n.f, n.step)
				n.step++
			}
			ok = ok && stackSize <= maxSize
			if n.step >= len(n.f) || n.f[n.step].Date() == BA.Never || !ok {
				evaluate(n)
				q2.put(n)
			} else {
				// Assertion: n.step < len(n.f) - 1 && !isCertif(n.f[n.step]) && !isCertif(n.f[n.step + 1]) && n.f[n.step + 1].date - n.f[n.step].date < int64(B.Pars().AvgGenTime)
				j := n.step + 2
				for j < len(n.f) && !isCertif(n.f[j]) && n.f[j].Date() - n.f[n.step].Date() < int64(B.Pars().AvgGenTime) {
					j++
				}
				n.sons = j - n.step
				for j := n.step; j < n.step + n.sons; j++ {
					m := &node{sons: 0, seenSons: 0, stack: n, f: copyFile(n.f, n.step, &stackSize)}
					stackSize += int64(unsafe.Sizeof(*m))
					m.f[j], m.f[n.step] = m.f[n.step], m.f[j]
					propagate(&m.f, n.step)
					m.step = n.step + 1
					m.sets = nil
					q1.put(m)
				}
				n.f = nil
			}
		}
		for !q2.isEmpty() {
			m := q2.get()
			if m.seenSons != m.sons {
				q2.put(m)
			} else {
				n := m.stack
				if n != nil {
					if n.sets == nil {
						n.sets = A.New()
						q2.put(n)
					}
					n.seenSons++
					addProba(n.sets, m.sets, n.sons)
				}
			}
		}
		sets = root.sets
		return
	}
	
	// CalcPermutations
	ok = true
	sets := calcRec(f)
	M.Assert(sets != nil, 60)
	return sets
}

// Merge the Set(s) returned by CalcPermutations and combine their probabilities; return the list of entries sorted by date(s) (occurDate with elements of type PropDate) or by id(s) (occurName with elements of type PropName)
func CalcEntries (f File) (sets, occurDate, occurName *A.Tree) {
	sets = CalcPermutations(f)
	
	// Computing of Propagation(s) with their proba(s)
	now := B.Now()
	occur := A.New()
	e := sets.Next(nil)
	for e != nil {
		s := e.Val().(*Set)
		ee := s.T.Next(nil)
		for ee != nil {
			p := ee.Val().(*Propagation)
			p.Date = M.Max64(p.Date, now)
			p.Proba = 0.
			eee, _, _ := occur.SearchIns(p)
			p = eee.Val().(*Propagation)
			p.Proba += s.Proba
			ee = s.T.Next(ee)
		}
		e = sets.Next(e)
	}
	
	// For each uid, with increasing date(s), gather all Propagation(s) following a Propagation with After = true together
	var pAfter *Propagation
	id := ""
	e = occur.Next(nil)
	for e != nil {
		ee := occur.Next(e)
		p := e.Val().(*Propagation)
		if p.Id != id {
			id = p.Id
			pAfter = nil
		}
		if pAfter == nil && p.After {
			pAfter = p
		} else if pAfter != nil {
			pAfter.Proba += p.Proba
			b := occur.Delete(p); M.Assert(b, 100)
		}
		e = ee
	}
	
	// Propagation -> PropDate
	occurDate = A.New()
	e = occur.Next(nil)
	for e != nil {
		p := e.Val().(*Propagation)
		_, b, _ := occurDate.SearchIns(&PropDate{Hash: p.Hash, Id: p.Id, Date: p.Date, After: p.After, Proba: p.Proba}); M.Assert(!b, 101)
		e = occur.Next(e)
	}
	
	// PropDate -> PropName
	occurName = A.New()
	e = occurDate.Next(nil)
	for e != nil {
		p := e.Val().(*PropDate)
		_, b, _ := occurName.SearchIns(&PropName{Hash: p.Hash, Id: p.Id, Date: p.Date, After: p.After, Proba: p.Proba}); M.Assert(!b, 102)
		e = occurDate.Next(e)
	}
	return
}

func lastEntryMTime (pubkey B.Pubkey) int64 {
	list, ok := B.JLPub(pubkey); M.Assert(ok, 100)
	block, _, ok := B.JLPubLNext(&list); M.Assert(ok, 101)
	mTime, _, ok := B.TimeOf(block); M.Assert(ok, 102)
	return mTime
}

/*
Q : Pour revenir, le membre doit refaire une demande d’adhésion et retrouver suffisamment de certifications.

R : Pour les certifications : s’il n’en avait pas assez, en effet, oui. Pour l’adhésion, oui, il en refaut une de façon systématique. On ne rentre pas sans demande d’adhésion.

Q : Est-ce que tout cela se passe comme pour une première adhésion ?

Demande d’adhésion en piscine dans membership ? Validité 2 mois ?
Certifications additionnelles en piscine, validité 2 mois.
Règle de distance, etc…

R : Oui, exactement de la même façon. A ceci près que les certifications existantes (celles récupérées lorsque l’on était membre) persistent en blockchain. Tu ne les trouveras donc pas en piscine, pourtant elles existent bien et font partie du dossier.

Q : J’ai bon ? J’oublie quelque chose ?

Non, je crois que tu as tout :slight_smile:

Règle de délai inter-renouvellements : cette information est facile à déduire par la dernière adhésion du membre (joiners|actives|leavers) + le délai (= msWindow)
*/

// Extract f from the blockchain (Duniter1Blockchain) and the sandbox (Duniter1Sandbox) and sort it; minCertifs is the minimum number of certifications by Dossier required (at least 1 if minCertifs < 1); keep only valid elements
func FillFile (minCertifs int) (f File, cNb, dNb int) {
	
	type (
		
		cdList struct {
			next *cdList
			cd CertOrDoss
		}
	
	)
	
	cdL := new(cdList); l := cdL // Chained list of Dossier and Certif
	dNb = 0 // Number of Dossier(s)
	useful := A.New()
	var el *A.Elem
	toHash, ok := S.IdNextHash(true, &el)
	for ok { // For all identity hash in sandbox
		idInBC, to, _, _, exp2, b := S.IdHash(toHash); M.Assert(b, 100)
		var minDate int64 = 0
		bb := !idInBC
		if !bb {
			var (member bool; exp int64)
			_, member, _, _, _, exp, bb = B.IdPubComplete(to); M.Assert(bb, 101)
			leTi := lastEntryMTime(to)
			minDate = leTi + int64(B.Pars().MsPeriod)
			bb = !member && exp >= 0 && exp2 > minDate
		}
		if bb { // identity in sandBox or (not member & not leaving & new membership application date later than previous one plus msPeriod)
			nbCertifs := 0; certs := (*pubList)(nil)
			var posBI B.CertPos
			if idInBC && B.CertTo(to, &posBI) {
				posB := posBI
				for {
					from, _, okP := posB.CertNextPos()
					if okP {
						_, member, _, _, _, _, bb := B.IdPubComplete(from)
						var posBF B.CertPos
						bb = bb && member && (!B.CertFrom(from, &posBF) || posBF.CertPosLen() < int(B.Pars().SigStock))
						if bb {
							_, exp, b := B.Cert(from, to); M.Assert(b, 103)
							bb = exp > minDate
						}
						if bb { // Don't consider certifications sent by a non-member or by a member who already has sent sigStock (100) certifications or certifications whose limit date is smaller than minDate
							nbCertifs++
							certs = &pubList{pub: &from, date: BA.Already, next: certs}
						}
					}
					if !okP {break}
				}
			}
			var posI S.CertPos
			if b = S.CertTo(toHash, &posI); b {
				pos := posI
				for {
					from, toHash, okP := pos.CertNextPos()
					if okP {
						c := certs
						for c != nil && *c.pub != from {
							c = c.next
						}
						if c != nil {
							continue
						}
					}
					if okP {
						_, member, _, _, _, _, bb := B.IdPubComplete(from)
						var posBF B.CertPos
						bb = bb && member && (!B.CertFrom(from, &posBF) || posBF.CertPosLen() < int(B.Pars().SigStock))
						if bb {
							_, _, exp, b := S.Cert(from, toHash); M.Assert(b, 104)
							date := fixCertNextDate(from)
							if M.Max64(date, minDate) <= exp { // Not-expired certification
								nbCertifs++
								certs = &pubList{pub: &from, date: date, next: certs}
							}
						}
					}
					if !okP {break}
				}
			}
			var (princCertif int; proportionOfSentries float64)
			bb := nbCertifs >= minCertifs
			if bb {
				princCertif, proportionOfSentries, b = notTooFar(&certs, nbCertifs)
				bb = b || minCertifs < int(B.Pars().SigQty)
			}
			if bb {
				dNb++
				d := &Dossier{MinDate: minDate, PrincCertif: princCertif, ProportionOfSentries: proportionOfSentries, Id: new(string), Hash: new(B.Hash), pub: new(B.Pubkey)}
				l.next = new(cdList); l = l.next
				l.cd = d
				var idInBC bool
				idInBC, *d.pub, *d.Id, _, d.limit, b = S.IdHash(toHash); M.Assert(b, 105)
				*d.Hash = toHash
				d.Certifs = make(File, nbCertifs)
				j := 0
				if idInBC {
					from, to, okP := posBI.CertNextPos()
					for okP {
						_, member, _, _, _, _, bb := B.IdPubComplete(from)
						var posBF B.CertPos
						bb = bb && member &&(!B.CertFrom(from, &posBF) || posBF.CertPosLen() < int(B.Pars().SigStock))
						if bb {
							_, exp, b := B.Cert(from, to); M.Assert(b, 106)
							if exp > minDate {
								c := &Certif{date: BA.Already, limit: exp, From: new(string), To: d.Id, ToH: d.Hash}
								c.fromP = new(B.Pubkey); *c.fromP = from
								*c.From, b = B.IdPub(from); M.Assert(b, 107)
								d.Certifs[j] = c
								j++
							}
						}
						from, to, okP = posBI.CertNextPos()
					}
				}
				k := j
				from, toHash, okP := posI.CertNextPos()
				for okP {
					i := 0
					for i < k && *d.Certifs[i].(*Certif).fromP != from {
						i++
					}
					if i >= k {
						_, b, _, _, _, _, bb := B.IdPubComplete(from)
						var posBF B.CertPos
						bb = bb && b &&  (!B.CertFrom(from, &posBF) || posBF.CertPosLen() < int(B.Pars().SigStock))
						if bb {
							_, _, exp, b := S.Cert(from, toHash); M.Assert(b, 106)
							date := fixCertNextDate(from)
							if M.Max64(date, minDate) <= exp {
								useful.SearchIns(&pubSet{p: from})
								c := &Certif{date: date, limit: exp, From: new(string), To: d.Id, ToH: d.Hash}
								c.fromP = new(B.Pubkey); *c.fromP = from
								*c.From, b = B.IdPub(from); M.Assert(b, 107)
								d.Certifs[j] = c
								j++
							}
						}
					}
					from, toHash, okP = posI.CertNextPos()
				}
			}
		}
		toHash, ok = S.IdNextHash(false, &el)
	}
	var (mt int64; b bool)
	n := B.LastBlock()
	if n <= 0 { // Encore utile ?
		mt = 0
	} else {
		mt, _, b = B.TimeOf(n - 1); M.Assert(b, 108)
	}
	cNb = 0; cNbU := 0
	var pos S.CertPos
	ok = S.CertNextTo(true, &pos, &el)
	for ok {
		from, toHash, ok2 := pos.CertNextPos()
		for ok2 {
			to, _, exp, b := S.Cert(from, toHash); M.Assert(b, 109)
			uid, member, _, _, _, _, bb := B.IdPubComplete(to)
			var posBF B.CertPos
			bb = bb && member && (!B.CertFrom(from, &posBF) || posBF.CertPosLen() < int(B.Pars().SigStock))
			if bb {
				date := fixCertNextDate(from)
				// Si la certif n'est pas encore passée au dernier bloc alors qu'elle aurait dû passer à ce bloc ou avant, il y a peu de chance qu'elle passe plus tard et il vaut mieux l'enlever ; il faut aussi l'enlever si elle a dépassé sa date limite ; peut-être inutile maintenant
				if date <= exp && date > mt {
					cNb++
					if _, b, _ := useful.Search(&pubSet{p: from}); b { // Keep only certifications whose sender is also a sender of a certification in SandBox in a dossier
						c := &Certif{fromP: &from, To: &uid, ToH: new(B.Hash), limit: exp, date: date, From: new(string)}
						*c.From, b = B.IdPub(from); M.Assert(b, 110)
						*c.ToH = toHash
						cNbU++
						l.next = new(cdList); l = l.next
						l.cd = c
					}
				}
			}
			from, toHash, ok2 = pos.CertNextPos()
		}
		ok = S.CertNextTo(false, &pos, &el)
	}
	l.next = nil; cdL = cdL.next
	if dNb == 0 {
		f = make(File, 0)
	} else {
		f = make(File, dNb + cNbU)
		i := 0;
		for cdL != nil {
			f[i] = cdL.cd
			i++
			cdL = cdL.next
		}
		sortFile(f, 0)
	}
	return
} //FillFile

// Calculate the current set of entries, sorted by dates (occur) and by names (invOccur)
func BuildEntries () (f File, cNb, dNb int, permutations, occurDate, occurName *A.Tree, duration int64) {
	
	byDate := func (tId *A.Tree) *A.Tree {
		tD :=A.New()
		e := tId.Next(nil)
		for e != nil {
			p := e.Val().(*Propagation)
			_, b, _ := tD.SearchIns(&PropDate{Hash: p.Hash, Id: p.Id, Date: p.Date, After: p.After}); M.Assert(!b, 101)
			e = tId.Next(e)
		}
		return tD
	} //byDate
	
	//BuildEntries
	ti := time.Now()
	f, cNb, dNb = FillFile(int(B.Pars().SigQty))
	n := int64(0)
	permutations, occurDate, occurName = CalcEntries(copyFile(f, 0, &n))
	e := permutations.Next(nil)
	for e != nil {
		s := e.Val().(*Set)
		s.T = byDate(s.T)
		e = permutations.Next(e)
	}
	duration = int64(math.Round(time.Since(ti).Seconds()))
	return
}

func MaxSize () int64 {
	return maxSize
}

func ChangeParameters (newMaxSize int64) {
	maxSize = newMaxSize
}
