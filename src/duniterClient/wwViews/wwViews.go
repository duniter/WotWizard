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
	S	"strconv"
		"strings"
		"html/template"

)

const (
	
	wwName = "00wwView"
	
	subByName = "By"
	subMetaName = "Meta"
	
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
							lastAppDate
							minDate
							expires_on:limit
							certifications {
								date
								certification {
									from {
										uid
									}
									expires_on
								}
							}
						}
					}
				}
			}
		}
	`
	
	html = `
		{{define "head"}}<title>WotWizard</title>{{end}}
		{{define "body"}}
			<h1>WotWizard</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<form action="" method="post">
				<p>
					<input type="radio" id="byName" name="display" value="0"{{if eq .Disp "0"}} checked{{end}}>
					<label for="byName">{{.WWByName}}</label>
					<input type="radio" id="meta" name="display" value="1"{{if eq .Disp "1"}} checked{{end}}>
					<label for="meta">{{.WWMeta}}</label>
					<input type="radio" id="byDate" name="display" value="2"{{if eq .Disp "2"}} checked{{end}}>
					<label for="byDate">{{.WWByDate}}</label>
				</p>
				<p>
					<input type="submit" value="{{.OK}}">
				</p>
			</form>
			{{if .F}}
				{{with .F}}
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
				{{end}}
			{{else}}
				{{range .M}}
					<p>
						{{.First}}
						{{if .Second}}
							<br>
							{{.Second}}
						{{end}}
						<blockquote>
							{{range .Certs}}
								{{.}}
								<br>
							{{end}}
						</blockquote>
					</p>
				{{end}}
			{{end}}
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
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
		LastAppDate,
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
	
	Disp struct {
		Now,
		Disp,
		WWByName,
		WWMeta,
		WWByDate,
		OK string
		F *ForeC
		M DossCertsT
	}

)

func printNow (now *NowT, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct, lang))
} //printNow

// Print the permutation occurD sorted by dates
func printByDates (fs Forecasts, lang *SM.Lang) CorpusA {
	proba := lang.Map("#duniterClient:Proba")
	var (d int64 = -1; a = false; c *CorpusT)
	cs := make(CorpusA, 0)
	for _, f := range fs {
		if f.Date != d || f.After != a {
			d = f.Date; a = f.After
			l := BA.Ts2s(d, lang)
			if a {
				l += "+"
			} else {
				l += " "
			}
			c = &CorpusT{Label: l, Seconds: make(Details, 0)}
			cs = append(cs, c)
		}
		s := fmt.Sprintf("%v%v%v%v%.f%%", f.Id.Uid, BA.SpL, proba, " = ", f.Proba * 100)
		c.Seconds = append(c.Seconds, s)
	}
	return cs
} //printByDates

// Print the permutation occurN sorted by names
func printByNames (fs Forecasts, lang *SM.Lang) CorpusA {
	proba := lang.Map("#duniterClient:Proba")
	var (uid = ""; c *CorpusT)
	cs := make(CorpusA, 0)
	for _, f := range fs {
		if f.Id.Uid != uid {
			uid = f.Id.Uid
			c = &CorpusT{Label: uid, Seconds: make(Details, 0)}
			cs = append(cs, c)
		}
		d := BA.Ts2s(f.Date, lang)
		if f.After {
			d += "+"
		} else {
			d += " "
		}
		s := fmt.Sprintf("%v%v%v%v%.f%%", d, BA.SpL, proba, " = ", f.Proba * 100)
		c.Seconds = append(c.Seconds, s)
	}
	return cs
} //printByNames

func printText (ww *WW, dateNameMeta string, lang *SM.Lang) *Disp {
	r := ww.Data.WWResult
	if r != nil {
		now := printNow(r.Now, lang)
		dnb := fmt.Sprint(r.Dossiers_nb, " ", lang.Map("#duniterClient:newcomers"))
		cnb := fmt.Sprint(r.Certifs_nb, " ", lang.Map("#duniterClient:intCertifs"))
		cd := fmt.Sprint(lang.Map("#duniterClient:Computation_duration"), " = ", r.Computation_duration, "s")
		pnb := fmt.Sprint(r.Permutations_nb, " ", lang.Map("#duniterClient:permutations"))
		var c CorpusA
		switch dateNameMeta {
		case "2":
			c = printByDates(r.ForecastsByDates, lang)
		case "0":
			c = printByNames(r.ForecastsByNames, lang)
		default:
			M.Halt(dateNameMeta, 100)
		}
		bn := lang.Map("#duniterClient:wwByName")
		bd := lang.Map("#duniterClient:wwByDate")
		m := lang.Map("#duniterClient:wwMeta")
		ok := lang.Map("#duniterClient:OK")
		return &Disp{Now: now, F: &ForeC{D_nb: dnb, C_nb: cnb, C_d: cd, P_nb: pnb,Firsts: c}, Disp: dateNameMeta, WWByName: bn, WWMeta: m, WWByDate: bd, OK: ok}
	} else {
		return &Disp{F: &ForeC{Firsts: make(CorpusA, 0)}}
	}
} //printText

