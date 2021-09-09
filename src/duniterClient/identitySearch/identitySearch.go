/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identitySearch

import (
	
	B	"duniterClient/blockchainPrint"
	BA	"duniterClient/basicPrint"
	GS	"duniterClient/gqlSender"
	J	"util/json"
	M	"util/misc"
	SM	"util/strMapping"
	W	"duniterClient/web"
		"fmt"
		"net/http"
		"util/sort"
		"strconv"
		"strings"
		"html/template"

)

const (
	
	explorerName = "1explorer"
	
	revokedIcon = '¶'
	missingIcon = '×'
	newcomerIcon = '°'
	
	queryNow = `
		query Now {
			now {
				number
				bct
			}
		}
	`
	
	queryFind = `
		query IdSearchFind ($hint: String, $statuses:  [Identity_Status!]) {
			now {
				number
				bct
			}
			idSearch (with: {hint: $hint, status_list: $statuses}) {
				revokedNb
				missingNb
				memberNb
				newcomerNb
				ids {
					uid
					status
					hash
				}
			}
		}
	`
	
	queryFix = `
		query IdSearchFix ($hash: Hash!, $dispDist: Boolean = false, $dispQual: Boolean = false, $dispCentr: Boolean = false) {
			now {
				number
				bct
			}
			idFromHash (hash: $hash) {
				hash
				uid
				pubkey
				id_written_block {
					number
					bct
				}
				limitDate
				status
				sentry
				membership_pending
				membership_pending_limitDate
				minDate
				minDatePassed
				history {
					in
					block {
						number
						bct
					}
				}
				received_certifications {
					certifications {
						from {
							uid
						}
						expires_on
						pending
					}
					limit
				}
				sent_certifications {
					to {
						uid
					}
					expires_on
					pending
				}
				all_certifiers {
					uid
				}
				all_certified {
					uid
				}
				all_certifiersIO {
					id {
						uid
					}
					hist {
						in
						block {
							number
							bct
						}
					}
				}
				all_certifiedIO {
					id {
						uid
					}
					hist {
						in
						block {
							number
							bct
						}
					}
				}
				distance @include(if: $dispDist) {
					value
					dist_ok
				}
				quality @include(if: $dispQual)
				centrality @include(if: $dispCentr)
			}
		}
	`
	
	html = `
		{{define "head"}}<title>{{.Start.Title}}</title>{{end}}
		{{define "body"}}
			{{with .Start}}
				<h1>{{.Title}}</h1>
				<p>
					<a href = "/">{{Map "index"}}</a>
				</p>
				<h3>
					{{.Now}}
				</h3>
			{{end}}
			<form action="" method="post">
				{{with .Start}}
					<p>
						<input type="text" name="hint" placeholder="{{.Placeholder}}" value="{{.Hint}}" size = 35/>
					</p>
					<input type="hidden" name="oldHint" value="{{.Hint}}"/>
					<p>
						<input type="checkbox" id="revoked" name="revoked" {{.RevokedChecked}}>
						<label for="revoked">{{.CheckRevoked}}</label>
						<input type="checkbox" id="missing" name="missing" {{.MissingChecked}}>
						<label for="missing">{{.CheckMissing}}</label>
						<input type="checkbox" id="member" name="member" {{.MemberChecked}}>
						<label for="member">{{.CheckMember}}</label>
						<input type="checkbox" id="newcomer" name="newcomer" {{.NewcomerChecked}}>
						<label for="newcomer">{{.CheckNewcomer}}</label>
					</p>
				{{end}}
				{{if .Find}}
					{{with .Find}}
						<p>
							{{.IdNumbers}}
						</p>
						{{if .Ids}}
							<p>
								<select name="idHash" id="idHash" >
									{{$s := .SelectedHash}}
									{{range $h := .Ids}}
										<option value="{{$h.Hash}}"{{if eq $h.Hash $s}} selected{{end}}>{{$h.Uid}}</option>
									{{end}}
								</select>
								<label for="idHash">{{.Select}}</label>
							</p>
							<input type="hidden" name="oldIdHash" value="{{.SelectedHash}}"/>
							<p>
								<input type="checkbox" id="dist" name="dist" {{.DistChecked}}>
								<label for="dist">{{.CheckDist}}</label>
								<input type="checkbox" id="qual" name="qual" {{.QualChecked}}>
								<label for="qual">{{.CheckQual}}</label>
								<input type="checkbox" id="centr" name="centr" {{.CentrChecked}}>
								<label for="centr">{{.CheckCentr}}</label>
							</p>
						{{end}}
					{{end}}
				{{end}}
				<p>
					<input type="submit" value="{{.OK}}">
				</p>
			</form>
			{{if .Fix}}
				{{with .Fix}}
					{{with .Id}}
						<h4>
							{{.Uid}}
						</h4>
						<p>
							{{.Pubkey}}
						</p>
						<p>
							{{.Hash}}
						</p>
						<p>
							{{.Member}}
						</p>
						{{if .Sentry}}
							<p>
								{{.Sentry}}
							</p>
						{{end}}
					{{end}}
					{{if .Dist}}
						<p>
							{{.Dist}}
						</p>
					{{end}}
					{{if .Qual}}
						<p>
							{{.Qual}}
						</p>
					{{end}}
					{{if .Centr}}
						<p>
							{{.Centr}}
						</p>
					{{end}}
					{{with .Id}}
						{{if .Block}}
							<p>
								{{.Block}}
							</p>
						{{end}}
						<p>
							{{.LimitDate}}
						</p>
						{{if .Availability}}
							<p>
								{{.Availability}}
							</p>
						{{end}}
						{{if .Pending}}
							<p>
								{{.Pending}}
							</p>
						{{end}}
						{{if .History}}
							{{with .History}}
								<h5>
									{{.Label}}
								</h5>
								<blockquote>
									<p>
										{{.Legend}}
									</p>
									{{range .List}}
										{{.}}
										<br>
									{{end}}
								</blockquote>
							{{end}}
						{{end}}
					{{end}}
					{{with .Cs}}
						<h5>
							{{.PresentCertifiers}}
						</h5>
						<blockquote>
							<p>
								{{range .ReceivedCerts}}
									{{.}}
									<br>
								{{end}}
							</p>
							<h5>
								{{.SortedByDateL}}
							</h5>
							<blockquote>
								<p>
									{{range .ReceivedByLimits}}
										{{.}}
										<br>
									{{end}}
								</p>
							</blockquote>
						</blockquote>
						{{if .AllCertifiers}}
							<h5>
								{{.AllCertifiers}}
							</h5>
							<p>
								<blockquote>
									{{range .ReceivedAllCerts}}
										{{.}}
										<br>
									{{end}}
								</blockquote>
							</p>
							<h5>
								{{.AllCertifiersIO}}
							</h5>
							<p>
								<blockquote>
									{{range .ReceivedAllCertsIO}}
										{{.Uid}}
										<blockquote>
											{{range .Hist}}
												{{.}}
												<br>
											{{end}}
										</blockquote>
										<br>
									{{end}}
								</blockquote>
							</p>
						{{end}}
						{{if .PresentCertified}}
							<h5>
								{{.PresentCertified}}
							</h5>
							<blockquote>
								<p>
									{{range .SentCerts}}
										{{.}}
										<br>
									{{end}}
								</p>
								<h5>
									{{.SortedByDate}}
								</h5>
								<blockquote>
									<p>
										{{range .SentByLimits}}
											{{.}}
											<br>
										{{end}}
									</p>
								</blockquote>
							</blockquote>
							<h5>
								{{.AllCertified}}
							</h5>
							<p>
								<blockquote>
									{{range .SentAllCerts}}
										{{.}}
										<br>
									{{end}}
								</blockquote>
							</p>
							<h5>
								{{.AllCertifiedIO}}
							</h5>
							<p>
								<blockquote>
									{{range .SentAllCertsIO}}
										{{.Uid}}
										<blockquote>
											{{range .Hist}}
												{{.}}
												<br>
											{{end}}
										</blockquote>
										<br>
									{{end}}
								</blockquote>
							</p>
						{{end}}
					{{end}}
				{{end}}
			{{end}}
			<p>
				<a href = "/">{{Map "index"}}</a>
			</p>
		{{end}}
	`

)

