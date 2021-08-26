/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package blockchain

import (
	
	A	"util/avl"
	B	"util/gbTree"
	BA	"duniter/basic"
	C	"strconv"
	F	"path/filepath"
	J	"util/json"
	M	"util/misc"
	Q	"database/sql"
	S	"util/sort"
	SC	"syscall"
	U	"util/sets2"
		"bytes"
		"fmt"
		"os"
		"os/signal"
		"sync"
		"time"
	_	"github.com/mattn/go-sqlite3"

)

const (
	
	// SQLite Driver
	driver = "sqlite3";
	
	syncName = "updating.txt";
	syncDelay = 15 * time.Second // Waiting time of Duniter after its creation of syncName
	verifyPeriod = 2 * time.Second // Minimum delay between two verifications of the presence of syncName
	secureDelay = 2 * time.Second // Security delay before the end of syncDelay
	addDelay = 5 * time.Second // Increment of syncDelay when approching the end and update is not finished
	
	// Default path to the Duniter database
	duniBaseDef = "$HOME/.config/duniter/duniter_default/wotwizard-export.db"
	
	// Directory of the WW database in the resource directory
	systemDef = "System"
	// Name of the Parameters json file
	dParsName = "DPars.json"
	// Name of WW main database
	dBaseName = "DBase.data"
	// Its copy
	dCopyName = "DBase.data.bak"
	// Its copy of copy
	dCopy1Name = "DBase.data.bak1"
	// Name of the Sandbox json database
	sBaseName = "SBase.json"
	
	blockchainName = "Blockchain"

)

var (
	
	 // Path to the Duniter synchronization file
	duniSync = F.Join(BA.DuniDir, syncName)
	
	// Directory of the WW database and names of files inside it
	system = F.Join(BA.RsrcDir(), systemDef)
	dPars = F.Join(system, dParsName)
	dBase = F.Join(system, dBaseName)
	dCopy =F.Join( system, dCopyName)
	dCopy1 =F.Join( system, dCopy1Name)
	sBase = F.Join(system, sBaseName)
	
	addDelayInt int64 = addDelay.Nanoseconds() / 1000000

)

const (
	
	// Max length of a Pubkey
	PubkeyLen = 44
	
	// Number of last blocks to be read again at every update, since they could have changed
	secureGap = 100
	
	// Number of pages used by UtilBTree
	pageNb = 256000
	
	// Numbers of the places of the indexes in dBase
	timePlace = iota // Index timeT
	timeMPlace // Index timeMT
	joinAndLeavePlace // Index joinAndLeaveT
	idPubPlace // Index idPubT
	idUidPlace // Index idUidT
	idHashPlace // Index idHashT
	idTimePlace // Index idTimeT
	certFromPlace // Index certFromT
	certToPlace // Index certToT
	certTimePlace // Index certTimeT
	undoListPlace // Head of the chained list of the operations to be undone before every update
	lastNPlace // Last read block
	idLenPlace // Number of actual members
	
	placeNb // Number of places

)

type (
	
	// Actions to do in Commands
	Actioner interface {
		Activate ()
		Name () string
	}
	
	// Procedure called at every update
	UpdateProc = func (... interface{})
	
	// Chained list of procedures called at every update
	updateListUpdtT struct {
		next *updateListUpdtT
		update UpdateProc
		params []interface{}
	}
	
	updateListT struct {
		next *updateListT
		name string
		update UpdateProc
		params []interface{}
	}
	
	Pubkey string
	Hash string
	
	StringArr []string
	
	CertEvent struct {
		Block int32
		InOut bool
	}
	
	CertEvents []CertEvent
	
	CertHist struct {
		Uid string
		Hist CertEvents
	}
	
	CertHists []CertHist
	
	// Duniter Parameters
	Parameters struct {
		
		// The %growth of the UD every [dtReeval] period = 0.0488 /  (6 months) = 4.88% / (6 months)
		C float64
		
		// Time period between two UD = 86400 s = 1 day
		Dt,
		
		// UD(0), i.e. initial Universal Dividend = 1000 cents = 10 Ğ1
		Ud0,
		
		// Minimum delay between 2 certifications of a same issuer = 432000 s = 5 days
		SigPeriod,
		
		// Maximum quantity of active certifications made by member = 100
		SigStock,
		
		// Maximum delay a certification can wait before being expired for non-writing = 5259600 s = 2 months
		SigWindow,
		
		// Maximum age of an active certification = 63115200 s = 2 years
		SigValidity,
		
		// Minimum delay before replaying a certification = 5259600 s = 2 months
		SigReplay,
		
		// Minimum quantity of signatures to be part of the WoT = 5
		SigQty,
		
		// Maximum delay an identity can wait before being expired for non-writing = 5259600 s = 2 months
		IdtyWindow,
		
		// Maximum delay a membership can wait before being expired for non-writing = 5259600 s = 2 months
		MsWindow,
		
		// Minimum delay between 2 memberships of a same issuer = 5259600 s = 2 months
		MsPeriod int32
		
		// Minimum percent of sentries to reach to match the distance rule = 0.8 = 80%
		Xpercent float64
		
		// Maximum age of an active membership = 31557600 s = 1 year
		MsValidity,
		
		// Maximum distance between a newcomer and [xpercent] of sentries = 5
		StepMax,
		
		// Number of blocks used for calculating median time = 24
		MedianTimeBlocks,
		
		// The average time for writing 1 block (wished time) = 300 s = 5 min
		AvgGenTime,
		
		// The number of blocks required to evaluate again PoWMin value = 12
		DtDiffEval int32
		
		// The percent of previous issuers to reach for personalized difficulty = 0.67 = 67%
		PercentRot float64
		
		// Time of first UD = 1488970800 s = 2017/03/08 11:00:00 UTC+0
		UdTime0,
		
		// Time of first reevaluation of the UD = 1490094000 s = 2017/03/21 11:00:00 UTC+0
		UdReevalTime0 int64
		
		// Time period between two re-evaluation of the UD = 15778800 s = 6 months
		DtReeval,
		
		// Maximum delay a transaction can wait before being expired for non-writing = 604800 s = 7 days
		TxWindow int32
	}
)

const (

	// Types of elements in undoListT
	timeList = iota
	idAddList
	joinList
	activeList
	leaveList
	idAddTimeList
	idRemoveTimeList
	certAddList
	certRemoveList
	remCertifiers
	remCertified

	HasNotLeaved = -1

)

type (
	
	Position = B.IndexReader
	
	// Chained list of the operations to be undone before every update
	undoListT struct {
		next B.FilePos
		typ byte
		// timeList -> timeTy; joinList, activeList, leaveList -> identity; certAddList, certRemoveList -> Certification
		ref B.FilePos
		aux, aux2 int64
	}
	
	// Factory of undoListT
	undoListFacT struct {
	}
	
	// Blocks and their times
	timeTy struct {
		bnb int32
		mTime,
		time int64
	}
	
	// Factory of timeTy
	timeFacT struct {
	}
	
	joinAndLeaveL struct { // stack of identity joining and leaving blocks
		next B.FilePos
		joiningBlock, // Block numbers
		leavingBlock int32
	}
	
	// Factory of joinAndLeaveL
	joinAndLeaveLFacT struct {
	}
	
	joinAndLeave struct {
		pubkey Pubkey
		list B.FilePos // JoinAndLeaveL
	}
	
	// Factory of joinAndLeave
	joinAndLeaveFacT struct {
	}
	
	certInOut struct { // stack of certification joining and leaving blocks
		next B.FilePos
		inBlock,
		outBlock int32
	}
	
	certInOutFacT struct {
	}
	
	identity struct {
		pubkey Pubkey
		uid string
		member bool
		hash Hash
		block_number, // Where the identity is written
		application int32 // block of last membership application (joiners, actives, leavers)
		expires_on int64
		certifiers, // Index of all non revoked certifiers uid, old or present, of this identity ; B.String -> nothing
		certified, // Index of all uid, old or present, and not revoked, certified by this identity ; B.String -> nothing
		certifiersIO,// Index of all certifiers uid, old or present, of this identity, with dates of validity ; B.String -> certInOut
		certifiedIO B.FilePos // Index of all uid, old or present, certified by this identity, with dates of validity ; B.String -> certInOut
	}
	
	// Factory of identity
	identityFacT struct {
	}
	
	certification struct {
		from,
		to Pubkey
		block_number int32 // Where the Certification is written
		expires_on int64
	}
	
	// Factory of certification
	certificationFacT struct {
	}
	
	certToFork struct { // Data of the certToT index
		byPub, // sub-index (pubKey -> certification)
		byExp B.FilePos // sub-index (filePosKey (certification) -> certification) sorted by reverse certification.expires_on; used to find the expiration date of the sigQtyth certification.
	}
	
	certToForkFacT struct{
	}
	
	// int32 index key
	intKey struct {
		ref int32
	}
	
	// Factory of intKey
	intKeyFacT struct {
	}
	
	// Manager of intKey
	intKeyManT struct {
	}
	
	// int64 index key
	lIntKey struct {
		ref int64
	}
	
	// Factory of lIntKey
	lIntKeyFacT struct {
	}
	
	// Manager of lIntKey
	lIntKeyManT struct {
	}
	
	// B.FilePos index key
	filePosKey struct {
		ref B.FilePos
	}
	
	// Factory of filePosKey
	filePosKeyFacT struct {
	}
	
	// Pubkey index key
	pubKey struct { // != Pubkey!
		ref Pubkey
	}
	
	// Factory of pubKey
	pubKeyFacT struct {
	}
	
	// Manager of pubKey
	pubKeyManT struct {
	}
	
	// Hashkey index key
	hashKey struct {
		ref Hash
	}
	
	// Factory of HashKey
	hashKeyFacT struct {
	}
	
	// Manager of HashKey
	hashKeyManT struct {
	}
	
	// Manager of string index (uid)
	uidKeyManT struct {
	}
	
	// Manager of identity sorted by expiration dates
	idKTimeManT struct {
	}
	
	// Manager of certification sorted by expiration dates
	certKTimeManT struct {
	}
	
	CertPos struct {
		posT *Position
	}
	
	// member, members, membersFind: Distance Rule
	member struct {
		p Pubkey
		links U.Set
	}
	
	membersT []member
	
	membersFinder struct {
		len int
		m membersT
	}
	
	PubkeysT []Pubkey
	
	poSET struct {
		pubkeys PubkeysT
		set_1,
		set_2 U.Set
		poS float64
	}

)

const (
	
	// Sizes of B.Data
	pubKeyS = (PubkeyLen + 1) * B.BYS
	
	// Sizes of keys
	timeKeyS = B.INS // intKey
	timeMKeyS = B.LIS // lIntKey
	idTimeKeyS = B.LIS // filePosKey
	certTimeKeyS = B.LIS // filePosKey

)

var (
	
	// Shared variables
	
	lg = BA.Lg
	
	pars Parameters // Duniter parameters
	parsJ J.Json
	
	mutex = new(sync.RWMutex)
	mutexCmds = new(sync.RWMutex)
	
	database *B.Database // duniter0 database
	
	// UtilBTree indexes
	timeT, // intKey -> timeTy; blocks sorted by bnb
	timeMT, // lIntKey -> timeTy; blocks sorted by mTime
	joinAndLeaveT, // pubKey -> joinAndLeave
	idPubT, // pubKey -> identity
	idUidT, // B.String -> identity
	idHashT, // HashKey -> identity
	idTimeT, // lIntKey -> nothing; addresses of identity sorted by expiration dates
	certFromT, // pubKey -> sub-index(pubKey -> certification)
	certToT *B.Index // pubKey -> certToFork
	
	lastBlock int32 = -1 // Last read & updated block
	now, rNow int64 = 0, 0 // Present medianTime and time
	idLenM = 0 // Number of members
	
	// Factories
	timeFac timeFacT
	joinAndLeaveLFac joinAndLeaveLFacT
	joinAndLeaveFac joinAndLeaveFacT
	certInOutFac certInOutFacT
	identityFac identityFacT
	certificationFac certificationFacT
	certToForkFac certToForkFacT
	intKeyFac intKeyFacT
	lIntKeyFac lIntKeyFacT
	filePosKeyFac filePosKeyFacT
	uidKeyFac B.StringFac
	pubKeyFac pubKeyFacT
	hashKeyFac hashKeyFacT
	
	// Data managers
	timeMan,
	joinAndLeaveLMan,
	joinAndLeaveMan,
	certInOutMan,
	idMan,
	certMan,
	certToForkMan *B.DataMan
	
	// Key managerers
	pubKeyManer pubKeyManT
	hashKeyManer hashKeyManT
	uidKeyManer uidKeyManT
	intKeyManer intKeyManT
	lIntKeyManer lIntKeyManT
	
	// Key managers
	pubKeyMan = B.MakeKM(pubKeyManer)
	hashKeyMan = B.MakeKM(hashKeyManer)
	uidKeyMan = B.MakeKM(uidKeyManer)
	intKeyMan = B.MakeKM(intKeyManer)
	lIntKeyMan = B.MakeKM(lIntKeyManer)
	
	// Update variables
	
	updateListUpdt *updateListUpdtT = nil // Updt
	updateList *updateListT = nil // Cmds // Head of updateListT
	sbFirstUpdt UpdateProc // Cmds
	
	doScan1 = true
	firstUpdate,
	startUpdate bool
	
	// UtilBTree indexes
	certTimeT *B.Index // lIntKey -> nothing; addresses of certification sorted by expiration dates in reverse order; used to erase expired certifications
	
	undoList B.FilePos = B.BNil // Head of undoListT
	
	// Factories
	undoListFac undoListFacT
	
	// Data managers
	undoListMan *B.DataMan
	
	// Key managerers
	idKTimeManer idKTimeManT
	certKTimeManer certKTimeManT
	
	// Key managers
	idKTimeMan = B.MakeKM(idKTimeManer)
	certKTimeMan = B.MakeKM(certKTimeManer)
	
	// Commands variables
	
	cmdsUpdateList *updateListT = nil // Head of updateListT
	
	// Sentries & Distance Rule
	sentriesS U.Set // Set of sentries
	poST *A.Tree
	poSTMut sync.RWMutex
	members membersFinder
	memberF = S.TF{Finder: &members}

)

