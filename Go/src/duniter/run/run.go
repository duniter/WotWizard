package run

import (
	
	B	"duniter/blockchain"
	BA	"duniter/basic"
	G	"duniter/gqlReceiver"
	S	"duniter/sandbox"
		"fmt"
	_	"duniter/centralities"
	_	"duniter/certifications"
	_	"duniter/events"
	_	"duniter/history"
	_	"duniter/identities"
	_	"duniter/identitySearchList"
	_	"duniter/members"
	_	"duniter/parameters"
	_	"duniter/qualities"
	_	"duniter/sandboxList"
	_	"duniter/sentries"
	_	"duniter/server"
	_	"duniter/tellLimits"
	_	"duniter/wotWizardList"
	
	_	"util/graphQL/staticComp"

)

const (
	
	version = "4.0.0"

)

func Start () {
	fmt.Println("WotWizard version", version, "\n")
	BA.Lg.Println("WotWizard version", version, "\n")
	B.Initialize()
	S.Initialize()
	G.Start()
}