type (
	
	Block struct {
		Number int
		Bct int64
	}
	
	NowRes struct {
		Data struct {
			Now *Block
		}
	}
	
	IdSearchOutput  struct {
		RevokedNb,
		MissingNb,
		MemberNb,
		NewcomerNb int
		Ids Identities
	}
	
	FindRes struct {
		Data struct {
			Now *Block
			IdSearch *IdSearchOutput
		}
	}
	
	HistoryEvent0 struct {
		In bool
		Block *Block
	}
	
	History []HistoryEvent0
	
	Certification0 struct {
		From,
		To *Identity
		Expires_on int64
		Pending bool
	}
	
	Certifications []*Certification0
	
	Received_certifications struct {
		Certifications Certifications
		Limit int64
	}
	
	Identities []*Identity
	
	CertEvent struct {
		In bool
		Block Block
	}
	
	CertEvents []CertEvent
	
	CertHist struct {
		Id *Identity
		Hist CertEvents
	}
	
	CertHists []CertHist
	
	Distance struct {
		Value float64
		Dist_ok bool
	}
	
	Identity struct {
		Hash B.Hash
		Uid string
		Pubkey B.Pubkey
		Id_written_block *Block
		LimitDate int64
		Status string
		Sentry bool
		Membership_pending bool
		Membership_pending_limitDate int64
		MinDate int64
		MinDatePassed bool
		History History
		Received_certifications *Received_certifications
		Sent_certifications Certifications
		All_certifiers,
		All_certified Identities
		All_certifiersIO,
		All_certifiedIO CertHists
		Distance *Distance
		Quality,
		Centrality float64
	}
	
	FixRes struct {
		Data struct {
			Now *Block
			IdFromHash *Identity
		}
	}
	
	expSort struct {
		ids ListS
		exps ListI
	}
	
	// Outputs
	
	Start struct {
		Title,
		Now,
		Placeholder,
		Hint,
		RevokedChecked,
		CheckRevoked,
		MissingChecked,
		CheckMissing,
		MemberChecked,
		CheckMember,
		NewcomerChecked,
		CheckNewcomer string
	}
	
	IdHash struct {
		Hash,
		Uid string
	}
	
	IdHashes []*IdHash
	
	Find struct {
		IdNumbers string
		Ids IdHashes
		Select,
		SelectedHash,
		DistChecked, 
		CheckDist,
		QualChecked,
		CheckQual,
		CentrChecked,
		CheckCentr string
	}
	
	ListS []string
	
	ListI []int64
	
	CH struct {
		Uid string
		Hist ListS
	}
	
	ListCH []CH
	
	Certifics struct {
		PresentCertifiers,
		PresentCertified string
		ReceivedCerts,
		SentCerts ListS
		SortedByDateL,
		SortedByDate string
		ReceivedByLimits,
		SentByLimits ListS
		AllCertifiers,
		AllCertified string
		ReceivedAllCerts,
		SentAllCerts ListS
		AllCertifiersIO,
		AllCertifiedIO string
		ReceivedAllCertsIO,
		SentAllCertsIO ListCH
	}
	
	Hist struct {
		Label,
		Legend string
		List ListS
	}
	
	Idty struct {
		Uid,
		Pubkey,
		Hash,
		Member,
		Sentry,
		Block, 
		LimitDate,
		Availability,
		Pending string
		History *Hist
	}
	
	Fix struct {
		Id *Idty
		Dist,
		Qual,
		Centr string
		Cs *Certifics
	}
	
	Out struct {
		Start *Start
		Find *Find
		Fix *Fix
		OK string
	}

)