// Print the metadata m with the help of f
func printMeta (cds Certifs_DossiersT, lang *SM.Lang) DossCertsT {
	
	PrintCertOrDoss := func (cd *Certif_DossierT) *DossCertT {
		
		PrintCertif := func (c *DatedCertificationT) string {
			cc := c.Certification
			return fmt.Sprint(cc.To.Uid, " ← ", cc.From.Uid, " ", BA.Ts2s(c.Date, lang), " (→ ", BA.Ts2s(cc.Expires_on, lang), ")")
		} //PrintCertif
		
		PrintDossier := func (d *DossierT)  *DossCertT {
			
			PrintCerts := func (cs DatedCertifications) CertsT {
				
				PrintCert := func (c *DatedCertificationT) string {
					cc := c.Certification
					return fmt.Sprint(cc.From.Uid, " ", BA.Ts2s(c.Date, lang), " (→ ", BA.Ts2s(cc.Expires_on, lang), ")")
				} //PrintCert
				
				//PrintCerts
				cc := make(CertsT, len(cs))
				for i, c := range cs {
					cc[i] = PrintCert(&c)
				}
				return cc
			} //PrintCerts
			
			//PrintDossier
			fi := fmt.Sprint(d.Newcomer.Uid, " ", BA.Ts2s(d.Date, lang), " (→ ", BA.Ts2s(d.Expires_on, lang), ") ", lang.Map("#duniterClient:distanceRule", S.Itoa(int(d.Newcomer.Distance.Value))))
			w := new(strings.Builder)
			fmt.Fprint(w, lang.Map("#duniterClient:requiredCertsNb", S.Itoa(len(d.Certifications)), S.Itoa(d.Main_certifs)))
			if d.Date == d.MinDate {
				fmt.Fprint(w, ". ", lang.Map("#duniterClient:minApplicationDate", BA.Ts2s(d.LastAppDate, lang)))
			}
			sd := w.String()
			return &DossCertT{First: fi, Second: sd, Certs: PrintCerts(d.Certifications)}
		} //PrintDossier
		
		//PrintCertOrDoss
		if cd.Dossier != nil {
			return PrintDossier(cd.Dossier)
		}
		PrintCertif(cd.DatedCertification)
		return &DossCertT{First: PrintCertif(cd.DatedCertification), Second: ""}
	} //PrintCertOrDoss
	
	//printMeta
	dcs := make(DossCertsT, len(cds))
	for i, cd := range cds {
		dcs[i] = *PrintCertOrDoss(&cd)
	}
	return dcs
} //printMeta

func printTextM (ww *WW, lang *SM.Lang) *Disp {
	f := ww.Data.WWFile
	if f != nil {
		now := printNow(f.Now, lang)
		dcs := printMeta(f.Certifs_dossiers, lang)
		bn := lang.Map("#duniterClient:wwByName")
		bd := lang.Map("#duniterClient:wwByDate")
		m := lang.Map("#duniterClient:wwMeta")
		ok := lang.Map("#duniterClient:OK")
		return &Disp{Now: now, M: dcs, Disp: "1", WWByName: bn, WWMeta: m, WWByDate: bd, OK: ok}
	} else {
		return &Disp{M: make(DossCertsT, 0)}
	}
} //printTextM

func runSubscription (name string) {
	GS.Send(nil, GS.ExtractDocument(name))
}

func end (name string, temp *template.Template, r *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	M.Assert(name == wwName, name, 100)
	disp := "0"
	if r.Method == "POST" {
		r.ParseForm()
		disp = r.PostFormValue("display")
	}
	var j J.Json
	switch disp {
	case "0", "2":
		for {
			j = GS.GetSub(subByName); M.Assert(j != nil, 101)
			f := j.(*J.Object).Fields
			M.Assert(len(f) > 0, 102)
			if f[0].Name != "errors" {break}
			s := f[0].Value.(*J.String).S
			M.Assert(strings.HasPrefix(s, "Unknown subscription"), s, 103)
			runSubscription(subscriptionBy)
		}
		ww := new(WW)
		J.ApplyTo(j, ww)
		temp.ExecuteTemplate(w, name, printText(ww, disp, lang))
	case "1":
		for {
			j = GS.GetSub(subMetaName); M.Assert(j != nil, 104)
			f := j.(*J.Object).Fields
			M.Assert(len(f) > 0, 105)
			if f[0].Name != "errors" {break}
			s := f[0].Value.(*J.String).S
			M.Assert(strings.HasPrefix(s, "Unknown subscription"), s, 106)
			runSubscription(subscriptionMeta)
		}
		ww := new(WW)
		J.ApplyTo(j, ww)
		temp.ExecuteTemplate(w, name, printTextM(ww, lang))
	default:
		M.Halt(disp, 101)
	}
} //end

func init () {
	runSubscription(subscriptionBy)
	runSubscription(subscriptionMeta)
	W.RegisterPackage(wwName, html, end, true)
}
