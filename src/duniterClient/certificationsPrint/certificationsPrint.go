/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package certificationsPrint

import (
	
	BA	"duniterClient/basicPrint"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"math"
		"strconv"
		"html/template"

)

const (
	
	certificationsFromName = "55certificationsFrom"
	certificationsToName = "56certificationsTo"
	
	queryCertifications = `
		query Certifications ($to: Boolean!) {
			now {
				number
				bct
			}
			identities(status: MEMBER) {
				uid
				sent_certifications @skip(if: $to) {
					other: to {
						uid
					}
					registration: block {
						bct
					}
					limit: expires_on
				}
				received_certifications @include(if: $to) {
					certifications {
						other: from {
							uid
						}
						registration: block {
							bct
						}
						limit: expires_on
					}
				}
			}
		}
	`
	
	htmlCertifications = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Block}}
			</h3>
			<p>
				{{.Number}}
			</p>
			{{$arrow := .Arrow}}
			{{range .First}}
				<p>
					{{.Name}} ({{.NbCerts}}) {{$arrow}}
					<br>
					<blockquote>
						{{range .Second}}
							{{.Name}}  {{.Registration}} → {{.Limit}}
							<br>
						{{end}}
					</blockquote>
				</p>
			{{end}}
			<p>
				{{.Mean}}
			<br>
				{{.Median}}
			<br>
				{{.StdDev}}
			<br>
				{{.DistName}}
			<br>
				{{range $nb, $val := .Distribution}}
					{{$nb}}  {{$val}}
					<br>
				{{end}}
			</p>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`

)

type (
	
	Dist []int
	
	NowT struct {
		Number int
		Bct int64
	}
	
	Cert struct {
		Other struct {
			Uid string
		}
		Registration struct {
			Bct int64
		}
		Limit int64
	}
	
	Certs []Cert
	
	Identity0 struct {
		Uid string
		Sent_certifications Certs
		Received_certifications struct {
			Certifications Certs
		}
	}
	
	IdentitiesT []Identity0
	
	DataT struct {
		Now *NowT
		Identities IdentitiesT
	}
	
	CertificationsT struct {
		Data *DataT
	}
	
	Disp3 struct {
		Name,
		Registration,
		Limit string
	}
	
	Disp2 []Disp3
	
	Disp1 struct {
		Name string
		NbCerts int
		Second Disp2
	}
	
	Disp0 []Disp1
	
	Disp struct {
		Title,
		Block,
		Number,
		Arrow string
		First Disp0
		Mean,
		Median,
		StdDev,
		DistName string
		Distribution Dist
	}

)
	
var (
	
	certificationsDoc = GS.ExtractDocument(queryCertifications)

)

func moments (certifications *CertificationsT, received bool) (d Dist, mean, sDev float64, nb, median int) {
	
	CountDist := func (ids IdentitiesT) (d Dist) {
		n := 0
		for _, id := range ids {
			var cs Certs
			if received {
				cs = id.Received_certifications.Certifications
			} else {
				cs = id.Sent_certifications
			}
			n = M.Max(n, len(cs))
		}
		d = make(Dist, n + 1)
		for _, id := range ids {
			var cs Certs
			if received {
				cs = id.Received_certifications.Certifications
			} else {
				cs = id.Sent_certifications
			}
			d[len(cs)]++
		}
		return
	} //CountDist

	//moments
	d = CountDist(certifications.Data.Identities)
	n := 0; nb = 0; nb2 := 0
	for i, dd := range d {
		n += dd
		nb += i * dd
		nb2 += i * i * dd
	}
	if n == 0 {
		mean = 0
		sDev = 0
		nb = 0
		median = 0
	} else {
		nf := float64(n)
		mean = float64(nb) / nf
		sDev = math.Sqrt(float64(nb2) / nf - mean * mean)
		median = - 1; q := 0
		for 2 * q < n {
			median++
			q += d[median]
		}
	}
	return
} //moments

func printMoments (certifications *CertificationsT, received bool) (d Dist, number, mean, stdDev, median, dName string) {
	d, m, sDev, nb, med := moments(certifications, received)
	number = strconv.Itoa(nb) + " " + SM.Map("#duniterClient:Certifications")
	mean = SM.Map("#duniterClient:Mean") + " = " + strconv.FormatFloat(m, 'f', -1, 64)
	median = SM.Map("#duniterClient:Median") + " = " + strconv.Itoa(med)
	stdDev = SM.Map("#duniterClient:SDev") + " = " + strconv.FormatFloat(sDev, 'f', -1, 64)
	dName = SM.Map("#duniterClient:Distribution")
	return
} //printMoments

func printNow (now *NowT) string {
	return fmt.Sprint(SM.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct))
} //printNow

func printCerts (cs Certs) Disp2 {

	PrintCert := func (c *Cert) *Disp3 {
		return &Disp3{c.Other.Uid, fmt.Sprint(BA.Ts2s(c.Registration.Bct)), fmt.Sprint(BA.Ts2s(c.Limit))}
	} //PrintCert

	//printCerts
	d := make(Disp2, len(cs))
	for i, c := range cs {
		d[i] = *PrintCert(&c)
	}
	return d
} //printCerts

func printIdentities (ids IdentitiesT, received bool) Disp0 {
	d := make(Disp0, len(ids))
	for i, id := range ids {
		d[i].Name = id.Uid
		if received {
			d[i].Second = printCerts(id.Received_certifications.Certifications)
		} else {
			d[i].Second = printCerts(id.Sent_certifications)
		}
		d[i].NbCerts = len(d[i].Second)
	}
	return d
} //printIdentities

func printFrom (certifications *CertificationsT) *Disp {
	d := certifications.Data
	dd := &Disp{Title: SM.Map("#duniterClient:certificationsFrom"), Block: printNow(d.Now), Arrow: "→", First: printIdentities(d.Identities, false)}
	dd.Distribution, dd.Number, dd.Mean, dd.StdDev, dd.Median, dd.DistName = printMoments(certifications, false)
	return dd
} //printFrom

func printTo (certifications *CertificationsT) *Disp {
	d := certifications.Data
	dd := &Disp{Title: SM.Map("#duniterClient:certificationsTo"), Block: printNow(d.Now), Arrow: "←", First: printIdentities(d.Identities, true)}
	dd.Distribution, dd.Number, dd.Mean, dd.StdDev, dd.Median, dd.DistName = printMoments(certifications, true)
	return dd
} //printTo

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	to := name == certificationsToName
	M.Assert(to || name == certificationsFromName, name, 100)
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushBoolean(to)
	mk.BuildField("to")
	mk.BuildObject()
	j := GS.Send(mk.GetJson(), certificationsDoc)
	certifications := new(CertificationsT)
	J.ApplyTo(j, certifications)
	var d *Disp
	if to {
		d = printTo(certifications)
	} else {
		d = printFrom(certifications)
	}
	temp.ExecuteTemplate(w, name, d)
} //end

func init() {
	W.RegisterPackage(certificationsFromName, htmlCertifications, end, true)
	W.RegisterPackage(certificationsToName, htmlCertifications, end, true)
} //init