func (m *membersFinder) Less (i, j int) bool {
	findMemberNumMutex.RLock()
	b := m.m[i].p < m.m[j].p
	findMemberNumMutex.RUnlock()
	return b
} //Less

// Data & Data factories procedures

func (t *timeTy) Read (r *B.Reader) {
	t.bnb = r.InInt32()
	t.mTime = r.InInt64()
	t.time = r.InInt64()
} //Read

func (t *timeTy) Write (w *B.Writer) {
	w.OutInt32(t.bnb)
	w.OutInt64(t.mTime)
	w.OutInt64(t.time)
} //Write

func (timeFacT) New (size int) B.Data {
	return new(timeTy)
} //New

func (jlL *joinAndLeaveL) Read (r *B.Reader) {
	jlL.next = r.InFilePos()
	jlL.joiningBlock = r.InInt32()
	jlL.leavingBlock = r.InInt32()
} //Read

func (jlL *joinAndLeaveL) Write (w *B.Writer) {
	w.OutFilePos(jlL.next)
	w.OutInt32(jlL.joiningBlock)
	w.OutInt32(jlL.leavingBlock)
} //Write

func (joinAndLeaveLFacT) New (size int) B.Data {
	return new(joinAndLeaveL)
} //New

func (jl *joinAndLeave) Read (r *B.Reader) {
	jl.pubkey = Pubkey(r.InString())
	jl.list = r.InFilePos()
} //Read

func (jl *joinAndLeave) Write (w *B.Writer) {
	w.OutString(string(jl.pubkey))
	w.OutFilePos(jl.list)
} //Write

func (joinAndLeaveFacT) New (size int) B.Data {
	return new(joinAndLeave)
} //New

func (cio *certInOut) Read (r *B.Reader) {
	cio.next = r.InFilePos()
	cio.inBlock = r.InInt32()
	cio.outBlock = r.InInt32()
} //Read

func (cio *certInOut) Write (w *B.Writer) {
	w.OutFilePos(cio.next)
	w.OutInt32(cio.inBlock)
	w.OutInt32(cio.outBlock)
} //Write

func (certInOutFacT) New (size int) B.Data {
	return new(certInOut)
} //New

func (id *identity) Read (r *B.Reader) {
	id.pubkey = Pubkey(r.InString())
	id.uid = r.InString()
	id.member = r.InBool()
	id.hash = Hash(r.InString())
	id.block_number = r.InInt32()
	id.application = r.InInt32()
	id.expires_on = r.InInt64()
	id.certifiers = r.InFilePos()
	id.certified = r.InFilePos()
	id.certifiersIO = r.InFilePos()
	id.certifiedIO = r.InFilePos()
} //Read

func (id *identity) Write (w *B.Writer) {
	w.OutString(string(id.pubkey))
	w.OutString(id.uid)
	w.OutBool(id.member)
	w.OutString(string(id.hash))
	w.OutInt32(id.block_number)
	w.OutInt32(id.application)
	w.OutInt64(id.expires_on)
	w.OutFilePos(id.certifiers)
	w.OutFilePos(id.certified)
	w.OutFilePos(id.certifiersIO)
	w.OutFilePos(id.certifiedIO)
} //Write

func (identityFacT) New (size int) B.Data {
	return new(identity)
} //New

func (c *certification) Read (r *B.Reader) {
	c.from = Pubkey(r.InString())
	c.to = Pubkey(r.InString())
	c.block_number = r.InInt32()
	c.expires_on = r.InInt64()
} //Read

func (c *certification) Write (w *B.Writer) {
	w.OutString(string(c.from))
	w.OutString(string(c.to))
	w.OutInt32(c.block_number)
	w.OutInt64(c.expires_on)
} //Write

func (certificationFacT) New (size int) B.Data {
	return new(certification)
} //New

func (c *certToFork) Read (r *B.Reader) {
	c.byPub = r.InFilePos()
	c.byExp = r.InFilePos()
} //Read

func (c *certToFork) Write (w *B.Writer) {
	w.OutFilePos(c.byPub)
	w.OutFilePos(c.byExp)
} //Write

func (certToForkFacT) New (size int) B.Data {
	return new(certToFork)
} //New

func (l *undoListT) Read (r *B.Reader) {
	l.next = r.InFilePos()
	l.typ = r.InByte()
	l.ref = r.InFilePos()
	l.aux = r.InInt64()
} //Read

func (l *undoListT) Write (w *B.Writer) {
	w.OutFilePos(l.next)
	w.OutByte(l.typ)
	w.OutFilePos(l.ref)
	w.OutInt64(l.aux)
} //Write

func (undoListFacT) New (size int) B.Data {
	return new(undoListT)
} //New

func (i *intKey) Read (r *B.Reader) {
	i.ref = r.InInt32()
} //Read

func (i *intKey) Write (w *B.Writer) {
	w.OutInt32(i.ref)
} //Write

func (intKeyFacT) New (size int) B.Data {
	M.Assert(size == 0 || size == B.INS, 20)
	if size == 0 {
		return nil
	}
	return new(intKey)
} //New

func (i *lIntKey) Read (r *B.Reader) {
	i.ref = r.InInt64()
} //Read

func (i *lIntKey) Write (w *B.Writer) {
	w.OutInt64(i.ref)
} //Write

func (lIntKeyFacT) New (size int) B.Data {
	M.Assert(size == 0 || size == B.LIS, 20)
	if size == 0 {
		return nil
	}
	return new(lIntKey)
} //New

func (i *filePosKey) Read (r *B.Reader) {
	i.ref = r.InFilePos()
} //Read

func (i *filePosKey) Write (w *B.Writer) {
	w.OutFilePos(i.ref)
} //Write

func (filePosKeyFacT) New (size int) B.Data {
	M.Assert(size == 0 || size == B.LIS, 20)
	if size == 0 {
		return nil
	}
	return new(filePosKey)
} //New

func (pub *pubKey) Read (r *B.Reader) {
	pub.ref = Pubkey(r.InString())
} //Read

func (pub *pubKey) Write (w *B.Writer) {
	w.OutString(string(pub.ref))
} //Write

func (pubKeyFacT) New (size int) B.Data {
	return new(pubKey)
} //New

func (hash *hashKey) Read (r *B.Reader) {
	hash.ref = Hash(r.InString())
} //Read

func (hash *hashKey) Write (w *B.Writer) {
	w.OutString(string(hash.ref))
} //Write

func (hashKeyFacT) New (size int) B.Data {
	return new(hashKey)
} //New

// Key managers procedures

func (intKeyManT) CompP (i1, i2 B.Data) B.Comp {
	if i1 == nil {
		if i2 == nil {
			return B.Eq
		}
		return B.Lt
	}
	if i2 == nil {
		return B.Gt
	}
	ii1 := i1.(*intKey); ii2 := i2.(*intKey)
	if ii1.ref < ii2.ref {
		return B.Lt
	}
	if ii1.ref > ii2.ref {
		return B.Gt
	}
	return B.Eq
} //CompP

func (intKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
} //PrefP

func (lIntKeyManT) CompP (i1, i2 B.Data) B.Comp {
	if i1 == nil {
		if i2 == nil {
			return B.Eq
		}
		return B.Lt
	}
	if i2 == nil {
		return B.Gt
	}
	ii1 := i1.(*lIntKey); ii2 := i2.(*lIntKey)
	if ii1.ref < ii2.ref {
		return B.Lt
	}
	if ii1.ref > ii2.ref {
		return B.Gt
	}
	return B.Eq
} //CompP

func (lIntKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
} //PrefP

func pkmCompP (p1, p2 B.Data) B.Comp {
	M.Assert(p1 != nil && p2 != nil, 20)
	pp1 := p1.(*pubKey); pp2 := p2.(*pubKey)
	if pp1.ref < pp2.ref {
		return B.Lt
	}
	if pp1.ref > pp2.ref {
		return B.Gt
	}
	return B.Eq
} //pkmCompP

func (pubKeyManT) CompP (p1, p2 B.Data) B.Comp {
	return pkmCompP(p1, p2)
} //CompP

func (pubKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
	M.Assert(p1 != nil && *p2 != nil, 20)
	M.Assert(pkmCompP(p1, *p2) == B.Lt, 21)
	pp2 := (*p2).(*pubKey)
	l2 := len(pp2.ref)
	p := new(pubKey)
	l := 0
	b := make([]byte, l, l2)
	p.ref = Pubkey(b)
	for l <= l2 && !(pkmCompP(p1, p) == B.Lt && pkmCompP(p, pp2) <= B.Eq) {
		l++
		b = b[:l]
		b[l - 1] = pp2.ref[l - 1]
		p.ref = Pubkey(b)
	}
	*p2 = p
} //PrefP

func hkmCompP (h1, h2 B.Data) B.Comp {
	M.Assert(h1 != nil && h2 != nil, 20)
	hh1 := h1.(*hashKey); hh2 := h2.(*hashKey)
	if hh1.ref < hh2.ref {
		return B.Lt
	}
	if hh1.ref > hh2.ref {
		return B.Gt
	}
	return B.Eq
} //hkmCompP

func (hashKeyManT) CompP (h1, h2 B.Data) B.Comp {
	return hkmCompP(h1, h2)
} //CompP

func (hashKeyManT) PrefP (h1 B.Data, h2 *B.Data) {
	M.Assert(h1 != nil && *h2 != nil, 20)
	M.Assert(hkmCompP(h1, *h2) == B.Lt, 21)
	hh2 := (*h2).(*hashKey)
	l2 := len(hh2.ref)
	h := new(hashKey)
	l := 0
	b := make([]byte, l, l2)
	h.ref = Hash(b)
	for l <= l2 && !(hkmCompP(h1, h) == B.Lt && hkmCompP(h, hh2) <= B.Eq) {
		l++
		b = b[:l]
		b[l - 1] = hh2.ref[l - 1]
		h.ref = Hash(b)
	}
	*h2 = h
} //PrefP

// Comparison method of two Strings. Use the lexical order.
func (uidKeyManT) CompP (key1, key2 B.Data) B.Comp {
	M.Assert(key1 != nil && key2 != nil, 20)
	k1 := key1.(*B.String); k2 := key2.(*B.String)
	return BA.CompP(k1.C, k2.C)
} //CompP

func (m uidKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
	*p2 = B.StringPrefP(p1.(*B.String), (*p2).(*B.String), m.CompP)
} //PrefP

// Comparison of identity(s) by expiration dates
func (idKTimeManT) CompP (l1, l2 B.Data) B.Comp {
	id1 := idMan.ReadData(l1.(*filePosKey).ref).(*identity)
	id2 := idMan.ReadData(l2.(*filePosKey).ref).(*identity)
	if M.Abs64(id1.expires_on) < M.Abs64(id2.expires_on) {
		return B.Gt // Inverse order, for the use of B.Index.Search to get all the expired identities
	}
	if M.Abs64(id1.expires_on) > M.Abs64(id2.expires_on) {
		return B.Lt
	}
	if id1.pubkey < id2.pubkey {
		return B.Lt
	}
	if id1.pubkey > id2.pubkey {
		return B.Gt
	}
	return B.Eq
} //CompP

func (idKTimeManT) PrefP (p1 B.Data, p2 *B.Data) {
} //PrefP