var (
	
	nowDoc = GS.ExtractDocument(queryNow)
	findDoc = GS.ExtractDocument(queryFind)
	fixDoc = GS.ExtractDocument(queryFix)

)

func doBlock (b *Block, lang *SM.Lang) string {
	return fmt.Sprint(lang.Map("#duniterClient:Block"), " ", b.Number, " ", BA.Ts2s(b.Bct, lang))
} //doBlock

func fixChecked (checked bool) string {
	if checked {
		return "checked"
	}
	return ""
} //fixChecked

func doStart (title, hint string, revokedC, missingC, memberC, newcomerC bool, now *Block, lang *SM.Lang) *Start {
	t := lang.Map(title)
	nowS := doBlock(now, lang)
	placeholder := lang.Map("#duniterClient:TypeUidOrPubkey")
	revokedChecked := fixChecked(revokedC)
	checkRevoked := lang.Map("#duniterClient:Revokeds") + " (" + string(revokedIcon) + ")"
	missingChecked := fixChecked(missingC)
	checkMissing := lang.Map("#duniterClient:Missings") + " (" + string(missingIcon) + ")"
	memberChecked := fixChecked(memberC)
	checkMember := lang.Map("#duniterClient:Members")
	newcomerChecked := fixChecked(newcomerC)
	checkNewcomer := lang.Map("#duniterClient:Newcomers") + " (" + string(newcomerIcon) + ")"
	return &Start{t, nowS, placeholder, hint, revokedChecked, checkRevoked, missingChecked, checkMissing, memberChecked, checkMember, newcomerChecked, checkNewcomer}
} //doStart

