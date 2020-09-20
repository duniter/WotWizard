/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package wwViews

// This package creates views displaying WotWizard forecasts
	
import (
	
	BA	"duniterClient/basicPrint"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"strings"
		"html/template"

)

const (
	
	byNameName = "wwByName"
	byDateName = "wwByDate"
	metaName = "wwMeta"
	
	subByName = "wwResult"
	subMetaName = "wwFile"
	
	subscriptionBy = `
		subscription By {
			wwResult {
				now {
					number
					bct
				}
				computation_duration
				permutations_nb
				dossiers_nb
				certifs_nb
				forecastsByNames {
					id {
						uid
					}
					date
					after
					proba
				}
				forecastsByDates {
					id {
						uid
					}
					date
					after
					proba
				}
			}
		}
	`
	
	subscriptionMeta = `
		subscription Meta {
			wwFile(full: false) {
				now {
					number
					bct
				}
				certifs_dossiers {
					...on MarkedDatedCertification {
						datedCertification {
							date
							certification {
								from {
									uid
								}
								to {
									uid
								}
								expires_on
							}
						}
					}
					...on MarkedDossier {
						dossier {
							main_certifs
							newcomer {
								uid
								distance {
									value
								}
							}
							date
							minDate
							expires_on:limit
							certifications {
								date
								certification {
									from {
										uid
									}
									pending
									expires_on
								}
							}
						}
					}
				}
			}
		}
	`
	
	htmlWW = `
		{{define "head"}}<title>WotWizard</title>{{end}}
		{{define "body"}}
			<h1>WotWizard</h1>
			<p>
				<a href = "/">index</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<h4>
				{{.D_nb}}
			</h4>
			<p>
				{{.C_d}}
				<br>
				{{.C_nb}}
				<br>
				{{.P_nb}}
			</p>
			{{range .Firsts}}
				<h5>
					{{.Label}}
				</h5>
				<p>
				<blockquote>
					{{range .Seconds}}
						{{.}}
						<br>
					{{end}}
				</blockquote>
				</p>
			{{end}}
			<p>
				<a href = "/">index</a>
			</p>
		{{end}}
	`
	
	htmlMeta = `
		{{define "head"}}<title>WotWizard</title>{{end}}
		{{define "body"}}
			<h1>WotWizard</h1>
			<p>
				<a href = "/">index</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			{{range .Dcs}}
				<p>
					{{.First}}
					{{if .Second}}
						<br>
						{{.Second}}
						<blockquote>
							{{range .Certs}}
								{{.}}
								<br>
							{{end}}
						</blockquote>
					{{end}}
				</p>
			{{end}}
			<p>
				<a href = "/">index</a>
			</p>
		{{end}}
	`
	
	htmlVoid = `
		{{define "head"}}{{end}}
		{{define "body"}}{{end}}
	`

)

type (
	
	NowT struct {
		Number int32
		Bct int64
	}
	
	ForecastT struct {
		Id struct {
			Uid string
		}
		Date int64
		After bool
		Proba float64
	}
	
	Forecasts []ForecastT
	
	WWResultT struct {
		Now *NowT
		Computation_duration,
		Permutations_nb,
		Dossiers_nb,
		Certifs_nb int
		ForecastsByDates,
		ForecastsByNames Forecasts
	}
	
	CertificationT struct {
		From struct {
			Uid string
		}
		To struct {
			Uid string
		}
		Pending bool
		Expires_on int64
	}
	
	DatedCertificationT struct {
		Date int64
		Certification *CertificationT
	}
	
	DatedCertifications []DatedCertificationT
	
	DossierT struct {
		Main_certifs int
		Newcomer struct {
			Uid string
			Distance struct {
				Value float64
			}
		}
		Date,
		MinDate,
		Expires_on int64
		Certifications DatedCertifications
	}
	
	Certif_DossierT struct {
		Dossier *DossierT
		DatedCertification *DatedCertificationT
	}
	
	Certifs_DossiersT []Certif_DossierT
	
	WWFileT struct {
		Now *NowT
		Certifs_dossiers Certifs_DossiersT
	}
	
	WW struct {
		Data struct {
			WWResult *WWResultT
			WWFile *WWFileT
		}
	}
	
	//Outputs
	
	Details []string
	
	CorpusT struct {
		Label string
		Seconds Details 
	}
	
	CorpusA []*CorpusT
	
	ForeC struct {
		Now,
		D_nb,
		C_nb,
		C_d,
		P_nb string
		Firsts CorpusA
	}
	
	CertT = string
	
	CertsT []CertT
	
	DossCertT struct{
		First,
		Second string
		Certs CertsT
	}
	
	DossCertsT []DossCertT
	
	DispM struct {
		Now string
		Dcs DossCertsT
	}

)