// Comparison of certification(s) by expiration dates
func (certKTimeManT) CompP (l1, l2 B.Data) B.Comp {
	c1 := certMan.ReadData(l1.(*filePosKey).ref).(*certification)
	c2 := certMan.ReadData(l2.(*filePosKey).ref).(*certification)
	if c1.expires_on < c2.expires_on {
		return B.Gt // Inverse order, for the use of B.Index.Search to get all the expired certifications
	}
	if c1.expires_on > c2.expires_on {
		return B.Lt
	}
	if c1.from < c2.from {
		return B.Lt
	}
	if c1.from > c2.from {
		return B.Gt
	}
	if c1.to < c2.to {
		return B.Lt
	}
	if c1.to > c2.to {
		return B.Gt
	}
	return B.Eq
} //CompP

func (certKTimeManT) PrefP (p1 B.Data, p2 *B.Data) {
} //PrefP

var findMemberNumMutex = new(sync.RWMutex)

func findMemberNum (p Pubkey) (int, bool) {
	n := members.len
	findMemberNumMutex.Lock()
	members.m[n].p = p
	findMemberNumMutex.Unlock()
	memberF.BinSearch(0, n - 1, &n)
	return n, n < members.len
} //findMemberNum

func Driver () string {
	return driver
} //Driver

func System () string {
	return system
} //System

func DPars () string {
	return dPars
} //DPars

func SBase () string {
	return sBase
} //SBase

func Pars () *Parameters {
	p := new(Parameters)
	*p = pars
	return p
} //Pars

func ParsJ () J.Json {
	return parsJ
} //ParsJ

func (p1 *poSET) Compare (p2 A.Comparer) BA.Comp {
	pp2 := p2.(*poSET)
	if len(p1.pubkeys) < len(pp2.pubkeys) {
		return BA.Lt
	}
	if len(p1.pubkeys) > len(pp2.pubkeys) {
		return BA.Gt
	}
	for i := 0; i < len(p1.pubkeys); i++ {
		if p1.pubkeys[i] < pp2.pubkeys[i] {
			return BA.Lt
		}
		if p1.pubkeys[i] > pp2.pubkeys[i] {
			return BA.Gt
		}
	}
	return BA.Eq
} //Compare

// Last read & updated block
func LastBlock () int32 {
	return lastBlock
} //LastBlock

// medianTime
func Now () int64 {
	return now
} //Now

// time
func RealNow () int64 {
	return rNow
} //RealNow

// Open the duniter0 database
func openB () {
	M.Assert(database == nil, 100)
	B.Fac.CloseBase(dBase)
	database = B.Fac.OpenBase(dBase, pageNb)
	if database == nil {
		b := B.Fac.CreateBase(dBase, placeNb); M.Assert(b, 101)
		database = B.Fac.OpenBase(dBase, pageNb); M.Assert(database != nil, 102)
		database.WritePlace(timePlace, int64(database.CreateIndex(timeKeyS)))
		database.WritePlace(timeMPlace, int64(database.CreateIndex(timeMKeyS)))
		database.WritePlace(joinAndLeavePlace, int64(database.CreateIndex(0)))
		database.WritePlace(idPubPlace, int64(database.CreateIndex(0)))
		database.WritePlace(idUidPlace, int64(database.CreateIndex(0)))
		database.WritePlace(idHashPlace, int64(database.CreateIndex(0)))
		database.WritePlace(idTimePlace, int64(database.CreateIndex(idTimeKeyS)))
		database.WritePlace(certFromPlace, int64(database.CreateIndex(0)))
		database.WritePlace(certToPlace, int64(database.CreateIndex(0)))
		database.WritePlace(certTimePlace, int64(database.CreateIndex(certTimeKeyS)))
		database.WritePlace(undoListPlace, int64(B.BNil))
		database.WritePlace(lastNPlace, -1)
		database.WritePlace(idLenPlace, 0)
		lg.Println("\"" + dBaseName + "\" created")
	}
	timeMan = database.CreateDataMan(timeFac)
	joinAndLeaveLMan = database.CreateDataMan(joinAndLeaveLFac)
	joinAndLeaveMan = database.CreateDataMan(joinAndLeaveFac)
	certInOutMan = database.CreateDataMan(certInOutFac)
	idMan = database.CreateDataMan(identityFac)
	certMan = database.CreateDataMan(certificationFac)
	certToForkMan = database.CreateDataMan(certToForkFac)
	undoListMan = database.CreateDataMan(undoListFac)
	timeT = database.OpenIndex(B.FilePos(database.ReadPlace(timePlace)), intKeyMan, intKeyFac)
	timeMT = database.OpenIndex(B.FilePos(database.ReadPlace(timeMPlace)), lIntKeyMan, lIntKeyFac)
	joinAndLeaveT = database.OpenIndex(B.FilePos(database.ReadPlace(joinAndLeavePlace)), pubKeyMan, pubKeyFac)
	idPubT = database.OpenIndex(B.FilePos(database.ReadPlace(idPubPlace)), pubKeyMan, pubKeyFac)
	idUidT = database.OpenIndex(B.FilePos(database.ReadPlace(idUidPlace)), uidKeyMan, uidKeyFac)
	idHashT = database.OpenIndex(B.FilePos(database.ReadPlace(idHashPlace)), hashKeyMan, hashKeyFac)
	idTimeT = database.OpenIndex(B.FilePos(database.ReadPlace(idTimePlace)), idKTimeMan, filePosKeyFac)
	certFromT = database.OpenIndex(B.FilePos(database.ReadPlace(certFromPlace)), pubKeyMan, pubKeyFac)
	certToT = database.OpenIndex(B.FilePos(database.ReadPlace(certToPlace)), pubKeyMan, pubKeyFac)
	certTimeT = database.OpenIndex(B.FilePos(database.ReadPlace(certTimePlace)), certKTimeMan, filePosKeyFac)
	lg.Println("\"" + dBaseName + "\" opened")
} //openB

// Close the duniter0 database
func closeB () {
	M.Assert(database != nil, 100)
	lg.Println("Closing \"" + dBaseName + "\"")
	database.CloseBase()
	database = nil
} //closeB

// Block number -> times
func TimeOf (bnb int32) (mTime, time int64, ok bool) {
	pst := timeT.NewReader()
	ok = pst.Search(&intKey{ref: bnb})
	if ok {
		t := timeMan.ReadData(pst.ReadValue()).(*timeTy)
		mTime = t.mTime
		time = t.time
	}
	return
} //TimeOf

// Median Time -> next Block Number
func BlockAfter (mTime int64) (bnb int32, ok bool) {
	pst := timeMT.NewReader()
	pst.Search(&lIntKey{ref: mTime})
	ok = pst.PosSet()
	if ok {
		bnb = timeMan.ReadData(pst.ReadValue()).(*timeTy).bnb
	}
	return
}

// Pubkey -> joining and leaving blocks (leavingBlock == HasNotLeaved if no leaving block)
func JLPub (pubkey Pubkey) (list B.FilePos, ok bool) {
	pst := joinAndLeaveT.NewReader()
	ok = pst.Search(&pubKey{ref: pubkey})
	if ok {
		list = joinAndLeaveMan.ReadData(pst.ReadValue()).(*joinAndLeave).list
	}
	return
} //JLPub

// Pubkey -> joining and leaving blocks (leavingBlock == HasNotLeaved if no leaving block)
func JLPubLNext (list *B.FilePos) (joiningBlock, leavingBlock int32, ok bool) {
	ok = *list != B.BNil
	if ok {
		jlL := joinAndLeaveLMan.ReadData(*list).(*joinAndLeaveL)
		*list = jlL.next
		joiningBlock = jlL.joiningBlock
		leavingBlock = jlL.leavingBlock
	}
	return
} //JLPubLNext

// Number of joinAndLeave
func JLLen () int {
	return joinAndLeaveT.NumberOfKeys()
} //JLLen

// Browse all joinAndLeave's pubkeys step by step
func JLNextPubkey (first bool, pst **Position) (pubkey Pubkey, ok bool) {
	if first {
		*pst = joinAndLeaveT.NewReader()
	}
	r := *pst
	r.Next()
	ok = r.PosSet()
	if ok {
		pubkey = r.CurrentKey().(*pubKey).ref
	}
	return
} //JLNextPubkey

// Pubkey -> uid
func IdPub (pubkey Pubkey) (uid string, ok bool) {
	pst := idPubT.NewReader()
	ok = pst.Search(&pubKey{ref: pubkey})
	if ok {
		uid = idMan.ReadData(pst.ReadValue()).(*identity).uid
	}
	return
} //IdPub

// Pubkey -> uid of member
func IdPubM (pubkey Pubkey) (uid string, ok bool) {
	pst := idPubT.NewReader()
	ok = pst.Search(&pubKey{ref: pubkey})
	if  ok{
		id := idMan.ReadData(pst.ReadValue()).(*identity)
		ok = id.member
		if ok {
			uid = id.uid
		}
	}
	return
} //IdPubM

// Pubkey -> identity
func IdPubComplete (pubkey Pubkey) (uid string, member bool, hash Hash, block_number, application int32, expires_on int64, ok bool) {
	pst := idPubT.NewReader()
	ok = pst.Search(&pubKey{ref: pubkey})
	if ok {
		id := idMan.ReadData(pst.ReadValue()).(*identity)
		uid = id.uid
		member = id.member
		hash = id.hash
		block_number = id.block_number
		application = id.application
		expires_on = id.expires_on
	}
	return
} //IdPubComplete

// uid -> identity
func IdUid (uid string) (pubkey Pubkey, ok bool) {
	pst := idUidT.NewReader()
	ok = pst.Search(&B.String{C: uid})
	if ok {
		pubkey = idMan.ReadData(pst.ReadValue()).(*identity).pubkey
	}
	return
} //IdUid

// uid -> Pubkey of member
func IdUidM (uid string) (pubkey Pubkey, ok bool) {
	pst := idUidT.NewReader()
	ok = pst.Search(&B.String{C: uid})
	if ok {
		id := idMan.ReadData(pst.ReadValue()).(*identity)
		ok = id.member
		if ok {
			pubkey = id.pubkey
		}
	}
	return
} //IdUidM

// uid -> identity
func IdUidComplete (uid string) (pubkey Pubkey, member bool, hash Hash, block_number, application int32, expires_on int64, ok bool) {
	pst := idUidT.NewReader()
	ok = pst.Search(&B.String{C: uid})
	if ok {
		id := idMan.ReadData(pst.ReadValue()).(*identity)
		pubkey = id.pubkey
		member = id.member
		hash = id.hash
		block_number = id.block_number
		application = id.application
		expires_on = id.expires_on
	}
	return
} //IdUidComplete

// Hash -> pubkey
func IdHash (hash Hash) (pub Pubkey, ok bool) {
	pst := idHashT.NewReader()
	ok = pst.Search(&hashKey{ref: hash})
	if ok {
		pub = idMan.ReadData(pst.ReadValue()).(*identity).pubkey
	}
	return
} //IdHash

// Number of identities
func IdLen () int {
	return idUidT.NumberOfKeys()
} //IdLen

// Number of members
func IdLenM () int {
	return idLenM
} //IdLenM

// Position next identity's pubkey for IdNextPubkey
func IdPosPubkey (pubkey Pubkey) *Position {
	pst := idPubT.NewReader()
	_ = pst.Search(&pubKey{ref: pubkey})
	pst.Previous()
	return pst
} //IdPosPubkey

// Browse all identity's pubkeys step by step
func IdNextPubkey (first bool, pst **Position) (pubkey Pubkey, ok bool) {
	if first {
		*pst = idPubT.NewReader()
	}
	r := *pst
	r.Next()
	ok = r.PosSet()
	if ok {
		pubkey = r.CurrentKey().(*pubKey).ref
	}
	return
} //IdNextPubkey

// Browse all members' pubkeys step by step
func IdNextPubkeyM (first bool, pst **Position) (pubkey Pubkey, ok bool) {
	if first {
		*pst = idPubT.NewReader()
	}
	r := *pst
	for {
		r.Next()
		ok = r.PosSet()
		if !ok {
			break
		}
		id := idMan.ReadData(r.ReadValue()).(*identity)
		if id.member {
			pubkey = id.pubkey
			break
		}
	}
	return
} //IdNextPubkeyM

// Position next identity's uid for IdNextUid
func IdPosUid (uid string) *Position {
	pst := idUidT.NewReader()
	pst.Search(&B.String{C: uid})
	pst.Previous()
	return pst
} //IdPosUid

// Browse all identity's uid(s) lexicographically step by step
func IdNextUid (first bool, pst **Position) (uid string, ok bool) {
	if first {
		*pst = idUidT.NewReader()
	}
	r := *pst
	r.Next()
	ok = r.PosSet()
	if ok {
		uid = r.CurrentKey().(*B.String).C
	}
	return
} //IdNextUid

