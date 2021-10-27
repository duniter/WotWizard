/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package wotWizardPrint

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
	
	wwFileName = "01wwFile"
	wwPermsName = "02wwPerms"
	
	queryWWFile = `
		query WWFile {
			now {
				number
				bct
			}
			parameter(name: sigQty) {
				sigQty:value
			}
			wwFile(full: true) {
				certifs_dossiers {
					... on MarkedDatedCertification {
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
					... on MarkedDossier {
						dossier {
							main_certifs
							newcomer {
								uid
								lastApplication {
									lastAppDate: bct
								}
								distance {
									value
									dist_ok
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
										quality
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
	
	htmlWWFile = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<p>
				{{with .Stats}}
					{{.Number}} {{.DossName}}
					{{$n := .Needed}} {{$d := .DossName}}
					<blockquote>
						{{range $i, $l := .List}}
							{{$i}} {{$n}} {{$l}} {{$d}}
							<br>
						{{end}}
					</blockquote>
				{{end}}
			</p>
			{{range .DossCerts}}
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
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`
	
	queryWWPerms = `
		query WWPerms {
			now {
				number
				bct
			}
			wwResult {
				permutations {
					proba
					permutation {
						id {
							uid
						}
						date
						after
					}
				}
			}
		}
	`
	
	htmlWWPerms = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			<p>
				{{.Number}}
			</p>
				{{range .Perms}}
					<p>
						{{.Proba}}
						<br>
						{{range .Props}}
							{{.}}
							<br>
						{{end}}
					</p>
				{{end}}
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`

)

type (
	
	NowT struct {
		Number int
		Bct int64
	}
	
	IdentityT struct {
		Uid string
		Quality float64
		LastApplication struct {
			LastAppDate int64
		}
		Distance struct {
			Value float64
			Dist_ok bool
		}
	}
	
	CertificationT struct {
		From,
		To *IdentityT
		Expires_on int64
	}
	
	DatedCertificationT struct {
		Date int64
		Certification *CertificationT
	}
	
	DatedCertifications []DatedCertificationT
	
	DossierT struct {
		Main_certifs int
		Newcomer *IdentityT
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
	
	DataT struct {
		Now *NowT
		Parameter struct {
			SigQty int
		}
		WWFile struct {
			Certifs_dossiers Certifs_DossiersT
		}
	}
	
	File struct {
		Data DataT
	}
	
	Propagation struct {
		Id *IdentityT
		Date int64
		After bool
	}
	
	PermutationES []Propagation
	
	WPermutation struct {
		Proba float64
		Permutation PermutationES
	}
	
	Permutations0 []WPermutation
	
	DataP struct {
		Now *NowT
		WWResult struct {
			Permutations Permutations0
		}
	}
	
	Perms struct {
		Data *DataP
	}
	
	//Outputs
	
	ListT []int
	
	StatsT struct {
		Number int
		DossName,
		Needed string
		List ListT
	}
	
	CertT = string
	
	CertsT []CertT
	
	DossCertT struct{
		First,
		Second string
		Certs CertsT
	}
	
	DossCertsT []DossCertT
	
	DispF struct {
		Title,
		Now string
		Stats *StatsT
		DossCerts DossCertsT
	}
	
	PropsT []string
	
	PermT struct {
		Proba string
		Props PropsT
	}
	
	PermsT []PermT
	
	DispP struct {
		Title,
		Now,
		Number string
		Perms PermsT
	}

)

var (
	
	wwFileDoc = GS.ExtractDocument(queryWWFile)
	wwPermsDoc = GS.ExtractDocument(queryWWPerms)

)

func printNow (now *NowT, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct, lang))
} //printNow

// Print f with fo, starting at the element of rank i0; if withNow, the output begins with the printing of the current date
func printFile (f *File, lang *SM.Lang) *DispF {
	
	PrintCertOrDoss := func (cd *Certif_DossierT, sigQty int) *DossCertT {
		
		PrintCertif := func (c *DatedCertificationT) string {
			cc := c.Certification
			return fmt.Sprint(cc.To.Uid, " ← ", cc.From.Uid, " ", BA.Ts2s(c.Date, lang), " (→ ", BA.Ts2s(cc.Expires_on, lang), ")")
		} //PrintCertif
		
		// Print d with fo
		PrintDossier := func (d *DossierT)  *DossCertT {
			
			PrintCerts := func (cs DatedCertifications) CertsT {
				
				PrintCert := func (c *DatedCertificationT) string {
					cc := c.Certification
					return fmt.Sprint(cc.From.Uid, " (", S.Itoa(int(cc.From.Quality)), "%) ", BA.Ts2s(c.Date, lang), " (→ ", BA.Ts2s(cc.Expires_on, lang), ")")
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
			if d.Main_certifs >= sigQty {
				fmt.Fprint(w, lang.Map("#duniterClient:requiredCertsNb", S.Itoa(len(d.Certifications)), S.Itoa(d.Main_certifs)))
			} else {
				fmt.Fprint(w, lang.Map("#duniterClient:certsNb", S.Itoa(len(d.Certifications))))
			}
			if d.Date == d.MinDate && d.Date >= f.Data.Now.Bct {
				fmt.Fprint(w, ". ", lang.Map("#duniterClient:minApplicationDate", BA.Ts2s(d.Newcomer.LastApplication.LastAppDate, lang)))
			}
			if d.Newcomer.Distance.Dist_ok && d.Main_certifs >= sigQty {
				fmt.Fprint(w, ". ", lang.Map("#duniterClient:OK"))
			} else {
				fmt.Fprint(w, ". ", lang.Map("#duniterClient:KO"))
				if !d.Newcomer.Distance.Dist_ok && d.Main_certifs >= sigQty {
					fmt.Fprint(w, " (", lang.Map("#duniterClient:Distance"), ")")
				}
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
	
	PrintDossiersNbs := func (cds Certifs_DossiersT) *StatsT {
		n := -1
		for _, cd := range cds {
			if cd.Dossier != nil {
				n = M.Max(n, cd.Dossier.Main_certifs)
			}
		}
		m := 0
		nbs := make(ListT, n + 1)
		if n >= 0 {
			for i := 0; i <= n; i++ {
				nbs[i] = 0
			}
			for _, cd := range cds {
				if cd.Dossier != nil {
					nbs[cd.Dossier.Main_certifs]++
					m++
				}
			}
		}
		return &StatsT{Number: m, List: nbs, DossName: lang.Map("#duniterClient:dossiers"), Needed: lang.Map("#duniterClient:neededCerts")}
	} //PrintDossiersNbs
	
	//printFile
	d := f.Data
	now := printNow(d.Now, lang)
	cds := d.WWFile.Certifs_dossiers
	dcs := make(DossCertsT, len(cds))
	for i, cd := range cds {
		dcs[i] = *PrintCertOrDoss(&cd, d.Parameter.SigQty)
	}
	return &DispF{Title: lang.Map("#duniterClient:ShowFile"), Now: now, Stats: PrintDossiersNbs(cds), DossCerts: dcs}
} //printFile

// Print permutations returned by CalcPermutations
func printPermutations (ps *Perms, lang *SM.Lang) *DispP {

	PrintPermutation := func (wp *WPermutation) *PermT {
	
		PrintPropagation := func (p *Propagation) string {
			w := new(strings.Builder)
			fmt.Fprint(w, BA.Ts2s(p.Date, lang))
			if p.After {
				fmt.Fprint(w, "+")
			} else {
				fmt.Fprint(w, BA.SpS)
			}
			fmt.Fprint(w, "\t", p.Id.Uid)
			return w.String()
		} //PrintPropagation
	
		//PrintPermutation
		p := fmt.Sprintf("%s = %10.6f%%", lang.Map("#duniterClient:Proba"), wp.Proba * 100)
		pp := make(PropsT, len(wp.Permutation))
		for i, pe := range wp.Permutation {
			pp[i] = PrintPropagation(&pe)
		}
		return &PermT{Proba: p, Props: pp}
	} //PrintPermutation

	//printPermutations
	d := ps.Data
	t := lang.Map("#duniterClient:Permutations")
	now := printNow(d.Now, lang)
	p := d.WWResult.Permutations
	n := fmt.Sprint(lang.Map("#duniterClient:PermutationsNb"), len(p))
	pp := make(PermsT, len(p))
	for i, pe := range p {
		pp[i] = *PrintPermutation(&pe)
	}
	return &DispP{Title: t, Now: now, Number: n, Perms: pp}
} //printPermutations

func endF (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	M.Assert(name == wwFileName, name, 100)
	j := GS.Send(nil, wwFileDoc)
	f := new(File)
	J.ApplyTo(j, f)
	temp.ExecuteTemplate(w, name, printFile(f, lang))
} //endF

func endP (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	M.Assert(name == wwPermsName, name, 100)
	j := GS.Send(nil, wwPermsDoc)
	p := new(Perms)
	J.ApplyTo(j, p)
	temp.ExecuteTemplate(w, name, printPermutations(p, lang))
} //endP

func init() {
	W.RegisterPackage(wwFileName, htmlWWFile, endF, true)
	W.RegisterPackage(wwPermsName, htmlWWPerms, endP, true)
} //init
