/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package wotWizardList

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	M	"util/misc"
	W	"duniter/wotWizard"
		/*
		"fmt"
		*/

)

var (
	
	fileStream = GQ.CreateStream("wwFile")
	resultStream = GQ.CreateStream("wwResult")

)

func wwFileStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return fileStream
} //wwFileStreamResolver

func wwServerStreamResolver (rootValue *G.OutputObjectValue, argumentValues *A.Tree) *G.EventStream { // *G.ValMapItem
	return resultStream
} //wwServerStreamResolver

func wwFileR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	var (v G.Value; full bool)
	ok := G.GetValue(argumentValues, "full", &v)
	M.Assert(ok, 100)
	switch v := v.(type) {
	case *G.BooleanValue:
		full = v.Boolean
	default:
		M.Halt(101)
	}
	q := 0;
	if !full {
		q = int(B.Pars().SigQty)
	}
	f, cNb, dNb := W.FillFile(q)
	return GQ.Wrap(B.LastBlock(), f, cNb, dNb)
} //wwFileR

func wwResultR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	_, cNb, dNb, permutations, occurDate, occurName, duration := W.BuildEntries()
	return GQ.Wrap(B.LastBlock(), permutations, occurDate, occurName, duration, dNb, cNb)
} //wwResultR

func resNowR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch b := GQ.Unwrap(rootValue, 0).(type) {
	case int32:
		return GQ.Wrap(b)
	default:
		M.Halt(b, 100)
		return nil
	}
} //resNowR

func resPermsNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch permutations := GQ.Unwrap(rootValue, 1).(type) {
	case *A.Tree:
		return G.MakeIntValue(permutations.NumberOfElems())
	default:
		M.Halt(permutations, 100)
		return nil
	}
} //resPermsNbR

func resPermsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch permutations := GQ.Unwrap(rootValue, 1).(type) {
	case *A.Tree:
		l := G.NewListValue()
		e := permutations.Next(nil)
		for e != nil {
			l.Append(GQ.Wrap(e.Val().(*W.Set)))
			e = permutations.Next(e)
		}
		return l
	default:
		M.Halt(permutations, 100)
		return nil
	}
} //resPermsR

func resDurationR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch duration := GQ.Unwrap(rootValue, 4).(type) {
	case int64:
		return G.MakeInt64Value(duration)
	default:
		M.Halt(duration, 100)
		return nil
	}
} //resDurationR

func resDossiersNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch dNb := GQ.Unwrap(rootValue, 5).(type) {
	case int:
		return G.MakeIntValue(dNb)
	default:
		M.Halt(dNb, 100)
		return nil
	}
} //resDossiersNbR

func resCertifsNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch cNb := GQ.Unwrap(rootValue, 6).(type) {
	case int:
		return G.MakeIntValue(cNb)
	default:
		M.Halt(cNb, 100)
		return nil
	}
} //resCertifsNbR

func resByDatesR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch occurDate := GQ.Unwrap(rootValue, 2).(type) {
	case *A.Tree:
		l := G.NewListValue()
		e := occurDate.Next(nil)
		for e != nil {
			l.Append(GQ.Wrap(e.Val().(*W.PropDate)))
			e = occurDate.Next(e)
		}
		return l
	default:
		M.Halt(occurDate, 100)
		return nil
	}
} //resByDatesR

func resByNamesR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch occurName := GQ.Unwrap(rootValue, 3).(type) {
	case *A.Tree:
		var (v G.Value; hint string)
		if G.GetValue(argumentValues, "with", &v) {
			switch v := v.(type) {
			case *G.StringValue:
				hint = v.String.S
			default:
				M.Halt(v, 100)
			}
		} else {
			M.Halt(101)
		}
		l := G.NewListValue()
		hint = BA.ToDown(hint)
		e, _, _ := occurName.SearchNext(&W.PropName{Id: hint})
		for e != nil && BA.Prefix(hint, BA.ToDown(e.Val().(*W.PropName).Id)) {
			l.Append(GQ.Wrap(e.Val().(*W.PropName)))
			e = occurName.Next(e)
		}
		return l
	default:
		M.Halt(occurName, 100)
		return nil
	}
} //resByNamesR

