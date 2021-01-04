/* 
duniterClient: WotWizard.

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package stats

import (
	
	BA	"duniterClient/basicPrint"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	S	"util/sets"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"html/template"

)

const (
	
	statsName = "60sentCertsHistory"
	
	queryStats = `
		query Stats {
			countMin {
				utc0
			}
			countMax {
				number
				bct
				utc0
			}
			idSearch(with:{status_list: [REVOKED, MISSING, MEMBER]}) {
				ids {
					history {
						block {
							utc0
						}
					}
					all_certifiedIO {
						hist {
							block {
								utc0
							}
						}
					}
				}
			}
		}
	`
	
	htmlStats = `
		{{define "head"}}<title>{{.Title}}</title>{{end}}
		{{define "body"}}
			<h1>{{.Title}}</h1>
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
			<h3>
				{{.Now}}
			</h3>
			{{$b := .Brackets}}
			{{$m := .Members}}
			{{$n := .NotMembers}}
			<p>
				{{.Day}}
				<br>
				{{range $i, $d := .Days}}
						{{$i}}
						,{{$b}}
						{{range $d}}
							,{{.Interval}}
						{{end}}
						<br>
						,{{$m}}
						{{range $d}}
							,{{index .S 0}}
						{{end}}
						<br>
						,{{$n}}
						{{range $d}}
							,{{index .S 1}}
						{{end}}
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
	
	Block struct {
		Number int
		Bct,
		Utc0 int64
	}
	
	HistoryEvent struct {
		Block Block
	}
	
	History []HistoryEvent
	
	CertHist struct {
		Hist History
	}
	
	CertHists []CertHist
	
	Identity struct {
		History History
		All_certifiedIO CertHists
	}
	
	Identities []*Identity
	
	Data struct {
		CountMin,
		CountMax *Block
		IdSearch struct {
			Ids Identities
		}
	}
	
	Stats struct {
		Data *Data
	}
	
	// Outputs
	
	Slice struct {
		Interval string
		S [2]int
	}
	
	Day []Slice
	
	Days []*Day
	
	Disp struct {
		Now,
		Title,
		Day,
		Brackets,
		Members,
		NotMembers string
		Days Days
	}

)

var (
	
	statsDoc = GS.ExtractDocument(queryStats)

)

// Je souhaiterais connaître , à un jour j, le nombre de membres par tranche de certifications émises (0, 1 à 10, 11 à 20, 21 à 30, …spécifique : ceux n’ayant jamais émis une certification) et pouvoir séparer en deux sous-groupes : membres actifs et membres exclus.
// 'calc' returns day by day the number of identities per brackets of sent certifications (0, 1-10, 11-20, etc...) for members and not members.
func calc (stats *Stats) Days {
	const (dayL = 24 * 60 * 60; slicesLength = 10)
	da := stats.Data
	t0 := da.CountMin.Utc0
	
	dayOf := func (t int64) int {
		return int((t - t0) / dayL)
	}
	
	n := dayOf(da.CountMax.Utc0)
	
	type (
		
		daysT []int
		
		MNM struct {
			org int
			days daysT
		}
		
		idcT [2]*MNM
		
		idcsT []*idcT
	
	)
	
	scan := func (h History, org int, i, min, max *int) bool {
		if *i >= len(h) {
			return false
		}
		*min = dayOf(h[*i].Block.Utc0) - org
		*i++
		*max = n
		if *i < len(h) {
			*max = dayOf(h[*i].Block.Utc0)
			*i++
		}
		*max -= org
		return true
	}
	
	ids := da.IdSearch.Ids
	idcs := make(idcsT, len(ids))
	for i, id := range ids {
		idc := new(idcT)
		h := id.History
		org := dayOf(h[0].Block.Utc0)
		isMember := S.NewSet()
		var min, max int
		j := 0
		for scan(h, org, &j, &min, &max) {
			isMember.Fill(min, max)
		}
		full := S.Interval(0, n - org)
		isNotMember := full.Diff(isMember)
		idc[0] = &MNM{org: org, days: make(daysT, max + 1)}
		if isNotMember.IsEmpty() {
			idc[1] = &MNM{days: make(daysT, 0)}
		} else {
			it := isNotMember.Attach()
			min, m, ok := it.First()
			for ok {
				max = m
				_, m, ok = it.Next()
			}
			idc[1] = &MNM{org: org + min, days: make(daysT, max - min + 1)}
		}
		for _, ch := range id.All_certifiedIO {
			h := ch.Hist
			set := S.NewSet()
			j := 0
			for scan(h, org, &j, &min, &max) {
				set.Fill(min, max)
			}
			setMember := set.Inter(isMember)
			iter := setMember.Attach()
			e, ok := iter.FirstE()
			for ok {
				idc[0].days[e]++
				e, ok = iter.NextE()
			}
			setNotMember := set.Inter(isNotMember)
			iter = setNotMember.Attach()
			e, ok = iter.FirstE()
			for ok {
				idc[1].days[e - (idc[1].org - idc[0].org)]++
				e, ok = iter.NextE()
			}
		}
		idcs[i] = idc
	}
	
	slice := func (n int) int {
		return (n + slicesLength - 1) / slicesLength
	}
	
	days := make(Days, n + 1)
	for i := range days {
		d := make(Day, 1)
		days[i] = &d
	}
	for _, idc := range idcs {
		for j, m := range idc {
			for i, d := range m.days {
				ds := days[i + m.org]
				s := slice(d)
				for len(*ds) <= s {
					*ds = append(*ds, Slice{})
				}
				(*ds)[s].S[j]++
			}
		}
	}
	for _, ds := range days {
		m := 0
		for j := range *ds {
			n := j * slicesLength
			(*ds)[j].Interval = fmt.Sprint(m, " - ", n)
			m = n + 1
		}
	}
	return days
}

func printNow (now *Block, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", now.Number, "\t", BA.Ts2s(now.Bct, lang))
} //printNow

func print (stats *Stats, title string, lang *SM.Lang) *Disp {
	now := printNow(stats.Data.CountMax, lang)
	t := lang.Map(title)
	day := lang.Map("#duniterClient:Day")
	brackets := lang.Map("#duniterClient:Brackets")
	members := lang.Map("#duniterClient:Members")
	notMembers := lang.Map("#duniterClient:NotMembers")
	return &Disp{now, t, day, brackets, members, notMembers, calc(stats)}
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	M.Assert(name == statsName, 100)
	j := GS.Send(nil, statsDoc)
	stats := new(Stats)
	J.ApplyTo(j, stats)
	temp.ExecuteTemplate(w, name, print(stats, "#duniterClient:SentCertsHistory", lang))
} //end

func init() {
	W.RegisterPackage(statsName, htmlStats, end, true)
} //init