// Browse all members' uid(s) lexicographically step by step
func IdNextUidM (first bool, pst **Position) (uid string, ok bool) {
	if first {
		*pst = idUidT.NewReader()
	}
	r := *pst
	for {
		r.Next()
		ok = r.PosSet()
		if !ok {
			break
		}
		id := idMan.ReadData(r.ReadValue()).(*identity)
		if id.member {
			uid = id.uid
			break
		}
	}
	return
} //IdNextUidM

// (Pubkey, Pubkey) -> certification
func Cert (from, to Pubkey) (bnb int32, expires_on int64, ok bool) {
	ir1 := certFromT.NewReader()
	ok = ir1.Search(&pubKey{ref: from})
	if ok {
		ir2 := database.OpenIndex(ir1.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
		ok = ir2.Search(&pubKey{ref: to})
		if ok {
			c := certMan.ReadData(ir2.ReadValue()).(*certification)
			bnb = c.block_number
			expires_on = c.expires_on
		}
	}
	return
} //Cert

// Pubkey -> head of sub-index
func CertFrom (from Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	pst := certFromT.NewReader()
	ok = pst.Search(&pubKey{ref: from})
	if ok {
		pos.posT = database.OpenIndex(pst.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertFrom

// Pubkey -> head of sub-index
func CertTo (to Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	pst := certToT.NewReader()
	ok = pst.Search(&pubKey{ref: to})
	if ok {
		pos.posT = database.OpenIndex(certToForkMan.ReadData(pst.ReadValue()).(*certToFork).byPub, pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertTo

// Pubkey -> head of sub-index
func CertToByExp (to Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	pst := certToT.NewReader()
	ok = pst.Search(&pubKey{ref: to})
	if ok {
		pos.posT = database.OpenIndex(certToForkMan.ReadData(pst.ReadValue()).(*certToFork).byExp, certKTimeMan, filePosKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertToByExp

// Number of keys in sub-index
func (pos *CertPos) CertPosLen () int {
	M.Assert(pos != nil, 20)
	if pos.posT == nil {
		return 0
	}
	return pos.posT.Ind().NumberOfKeys()
} //CertPosLen

// Browse all certification's pairs of Pubkey in a sub-index step by step
func (pos *CertPos) CertNextPos () (from, to Pubkey, ok bool) {
	ok = pos.posT != nil
	if ok {
		pst := pos.posT
		pst.Next()
		ok = pst.PosSet()
		if ok {
			c := certMan.ReadData(pst.ReadValue()).(*certification)
			from = c.from
			to = c.to
		}
	}
	return
} //CertNextPos

// Position next sub-index to pubKey for CertNextFrom
func CertPosFrom (pubkey Pubkey) *Position {
	pst := certFromT.NewReader()
	_ = pst.Search(&pubKey{ref: pubkey})
	pst.Previous()
	return pst
} //CertPosFrom

// Browse all sub-indexes step by step in the lexicographic order of the from Pubkey
func CertNextFrom (first bool, pos *CertPos, pst **Position) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(pst != nil, 21)
	if first {
		*pst = certFromT.NewReader()
	}
	r := *pst
	M.Assert(r != nil, 22)
	r.Next()
	ok = r.PosSet()
	if ok {
		pos.posT = database.OpenIndex(r.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertNextFrom

// Position next sub-index to pubKey for CertNextTo
func CertPosTo (pubkey Pubkey) *Position {
	pst := certToT.NewReader()
	_ = pst.Search(&pubKey{ref: pubkey})
	pst.Previous()
	return pst
} //CertPosTo

// Browse all sub-indexes step by step in the lexicographic order of the to Pubkey
func CertNextTo (first bool, pos *CertPos, pst **Position) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(pst != nil, 21)
	if first {
		*pst = certToT.NewReader()
	}
	r := *pst
	M.Assert(r != nil, 22)
	r.Next()
	ok = r.PosSet()
	if ok {
		pos.posT = database.OpenIndex(certToForkMan.ReadData(r.ReadValue()).(*certToFork).byPub, pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertNextTo

// Browse all sub-indexes step by step in the lexicographic order of the to Pubkey
func CertNextToByExp (first bool, pos *CertPos, pst **Position) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(pst != nil, 21)
	if first {
		*pst = certToT.NewReader()
	}
	r := *pst
	M.Assert(r != nil, 22)
	r.Next()
	ok = r.PosSet()
	if ok {
		pos.posT = database.OpenIndex(certToForkMan.ReadData(r.ReadValue()).(*certToFork).byExp, certKTimeMan, filePosKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
} //CertNextToByExp

func AllCertifiers (to string) StringArr {
	pst := idUidT.NewReader()
	if !pst.Search(&B.String{C: to}) {
		return make(StringArr, 0)
	}
	id := idMan.ReadData(pst.ReadValue()).(*identity)
	if id.certifiers == B.BNil {
		return make(StringArr, 0)
	}
	ind := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac)
	from := make(StringArr, ind.NumberOfKeys())
	pst = ind.NewReader()
	pst.Next()
	i := 0
	for pst.PosSet() {
		from[i] = pst.CurrentKey().(*B.String).C
		i++
		pst.Next()
	}
	M.Assert(i == len(from), 60)
	return from
} //AllCertifiers

func AllCertified (from string) StringArr {
	pst := idUidT.NewReader()
	if !pst.Search(&B.String{C: from}) {
		return make(StringArr, 0)
	}
	id := idMan.ReadData(pst.ReadValue()).(*identity)
	if id.certified == B.BNil {
		return make(StringArr, 0)
	}
	ind := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac)
	to := make(StringArr, ind.NumberOfKeys())
	pst = ind.NewReader()
	pst.Next()
	i := 0
	for pst.PosSet() {
		to[i] = pst.CurrentKey().(*B.String).C
		i++
		pst.Next()
	}
	M.Assert(i == len(to), 60)
	return to
} //AllCertified

type
	hList struct {
		next *hList
		inBlock,
		outBlock int32
	}

func AllCertifiersIO (to string) CertHists {
	pst := idUidT.NewReader()
	if !pst.Search(&B.String{C: to}) {
		return make(CertHists, 0)
	}
	id := idMan.ReadData(pst.ReadValue()).(*identity)
	if id.certifiersIO == B.BNil {
		return make(CertHists, 0)
	}
	ind := database.OpenIndex(id.certifiersIO, uidKeyMan, uidKeyFac)
	from := make(CertHists, ind.NumberOfKeys())
	pst = ind.NewReader()
	pst.Next()
	i := 0
	for pst.PosSet() {
		from[i].Uid = pst.CurrentKey().(*B.String).C
		var l *hList = nil
		n := 0
		cioR := pst.ReadValue()
		for cioR != B.BNil {
			cio := certInOutMan.ReadData(cioR).(*certInOut)
			l = &hList{next: l, inBlock: cio.inBlock, outBlock: cio.outBlock}
			n++
			if l.outBlock != HasNotLeaved {
				n++
			}
			cioR = cio.next
		}
		h := make(CertEvents, n)
		n = 0
		for l != nil {
			h[n] = CertEvent{l.inBlock, true}
			n++
			if l.outBlock != HasNotLeaved {
				h[n] = CertEvent{l.outBlock, false}
				n++
			}
			l = l.next
		}
		from[i].Hist = h
		i++
		pst.Next()
	}
	M.Assert(i == len(from), 60)
	return from
} //AllCertifiersIO

func AllCertifiedIO (from string) CertHists {
	pst := idUidT.NewReader()
	if !pst.Search(&B.String{C: from}) {
		return make(CertHists, 0)
	}
	id := idMan.ReadData(pst.ReadValue()).(*identity)
	if id.certifiedIO == B.BNil {
		return make(CertHists, 0)
	}
	ind := database.OpenIndex(id.certifiedIO, uidKeyMan, uidKeyFac)
	to := make(CertHists, ind.NumberOfKeys())
	pst = ind.NewReader()
	pst.Next()
	i := 0
	for pst.PosSet() {
		to[i].Uid = pst.CurrentKey().(*B.String).C
		var l *hList = nil
		n := 0
		cioR := pst.ReadValue()
		for cioR != B.BNil {
			cio := certInOutMan.ReadData(cioR).(*certInOut)
			l = &hList{next: l, inBlock: cio.inBlock, outBlock: cio.outBlock}
			n++
			if l.outBlock != HasNotLeaved {
				n++
			}
			cioR = cio.next
		}
		h := make(CertEvents, n)
		n = 0
		for l != nil {
			h[n] = CertEvent{l.inBlock, true}
			n++
			if l.outBlock != HasNotLeaved {
				h[n] = CertEvent{l.outBlock, false}
				n++
			}
			l = l.next
		}
		to[i].Hist = h
		i++
		pst.Next()
	}
	M.Assert(i == len(to), 60)
	return to
} //AllCertifiedIO

func IsSentry (pubkey Pubkey) bool {
	e, ok := findMemberNum(pubkey)
	return ok && sentriesS.In(e)
} //IsSentry

// Return in pubkey the next sentry's pubkey if !first or the first one if first; return false if there is no more sentry
func NextSentry (first bool, sentriesI **U.SetIterator) (pubkey Pubkey, ok bool) {
	var sentryCur int
	if first {
		*sentriesI = sentriesS.Attach()
		sentryCur, ok = (*sentriesI).FirstE()
	} else {
		sentryCur, ok = (*sentriesI).NextE()
	}
	if ok {
		pubkey = members.m[sentryCur].p
	}
	return
} //NextSentry

// Return the number of sentries
func SentriesLen () int {
	return sentriesS.NbElems()
} //SentriesLen

// Array of certifiers' pubkeys -> % of sentries reached in pars.stepMax - 1 steps
func PercentOfSentriesS (pubkeys PubkeysT) (set_1, set_2 U.Set, poS float64) {
	
	sort := func (pubkeys []Pubkey) {
		for i := 1; i < len(pubkeys); i++ {
			p := pubkeys[i]
			j := i
			for j > 0 && p < pubkeys[j - 1] {
				pubkeys[j] = pubkeys[j - 1]
				j--
			}
			pubkeys[j] = p
		}
	} //sort

	find := func (poSE *poSET) (set_1, set_2 U.Set, poS float64, ok bool) {
		(&poSTMut).RLock()
		el, ok, _ := poST.Search(poSE)
		(&poSTMut).RUnlock()
		if ok {
			p := el.Val().(*poSET)
			set_1 = p.set_1
			set_2 = p.set_2
			poS = p.poS
		}
		return
	} //find

	store := func (poSE *poSET, set_1, set_2 U.Set, poS float64) {
		poSE.set_1 = set_1
		poSE.set_2 = set_2
		poSE.poS = poS
		(&poSTMut).Lock()
		poST.SearchIns(poSE)
		(&poSTMut).Unlock()
	} //store
	
	// PercentOfSentriesS
	sort(pubkeys)
	poSE := &poSET{pubkeys: pubkeys}
	set_1, set_2, poS, ok := find(poSE)
	if !ok {
		set := U.NewSet()
		frontier := U.NewSet()
		for i := 0; i < len(pubkeys); i++ {
			if e, b := findMemberNum(pubkeys[i]); b {
				set.Incl(e)
				frontier.Incl(e)
			}
		}
		for i := 1; i < int(pars.StepMax); i++ {
			newFrontier := U.NewSet()
			frontierI := frontier.Attach()
			e, ok := frontierI.FirstE()
			for ok {
				newFrontier.Add(members.m[e].links)
				e, ok = frontierI.NextE()
			}
			frontier = newFrontier
			set.Add(frontier)
			if i == int(pars.StepMax) - 2 {
				set_2 = set.Inter(sentriesS)
			}
		}
		set_1 = set.Inter(sentriesS)
		poS = float64(set_1.NbElems()) / float64(SentriesLen())
		store(poSE, set_1, set_2, poS)
	}
	return
} //PercentOfSentriesS

// Array of certifiers' pubkeys -> % of sentries reached in pars.stepMax - 1 steps
func PercentOfSentries (pubkeys PubkeysT) float64 {
	_, _, poS := PercentOfSentriesS(pubkeys)
	return poS
} //PercentOfSentries

// Verify the distance rule for a set of certifiers' pubkeys
func DistanceRuleOk (pubkeys PubkeysT) bool {
	return PercentOfSentries(pubkeys) >= pars.Xpercent
} //DistanceRuleOk

// Updt
// Scan the string s from position i to the position of stop excluded; update i and return the scanned string in sub
func scanS (s []rune, stop rune, i *int) string {
	sub := new(bytes.Buffer)
	for *i < len(s) && s[*i] != stop {
		sub.WriteRune(s[*i])
		*i++
	}
	*i++
	return string(sub.Bytes())
} //scanS

// Updt
// Skip the string s from position i to the position of stop excluded; update i
func skipS (s []rune, stop rune, i *int) {
	for *i < len(s) && s[*i] != stop {
		*i++
	}
	*i++
} //skipS

// Updt
// Extract Duniter parameters from block 0
func paramsUpdt (d *Q.DB) {
	
	const
		txWindow = 60 * 60 * 24 * 7
	
	lg.Println("Reading money parameters")
	row := d.QueryRow("SELECT parameters FROM block WHERE number == 0 AND NOT fork")
	var ns Q.NullString
	err := row.Scan(&ns)
	M.Assert(err == nil, err, 100)
	M.Assert(ns.Valid, 101)
	ss := bytes.Runes([]byte(ns.String))
	var n int
	i := 0
	s := scanS(ss, ':', &i); pars.C, err = C.ParseFloat(s, 64); M.Assert(err == nil, err, 102)
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 103); pars.Dt = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 104); pars.Ud0 = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 105); pars.SigPeriod = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 106); pars.SigStock = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 107); pars.SigWindow = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 108); pars.SigValidity = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 109); pars.SigQty = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 110); pars.IdtyWindow = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 111); pars.MsWindow = int32(n);
	s = scanS(ss, ':', &i); pars.Xpercent, err = C.ParseFloat(s, 64); M.Assert(err == nil, err, 112)
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 113); pars.MsValidity = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 114); pars.StepMax = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 115); pars.MedianTimeBlocks = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 116); pars.AvgGenTime = int32(n);
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 117); pars.DtDiffEval = int32(n);
	s = scanS(ss, ':', &i); pars.PercentRot, err = C.ParseFloat(s, 64); M.Assert(err == nil, err, 118)
	s = scanS(ss, ':', &i); pars.UdTime0, err = C.ParseInt(s, 10, 64); M.Assert(err == nil, err, 119)
	s = scanS(ss, ':', &i); pars.UdReevalTime0, err = C.ParseInt(s, 10, 64); M.Assert(err == nil, err, 120)
	s = scanS(ss, ':', &i); n, err = C.Atoi(s); M.Assert(err == nil, err, 121); pars.DtReeval = int32(n);
	pars.TxWindow = txWindow
	pars.MsPeriod = pars.MsWindow
	pars.SigReplay = pars.MsPeriod
} //paramsUpdt