func printNow (now *NowT) string {
	return fmt.Sprint(SM.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct))
} //printNow

// Print the permutation occurD sorted by dates
func printByDates (fs Forecasts) CorpusA {
	proba := SM.Map("#duniterClient:Proba")
	var (d int64 = -1; a = false; c *CorpusT)
	cs := make(CorpusA, 0)
	for _, f := range fs {
		if f.Date != d || f.After != a {
			d = f.Date; a = f.After
			l := BA.Ts2s(d)
			if a {
				l += "+"
			} else {
				l += " "
			}
			c = &CorpusT{Label: l, Seconds: make(Details, 0)}
			cs = append(cs, c)
		}
		s := fmt.Sprintf("%v%v%v%v%.f%%", f.Id.Uid, "    ", proba, " = ", f.Proba * 100)
		c.Seconds = append(c.Seconds, s)
	}
	return cs
} //printByDates

// Print the permutation occurN sorted by names
func printByNames (fs Forecasts) CorpusA {
	proba := SM.Map("#duniterClient:Proba")
	var (uid = ""; c *CorpusT)
	cs := make(CorpusA, 0)
	for _, f := range fs {
		if f.Id.Uid != uid {
			uid = f.Id.Uid
			c = &CorpusT{Label: uid, Seconds: make(Details, 0)}
			cs = append(cs, c)
		}
		d := BA.Ts2s(f.Date)
		if f.After {
			d += "+"
		} else {
			d += " "
		}
		s := fmt.Sprintf("%v%v%v%v%.f%%", d, "    ", proba, " = ", f.Proba * 100)
		c.Seconds = append(c.Seconds, s)
	}
	return cs
} //printByNames

func printText (ww *WW, dateNameMeta string) *ForeC {
	r := ww.Data.WWResult
	if r != nil {
		now := printNow(r.Now)
		dnb := fmt.Sprint(r.Dossiers_nb, " ", SM.Map("#duniterClient:newcomers"))
		cnb := fmt.Sprint(r.Certifs_nb, " ", SM.Map("#duniterClient:extCertifs"))
		cd := fmt.Sprint(SM.Map("#duniterClient:Computation_duration"), " = ", r.Computation_duration, "s")
		pnb := fmt.Sprint(r.Permutations_nb, " ", SM.Map("#duniterClient:permutations"))
		var c CorpusA
		switch dateNameMeta {
		case byDateName:
			c = printByDates(r.ForecastsByDates)
		case byNameName:
			c = printByNames(r.ForecastsByNames)
		default:
			M.Halt(dateNameMeta, 100)
		}
		return &ForeC{Now: now, D_nb: dnb, C_nb: cnb, C_d: cd, P_nb: pnb,Firsts: c}
	} else {
		return &ForeC{Firsts: make(CorpusA, 0)}
	}
} //printText