func doFind (idso *IdSearchOutput, selectedHash string, distC, qualC, centrC bool, lang *SM.Lang) *Find {
	
	SearchId := func (id *Identity) *IdHash {
		var uid string
		switch id.Status {
		case "NEWCOMER":
			uid = string(newcomerIcon) + " " + id.Uid
		case "MEMBER":
			uid = id.Uid
		case "MISSING":
			uid = string(missingIcon) + " " + id.Uid
		case "REVOKED":
			uid = string(revokedIcon) + " " + id.Uid
		default:
			M.Halt(id.Status, 100)
		}
		return &IdHash{string(id.Hash), uid}
	} //SearchId
	
	//doFind
	if idso == nil {
		return nil
	}
	idNumbers := fmt.Sprint(idso.RevokedNb, " ", lang.Map("#duniterClient:Revokeds"), BA.SpL, idso.MissingNb, " ", lang.Map("#duniterClient:Missings"), BA.SpL, idso.MemberNb, " ", lang.Map("#duniterClient:Members"), BA.SpL, idso.NewcomerNb, " ", lang.Map("#duniterClient:Newcomers"))
	selectIt := lang.Map("#duniterClient:Select")
	distChecked := fixChecked(distC)
	checkDist := lang.Map("#duniterClient:Distance")
	qualChecked := fixChecked(qualC)
	checkQual := lang.Map("#duniterClient:Quality")
	centrChecked := fixChecked(centrC)
	checkCentr := lang.Map("#duniterClient:Centrality")
	ids := idso.Ids
	idHs := make(IdHashes, len(ids))
	for i, id := range ids {
		idHs[i] = SearchId(id)
	}
	return &Find{idNumbers, idHs, selectIt, selectedHash, distChecked, checkDist, qualChecked, checkQual, centrChecked, checkCentr}
} //doFind

func (e *expSort) Less (i, j int) bool {
	return e.exps[i] < e.exps[j] || e.exps[i] == e.exps[j] && BA.CompP(e.ids[i], e.ids[j]) == BA.Lt
} //Less

func (e *expSort) Swap (i, j int) {
	e.ids[i], e.ids[j] = e.ids[j], e.ids[i]
	e.exps[i], e.exps[j] = e.exps[j], e.exps[i]
} //Swap