func wPermProbaR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch set := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Set:
		return G.MakeFloat64Value(set.Proba)
	default:
		M.Halt(set, 100)
		return nil
	}
} //wPermProbaR

func wPermPermR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch set := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Set:
		perm := set.T
		l := G.NewListValue()
		e := perm.Next(nil)
		for e != nil {
			l.Append(GQ.Wrap(e.Val().(*W.PropDate)))
			e = perm.Next(e)
		}
		return l
	default:
		M.Halt(set, 100)
		return nil
	}
} //wPermPermR

func permEIdR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch permE := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return GQ.Wrap(permE.Hash)
	default:
		M.Halt(permE, 100)
		return nil
	}
} //permEIdR

func permEDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch permE := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return G.MakeInt64Value(permE.Date)
	default:
		M.Halt(permE, 100)
		return nil
	}
} //permEDateR

func permEAfterR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch permE := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return G.MakeBooleanValue(permE.After)
	default:
		M.Halt(permE, 100)
		return nil
	}
} //permEAfterR

func forecastIdR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch forecast := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return GQ.Wrap(forecast.Hash)
	case *W.PropName:
		return GQ.Wrap(forecast.Hash)
	default:
		M.Halt(forecast, 100)
		return nil
	}
} //forecastIdR

func forecastDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch forecast := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return G.MakeInt64Value(forecast.Date)
	case *W.PropName:
		return G.MakeInt64Value(forecast.Date)
	default:
		M.Halt(forecast, 100)
		return nil
	}
} //forecastDateR

func forecastAfterR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch forecast := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return G.MakeBooleanValue(forecast.After)
	case *W.PropName:
		return G.MakeBooleanValue(forecast.After)
	default:
		M.Halt(forecast, 100)
		return nil
	}
} //forecastAfterR

func forecastProbaR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch forecast := GQ.Unwrap(rootValue, 0).(type) {
	case *W.PropDate:
		return G.MakeFloat64Value(forecast.Proba)
	case *W.PropName:
		return G.MakeFloat64Value(forecast.Proba)
	default:
		M.Halt(forecast, 100)
		return nil
	}
} //forecastProbaR

func fileNowR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch b := GQ.Unwrap(rootValue, 0).(type) {
	case int32:
		return GQ.Wrap(b)
	default:
		M.Halt(b, 100)
		return nil
	}
} //fileNowR

func fileCDR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch f := GQ.Unwrap(rootValue, 1).(type) {
	case W.File:
		l := G.NewListValue()
		for _, cd := range f {
			l.Append(GQ.Wrap(cd))
		}
		return l
	default:
		M.Halt(f, 100)
		return nil
	}
} //fileCDR

func fileCNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch cNb := GQ.Unwrap(rootValue, 2).(type) {
	case int:
		return G.MakeIntValue(cNb)
	default:
		M.Halt(cNb, 100)
		return nil
	}
} //fileCNbR

func fileDNbR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch dNb := GQ.Unwrap(rootValue, 3).(type) {
	case int:
		return G.MakeIntValue(dNb)
	default:
		M.Halt(dNb, 100)
		return nil
	}
} //fileDNbR

func mCertCertR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch GQ.Unwrap(rootValue, 0).(type) {
	case *W.Certif:
		return rootValue
	default:
		M.Halt(100)
		return nil
	}
} //mCertCertR

func certCertR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch c := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Certif:
		_, _, from, _, _, _, ok := B.IdUidComplete(*c.From); M.Assert(ok, 101)
		return GQ.Wrap(from, *c.ToH, true)
	default:
		M.Halt(c, 100)
		return nil
	}
} //certCertR

func certDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch c := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Certif:
		return G.MakeInt64Value(c.Date())
	default:
		M.Halt(c, 100)
		return nil
	}
} //certDateR

func mDossierDossierR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return rootValue
	default:
		M.Halt(100)
		return nil
	}
} //mDossierDossierR