// Cmds
// Extract Duniter parameters from JSON file
func params () {
	parsJ = J.ReadFile(dPars); M.Assert(parsJ != nil, 100)
	J.ApplyTo(parsJ, &pars)
} //params

// Updt
// Add a block in timeT and timeMT
func times (withList bool, bnb int, mTime, time int64) {
	t := &timeTy{bnb: int32(bnb), mTime: mTime, time: time}
	tRef := timeMan.WriteAllocateData(t)
	iw := timeT.Writer()
	b := iw.SearchIns(&intKey{ref: int32(bnb)}); M.Assert(!b, bnb, 100)
	iw.WriteValue(tRef)
	iw = timeMT.Writer()
	b = iw.SearchIns(&lIntKey{ref: mTime})
	var aux int64 = 0
	if !b { // Different blocks may have the same mTime
		iw.WriteValue(tRef)
		aux = 1
	}
	if withList {
		tL := &undoListT{next: undoList, typ: timeList, ref: tRef, aux: aux}
		undoList = undoListMan.WriteAllocateData(tL)
	}
	now = M.Max64(now, mTime)
	rNow = M.Max64(rNow, time)
} //times

// Updt
func removeCertifiersCertified (withList bool, idRef B.FilePos, id *identity) {
	idU1 := &B.String{C: id.uid}
	
	if id.certifiers != B.BNil {
		pst := idUidT.NewReader()
		c1 := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).NewReader()
		c1.Next()
		for c1.PosSet() {
			idU2 := c1.CurrentKey().(*B.String)
			b := pst.Search(idU2); M.Assert(b, idU2.C, 100)
			id2Ref := pst.ReadValue()
			id2 := idMan.ReadData(id2Ref).(*identity)
			M.Assert(id2.certified != B.BNil, 101)
			ind := database.OpenIndex(id2.certified, uidKeyMan, uidKeyFac)
			c2 := ind.Writer()
			b = c2.Erase(idU1); M.Assert(b, idU1.C, 102)
			if ind.IsEmpty() {
				database.DeleteIndex(id2.certified); id2.certified = B.BNil
				idMan.WriteData(id2Ref, id2)
			}
			if withList {
				rL := &undoListT{next: undoList, typ: remCertified, ref: id2Ref, aux: int64(idRef)}
				undoList = undoListMan.WriteAllocateData(rL)
			}
			c1.Next()
		}
	}
	
	if id.certified != B.BNil {
		pst := idUidT.NewReader()
		c1 := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).NewReader()
		c1.Next()
		for c1.PosSet() {
			idU2 := c1.CurrentKey().(*B.String)
			b := pst.Search(idU2); M.Assert(b, idU2.C, 103)
			id2Ref := pst.ReadValue()
			id2 := idMan.ReadData(id2Ref).(*identity)
			M.Assert(id2.certifiers != B.BNil, 104)
			ind := database.OpenIndex(id2.certifiers, uidKeyMan, uidKeyFac)
			c2 := ind.Writer()
			b = c2.Erase(idU1); M.Assert(b, idU1.C, 105)
			if ind.IsEmpty() {
				database.DeleteIndex(id2.certifiers); id2.certifiers = B.BNil
				idMan.WriteData(id2Ref, id2)
			}
			if withList {
				rL := &undoListT{next: undoList, typ: remCertifiers, ref: id2Ref, aux: int64(idRef)}
				undoList = undoListMan.WriteAllocateData(rL)
			}
			c1.Next()
		}
	}
} //removeCertifiersCertified

// Updt
func revokeId (withList bool, p Pubkey) {
	pst := idPubT.NewReader()
	b := pst.Search(&pubKey{ref: p}); M.Assert(b, p, 100)
	idRef := pst.ReadValue()
	id := idMan.ReadData(idRef).(*identity)
	if withList {
		idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: int64(id.application)}
		undoList = undoListMan.WriteAllocateData(idL)
	}
	id.expires_on = BA.Revoked
	removeCertifiersCertified(withList, idRef, id)
	idMan.WriteData(idRef, id)
} //revokeId

// Updt
// For one block, add joining & leaving identities in joinAndLeaveT and updade identities in idPubT and idUidT; update certFromT & certToT too
func identities (withList bool, ssJ, ssA, ssL, ssR, ssE string, nb int, d *Q.DB) {
	
	iwP := idPubT.Writer()
	iwU := idUidT.Writer()
	iwH := idHashT.Writer()
	iwJ := joinAndLeaveT.Writer()
	
	var b bool
	ss := bytes.Runes([]byte(ssJ))
	i := 1
	for ss[i] != ']' { // joiners : Insert id
		i++
		idLenM++
		id := &identity{member: true}
		id.pubkey = Pubkey(scanS(ss, ':', &i))
		skipS(ss, ':', &i)
		s := scanS(ss, '-', &i); n, err := C.Atoi(s); M.Assert(err == nil, err, 100)
		id.application = int32(n)
		id.expires_on, _, b = TimeOf(int32(n)); M.Assert(b, n, 102)
		id.expires_on += int64(pars.MsValidity)
		skipS(ss, ':', &i)
		skipS(ss, ':', &i)
		id.uid = scanS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		rows, err := d.Query("SELECT hash FROM i_index WHERE pub == '" + string(id.pubkey) + "' ORDER BY writtenOn ASC")
		M.Assert(err == nil, err, 103)
		id.hash = ""
		for rows.Next() {
			var s Q.NullString
			err = rows.Scan(&s)
			M.Assert(err == nil, err, 104)
			if s.Valid {
				id.hash = Hash(s.String)
			}
		}
		M.Assert(rows.Err() == nil, rows.Err(), 60)
		M.Assert(id.hash != "", 105)
		bnb := int32(nb)
		id.block_number = bnb
		idU := &B.String{C: id.uid}
		idP := &pubKey{ref: id.pubkey}
		idH := &hashKey{ref: id.hash}
		var idRef B.FilePos
		if iwP.SearchIns(idP) {
			idRef = iwP.ReadValue()
			oldId := idMan.ReadData(idRef).(*identity)
			M.Assert(!oldId.member && oldId.uid == id.uid && oldId.hash == id.hash, 106)
			idTimeT.Writer().Erase(&filePosKey{ref: idRef})
			if  withList {
				idL := &undoListT{next: undoList, typ: idRemoveTimeList, ref: idRef, aux: 0, aux2: 0}
				undoList = undoListMan.WriteAllocateData(idL)
			}
			id.block_number = oldId.block_number
			id.certifiers = oldId.certifiers; id.certified = oldId.certified
			id.certifiersIO = oldId.certifiersIO; id.certifiedIO = oldId.certifiedIO
			idMan.WriteData(idRef, id)
			b = iwU.SearchIns(idU); M.Assert(b, idU.C, 107)
			M.Assert(iwU.ReadValue() == idRef, 108)
			b = iwH.SearchIns(idH); M.Assert(b, idH.ref, 109)
			M.Assert(iwH.ReadValue() == idRef, 110)
			if withList {
				idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: oldId.expires_on, aux2: int64(oldId.application)}
				undoList = undoListMan.WriteAllocateData(idL)
			}
		} else {
			id.certifiers = B.BNil; id.certified = B.BNil
			id.certifiersIO = B.BNil; id.certifiedIO = B.BNil
			idRef = idMan.WriteAllocateData(id)
			iwP.WriteValue(idRef)
			b = iwU.SearchIns(idU); M.Assert(!b, idU.C, 111)
			iwU.WriteValue(idRef)
			b = iwH.SearchIns(idH); M.Assert(!b, idH.ref, 112)
			iwH.WriteValue(idRef)
			if withList {
				idL := &undoListT{next: undoList, typ: idAddList, ref: idRef, aux: 0}
				undoList = undoListMan.WriteAllocateData(idL)
			}
		}
		var (jlRef B.FilePos; jl *joinAndLeave)
		if iwJ.SearchIns(idP) {
			jlRef = iwJ.ReadValue()
			jl = joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
		} else {
			jl = &joinAndLeave{pubkey: id.pubkey, list: B.BNil}
			jlRef = joinAndLeaveMan.AllocateData(jl)
			iwJ.WriteValue(jlRef)
		}
		jlL := &joinAndLeaveL{next: jl .list, joiningBlock: bnb, leavingBlock: HasNotLeaved}
		jl.list = joinAndLeaveLMan.WriteAllocateData(jlL)
		joinAndLeaveMan.WriteData(jlRef, jl)
		if withList {
			idL := &undoListT{next: undoList, typ: joinList, ref: idRef, aux: 0}
			undoList = undoListMan.WriteAllocateData(idL)
		}
	}
	
	ss = bytes.Runes([]byte(ssA))
	i = 1
	for ss[i] != ']' { // actives
		i++
		idP := new(pubKey)
		idP.ref = Pubkey(scanS(ss, ':', &i))
		skipS(ss, ':', &i)
		s := scanS(ss, '-', &i); n, err := C.Atoi(s); M.Assert(err == nil, err, 113)
		skipS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		b = iwP.Search(idP); M.Assert(b, idP.ref, 114)
		idRef := iwP.ReadValue()
		id := idMan.ReadData(idRef).(*identity)
		M.Assert(id.member, 115)
		if withList {
			idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: int64(id.application)}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		id.application = int32(n)
		id.expires_on, _, b = TimeOf(int32(n)); M.Assert(b, n, 117)
		id.expires_on += int64(pars.MsValidity)
		idMan.WriteData(idRef, id)
	}
	
	ss = bytes.Runes([]byte(ssL))
	i = 1
	for ss[i] != ']' { // leavers
		i++
		idP := new(pubKey)
		idP.ref = Pubkey(scanS(ss, ':', &i))
		skipS(ss, ':', &i)
		s := scanS(ss, '-', &i); n, err := C.Atoi(s); M.Assert(err == nil, err, 118)
		skipS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		b = iwP.Search(idP); M.Assert(b, idP.ref, 119)
		idRef := iwP.ReadValue()
		id := idMan.ReadData(idRef).(*identity)
		if withList {
			idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: int64(id.application)}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		id.application = int32(n)
		id.expires_on = - M.Abs64(id.expires_on) // id.expires_on < 0 if leaving
		idMan.WriteData(idRef, id)
	}
	
	ss = bytes.Runes([]byte(ssR))
	i = 1
	for ss[i] != ']' { // revoked
		i++
		p := Pubkey(scanS(ss, ':', &i))
		skipS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		idP := &pubKey{ref: p}
		b = iwP.Search(idP); M.Assert(b, idP.ref, 120)
		idRef := iwP.ReadValue()
		if idTimeT.Writer().Erase(&filePosKey{ref: idRef}) && withList {
			idL := &undoListT{next: undoList, typ: idRemoveTimeList, ref: idRef, aux: 0}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		revokeId(withList, p);
	}
	
	ss = bytes.Runes([]byte(ssE))
	i = 1
	for ss[i] != ']' { // excluded
		i++
		idLenM--
		idP := new(pubKey)
		idP.ref = Pubkey(scanS(ss, '"', &i))
		if ss[i] != ']' {
			i++
		}
		b = iwP.Search(idP); M.Assert(b, idP.ref, 121)
		idRef := iwP.ReadValue()
		id := idMan.ReadData(idRef).(*identity)
		M.Assert(id.member, 122)
		id.member = false
		idMan.WriteData(idRef, id)
		b = iwJ.Search(idP); M.Assert(b, idP.ref, 123)
		jlRef := iwJ.ReadValue()
		jl := joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
		jlL := joinAndLeaveLMan.ReadData(jl.list).(*joinAndLeaveL)
		M.Assert(jlL.leavingBlock == HasNotLeaved, 124)
		jlL.leavingBlock = int32(nb)
		joinAndLeaveMan.WriteData(jl.list, jlL)
		if withList {
			idL := &undoListT{next: undoList, typ: leaveList, ref: idRef, aux: 0}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		if id.expires_on != BA.Revoked {
			if withList {
				idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: int64(id.application)}
				undoList = undoListMan.WriteAllocateData(idL)
			}
			if id.expires_on >= 0 { // !leaving
				id.expires_on += int64(pars.MsValidity)
			} else { // leaving
				id.expires_on -= int64(pars.MsValidity)
			}
			idMan.WriteData(idRef, id)
			b = idTimeT.Writer().SearchIns(&filePosKey{ref: idRef}); M.Assert(!b, 125)
			if withList {
				idL := &undoListT{next: undoList, typ: idAddTimeList, ref: idRef, aux: 0}
				undoList = undoListMan.WriteAllocateData(idL)
			}
		}
	}
} //identities