func certs (res *Identity, lang *SM.Lang) *Certifics {

	countCerts := func (cs Certifications) (nb, futNb int) {
		nb = 0
		futNb = 0
		for _, c := range cs {
			if c.Pending {
				futNb++
			} else {
				nb++
			}
		}
		return
	} //countCerts

	//certs
	var ts sort.TS
	sortedByDate := lang.Map("#duniterClient:SortedByCExpDates")
	sortedByDateL := sortedByDate
	if res.Status == "MISSING" && res.Received_certifications.Limit != 0 || res.Status == "MEMBER" {
		sortedByDateL = lang.Map("#duniterClient:SortedByCExpDatesL")
	}
	es := new(expSort)
	ts.Sorter = es
	
	presentCertified := ""
	allCertified := ""
	allCertifiedIO := ""
	var (sentCerts, sentByLimits, sentAllCerts ListS; sentAllCertsIO ListCH)
	if res.Status != "NEWCOMER" {
		certifs := res.Sent_certifications
		sentCertsNb, sentCertsFutNb := countCerts(certifs)
		presentCertified = fmt.Sprint(lang.Map("#duniterClient:PresentCertified"), " (", sentCertsNb, " + ", sentCertsFutNb, ")")
		sentCerts = make(ListS, sentCertsNb + sentCertsFutNb)
		es.exps = make(ListI, len(sentCerts))
		for i, c := range certifs {
			if c.Pending {
				sentCerts[i] = string(newcomerIcon) + " " + c.To.Uid
			} else {
				sentCerts[i] = c.To.Uid
			}
			es.exps[i] = c.Expires_on
		}
		es.ids = make(ListS, len(es.exps))
		copy(es.ids, sentCerts)
		ts.QuickSort(0, len(es.ids) - 1)
		sentByLimits = make(ListS, len(es.ids))
		for i := 0; i < len(es.ids); i++ {
			sentByLimits[i] = fmt.Sprint(BA.Ts2s(es.exps[i], lang), BA.SpL, es.ids[i])
		}
		allCertified = lang.Map("#duniterClient:AllCertified")
		sentAllCerts = make(ListS, len(res.All_certified))
		for i, a := range res.All_certified {
			sentAllCerts[i] = a.Uid
		}
		allCertifiedIO = lang.Map("#duniterClient:AllCertifiedIO")
		sentAllCertsIO = make(ListCH, len(res.All_certifiedIO))
		for i, a := range res.All_certifiedIO {
			sentAllCertsIO[i].Uid = a.Id.Uid
			h := make(ListS, len(a.Hist))
			for j, ce := range a.Hist {
				w := new(strings.Builder)
				if ce.In {
					fmt.Fprint(w, "↑")
				} else {
					fmt.Fprint(w, "↓")
				}
				b := ce.Block
				fmt.Fprint(w, BA.SpL, b.Number, " ", BA.Ts2s(b.Bct, lang))
				h[j] = w.String()
			}
			sentAllCertsIO[i].Hist = h
		}
	}
	
	certifs := res.Received_certifications.Certifications
	receivedCertsNb, receivedCertsFutNb := countCerts(certifs)
	presentCertifiers := fmt.Sprint(lang.Map("#duniterClient:PresentCertifiers"), " (", receivedCertsNb, " + ", receivedCertsFutNb, ")")
	receivedCerts := make(ListS, receivedCertsNb + receivedCertsFutNb)
	es.exps = make(ListI, len(receivedCerts))
	for i, c := range certifs {
		if c.Pending {
			receivedCerts[i] = string(newcomerIcon) + " " + c.From.Uid
		} else {
			receivedCerts[i] = c.From.Uid
		}
		es.exps[i] = c.Expires_on
	}
	es.ids = make(ListS, len(es.exps))
	copy(es.ids, receivedCerts)
	ts.QuickSort(0, len(es.ids) - 1)
	receivedByLimits := make(ListS, len(es.ids))
	for i := 0; i < len(es.ids); i++ {
		receivedByLimits[i] = fmt.Sprint(BA.Ts2s(es.exps[i], lang), BA.SpL, es.ids[i])
		if es.exps[i] == res.Received_certifications.Limit {
			receivedByLimits[i] = "→ " + receivedByLimits[i]
		}
	}
	allCertifiers := ""
	allCertifiersIO := ""
	var (receivedAllCerts ListS; receivedAllCertsIO ListCH)
	if res.Status != "NEWCOMER" {
		allCertifiers = lang.Map("#duniterClient:AllCertifiers")
		receivedAllCerts = make(ListS, len(res.All_certifiers))
		for i, a := range res.All_certifiers {
			receivedAllCerts[i] = a.Uid
		}
		allCertifiersIO = lang.Map("#duniterClient:AllCertifiersIO")
		receivedAllCertsIO = make(ListCH, len(res.All_certifiersIO))
		for i, a := range res.All_certifiersIO {
			receivedAllCertsIO[i].Uid = a.Id.Uid
			h := make(ListS, len(a.Hist))
			for j, ce := range a.Hist {
				w := new(strings.Builder)
				if ce.In {
					fmt.Fprint(w, "↑")
				} else {
					fmt.Fprint(w, "↓")
				}
				b := ce.Block
				fmt.Fprint(w, BA.SpL, b.Number, " ", BA.Ts2s(b.Bct, lang))
				h[j] = w.String()
			}
			receivedAllCertsIO[i].Hist = h
		}
	}
	
	return &Certifics{presentCertifiers, presentCertified, receivedCerts, sentCerts, sortedByDateL, sortedByDate, receivedByLimits, sentByLimits, allCertifiers, allCertified, receivedAllCerts, sentAllCerts, allCertifiersIO, allCertifiedIO, receivedAllCertsIO, sentAllCertsIO}
} //certs

