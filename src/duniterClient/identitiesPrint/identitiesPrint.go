/* 
duniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package identitiesPrint

import (
	
	B	"duniterClient/blockchainPrint"
	BA	"duniterClient/basicPrint"
	G	"util/graphQL"
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
	
	revokedName = "51revokedIdentities"
	missingName = "52missingIdentities"
	memberName = "53memberIdentities"
	newcomerName = "54newcomerIdentities"
	
	queryIdentities = `
		query Identities ($status: Identity_Status) {
			now {
				number
				bct
			}
			identities(status: $status) {
				uid
				pubkey
				hash
				block: id_written_block {
					bct
				}
				limitDate
			}
		}
	`
	
	htmlIdentities = `
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
			{{range .Ids}}
				<p>
				{{range .}}
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
	
	Block0 struct {
		Number int
		Bct int64
	}
	
	Block1 struct {
		Bct int64
	}
	
	Identity0 struct {
		Uid string
		Pubkey B.Pubkey
		Hash B.Hash
		Block *Block1
		LimitDate int64
	}
	
	Ids struct {
		Now *Block0
		Identities []Identity0
	}
	
	IdentitiesT struct {
		Data *Ids
	}
	
	Disp1 [3]string
	
	Disp0 []Disp1
	
	Disp struct {
		Title,
		Block,
		Number string
		Ids Disp0
	}

)

var (
	
	identitiesDoc *G.Document

)

func print (title, nbMsg string, identities *IdentitiesT) *Disp {
	
	hashId := SM.Map("#duniterClient:Hash")

	PrintIds := func (ids []Identity0) Disp0 {
		d := make(Disp0, len(ids))
		w := new(strings.Builder)
		for i, id := range ids {
			var dd Disp1
			dd[0] = fmt.Sprint(id.Uid, " ", id.Pubkey)
			dd[1] = fmt.Sprint(hashId, " ", id.Hash)
			w.Reset()
			if id.Block != nil {
				fmt.Fprint(w, BA.Ts2s(id.Block.Bct))
			}
			if id.LimitDate != 0 {
				if id.Block != nil {
					fmt.Fprint(w, " ")
				}
				fmt.Fprint(w, "→ ", BA.Ts2s(id.LimitDate))
			}
			dd[2] = w.String()
			d[i] = dd
		}
		return d
	} //PrintIds

	//print
	d := new(Disp)
	ids := identities.Data
	d.Title = title
	d.Block = fmt.Sprint(SM.Map("#duniterClient:Block"), " ", ids.Now.Number, "\t", BA.Ts2s(ids.Now.Bct))
	d.Number = fmt.Sprint(nbMsg, " = ", len(ids.Identities))
	d.Ids = PrintIds(ids.Identities)
	return d
} //print

func end (name string, temp *template.Template, _ *http.Request, w http.ResponseWriter) {
	var status, title, nbMsg string
	switch name {
	case revokedName:
		status = "REVOKED"
		title = SM.Map("#duniterClient:RevokedM")
		nbMsg = SM.Map("#duniterClient:RevokedNb")
	case missingName:
		status = "MISSING"
		title = SM.Map("#duniterClient:Missing")
		nbMsg = SM.Map("#duniterClient:MissingNb")
	case memberName:
		status = "MEMBER"
		title = SM.Map("#duniterClient:Members")
		nbMsg = SM.Map("#duniterClient:MembersNb")
	case newcomerName:
		status = "NEWCOMER"
		title = SM.Map("#duniterClient:Newcomers")
		nbMsg = SM.Map("#duniterClient:NewcomersNb")
	default:
		M.Halt(name, 100)
	}
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushString(status)
	mk.BuildField("status")
	mk.BuildObject()
	j := GS.Send(mk.GetJson(), identitiesDoc)
	identities := new(IdentitiesT)
	J.ApplyTo(j, identities)
	temp.ExecuteTemplate(w, name, print(title, nbMsg, identities))
} //end

func init() {
	identitiesDoc = GS.ExtractDocument(queryIdentities)
	W.RegisterPackage(revokedName, htmlIdentities, end, true)
	W.RegisterPackage(missingName, htmlIdentities, end, true)
	W.RegisterPackage(memberName, htmlIdentities, end, true)
	W.RegisterPackage(newcomerName, htmlIdentities, end, true)
} //init