func dossierNewcomerR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return GQ.Wrap(*d.Hash)
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierNewcomerR

func dossierMainCR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return G.MakeIntValue(d.PrincCertif)
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierMainCR

func dossierCertsR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		l := G.NewListValue()
		for _, cd := range d.Certifs {
			l.Append(GQ.Wrap(cd.(*W.Certif)))
		}
		return l
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierCertsR

func dossierMinDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return G.MakeInt64Value(d.MinDate)
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierMinDateR

func dossierDateR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return G.MakeInt64Value(d.Date())
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierDateR

func dossierLimitR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	switch d := GQ.Unwrap(rootValue, 0).(type) {
	case *W.Dossier:
		return G.MakeInt64Value(d.Limit())
	default:
		M.Halt(d, 100)
		return nil
	}
} //dossierLimitR

func fixFieldResolvers (ts G.TypeSystem) {
	ts.FixFieldResolver("Query", "wwFile", wwFileR)
	ts.FixFieldResolver("Query", "wwResult", wwResultR)
	ts.FixFieldResolver("FileS", "now", fileNowR)
	ts.FixFieldResolver("FileS", "certifs_dossiers", fileCDR)
	ts.FixFieldResolver("FileS", "certifs_nb", fileCNbR)
	ts.FixFieldResolver("FileS", "dossiers_nb", fileDNbR)
	ts.FixFieldResolver("MarkedDatedCertification", "datedCertification", mCertCertR)
	ts.FixFieldResolver("DatedCertification", "certification", certCertR)
	ts.FixFieldResolver("DatedCertification", "date", certDateR)
	ts.FixFieldResolver("MarkedDossier", "dossier", mDossierDossierR)
	ts.FixFieldResolver("Dossier", "newcomer", dossierNewcomerR)
	ts.FixFieldResolver("Dossier", "main_certifs", dossierMainCR)
	ts.FixFieldResolver("Dossier", "certifications", dossierCertsR)
	ts.FixFieldResolver("Dossier", "minDate", dossierMinDateR)
	ts.FixFieldResolver("Dossier", "date", dossierDateR)
	ts.FixFieldResolver("Dossier", "limit", dossierLimitR)
	ts.FixFieldResolver("WWResultS", "now", resNowR)
	ts.FixFieldResolver("WWResultS", "computation_duration", resDurationR)
	ts.FixFieldResolver("WWResultS", "permutations_nb", resPermsNbR)
	ts.FixFieldResolver("WWResultS", "dossiers_nb", resDossiersNbR)
	ts.FixFieldResolver("WWResultS", "certifs_nb", resCertifsNbR)
	ts.FixFieldResolver("WWResultS", "permutations", resPermsR)
	ts.FixFieldResolver("WWResultS", "forecastsByDates", resByDatesR)
	ts.FixFieldResolver("WWResultS", "forecastsByNames", resByNamesR)
	ts.FixFieldResolver("WeightedPermutation", "proba", wPermProbaR)
	ts.FixFieldResolver("WeightedPermutation", "permutation", wPermPermR)
	ts.FixFieldResolver("PermutationElem", "id", permEIdR)
	ts.FixFieldResolver("PermutationElem", "date", permEDateR)
	ts.FixFieldResolver("PermutationElem", "after", permEAfterR)
	ts.FixFieldResolver("Forecast", "id", forecastIdR)
	ts.FixFieldResolver("Forecast", "date", forecastDateR)
	ts.FixFieldResolver("Forecast", "after", forecastAfterR)
	ts.FixFieldResolver("Forecast", "proba", forecastProbaR)
	ts.FixFieldResolver("Subscription", "wwFile", wwFileR)
	ts.FixFieldResolver("Subscription", "wwResult", wwResultR)
} //fixFieldResolvers

func fixStreamResolvers (ts G.TypeSystem) {
	ts.FixStreamResolver("wwFile", wwFileStreamResolver)
	ts.FixStreamResolver("wwResult", wwServerStreamResolver)
}

func init () {
	ts := GQ.TS()
	fixFieldResolvers(ts)
	fixStreamResolvers(ts)
}