func printHistory (h History, lang *SM.Lang) *Hist {
	if len(h) == 0 {
		return nil
	}
	la := lang.Map("#duniterClient:history")
	lg := fmt.Sprint("↑  ", lang.Map("#duniterClient:In"), BA.SpL, "↓  ", lang.Map("#duniterClient:Out"))
	ls := make(ListS, len(h))
	for i, hi := range h {
		w := new(strings.Builder)
		if hi.In {
			fmt.Fprint(w, "↑")
		} else {
			fmt.Fprint(w, "↓")
		}
		b := hi.Block
		fmt.Fprint(w, BA.SpL, b.Number, " ", BA.Ts2s(b.Bct, lang))
		ls[i] = w.String()
	}
	return &Hist{la, lg, ls}
} //printHistory

func get (res *Identity, lang *SM.Lang) *Idty {
	yes := lang.Map("#duniterClient:yes")
	no := lang.Map("#duniterClient:no")
	uid := fmt.Sprint(lang.Map("#duniterClient:Nickname"), BA.SpL, res.Uid)
	hash := fmt.Sprint(lang.Map("#duniterClient:Hash"), BA.SpL, string(res.Hash))
	pubkey := fmt.Sprint(lang.Map("#duniterClient:Pubkey"), BA.SpL, string(res.Pubkey))
	member := fmt.Sprint(lang.Map("#duniterClient:Member"), BA.SpL)
	sentry := ""
	availability := ""
	if res.Status == "MEMBER" {
		member += yes
		sentry = fmt.Sprint(lang.Map("#duniterClient:Sentry"), BA.SpL)
		if res.Sentry {
			sentry += yes
		} else {
			sentry += no
		}
		if res.MinDate != 0 {
			availability = BA.Ts2s(res.MinDate, lang) + BA.SpL
		}
		if res.MinDatePassed {
			availability += lang.Map("#duniterClient:OK")
		} else {
			availability += lang.Map("#duniterClient:KO")
		}
		if availability != "" {
			availability = fmt.Sprint(lang.Map("#duniterClient:Availability"), BA.SpL, availability)
		}
	} else {
		member += no
	}
	var block string
	if res.Id_written_block == nil {
		block = ""
	} else {
		block = fmt.Sprint(lang.Map("#duniterClient:Written_block"), BA.SpL, doBlock(res.Id_written_block, lang))
	}
	var limitDate string
	switch res.Status {
	case "REVOKED":
		limitDate = BA.Ts2s(BA.Revoked, lang)
	case "MISSING":
		limitDate = fmt.Sprint(lang.Map("#duniterClient:AppRLimitDate"), BA.SpL, BA.Ts2s(res.LimitDate, lang))
	case "MEMBER":
		limitDate = fmt.Sprint(lang.Map("#duniterClient:AppMLimitDate"), BA.SpL, BA.Ts2s(res.LimitDate, lang))
	case "NEWCOMER":
		limitDate = fmt.Sprint(lang.Map("#duniterClient:AppNLimitDate"), BA.SpL, BA.Ts2s(res.LimitDate, lang))
	default:
		M.Halt(res.Status, 101)
	}
	var pending string
	isPending := res.Membership_pending && (res.Status == "MISSING" || res.Status == "MEMBER")
	if isPending {
		pending = fmt.Sprint(lang.Map("#duniterClient:pending"), BA.SpL, lang.Map("#duniterClient:LimitDate"), BA.SpL, BA.Ts2s(res.Membership_pending_limitDate, lang))
	} else {
		pending = ""
	}
	history := printHistory(res.History, lang)
	return &Idty{uid, pubkey, hash, member, sentry, block, limitDate, availability, pending, history}
} //get