// Updt
// Add certifications of one block in certFromT, certToT and certTimeT
func certifications (withList bool, ssC string, nb int) {
	
	iwF := certFromT.Writer()
	iwT := certToT.Writer()
	iwTi := certTimeT.Writer()
	iwP := idPubT.Writer()
	
	ss := bytes.Runes([]byte(ssC))
	i := 1
	for ss[i] != ']' {
		i++
		c := new(certification)
		c.from = Pubkey(scanS(ss, ':', &i))
		c.to = Pubkey(scanS(ss, ':', &i))
		s := scanS(ss, ':', &i)
		skipS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		c.block_number = int32(nb)
		n, err := C.Atoi(s); M.Assert(err == nil, err, 100)
		var b bool
		c.expires_on, _, b = TimeOf(int32(n)); M.Assert(b, n, 101)
		c.expires_on += int64(pars.SigValidity)
		pC := certMan.WriteAllocateData(c)
		var v, vE B.FilePos
		// Insert into certFromT
		idP := &pubKey{ref: c.from}
		if iwF.SearchIns(idP) {
			v = iwF.ReadValue()
		} else {
			v = database.CreateIndex(pubKeyS)
			iwF.WriteValue(v)
		}
		idP.ref = c.to
		iw := database.OpenIndex(v, pubKeyMan, pubKeyFac).Writer()
		var oldPC B.FilePos = B.BNil
		if iw.SearchIns(idP) {
			oldPC = iw.ReadValue()
		}
		iw.WriteValue(pC)
		// Insert into certToT
		idP.ref = c.to
		if iwT.SearchIns(idP) {
			ctf := certToForkMan.ReadData(iwT.ReadValue()).(*certToFork)
			v = ctf.byPub
			vE = ctf.byExp
		} else {
			v = database.CreateIndex(pubKeyS)
			vE = database.CreateIndex(certTimeKeyS)
			ctf := &certToFork{byPub: v, byExp: vE}
			iwT.WriteValue(certToForkMan.WriteAllocateData(ctf))
		}
		// Into byPub (pubKey -> certification)
		idP.ref = c.from
		iw = database.OpenIndex(v, pubKeyMan, pubKeyFac).Writer()
		iw.SearchIns(idP)
		iw.WriteValue(pC)
		// Into byExp (filePosKey (certification) -> certification sorted by reverse certification.expires_on)
		iw = database.OpenIndex(vE, certKTimeMan, filePosKeyFac).Writer()
		b = iw.SearchIns(&filePosKey{ref: pC}); M.Assert(!b, 102)
		iw.WriteValue(pC)
		// Erase old certification, if any, in byExp
		if oldPC != B.BNil {
			b = iw.Erase(&filePosKey{ref: oldPC}); M.Assert(b, 103)
		}
		if oldPC == B.BNil {
			idU := new(B.String)
			idP.ref = c.from
			b = iwP.Search(idP); M.Assert(b, idP.ref, 104)
			idRef := iwP.ReadValue()
			id := idMan.ReadData(idRef).(*identity)
			idU.C, b = IdPub(c.to); M.Assert(b, c.to, 105)
			// Add c.to to id.certified, where id is the identity of c.from
			if id.certified == B.BNil {
				id.certified = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).Writer()
			iw.SearchIns(idU)
			// Add c to id.certifiedIO, where id is the identity of c.from
			if id.certifiedIO == B.BNil {
				id.certifiedIO = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certifiedIO, uidKeyMan, uidKeyFac).Writer()
			if iw.SearchIns(idU) {
				cioR := iw.ReadValue()
				cio := certInOutMan.ReadData(cioR).(*certInOut)
				M.Assert(cio.outBlock != HasNotLeaved, 106)
				cio = &certInOut{next: cioR, inBlock: int32(nb), outBlock: HasNotLeaved}
				cioR = certInOutMan.WriteAllocateData(cio)
				iw.WriteValue(cioR)
			} else {
				cio := &certInOut{next: B.BNil, inBlock: int32(nb), outBlock: HasNotLeaved}
				cioR := certInOutMan.WriteAllocateData(cio)
				iw.WriteValue(cioR)
			}
			
			idP.ref = c.to
			b = iwP.Search(idP); M.Assert(b, idP.ref, 107)
			idRef = iwP.ReadValue()
			id = idMan.ReadData(idRef).(*identity)
			idU.C, b = IdPub(c.from); M.Assert(b, c.from, 108)
			// Add c.from to id.certifiers, where id is the identity of c.to
			if id.certifiers == B.BNil {
				id.certifiers = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).Writer()
			iw.SearchIns(idU)
			// Add c to id.certifiersIO, where id is the identity of c.to
			if id.certifiersIO == B.BNil {
				id.certifiersIO = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certifiersIO, uidKeyMan, uidKeyFac).Writer()
			if iw.SearchIns(idU) {
				cioR := iw.ReadValue()
				cio := certInOutMan.ReadData(cioR).(*certInOut)
				M.Assert(cio.outBlock != HasNotLeaved, 109)
				cio = &certInOut{next: cioR, inBlock: int32(nb), outBlock: HasNotLeaved}
				cioR = certInOutMan.WriteAllocateData(cio)
				iw.WriteValue(cioR)
			} else {
				cio := &certInOut{next: B.BNil, inBlock: int32(nb), outBlock: HasNotLeaved}
				cioR := certInOutMan.WriteAllocateData(cio)
				iw.WriteValue(cioR)
			}
		} else {
			b = iwTi.Erase(&filePosKey{ref: oldPC}); M.Assert(b, 110)
		}
		b = iwTi.SearchIns(&filePosKey{ref: pC}); M.Assert(!b, 111)
		if withList {
			cL := &undoListT{next: undoList, typ: certAddList, ref: pC, aux: int64(oldPC)}
			undoList = undoListMan.WriteAllocateData(cL)
		} else if oldPC != B.BNil {
			certMan.EraseData(oldPC)
		}
	}
} //certifications

// Updt
// Remove c keys from certFromT and certToT
func removeCertSimply (c *certification, pC B.FilePos) {
	pKFrom := &pubKey{ref: c.from}
	pKTo := &pubKey{ref: c.to}
	iK := &filePosKey{ref: pC}
	
	iw1 := certFromT.Writer()
	b := iw1.Search(pKFrom); M.Assert(b, 100)
	n := iw1.ReadValue()
	ind := database.OpenIndex(n, pubKeyMan, pubKeyFac)
	iw2 := ind.Writer()
	b = iw2.Erase(pKTo); M.Assert(b, 101)
	if ind.IsEmpty() {
		database.DeleteIndex(n)
		b = iw1.Erase(pKFrom); M.Assert(b, 102)
	}
	
	iw1 = certToT.Writer()
	b = iw1.Search(pKTo); M.Assert(b, 103)
	nF := iw1.ReadValue()
	ctf := certToForkMan.ReadData(nF).(*certToFork)
	n = ctf.byPub
	nE := ctf.byExp
	ind = database.OpenIndex(n, pubKeyMan, pubKeyFac)
	iw2 = ind.Writer()
	b = iw2.Erase(pKFrom); M.Assert(b, 104)
	empty := ind.IsEmpty()
	ind = database.OpenIndex(nE, certKTimeMan, filePosKeyFac)
	iw2 = ind.Writer()
	b = iw2.Erase(iK)
	M.Assert(ind.IsEmpty() == empty, 105)
	if empty {
		database.DeleteIndex(n)
		database.DeleteIndex(nE)
		certToForkMan.EraseData(nF)
		b = iw1.Erase(pKTo); M.Assert(b, 106)
	}
} //removeCertSimply

// Updt
// Remove c keys from certFromT and certToT
func removeCert (c *certification, pC B.FilePos) {
	removeCertSimply(c, pC)
	
	iwP := idPubT.Writer()
	idP := &pubKey{ref: c.from}
	b := iwP.Search(idP); M.Assert(b, idP.ref, 107)
	idFRef := iwP.ReadValue()
	idF := idMan.ReadData(idFRef).(*identity)
	idUF := &B.String{C: idF.uid}
	idP.ref = c.to
	b = iwP.Search(idP); M.Assert(b, idP.ref, 108)
	idTRef := iwP.ReadValue()
	idT := idMan.ReadData(idTRef).(*identity)
	idUT := &B.String{C: idT.uid}
	
	iw := database.OpenIndex(idF.certifiedIO, uidKeyMan, uidKeyFac).Writer()
	b = iw.SearchIns(idUT); M.Assert(b, idF.uid, idUT.C, 109)
	cioR := iw.ReadValue()
	cio := certInOutMan.ReadData(cioR).(*certInOut)
	M.Assert(cio.outBlock == HasNotLeaved, 110)
	cio.outBlock, b = BlockAfter(c.expires_on); M.Assert(b, c.expires_on, 111)
	certInOutMan.WriteData(cioR, cio)
	
	iw = database.OpenIndex(idT.certifiersIO, uidKeyMan, uidKeyFac).Writer()
	b = iw.SearchIns(idUF); M.Assert(b, idT.uid, idUF.C, 112)
	cioR = iw.ReadValue()
	cio = certInOutMan.ReadData(cioR).(*certInOut)
	M.Assert(cio.outBlock == HasNotLeaved, 113)
	cio.outBlock, b = BlockAfter(c.expires_on); M.Assert(b, c.expires_on, 114)
	certInOutMan.WriteData(cioR, cio)
} //removeCert

// Updt
// Remove expired certifications from certFromT and certToT
func removeExpiredCerts (now, secureNow int64) {
	iw := certTimeT.Writer()
	c := &certification{expires_on: now, from: ""}
	pC := certMan.WriteAllocateData(c)
	iw.Search(&filePosKey{ref: pC})
	certMan.EraseData(pC)
	for iw.PosSet() {
		pC = iw.CurrentKey().(*filePosKey).ref
		iw.Next()
		c = certMan.ReadData(pC).(*certification)
		removeCert(c, pC)
		withList := c.expires_on >= secureNow
		b := iw.Erase(&filePosKey{ref: pC}); M.Assert(b, 100)
		if withList {
			cL := &undoListT{next: undoList, typ: certRemoveList, ref: pC, aux: 0}
			undoList = undoListMan.WriteAllocateData(cL)
		} else {
			certMan.EraseData(pC)
		}
	}
} //removeExpiredCerts

