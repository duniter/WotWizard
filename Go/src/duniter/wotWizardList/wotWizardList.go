/* 
Duniter1: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package wotWizardList

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	W	"duniter/wotWizard"

)

const (
	
	fileName = "WotWizardListFile"
	permName = "WotWizardListPerm"
	
	fileA = iota
	permA

)

type (
	
	action struct {
		what int
		output string
	}

)

func listFile (f W.File, withTo bool, mk *J.Maker) {

	var now int64
	
	listCertOrDoss := func (cd W.CertOrDoss) {
		
		listCertOrDossEnd := func (cd W.CertOrDoss) {
			mk.PushInteger(cd.Date())
			mk.BuildField("date")
			mk.PushInteger(cd.Limit())
			mk.BuildField("limit")
		}
		
		// List c
		listCertif := func (c *W.Certif) {
			mk.StartObject()
			mk.PushString(c.From)
			mk.BuildField("from")
			if withTo {
				mk.PushString(c.To)
				mk.BuildField("to")
			}
			mk.PushBoolean(c.Date() < now && now <= c.Limit())
			mk.BuildField("ok")
			listCertOrDossEnd(c)
			mk.BuildObject()
		}
		
		// List d
		listDossier := func (d *W.Dossier) {
			mk.StartObject()
			mk.PushString(d.Id)
			mk.BuildField("newcomer")
			mk.PushInteger(int64(d.PrincCertif))
			mk.BuildField("main_certifs")
			mk.PushFloat(d.ProportionOfSentries)
			mk.BuildField("proportion_of_sentries")
			mk.PushBoolean(d.PrincCertif >= int(B.Pars().SigQty) && now <= d.Limit() && d.ProportionOfSentries >= B.Pars().Xpercent)
			mk.BuildField("ok")
			listCertOrDossEnd(d)
			mk.PushInteger(d.MinDate)
			mk.BuildField("minimum_date")
			listFile(d.Certifs, false, mk)
			mk.BuildField("certifs")
			mk.BuildObject()
		}
		
		// listCertOrDoss
		switch cdd := cd.(type) {
		case *W.Certif:
			if withTo {
				mk.StartObject()
			}
			listCertif(cdd)
			if withTo {
				mk.BuildField("certif")
				mk.BuildObject()
			}
		case *W.Dossier:
			mk.StartObject()
			listDossier(cdd)
			mk.BuildField("dossier")
			mk.BuildObject()
		}
	}
	
	// listFile
	mk.StartArray()
	if f != nil {
		now = B.Now()
		for _, cd := range f {
			listCertOrDoss(cd)
		}
	}
	mk.BuildArray()
}

// List f with fo, starting at the element of rank i0; if withNow, the output begins with the listing of the current date
func doListFile (f W.File) J.Json {
	mk := J.NewMaker()
	mk.StartObject()
	listFile(f, true, mk)
	mk.BuildField("file")
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
	mk.BuildObject()
	return mk.GetJson()
}

// Print the current W.File
func doShowFile () J.Json {
	f, _, _ := W.FillFile(0)
	return doListFile(f)
}

// List permutations returned by CalcPermutations
func listPermutations (f W.File) J.Json {

	byDate := func (tId *A.Tree) *A.Tree {
		tD :=A.New()
		e := tId.Next(nil)
		for e != nil {
			p := e.Val().(*W.Propagation)
			_, b, _ := tD.SearchIns(&W.PropDate{Id: p.Id, Date: p.Date, After: p.After}); M.Assert(!b, 101)
			e = tId.Next(e)
		}
		return tD
	}

	//listPermutations
	mk := J.NewMaker()
	mk.StartObject()
	mk.StartArray()
	if sets, ok := W.CalcPermutations(f); ok {
		e := sets.Next(nil)
		for e != nil {
			mk.StartObject()
			s := e.Val().(*W.Set)
			mk.PushFloat(s.Proba)
			mk.BuildField("proba")
			tD := byDate(s.T)
			mk.StartArray()
			ee := tD.Next(nil)
			for ee != nil {
				mk.StartObject()
				p := ee.Val().(*W.PropDate)
				mk.PushString(p.Id)
				mk.BuildField("id")
				mk.PushInteger(p.Date)
				mk.BuildField("date")
				mk.PushBoolean(p.After)
				mk.BuildField("after")
				mk.BuildObject()
				ee = tD.Next(ee)
			}
			mk.BuildArray()
			mk.BuildField("permutation")
			mk.BuildObject()
			e = sets.Next(e)
		}
	}
	mk.BuildArray()
	mk.BuildField("permutations")
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
	mk.BuildObject()
	return mk.GetJson()
}

// Print the set of current possible permutations of entries
func doPermutations () J.Json {
	f, _, _ := W.FillFile(int(B.Pars().SigQty))
	return listPermutations(f)
}

func (a *action) Name () string {
	var s string
	switch a.what {
	case fileA:
		s = fileName
	case permA:
		s = permName
	}
	return s
}

func (a *action) Activate () {
	switch a.what {
	case fileA:
		G.Json(doShowFile(), a.output)
	case permA:
		G.Json(doPermutations(), a.output)
	}
}

func file (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: fileA, output: output}
}

func permutations (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{what: permA, output: output}
}

func init () {
	G.AddAction(fileName, file, G.Arguments{})
	G.AddAction(permName, permutations, G.Arguments{})
}