// Print the metadata m with the help of f
func printMeta (cds Certifs_DossiersT) DossCertsT {
	
	PrintCertOrDoss := func (cd *Certif_DossierT) *DossCertT {
		
		PrintCertif := func (c *DatedCertificationT) string {
			cc := c.Certification
			return fmt.Sprint(cc.To.Uid, " ← ", cc.From.Uid, " ", BA.Ts2s(c.Date), " (→ ", BA.Ts2s(cc.Expires_on), ")")
		} //PrintCertif
		
		PrintDossier := func (d *DossierT)  *DossCertT {
			
			PrintCerts := func (cs DatedCertifications) CertsT {
				
				PrintCert := func (c *DatedCertificationT) string {
					cc := c.Certification
					return fmt.Sprint(cc.From.Uid, " ", BA.Ts2s(c.Date), " (→ ", BA.Ts2s(cc.Expires_on), ")")
				} //PrintCert
				
				//PrintCerts
				cc := make(CertsT, len(cs))
				for i, c := range cs {
					cc[i] = PrintCert(&c)
				}
				return cc
			} //PrintCerts
			
			//PrintDossier
			fi := fmt.Sprint(d.Main_certifs, " ", d.Newcomer.Uid, " (", BA.Ts2s(d.Date), " ≥ ", BA.Ts2s(d.MinDate), ")")
			sd := fmt.Sprint("(→ ", BA.Ts2s(d.Expires_on), ") (", int(d.Newcomer.Distance.Value), "%) |")
			return &DossCertT{First: fi, Second: sd, Certs: PrintCerts(d.Certifications)}
		} //PrintDossier
		
		//PrintCertOrDoss
		if cd.Dossier != nil {
			return PrintDossier(cd.Dossier)
		}
		PrintCertif(cd.DatedCertification)
		return &DossCertT{First: PrintCertif(cd.DatedCertification), Second: ""}
	} //PrintCertOrDoss
	
	//printFile
	dcs := make(DossCertsT, len(cds))
	for i, cd := range cds {
		dcs[i] = *PrintCertOrDoss(&cd)
	}
	return dcs
} //printMeta

func printTextM (ww *WW) *DispM {
	f := ww.Data.WWFile
	if f != nil {
		now := printNow(f.Now)
		dcs := printMeta(f.Certifs_dossiers)
		return &DispM{Now: now, Dcs: dcs}
	} else {
		return &DispM{Dcs: make(DossCertsT, 0)}
	}
} //printTextM

func runSubscription (name string) {
	GS.Send(nil, GS.ExtractDocument(name))
}

func endW (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	M.Assert(name == byNameName || name == byDateName, name, 100)
	var j J.Json
	for {
		j = GS.SendSub(subByName)
		M.Assert(j != nil, 101)
		f := j.(*J.Object).Fields
		M.Assert(len(f) > 0, 102)
		if f[0].Name != "errors" {break}
		s := f[0].Value.(*J.String).S
		M.Assert(strings.HasPrefix(s, "Unknown subscription"), s, 103)
		runSubscription(subscriptionBy)
	}
	ww := new(WW)
	J.ApplyTo(j, ww)
	temp.ExecuteTemplate(w, name, printText(ww, name))
} //endW

func endM (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	M.Assert(name == metaName, name, 100)
	var j J.Json
	for {
		j = GS.SendSub(subMetaName)
		M.Assert(j != nil, 101)
		f := j.(*J.Object).Fields
		M.Assert(len(f) > 0, 102)
		if f[0].Name != "errors" {break}
		s := f[0].Value.(*J.String).S
		M.Assert(strings.HasPrefix(s, "Unknown subscription"), s, 103)
		runSubscription(subscriptionMeta)
	}
	ww := new(WW)
	J.ApplyTo(j, ww)
	temp.ExecuteTemplate(w, name, printTextM(ww))
} //endW

func init () {
	runSubscription(subscriptionBy)
	runSubscription(subscriptionMeta)
	W.RegisterPackage(byNameName, htmlWW, endW, true)
	W.RegisterPackage(byDateName, htmlWW, endW, true)
	W.RegisterPackage(metaName, htmlMeta, endM, true)
}