func notTooFar (res *Identity, lang *SM.Lang) string {
	if res.Distance == nil {
		return ""
	}
	d := res.Distance
	b := new(strings.Builder)
	fmt.Fprint(b, lang.Map("#duniterClient:Distance"), BA.SpL, strconv.FormatFloat(d.Value, 'f', 2, 64), "%", BA.SpL)
	if d.Dist_ok {
		fmt.Fprint(b, lang.Map("#duniterClient:OK"))
	} else {
		fmt.Fprint(b, lang.Map("#duniterClient:KO"))
	}
	return b.String()
} //notTooFar

func calcQuality (res *Identity, lang *SM.Lang) string {
	if res.Quality == 0 {
		return ""
	}
	return fmt.Sprint(lang.Map("#duniterClient:Quality"), BA.SpL, strconv.FormatFloat(res.Quality, 'f', 2, 64), "%")
} //calcQuality

func calcCentrality (res *Identity, lang *SM.Lang) string {
	if res.Centrality == 0 {
		return ""
	}
	return fmt.Sprint(lang.Map("#duniterClient:Centrality"), BA.SpL, strconv.FormatFloat(res.Centrality, 'f', 2, 64), "%")
} //calcCentrality

func doFix (res *Identity, lang *SM.Lang) *Fix {
	if res == nil {
		return nil
	}
	id := get(res, lang)
	cs := certs(res, lang)
	dist := notTooFar(res, lang)
	qual := calcQuality(res, lang)
	centr := calcCentrality(res, lang)
	return &Fix{id, dist, qual, centr, cs}
} //doFix

func printStart (t, hint string, reC, miC, meC, neC bool, now *NowRes, lang *SM.Lang) *Out {
	start := doStart(t, hint, reC, miC, meC, neC, now.Data.Now, lang)
	return &Out{Start: start, OK: lang.Map("#duniterClient: OK")}
} //printStart

func printFind (t, hint, selHash string, reC, miC, meC, neC, dC, qC, cC bool, find *FindRes, lang *SM.Lang) *Out {
	start := doStart(t, hint, reC, miC, meC, neC, find.Data.Now, lang)
	findS := doFind(find.Data.IdSearch, selHash, dC, qC, cC, lang)
	return &Out{Start: start, Find: findS, OK: lang.Map("#duniterClient:OK")}
} //printFind

