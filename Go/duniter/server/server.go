/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package server

// Server of the WotWizard's window

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	F	"path/filepath"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"
	W	"duniter/wotWizard"
		"fmt"
		"math"
		"os"
		"time"

)

const (
	
	parametersName = "parameters.txt" // File containing the greatest allowed allocated memory size for W.CalcPermutations
	
	wwStartName = "WWServerStart"
	wwStopName = "WWServerStop"
	
	wwUpdateName = "WWServerUpdate"
	
	startAction = iota
	stopAction

)

type (
	
	action struct {
		what int
	}
	
	updateA struct {
		output string
	}

)

var (
	
	wwParameters = F.Join(BA.RsrcDir(), parametersName)

)

func writeFile (f W.File, withTo bool, mk *J.Maker) {

	writeCertOrDoss := func (cd W.CertOrDoss) {
		
		writeCertOrDossEnd := func (cd W.CertOrDoss) {
			mk.PushInteger(cd.Date())
			mk.BuildField("date")
			mk.PushInteger(cd.Limit())
			mk.BuildField("limit")
		}
		
		writeCertif := func (c *W.Certif) {
			mk.StartObject()
			mk.PushString(c.From)
			mk.BuildField("from")
			if withTo {
				mk.PushString(c.To)
				mk.BuildField("to")
			}
			writeCertOrDossEnd(c)
			mk.BuildObject()
		}
		
		writeDossier := func (d *W.Dossier) {
			mk.StartObject()
			mk.PushString(d.Id)
			mk.BuildField("newcomer")
			mk.PushInteger(int64(d.PrincCertif))
			mk.BuildField("main_certifs")
			mk.PushFloat(d.ProportionOfSentries)
			mk.BuildField("proportion_of_sentries")
			writeCertOrDossEnd(d)
			mk.PushInteger(d.MinDate)
			mk.BuildField("minimum_date")
			writeFile(d.Certifs, false, mk)
			mk.BuildField("certifs")
			mk.BuildObject()
		}
	
		//writeCertOrDoss
		switch cdd := cd.(type) {
		case *W.Certif:
			if withTo {
				mk.StartObject()
			}
			writeCertif(cdd)
			if withTo {
				mk.BuildField("certif")
				mk.BuildObject()
			}
		case *W.Dossier:
			mk.StartObject()
			writeDossier(cdd)
			mk.BuildField("dossier")
			mk.BuildObject()
		}
	}

	//writeFile
	mk.StartArray()
	if f != nil {
		for i := 0; i < len(f); i++ {
			writeCertOrDoss(f[i])
		}
	}
	mk.BuildArray()
}

// Write the metadata duration, f, cNb & dNb with the help of fo, in json format; duration is the computation duration, f is the Duniter1WotWizard.File structure, cNb and dNb are respectively the numbers of internal certifications and of external dossiers
func writeMeta (duration int64, f W.File, permutations, cNb, dNb int, mk *J.Maker) {
	mk.StartObject()
	mk.PushInteger(int64(permutations))
	mk.BuildField("permutations")
	mk.PushInteger(duration)
	mk.BuildField("computation_duration")
	mk.PushInteger(int64(cNb))
	mk.BuildField("certNb")
	mk.PushInteger(int64(dNb))
	mk.BuildField("dossNb")
	writeFile(f, true, mk)
	mk.BuildField("certifs_dossiers")
	mk.BuildObject()
}

// Print the entries sorted by dates in json format
func byDates (occur *A.Tree, mk *J.Maker) {
	date := int64(-1)
	var after bool
	e := occur.Next(nil)
	for e != nil {
		el := e.Val().(*W.PropDate)
		if (el.Date != date) || (el.After != after) {
			if date >= 0 {
				mk.BuildArray()
				mk.BuildField("names")
				mk.BuildObject()
			}
			date = el.Date; after = el.After
			mk.StartObject()
			mk.PushBoolean(el.After)
			mk.BuildField("after")
			mk.PushInteger(el.Date)
			mk.BuildField("date")
			mk.StartArray()
		}
		mk.StartObject()
		mk.PushString(el.Id)
		mk.BuildField("name")
		mk.PushFloat(el.Proba)
		mk.BuildField("proba")
		mk.BuildObject()
		e = occur.Next(e)
	}
	if date >= 0 {
		mk.BuildArray()
		mk.BuildField("names")
		mk.BuildObject()
	}
}

// Print the entries sorted by names in json format
func byNames (occur *A.Tree, mk *J.Maker) {
	id := ""
	e := occur.Next(nil)
	for e != nil {
		el := e.Val().(*W.PropName)
		if el.Id != id {
			if id != "" {
				mk.BuildArray()
				mk.BuildField("dates")
				mk.BuildObject()
			}
			id = el.Id
			mk.StartObject()
			mk.PushString(el.Id)
			mk.BuildField("name")
			mk.StartArray()
		}
		mk.StartObject()
		mk.PushBoolean(el.After)
		mk.BuildField("after")
		mk.PushInteger(el.Date)
		mk.BuildField("date")
		mk.PushFloat(el.Proba)
		mk.BuildField("proba")
		mk.BuildObject()
		e = occur.Next(e)
	}
	if id != "" {
		mk.BuildArray()
		mk.BuildField("dates")
		mk.BuildObject()
	}
}

// Build the WotWizard lists in json format
func list () J.Json {
	ti := time.Now()
	var (
		f W.File
		permutations,
		cNb,
		dNb int
		occurDate,
		occurName *A.Tree
		ok bool
	)
	if f, permutations, cNb, dNb, occurDate, occurName, ok = W.BuildEntries(); !ok {
		occurDate = A.New()
		occurName = A.New()
	}
	duration := int64(math.Round(time.Since(ti).Seconds()))
	mk := J.NewMaker()
	mk.StartObject()
	writeMeta(duration, f, permutations, cNb, dNb, mk)
	mk.BuildField("meta")
	mk.StartArray()
	byDates(occurDate, mk)
	mk.BuildArray()
	mk.BuildField("dates")
	mk.StartArray()
	byNames(occurName, mk)
	mk.BuildArray()
	mk.BuildField("names")
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("now")
	mk.BuildObject()
	return mk.GetJson()
}

// Save the WotWizard lists on disk, in file wwResult in json format
func storeText (name ... interface{}) {
	f, err := M.InstantCreate(name[0].(string)); M.Assert(err == nil, err, 100)
	defer M.InstantClose(f)
	J.Fprint(f, list())
}

func (u *updateA) Name () string {
	return wwUpdateName
}

func (u *updateA) Activate () {
	storeText(u.output)
}

func start (output string, newAction chan<- B.Actioner, fields ...string) {
	B.AddUpdateProc(wwUpdateName, storeText, output)
	newAction <- &updateA{output: output}
}

func stop (output string, newAction chan<- B.Actioner, fields ...string) {
	B.RemoveUpdateProc(wwUpdateName)
}

func storeParameters (maxSize int64) {
	f, err := os.Create(wwParameters); M.Assert(err == nil, err, 100)
	defer f.Close()
	fmt.Fprintln(f, maxSize)
}

func init () {
	f, err := os.Open(wwParameters); M.Assert(err == nil || os.IsNotExist(err), err, 100)
	if err != nil {
		storeParameters(W.MaxSize())
	} else {
		defer f.Close()
		var maxSize int64
		_, err := fmt.Fscan(f, &maxSize); M.Assert(err == nil, err, 101)
		W.ChangeParameters(maxSize)
	}
	G.AddAction(wwStartName, start, G.Arguments{})
	G.AddAction(wwStopName, stop, G.Arguments{})
}