// Updt
func revokeExpiredIds (now, secureNow int64) {
	id := &identity{expires_on: now, pubkey:"", uid: ""}
	pId := idMan.WriteAllocateData(id)
	iw := idTimeT.Writer()
	iw.Search(&filePosKey{ref: pId})
	idMan.EraseData(pId)
	for iw.PosSet() {
		pId = iw.CurrentKey().(*filePosKey).ref
		iw.Next()
		id = idMan.ReadData(pId).(*identity)
		b := iw.Erase(&filePosKey{ref: pId}); M.Assert(b, 100)
		withList := M.Abs64(id.expires_on) >= secureNow
		if withList {
			idL := &undoListT{next: undoList, typ: idRemoveTimeList, ref: pId, aux: 0}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		revokeId(withList, id.pubkey)
	}
} //revokeExpiredIds

// Updt
// Undo the last operations done from the secureGap last blocks
func removeSecureGap () {
	for undoList != B.BNil {
		l := undoListMan.ReadData(undoList).(*undoListT)
		switch l.typ {
			case timeList:
				// Erase the timeTy data pointed by l.ref and the corresponding keys in timeT and timeMT
				t := timeMan.ReadData(l.ref).(*timeTy)
				b := timeT.Writer().Erase(&intKey{ref: t.bnb}); M.Assert(b, t.bnb, 100)
				if l.aux == 1 {
					b = timeMT.Writer().Erase(&lIntKey{ref: t.mTime}); M.Assert(b, t.mTime, 101)
				}
				timeMan.EraseData(l.ref)
			case idAddList:
				// Erase the identity data pointed by l.ref and the corresponding keys in idPubT, idHashT and idUidT
				id := idMan.ReadData(l.ref).(*identity)
				b := idPubT.Writer().Erase(&pubKey{ref: id.pubkey}); M.Assert(b, id.pubkey, 102)
				b = idUidT.Writer().Erase(&B.String{C: id.uid}); M.Assert(b, id.uid, 103)
				b = idHashT.Writer().Erase(&hashKey{ref: id.hash}); M.Assert(b, id.hash, 104)
				idMan.EraseData(l.ref)
			case joinList:
				// Let the identity no more be member; erase the last joinAndLeaveL data corresponding to l.ref; if this is also the first data, erase the corresponding joinAndLeave data and its key in joinAndLeaveT
				id := idMan.ReadData(l.ref).(*identity)
				id.member = false
				idMan.WriteData(l.ref, id)
				idLenM--
				p := &pubKey{ref: id.pubkey}
				iw := joinAndLeaveT.Writer()
				b := iw.Search(p); M.Assert(b, 105)
				jlRef := iw.ReadValue()
				jl := joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
				jlLRef := jl.list
				jlL := joinAndLeaveLMan.ReadData(jlLRef).(*joinAndLeaveL)
				M.Assert(jlL.leavingBlock == HasNotLeaved, 106)
				if jlL.next == B.BNil {
					joinAndLeaveMan.EraseData(jlRef)
					b = iw.Erase(p); M.Assert(b, 107)
				} else {
					jl.list = jlL.next
					joinAndLeaveMan.WriteData(jlRef, jl)
				}
				joinAndLeaveLMan.EraseData(jlLRef)
			case activeList:
				// Undo the identity.expires_on and identity.application updates
				id := idMan.ReadData(l.ref).(*identity)
				id.expires_on = l.aux
				id.application = int32(l.aux2)
				idMan.WriteData(l.ref, id)
			case leaveList:
				// Update the last joinAndLeaveL data corresponding to l.ref
				id := idMan.ReadData(l.ref).(*identity)
				id.member = true
				idMan.WriteData(l.ref, id)
				idLenM++
				pst := joinAndLeaveT.NewReader()
				b := pst.Search(&pubKey{ref: id.pubkey}); M.Assert(b, 108)
				jlRef := pst.ReadValue()
				jl := joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
				jlL := joinAndLeaveLMan.ReadData(jl.list).(*joinAndLeaveL)
				M.Assert(jlL.leavingBlock != HasNotLeaved, 109)
				jlL.leavingBlock = HasNotLeaved
				joinAndLeaveLMan.WriteData(jl.list, jlL)
			case idAddTimeList:
				b := idTimeT.Writer().Erase(&filePosKey{ref: l.ref}); M.Assert(b, 110)
			case idRemoveTimeList:
				b := idTimeT.Writer().SearchIns(&filePosKey{ref: l.ref}); M.Assert(!b, 111)
			case certAddList:
				// Erase the keys corresponding to the certification pointed by l.ref in certFromT and certToT, or, if l.aux # B.BNil, update them; modify identity. certifiers, identity.certified, identity.certifiersIO and identity.certifiedIO as needed
				
				remCertifiedrs := func (idRef B.FilePos, id *identity, certifiedrs *B.FilePos, key *B.String) {
					M.Assert(*certifiedrs != B.BNil, 200)
					ind := database.OpenIndex(*certifiedrs, uidKeyMan, uidKeyFac)
					iw := ind.Writer()
					b := iw.Erase(key); M.Assert(b, 201)
					if ind.IsEmpty() {
						database.DeleteIndex(*certifiedrs); *certifiedrs = B.BNil
						idMan.WriteData(idRef, id)
					}
				} //remCertifiedrs
				
				remCertifiedrsIO := func (idRef B.FilePos, id *identity, certifiedrsIO *B.FilePos, key *B.String) {
					M.Assert(*certifiedrsIO != B.BNil, 300)
					ind := database.OpenIndex(*certifiedrsIO, uidKeyMan, uidKeyFac)
					iw := ind.Writer()
					b := iw.Search(key); M.Assert(b, 301)
					cioR := iw.ReadValue()
					cio := certInOutMan.ReadData(cioR).(*certInOut)
					M.Assert(cio.outBlock == HasNotLeaved, 302)
					certInOutMan.EraseData(cioR)
					cioR = cio.next
					if cioR == B.BNil {
						b = iw.Erase(key); M.Assert(b, 303)
						if ind.IsEmpty() {
							database.DeleteIndex(*certifiedrsIO); *certifiedrsIO = B.BNil
							idMan.WriteData(idRef, id)
						}
					} else {
						iw.WriteValue(cioR)
					}
				} //remCertifiedrsIO
				
				c := certMan.ReadData(l.ref).(*certification)
				iwT := certTimeT.Writer()
				b := iwT.Erase(&filePosKey{ref: l.ref}); M.Assert(b, 112)
				pst := idPubT.NewReader()
				p := new(pubKey)
				if B.FilePos(l.aux) == B.BNil {
					
					removeCertSimply(c, l.ref)
					u := new(B.String)
					
					p.ref = c.from
					b = pst.Search(p); M.Assert(b, 113)
					idRef := pst.ReadValue()
					id := idMan.ReadData(idRef).(*identity)
					u.C, b = IdPub(c.to); M.Assert(b, 114)
					remCertifiedrs(idRef, id, &id.certified, u)
					remCertifiedrsIO(idRef, id, &id.certifiedIO, u)
					
					p.ref = c.to
					b = pst.Search(p); M.Assert(b, 115)
					idRef = pst.ReadValue()
					id = idMan.ReadData(idRef).(*identity)
					u.C, b = IdPub(c.from); M.Assert(b, 116)
					remCertifiedrs(idRef, id, &id.certifiers, u)
					remCertifiedrsIO(idRef, id, &id.certifiersIO, u)
					
				}else {
					p.ref = c.from
					irF := certFromT.NewReader()
					b = irF.Search(p); M.Assert(b, 117)
					n := irF.ReadValue()
					iw := database.OpenIndex(n, pubKeyMan, pubKeyFac).Writer()
					p.ref = c.to
					b = iw.Search(p); M.Assert(b, 118)
					iw.WriteValue(B.FilePos(l.aux))
					p.ref = c.to
					irT := certToT.NewReader()
					b = irT.Search(p); M.Assert(b, 119)
					ctf := certToForkMan.ReadData(irT.ReadValue()).(*certToFork)
					iw = database.OpenIndex(ctf.byPub, pubKeyMan, pubKeyFac).Writer()
					p.ref = c.from
					b = iw.Search(p); M.Assert(b, 120)
					iw.WriteValue(B.FilePos(l.aux))
					iw = database.OpenIndex(ctf.byExp, certKTimeMan, filePosKeyFac).Writer()
					b = iw.Erase(&filePosKey{ref: B.FilePos(l.ref)}); M.Assert(b, l.ref, 121)
					b = iw.SearchIns(&filePosKey{ref: B.FilePos(l.aux)}); M.Assert(!b, l.aux, 122)
					iw.WriteValue(B.FilePos(l.aux))
					b = iwT.SearchIns(&filePosKey{ref: B.FilePos(l.aux)}); M.Assert(!b, 123)
				}
				certMan.EraseData(l.ref)
			case certRemoveList:
				// Insert the keys corresponding to the certification pointed by l.ref into certFromT, certToT and certTimeT; modify identity.certifiersIO and identity.certifiedIO as needed
				c := certMan.ReadData(l.ref).(*certification)
				p := &pubKey{ref: c.from}
				var n B.FilePos
				iwF := certFromT.Writer()
				if iwF.SearchIns(p) {
					n = iwF.ReadValue()
				} else {
					n = database.CreateIndex(pubKeyS)
					iwF.WriteValue(n)
				}
				p.ref = c.to
				iw := database.OpenIndex(n, pubKeyMan, pubKeyFac).Writer()
				b := iw.SearchIns(p); M.Assert(!b, 124)
				iw.WriteValue(l.ref)
				iwT := certToT.Writer()
				var ctf *certToFork
				if iwT.SearchIns(p) {
					ctf = certToForkMan.ReadData(iwT.ReadValue()).(*certToFork)
				} else {
					ctf = new(certToFork)
					ctf.byPub = database.CreateIndex(pubKeyS)
					ctf.byExp = database.CreateIndex(certTimeKeyS)
					iwT.WriteValue(certToForkMan.WriteAllocateData(ctf))
				}
				p.ref = c.from
				iw = database.OpenIndex(ctf.byPub, pubKeyMan, pubKeyFac).Writer()
				b = iw.SearchIns(p); M.Assert(!b, 125)
				iw.WriteValue(l.ref)
				iw = database.OpenIndex(ctf.byExp, certKTimeMan, filePosKeyFac).Writer()
				b = iw.SearchIns(&filePosKey{ref: l.ref}); M.Assert(!b, 126)
				iw.WriteValue(l.ref)
				b = certTimeT.Writer().SearchIns(&filePosKey{ref: l.ref}); M.Assert(!b, 127)
				
				iwP := idPubT.Writer()
				idP := &pubKey{ref: c.from}
				b = iwP.Search(idP); M.Assert(b, idP.ref, 128)
				idFRef := iwP.ReadValue()
				idF := idMan.ReadData(idFRef).(*identity)
				idUF := &B.String{C: idF.uid}
				idP.ref = c.to
				b = iwP.Search(idP); M.Assert(b, idP.ref, 129)
				idTRef := iwP.ReadValue()
				idT := idMan.ReadData(idTRef).(*identity)
				idUT := &B.String{C: idT.uid}
				
				iw = database.OpenIndex(idF.certifiedIO, uidKeyMan, uidKeyFac).Writer()
				b = iw.SearchIns(idUT); M.Assert(b, idF.uid, idUT.C, 130)
				cioR := iw.ReadValue()
				cio := certInOutMan.ReadData(cioR).(*certInOut)
				M.Assert(cio.outBlock != HasNotLeaved, 131)
				cio.outBlock = HasNotLeaved
				certInOutMan.WriteData(cioR, cio)
				
				iw = database.OpenIndex(idT.certifiersIO, uidKeyMan, uidKeyFac).Writer()
				b = iw.SearchIns(idUF); M.Assert(b, idT.uid, idUF.C, 132)
				cioR = iw.ReadValue()
				cio = certInOutMan.ReadData(cioR).(*certInOut)
				M.Assert(cio.outBlock != HasNotLeaved, 133)
				cio.outBlock = HasNotLeaved
				certInOutMan.WriteData(cioR, cio)
			case remCertifiers:
				id := idMan.ReadData(B.FilePos(l.aux)).(*identity)
				u := &B.String{C: id.uid}
				id = idMan.ReadData(l.ref).(*identity)
				if id.certifiers == B.BNil {
					id.certifiers = database.CreateIndex(0)
					idMan.WriteData(l.ref, id)
				}
				iw := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).Writer()
				b := iw.SearchIns(u); M.Assert(!b, 134)
			case remCertified:
				id := idMan.ReadData(B.FilePos(l.aux)).(*identity)
				u := &B.String{C: id.uid}
				id = idMan.ReadData(l.ref).(*identity)
				if id.certified == B.BNil {
					id.certified = database.CreateIndex(0)
					idMan.WriteData(l.ref, id)
				}
				iw := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).Writer()
				b := iw.SearchIns(u); M.Assert(!b, 135)
			default:
				M.Halt(136)
		}
		undoListMan.EraseData(undoList)
		undoList = l.next
	}
} //removeSecureGap

//Cmds
func threshold (m, s int) int {
	
	pow := func (x, y int) int {
		z := 1
		for {
			if y & 1 != 0 {
				z *= x
			}
			if y >>= 1; y == 0 {
				break
			}
			x *= x
		}
		return z
	}
	
	n := 0
	for  pow(n, s) < m{
		n++
	}
	return n
} //threshold

func SentryThreshold () int {
	return threshold(IdLenM(), int(pars.StepMax))
} //SentryThreshold