func printFix (t, hint, selHash string, reC, miC, meC, neC, dC, qC, cC bool, find *FindRes, fix *FixRes, lang *SM.Lang) *Out {
	start := doStart(t, hint, reC, miC, meC, neC, fix.Data.Now, lang)
	findS := doFind(find.Data.IdSearch, selHash, dC, qC, cC, lang)
	// if 'find.Data.IdSearch' is nil or void, don't display it
	fixS := (*Fix)(nil)
	if find.Data.IdSearch != nil && len(find.Data.IdSearch.Ids) > 0 {
		fixS = doFix(fix.Data.IdFromHash, lang)
	}
	return &Out{Start: start, Find: findS, Fix: fixS, OK:lang.Map("#duniterClient:OK")}
} //printFix

func end (name string, temp *template.Template, r *http.Request, w http.ResponseWriter, lang *SM.Lang) {
	
	const (
		
		defaultHint = ""
		
		defaultRevokedC = true
		defaultMissingC = true
		defaultMemberC = true
		defaultNewcomerC = true
		
		defaultDist = true
		defaultQual = false
		defaultCentr = false
	
	)
	
	M.Assert(name == explorerName, name, 100)
	t := "#duniterClient:Explorer"
	dC := defaultDist
	qC := defaultQual
	cC := defaultCentr
	if r.Method == "GET" {
		j := GS.Send(nil, nowDoc)
		n := new(NowRes)
		J.ApplyTo(j, n)
		out := printStart(t, defaultHint, defaultRevokedC, defaultMissingC, defaultMemberC, defaultNewcomerC, n, lang)
		err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 101)
		return
	}
	r.ParseForm()
	hint := r.PostFormValue("hint")
	oldHint := r.PostFormValue("oldHint")
	selHash := r.PostFormValue("idHash")
	isFix := hint == oldHint && selHash != ""
	reC := r.PostFormValue("revoked") != ""
	miC := r.PostFormValue("missing") != ""
	meC := r.PostFormValue("member") != ""
	neC := r.PostFormValue("newcomer") != ""
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushString(hint)
	mk.BuildField("hint")
	mk.StartArray()
	if reC {
		mk.PushString("REVOKED")
	}
	if miC {
		mk.PushString("MISSING")
	}
	if meC {
		mk.PushString("MEMBER")
	}
	if neC {
		mk.PushString("NEWCOMER")
	}
	mk.BuildArray()
	mk.BuildField("statuses")
	mk.BuildObject()
	j := mk.GetJson()
	j = GS.Send(j, findDoc)
	fd := new(FindRes)
	J.ApplyTo(j, fd)
	if !isFix {
		if fd.Data.IdSearch == nil || len(fd.Data.IdSearch.Ids) != 1 {
			out := printFind(t, hint, selHash, reC, miC, meC, neC, dC, qC, cC, fd, lang)
			err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 102)
			return
		}
		selHash = string(fd.Data.IdSearch.Ids[0].Hash)
	} else {
		dC = r.PostFormValue("dist") != ""
		qC = r.PostFormValue("qual") != ""
		cC = r.PostFormValue("centr") != ""
	}
	mk = J.NewMaker()
	mk.StartObject()
	mk.PushString(selHash)
	mk.BuildField("hash")
	mk.PushBoolean(dC)
	mk.BuildField("dispDist")
	mk.PushBoolean(qC)
	mk.BuildField("dispQual")
	mk.PushBoolean(cC)
	mk.BuildField("dispCentr")
	mk.BuildObject()
	j = mk.GetJson()
	j = GS.Send(j, fixDoc)
	fx := new(FixRes)
	J.ApplyTo(j, fx)
	out := printFix(t, hint, selHash, reC, miC, meC, neC, dC, qC, cC, fd, fx, lang)
	err := temp.ExecuteTemplate(w, name, out); M.Assert(err == nil, err, 103)
} //end

func init() {
	W.RegisterPackage(explorerName, html, end, true)
} //init
