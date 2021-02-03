/* 
WotWizard

Copyright (C) 2017-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package run

import (
	
	_	"babel/static"
	_	"util/json/static"
	_	"util/graphQL/static"
	_	"duniter/static"
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"util/graphQL"
	GQ	"duniter/gqlReceiver"
	S	"duniter/sandbox"
	
	_	"duniter/blocks"
	_	"duniter/certifications"
	_	"duniter/events"
	_	"duniter/history"
	_	"duniter/identities"
	_	"duniter/members"
	_	"duniter/parameters"
	_	"duniter/sentries"
	_	"duniter/wotWizardList"
	
		"fmt"

)

const (
	
	version = "5.1.4"

)

func Start () {
	fmt.Println("WotWizard version", version, "\n")
	BA.Lg.Println("WotWizard version", version, "\n")
	B.Initialize()
	S.Initialize()
	GQ.Start()
}

func versionR (rootValue *G.OutputObjectValue, argumentValues *A.Tree) G.Value {
	return G.MakeStringValue(version)
} //nowR

func init () {
	ts := GQ.TS()
	ts.FixFieldResolver("Query", "version", versionR)
} //init