// Cmds
// Initialize members and sentriesS
func calculateSentries (... interface{}) {
	members.len = IdLen()
	members.m = make(membersT, members.len + 1)
	var (pst *Position; pos CertPos)
	i := 0
	p, ok := IdNextPubkey(true, &pst)
	for ok {
		M.Assert(i == 0 || p > members.m[i - 1].p, 100)
		members.m[i].p = p
		members.m[i].links = U.NewSet()
		i++
		p, ok = IdNextPubkey(false, &pst)
	}
	M.Assert(i == members.len, 101)
	for i = 0; i < members.len; i++ {
		if CertTo(members.m[i].p, &pos) {
			p, _, ok = pos.CertNextPos()
			for ok {
				e, b := findMemberNum(p); M.Assert(b, 102)
				members.m[i].links.Incl(e)
				p, _, ok = pos.CertNextPos()
			}
		}
	}
	
	sentriesS = U.NewSet()
	n := SentryThreshold()
	if n == 0 {
		return
	}
	p, ok = IdNextPubkeyM(true, &pst)
	for ok {
		if CertFrom(p, &pos) && pos.CertPosLen() >= n && CertTo(p, &pos) && pos.CertPosLen() >= n {
			e, b := findMemberNum(p); M.Assert(b, 103)
			sentriesS.Incl(e)
		}
		p, ok = IdNextPubkeyM(false, &pst)
	}
	
	poST = A.New()
} //calculateSentries

// Updt
// Insert datas from all the blocks from the secureGapth block before the last read
func scanBlocksUpdt (d *Q.DB) {
	lg.Println("Updating \"" + dBaseName + "\"")
	idLenM = int(database.ReadPlace(idLenPlace))
	undoList = B.FilePos(database.ReadPlace(undoListPlace))
	removeSecureGap();
	lastBlock = int32(database.ReadPlace(lastNPlace))
	maxN := -1
	row := d.QueryRow("SELECT max(number) FROM block WHERE NOT fork")
	var r int
	err := row.Scan(&r)
	if err == nil {
		maxN = r
	}
	var secureNow int64 = M.MaxInt64
	var n = 0
	if maxN >= secureGap {
		n = maxN - secureGap + 1
	}
	s := C.Itoa(n)
	row = d.QueryRow("SELECT medianTime FROM block WHERE NOT fork AND number = " + s)
	var m time.Time
	err = row.Scan(&m)
	if err == nil {
		secureNow = m.Unix()
	}
	s = C.Itoa(int(lastBlock - secureGap + 1))
	rs, err := d.Query("SELECT number, medianTime, time, joiners, actives, leavers, revoked, excluded, certifications FROM block WHERE NOT fork AND number >= " + s + " ORDER BY number ASC")
	M.Assert(err == nil, err, 100)
	defer rs.Close()
	for rs.Next() {
		var (
			number int
			t time.Time
			j,
			a,
			l,
			r,
			e,
			c Q.NullString
		)
		err = rs.Scan(&number, &m, &t, &j, &a, &l, &r, &e, &c)
		M.Assert(err == nil, err, 101)
		medianTime := m.Unix()
		time := t.Unix()
		M.Assert(j.Valid, 102); joiners := j.String
		M.Assert(a.Valid, 103); actives := a.String
		M.Assert(l.Valid, 104); leavers := l.String
		M.Assert(r.Valid, 105); revoked := r.String
		M.Assert(e.Valid, 106); excluded := e.String
		M.Assert(c.Valid, 107); certificationList := c.String
		if number > maxN - 10 || number % 5000 == 0 {
			lg.Println("Added block ", number)
		}
		withList := number >= n
		times(withList , number, medianTime, time)
		revokeExpiredIds(medianTime, secureNow)
		identities(withList, joiners, actives, leavers, revoked, excluded, number, d)
		removeExpiredCerts(medianTime, secureNow) // Élimine toutes les certifications expirées avec réversibilité dans secureGap
		certifications(withList, certificationList, number)
	}
	M.Assert(rs.Err() == nil, rs.Err(), 60)
	database.WritePlace(undoListPlace, int64(undoList))
	lastBlock = int32(maxN)
	database.WritePlace(lastNPlace, int64(lastBlock))
	database.WritePlace(idLenPlace, int64(idLenM))
	lg.Println("\"" + dBaseName + "\" updated")
	lg.Println("Median Time:", m.Local().Format("2/01/2006 15:04:05"))
	lg.Println("Number of members: ", idLenM)
} //scanBlocksUpdt

// Updt
func AddUpdateProcUpdt (updateProc UpdateProc, params ... interface{}) {
	l := updateListUpdt
	var m *updateListUpdtT = nil
	for l != nil {
		m = l
		l = l.next
	}
	l = &updateListUpdtT{next: nil, update: updateProc, params: params}
	if m == nil {
		updateListUpdt = l
	} else {
		m.next = l
	}
} //AddUpdateProcUpdt

// Updt
// Scan the Duniter database
func scan (... interface{}) {
	lg.Println("Opening Duniter database (bis)")
	d, err := Q.Open(driver, BA.DuniBase)
	M.Assert(err == nil, err, 100)
	defer d.Close()
	scanBlocksUpdt(d)
} //scan

// Updt
// Scan the Duniter parameters in block 0
func scan1 () {
	lg.Println("Opening Duniter database")
	d, err := Q.Open(driver, BA.DuniBase)
	M.Assert(err == nil, err, 100)
	defer d.Close()
	paramsUpdt(d)
} //scan1

// Updt
func exportParameters () {
	lg.Println("Exporting money parameters")
	f, err := os.Create(dPars); M.Assert(err == nil, err, 100)
	defer f.Close()
	j := J.BuildJsonFrom(&pars); M.Assert(j != nil, 101)
	j.Write(f)
	lg.Println("Money parameters exported")
} //exportParameters

// Updt
func doUpdates (done, updateReady chan<- bool) {
	mutex.Lock()
	lg.Println("Updating WotWizard database")
	if doScan1 {
		doScan1 = false
		scan1()
		exportParameters()
	}
	l := updateListUpdt
	for l != nil {
		l.update(l.params...)
		l = l.next
	}
	database.UpdateBase()
	mutex.Unlock()
	lg.Println("WotWizard database updated")
	done <- true
	updateReady <- true
} //doUpdates

// Updt
func readSyncTime () int64 {
	f, err := os.Open(duniSync); M.Assert(err == nil, err, 100)
	defer f.Close()
	var t int64
	_, err = fmt.Fscanf(f, "%d", &t); M.Assert(err == nil, err, 101)
	return t
} //readSyncTime

// Updt
func writeSyncTime (t int64) {
	f, err := os.Create(duniSync); M.Assert(err == nil, err, 100)
	defer f.Close()
	_, err = fmt.Fprintf(f, "%d", t); M.Assert(err == nil, err, 101)
} //writeSyncTime

// Updt
func updateAllUpdt (stopProg <-chan os.Signal, updateReady chan<- bool) {
	for {
		err := os.Remove(duniSync)
		M.Assert(err == nil || os.IsNotExist(err), err, 100)
		lg.Println("\"" +  syncName + "\" erased")
		lg.Println("Looking for", duniSync); lg.Println()
		f, err := os.Open(duniSync)
		for os.IsNotExist(err) {
			select {
			case <-stopProg:
				lg.Println("Halting"); lg.Println()
				mutex.Lock()
				closeB()
				return
			default:
			}
			time.Sleep(verifyPeriod)
			f, err = os.Open(duniSync)
		}
		M.Assert(err == nil, err, 101)
		f.Close()
		lg.Println("\"" + syncName + "\" seen; reading it")
		var done = make(chan bool)
		t0 := readSyncTime()
		ct1 := time.NewTicker(syncDelay - verifyPeriod - addDelay - secureDelay)
		go doUpdates(done, updateReady)
		select {
		case <- done:
			ct1.Stop()
		case <- ct1.C:
			ct1.Reset(addDelay)
			innerLoop:
			for {
				select {
				case <- done:
					ct1.Stop()
					break innerLoop
				case <- ct1.C:
					t0 += addDelayInt
					writeSyncTime(t0)
				}
			}
		}
	}
} //updateAllUpdt

// Updt
func saveBase () {
	os.Remove(dCopy1)
	os.Rename(dCopy, dCopy1)
	const bufferSize = 0X800
	f, err := os.Open(dBase)
	if err == nil {
		defer f.Close()
		lg.Println("Making a copy of \"" + dBaseName + "\"")
		fC, err := os.Create(dCopy); M.Assert(err == nil, err, 100)
		defer fC.Close()
		buf := make([]byte, bufferSize)
		n, err := f.Read(buf)
		for err == nil {
			_, err = fC.Write(buf[:n]); M.Assert(err == nil, err, 101)
			n, err = f.Read(buf)
		}
		lg.Println("Copy made")
	}
} //saveBase

var updateProMutex = new(sync.Mutex)

// Cmds
func AddUpdateProc (name string, updateProc UpdateProc, params ... interface{}) {
	updateProMutex.Lock()
	l := updateList
	var m *updateListT = nil
	for l != nil && l.name != name {
		m = l
		l = l.next
	}
	if l == nil {
		lg.Println("Adding", name, "to updateList")
		l = &updateListT{next: nil, name: name}
		if m == nil {
			updateList = l
		} else {
			m.next = l
		}
	} else {
		lg.Println("Updating", name, "in updateList")
	}
	l.update = updateProc
	l.params = params
	updateProMutex.Unlock()
} //AddUpdateProc

// Cmds
func RemoveUpdateProc (name string) {
	updateProMutex.Lock()
	l := updateList
	var m *updateListT = nil
	for l != nil && l.name != name {
		m = l
		l = l.next
	}
	if l != nil {
		lg.Println("Removing", name, "from updateList")
		if m == nil {
			updateList = l.next
		} else {
			m.next = l.next
		}
	}
	updateProMutex.Unlock()
} //RemoveUpdateProc

// Cmds
func FixSandBoxFUpdt (updateProc UpdateProc) {
	sbFirstUpdt = updateProc
} //FixSandBoxFUpdt

// Cmds
func updateAll () {
	l := updateList
	for l != nil {
		l.update(l.params...)
		l = l.next
	}
} //updateAll

// Cmds
func updateCmds () {
	lg.Println("Starting update of commands")
	if firstUpdate {
		params()
		sbFirstUpdt()
		firstUpdate = false
	}
	updateAll()
	lg.Println("Update of commands done")
} //updateCmds

// Cmds
func updateFirstCmds () {
	lg.Println("Starting first update")
	params()
	idLenM = int(database.ReadPlace(idLenPlace))
	lastBlock = int32(database.ReadPlace(lastNPlace))
	if lastBlock >= 0 {
		var b bool
		now, rNow, b = TimeOf(lastBlock); M.Assert(b, lastBlock, 100)
	}
	calculateSentries()
	sbFirstUpdt()
	lg.Println("First update done")
} //updateFirstCmds

// Cmds
func doAction (a Actioner) {
	lg.Println("Starting action", a.Name())
	mutexCmds.RLock()
	mutex.RLock()
	a.Activate()
	mutex.RUnlock()
	mutexCmds.RUnlock()
	lg.Println("Action", a.Name(), "done")
} //doAction

// Cmds
func dispatchActions (updateReady <-chan bool, newAction chan Actioner) {
	
	type
		
		actionerStack struct {
			next *actionerStack
			a Actioner
		}
	
	if startUpdate {
		startUpdate = false
		mutexCmds.Lock()
		mutex.RLock()
		updateFirstCmds()
		mutex.RUnlock()
		mutexCmds.Unlock()
	}
	
	var s *actionerStack = nil
	for {
		select {
		case <-updateReady:
			mutexCmds.Lock()
			mutex.RLock()
			updateCmds()
			mutex.RUnlock()
			mutexCmds.Unlock()
		case a := <-newAction:
			s = &actionerStack{next: s, a: a}
		}
		if !firstUpdate {
			for s != nil {
				go doAction(s.a)
				s = s.next
			}
		}
	}
} //dispatchActions

func Start (newAction chan Actioner) {
	lg.Println("Starting"); lg.Println()
	saveBase()
	openB()
	stopProg := make(chan os.Signal, 1)
	signal.Notify(stopProg, SC.SIGHUP, SC.SIGINT, SC.SIGTERM)
	updateReady := make(chan bool)
	go dispatchActions(updateReady, newAction)
	updateAllUpdt(stopProg, updateReady)
} //Start

func virgin () bool {
	f, err := os.Open(dPars)
	if err == nil {
		f.Close()
		f, err = os.Open(dBase)
	}
	if err == nil {
		f.Close()
		f, err = os.Open(sBase)
	}
	if err == nil {
		f.Close()
	}
	return err != nil
} //virgin

func Initialize () {
	AddUpdateProcUpdt(scan)
	firstUpdate = virgin()
	startUpdate = !firstUpdate	
	AddUpdateProc(blockchainName, calculateSentries)
} //Initialize

func init() {
	os.MkdirAll(system, 0777)
} //init
