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
	U	"util/sets2"
		"bytes"
		"fmt"
		"math"
		"os"
		"sync"
		"time"
	_	"github.com/mattn/go-sqlite3"

)

const (
	
	// SQLite Driver
	driver = "sqlite3";
	
	syncName = "updating.txt";
	syncDelay = 15 * time.Second
	verifyPeriod = 2 * time.Second // Minimum delay between two verifications of the presence of syncName
	checkPeriod = 200 * time.Millisecond
	secureDelay = 2 * time.Second
	addDelay = 5 * time.Second
	stopPeriod = 500 * time.Millisecond
	
	// Default path to the Duniter database
	duniBaseDef = "$HOME/.config/duniter/duniter_default/wotwizard-export.db"
	
	// Directory of the WW database in the resource directory
	systemDef = "System"
	// Working directory
	workDef = "Work"
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
	// Name of the JSON file containing the current block number
	updtStatusName = "Status.json"
	// Name of the stopping file
	stopName = "Stop.txt"
	// Name of the directory where queries must be written
	queryName = "Query"
	// Name of the directory where answers are written
	jsonName = "Json"
	
	blockchainName = "Blockchain"

)

var (
	
	 // Path to the Duniter synchronization file
	duniSync = F.Join(BA.DuniDir, syncName)
	
	// Directory of the WW database and names of files inside it
	system = F.Join(BA.RsrcDir(), systemDef)
	work = F.Join(BA.RsrcDir(), workDef)
	dPars = F.Join(system, dParsName)
	dBase = F.Join(system, dBaseName)
	dCopy =F.Join( system, dCopyName)
	dCopy1 =F.Join( system, dCopy1Name)
	sBase = F.Join(system, sBaseName)
	stop = F.Join(work, stopName)
	qdir = F.Join(work, queryName)
	jdir = F.Join(work, jsonName)
	status = F.Join(jdir, updtStatusName)
	
	addDelayInt int64 = addDelay.Nanoseconds() / 1000000

)

const (
	
	// Max length of a Pubkey
	PubkeyLen = 44
	
	// Number of last blocks to be read again at every update, since they could have changed
	secureGap = 100
	
	// Number of pages used by UtilBTree
	pageNb = 2000
	
	// Numbers of the places of the indexes in dBase
	timePlace = iota // Index timeT
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
	
	action struct {
		next *action
		Actioner
	}
	
	// Queue of actions to do in Commands
	actionQueue struct {
		end *action
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
	
	StringArr = []string
	
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
	
	// Chained list of the operations to be undone before every update
	undoListT struct {
		dataType byte
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
		dataType byte
		bnb int32
		mTime,
		time int64
	}
	
	// Factory of timeTy
	timeFacT struct {
	}
	
	joinAndLeaveL struct {
		dataType byte
		next B.FilePos
		joiningBlock, // Block numbers
		leavingBlock int32
	}
	
	// Factory of joinAndLeaveL
	joinAndLeaveLFacT struct {
	}
	
	joinAndLeave struct {
		dataType byte
		pubkey Pubkey
		list B.FilePos // JoinAndLeaveL
	}
	
	// Factory of joinAndLeave
	joinAndLeaveFacT struct {
	}
	
	identity struct {
		dataType byte
		pubkey Pubkey
		uid string
		member bool
		hash Hash
		block_number int32 // Where the identity is written
		application, // Date of last membership application (joiners, actives, leavers)
		expires_on int64
		certifiers, // Index of all certifiers uid, old or present, of this identity ; B.String -> nothing
		certified B.FilePos // Index of all uid, old or present, certified by this identity ; B.String -> nothing
	}
	
	// Factory of identity
	identityFacT struct {
	}
	
	certification struct {
		dataType byte
		from,
		to Pubkey
		block_number int32 // Where the Certification is written
		expires_on int64
	}
	
	// Factory of certification
	certificationFacT struct {
	}
	
	// int32 index key
	intKey struct {
		dataType byte
		ref int32
	}
	
	// Factory of intKey
	intKeyFacT struct {
	}
	
	// Manager of intKey
	intKeyManT struct {
	}
	
	// B.FilePos index key
	filePosKey struct {
		dataType byte
		ref B.FilePos
	}
	
	// Factory of filePosKey
	filePosKeyFacT struct {
	}
	
	// Pubkey index key
	pubKey struct { // != Pubkey!
		dataType byte
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
		dataType byte
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
		posT *B.IndexReader
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
	
	// Types of Data
	undoListType = iota + 1
	timeType
	joinAndLeaveLType
	joinAndLeaveType
	identityType
	certificationType
	intKeyType
	lIntKeyType
	filePosKeyType
	pubKeyType
	hashKeyType
	
	// Sizes of B.Data
	pubKeyS = (PubkeyLen + 1) * B.BYS
	
	// Sizes of keys
	timeKeyS = B.INS + B.BYS
	idTimeKeyS = B.LIS + B.BYS
	certTimeKeyS = B.LIS + B.BYS

)

var (
	
	// Shared variables
	
	lg = BA.Lg
	
	pars Parameters // Duniter parameters
	
	mutex = new(sync.RWMutex)
	mutexCmds = new(sync.RWMutex)
	
	database *B.Database // duniter0 database
	
	// UtilBTree indexes
	timeT, // intKey -> timeTy
	joinAndLeaveT, // PubKey -> joinAndLeave
	idPubT, // PubKey -> identity
	idUidT, // B.String -> identity
	idHashT, // HashKey -> identity
	idTimeT, // lIntKey -> nothing; addresses of identity sorted by expiration dates
	certFromT, certToT *B.Index // PubKey -> sub-index(PubKey -> certification)
	
	lastBlock int32 = -1 // Last read & updated block
	now, rNow int64 = 0, 0 // Present medianTime and time
	idLenM = 0 // Number of members
	
	// Factories
	timeFac timeFacT
	joinAndLeaveLFac joinAndLeaveLFacT
	joinAndLeaveFac joinAndLeaveFacT
	identityFac identityFacT
	certificationFac certificationFacT
	intKeyFac intKeyFacT
	filePosKeyFac filePosKeyFacT
	uidKeyFac B.StringFac
	pubKeyFac pubKeyFacT
	hashKeyFac hashKeyFacT
	
	// Data managers
	timeMan,
	joinAndLeaveLMan,
	joinAndLeaveMan,
	idMan,
	certMan *B.DataMan
	
	// Key managerers
	pubKeyManer pubKeyManT
	hashKeyManer hashKeyManT
	uidKeyManer uidKeyManT
	intKeyManer intKeyManT
	
	// Key managers
	pubKeyMan = B.MakeKM(pubKeyManer)
	hashKeyMan = B.MakeKM(hashKeyManer)
	uidKeyMan = B.MakeKM(uidKeyManer)
	intKeyMan = B.MakeKM(intKeyManer)
	
	// Update variables
	
	updateListUpdt *updateListUpdtT = nil // Updt
	updateList *updateListT = nil // Cmds // Head of updateListT
	sbFirstUpdt UpdateProc // Cmds
	
	doScan1 = true
	firstUpdate,
	startUpdate bool
	
	// UtilBTree indexes
	certTimeT *B.Index // lIntKey -> nothing; addresses of certification sorted by expiration dates
	
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
	members membersFinder
	memberF = S.TF{Finder: &members}

)

func (m *membersFinder) Less (i, j int) bool {
	return m.m[i].p < m.m[j].p
}

// Data & Data factories procedures

func (t *timeTy) Read (r *B.Reader) {
	t.dataType = r.InByte(); M.Assert(t.dataType == timeType, 100)
	t.bnb = r.InInt32()
	t.mTime = r.InInt64()
	t.time = r.InInt64()
}

func (t *timeTy) Write (w *B.Writer) {
	t.dataType = timeType; w.OutByte(t.dataType)
	w.OutInt32(t.bnb)
	w.OutInt64(t.mTime)
	w.OutInt64(t.time)
}

func (timeFacT) New (size int) B.Data {
	return new(timeTy)
}

func (jlL *joinAndLeaveL) Read (r *B.Reader) {
	jlL.dataType = r.InByte(); M.Assert(jlL.dataType == joinAndLeaveLType, 100)
	jlL.next = r.InFilePos()
	jlL.joiningBlock = r.InInt32()
	jlL.leavingBlock = r.InInt32()
}

func (jlL *joinAndLeaveL) Write (w *B.Writer) {
	jlL.dataType = joinAndLeaveLType; w.OutByte(jlL.dataType)
	w.OutFilePos(jlL.next)
	w.OutInt32(jlL.joiningBlock)
	w.OutInt32(jlL.leavingBlock)
}

func (joinAndLeaveLFacT) New (size int) B.Data {
	return new(joinAndLeaveL)
}

func (jl *joinAndLeave) Read (r *B.Reader) {
	jl.dataType = r.InByte(); M.Assert(jl.dataType == joinAndLeaveType, 100)
	jl.pubkey = Pubkey(r.InString())
	jl.list = r.InFilePos()
}

func (jl *joinAndLeave) Write (w *B.Writer) {
	jl.dataType = joinAndLeaveType; w.OutByte(jl.dataType)
	w.OutString(string(jl.pubkey))
	w.OutFilePos(jl.list)
}

func (joinAndLeaveFacT) New (size int) B.Data {
	return new(joinAndLeave)
}

func (id *identity) Read (r *B.Reader) {
	id.dataType = r.InByte(); M.Assert(id.dataType == identityType, 100)
	id.pubkey = Pubkey(r.InString())
	id.uid = r.InString()
	id.member = r.InBool()
	id.hash = Hash(r.InString())
	id.block_number = r.InInt32()
	id.application = r.InInt64()
	id.expires_on = r.InInt64()
	id.certifiers = r.InFilePos()
	id.certified = r.InFilePos()
}

func (id *identity) Write (w *B.Writer) {
	id.dataType = identityType; w.OutByte(id.dataType)
	w.OutString(string(id.pubkey))
	w.OutString(id.uid)
	w.OutBool(id.member)
	w.OutString(string(id.hash))
	w.OutInt32(id.block_number)
	w.OutInt64(id.application)
	w.OutInt64(id.expires_on)
	w.OutFilePos(id.certifiers)
	w.OutFilePos(id.certified)
}

func (identityFacT) New (size int) B.Data {
	return new(identity)
}

func (c *certification) Read (r *B.Reader) {
	c.dataType = r.InByte(); M.Assert(c.dataType == certificationType, 100)
	c.from = Pubkey(r.InString())
	c.to = Pubkey(r.InString())
	c.block_number = r.InInt32()
	c.expires_on = r.InInt64()
}

func (c *certification) Write (w *B.Writer) {
	c.dataType = certificationType; w.OutByte(c.dataType)
	w.OutString(string(c.from))
	w.OutString(string(c.to))
	w.OutInt32(c.block_number)
	w.OutInt64(c.expires_on)
}

func (certificationFacT) New (size int) B.Data {
	return new(certification)
}

func (l *undoListT) Read (r *B.Reader) {
	l.dataType = r.InByte(); M.Assert(l.dataType == undoListType, 100)
	l.next = r.InFilePos()
	l.typ = r.InByte()
	l.ref = r.InFilePos()
	l.aux = r.InInt64()
}

func (l *undoListT) Write (w *B.Writer) {
	l.dataType = undoListType; w.OutByte(l.dataType)
	w.OutFilePos(l.next)
	w.OutByte(l.typ)
	w.OutFilePos(l.ref)
	w.OutInt64(l.aux)
}

func (undoListFacT) New (size int) B.Data {
	return new(undoListT)
}

func (i *intKey) Read (r *B.Reader) {
	i.dataType = r.InByte(); M.Assert(i.dataType == intKeyType, 100)
	i.ref = r.InInt32()
}

func (i *intKey) Write (w *B.Writer) {
	i.dataType = intKeyType; w.OutByte(i.dataType)
	w.OutInt32(i.ref)
}

func (intKeyFacT) New (size int) B.Data {
	M.Assert(size == 0 || size == B.BYS + B.INS, 20)
	if size == 0 {
		return nil
	}
	return new(intKey)
}

func (i *filePosKey) Read (r *B.Reader) {
	i.dataType = r.InByte(); M.Assert(i.dataType == filePosKeyType, 100)
	i.ref = r.InFilePos()
}

func (i *filePosKey) Write (w *B.Writer) {
	i.dataType = filePosKeyType; w.OutByte(i.dataType)
	w.OutFilePos(i.ref)
}

func (filePosKeyFacT) New (size int) B.Data {
	M.Assert(size == 0 || size == B.BYS + B.LIS, 20)
	if size == 0 {
		return nil
	}
	return new(filePosKey)
}

func (pub *pubKey) Read (r *B.Reader) {
	pub.dataType = r.InByte(); M.Assert(pub.dataType == pubKeyType, 100)
	pub.ref = Pubkey(r.InString())
}

func (pub *pubKey) Write (w *B.Writer) {
	pub.dataType = pubKeyType; w.OutByte(pub.dataType)
	w.OutString(string(pub.ref))
}

func (pubKeyFacT) New (size int) B.Data {
	return new(pubKey)
}

func (hash *hashKey) Read (r *B.Reader) {
	hash.dataType = r.InByte(); M.Assert(hash.dataType == hashKeyType, 100)
	hash.ref = Hash(r.InString())
}

func (hash *hashKey) Write (w *B.Writer) {
	hash.dataType = hashKeyType; w.OutByte(hash.dataType)
	w.OutString(string(hash.ref))
}

func (hashKeyFacT) New (size int) B.Data {
	return new(hashKey)
}

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
}

func (intKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
}

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
}

func (pubKeyManT) CompP (p1, p2 B.Data) B.Comp {
	return pkmCompP(p1, p2)
}

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
}

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
}

func (hashKeyManT) CompP (h1, h2 B.Data) B.Comp {
	return hkmCompP(h1, h2)
}

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
}

// Comparison method of two Strings. Use the lexical order.
func (uidKeyManT) CompP (key1, key2 B.Data) B.Comp {
	M.Assert(key1 != nil && key2 != nil, 20)
	k1 := key1.(*B.String); k2 := key2.(*B.String)
	return BA.CompP(k1.C, k2.C)
}

func (m uidKeyManT) PrefP (p1 B.Data, p2 *B.Data) {
	*p2 = B.StringPrefP(p1.(*B.String), (*p2).(*B.String), m.CompP)
}

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
}

func (idKTimeManT) PrefP (p1 B.Data, p2 *B.Data) {
}

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
}

func (certKTimeManT) PrefP (p1 B.Data, p2 *B.Data) {
}

func findMemberNum (p Pubkey) (int, bool) {
	n := members.len
	members.m[n].p = p
	memberF.BinSearch(0, n - 1, &n)
	return n, n < members.len
}

func Driver () string {
	return driver
}

func System () string {
	return system
}

func Work () string {
	return work
}

func Qdir () string {
	return qdir
}

func Jdir () string {
	return jdir
}

func SBase () string {
	return sBase
}

func Pars () *Parameters {
	p := new(Parameters)
	*p = pars
	return p
}

// Cmds
func (q *actionQueue) isEmpty () bool {
	return q.end == nil
}

// Cmds
func (q *actionQueue) put (a Actioner) {
	if q.end == nil {
		q.end = &action{Actioner: a}
		q.end.next = q.end
	} else {
		p := &action{Actioner: a, next: q.end.next}
		q.end.next = p
		q.end = p
	}
}

// Cmds
func (q *actionQueue) get () Actioner {
	M.Assert(q.end != nil, 100)
	p := q.end.next
	q.end.next = p.next
	if q.end == p {
		q.end = nil
	}
	return p.Actioner
}

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
}

// Last read & updated block
func LastBlock () int32 {
	return lastBlock
}

// medianTime
func Now () int64 {
	return now
}

// time
func RealNow () int64 {
	return rNow
}

// Open the duniter0 database
func openB () {
	M.Assert(database == nil, 101)
	B.Fac.CloseBase(dBase)
	database = B.Fac.OpenBase(dBase, pageNb)
	if database == nil {
		b := B.Fac.CreateBase(dBase, placeNb); M.Assert(b, 102)
		database = B.Fac.OpenBase(dBase, pageNb); M.Assert(database != nil, 103)
		database.WritePlace(timePlace, int64(database.CreateIndex(timeKeyS)))
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
		lg.Println("\"", dBaseName, "\" created")
	}
	timeMan = database.CreateDataMan(timeFac)
	joinAndLeaveLMan = database.CreateDataMan(joinAndLeaveLFac)
	joinAndLeaveMan = database.CreateDataMan(joinAndLeaveFac)
	idMan = database.CreateDataMan(identityFac)
	certMan = database.CreateDataMan(certificationFac)
	undoListMan = database.CreateDataMan(undoListFac)
	timeT = database.OpenIndex(B.FilePos(database.ReadPlace(timePlace)), intKeyMan, intKeyFac)
	joinAndLeaveT = database.OpenIndex(B.FilePos(database.ReadPlace(joinAndLeavePlace)), pubKeyMan, pubKeyFac)
	idPubT = database.OpenIndex(B.FilePos(database.ReadPlace(idPubPlace)), pubKeyMan, pubKeyFac)
	idUidT = database.OpenIndex(B.FilePos(database.ReadPlace(idUidPlace)), uidKeyMan, uidKeyFac)
	idHashT = database.OpenIndex(B.FilePos(database.ReadPlace(idHashPlace)), hashKeyMan, hashKeyFac)
	idTimeT = database.OpenIndex(B.FilePos(database.ReadPlace(idTimePlace)), idKTimeMan, filePosKeyFac)
	certFromT = database.OpenIndex(B.FilePos(database.ReadPlace(certFromPlace)), pubKeyMan, pubKeyFac)
	certToT = database.OpenIndex(B.FilePos(database.ReadPlace(certToPlace)), pubKeyMan, pubKeyFac)
	certTimeT = database.OpenIndex(B.FilePos(database.ReadPlace(certTimePlace)), certKTimeMan, filePosKeyFac)
	lg.Println("\"", dBaseName, "\" opened")
}

// Close the duniter0 database
func closeB () {
	M.Assert(database != nil, 100)
	lg.Println("Closing \"", dBaseName, "\"")
	database.CloseBase()
	database = nil
}

// Block number -> times
func TimeOf (bnb int32) (mTime, time int64, ok bool) {
	ir := timeT.NewReader()
	ok = ir.Search(&intKey{ref: bnb})
	if ok {
		t := timeMan.ReadData(ir.ReadValue()).(*timeTy)
		mTime = t.mTime
		time = t.time
	}
	return
}

// Pubkey -> joining and leaving blocks (leavingBlock == HasNotLeaved if no leaving block)
func JLPub (pubkey Pubkey) (list B.FilePos, ok bool) {
	ir := joinAndLeaveT.NewReader()
	ok = ir.Search(&pubKey{ref: pubkey})
	if ok {
		list = joinAndLeaveMan.ReadData(ir.ReadValue()).(*joinAndLeave).list
	}
	return
}

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
}

// Number of joinAndLeave
func JLLen () int {
	return joinAndLeaveT.NumberOfKeys()
}

// Browse all joinAndLeave's pubkeys step by step
func JLNextPubkey (first bool, ir **B.IndexReader) (pubkey Pubkey, ok bool) {
	if first {
		*ir = joinAndLeaveT.NewReader()
	}
	r := *ir
	r.Next()
	ok = r.PosSet()
	if ok {
		pubkey = r.CurrentKey().(*pubKey).ref
	}
	return
}

// Pubkey -> uid
func IdPub (pubkey Pubkey) (uid string, ok bool) {
	ir := idPubT.NewReader()
	ok = ir.Search(&pubKey{ref: pubkey})
	if ok {
		uid = idMan.ReadData(ir.ReadValue()).(*identity).uid
	}
	return
}

// Pubkey -> uid of member
func IdPubM (pubkey Pubkey) (uid string, ok bool) {
	ir := idPubT.NewReader()
	ok = ir.Search(&pubKey{ref: pubkey})
	if  ok{
		id := idMan.ReadData(ir.ReadValue()).(*identity)
		ok = id.member
		if ok {
			uid = id.uid
		}
	}
	return
}

// Pubkey -> identity
func IdPubComplete (pubkey Pubkey) (uid string, member bool, hash Hash, block_number int32, application, expires_on int64, ok bool) {
	ir := idPubT.NewReader()
	ok = ir.Search(&pubKey{ref: pubkey})
	if ok {
		id := idMan.ReadData(ir.ReadValue()).(*identity)
		uid = id.uid
		member = id.member
		hash = id.hash
		block_number = id.block_number
		application = id.application
		expires_on = id.expires_on
	}
	return
}

// uid -> identity
func IdUid (uid string) (pubkey Pubkey, ok bool) {
	ir := idUidT.NewReader()
	ok = ir.Search(&B.String{C: uid})
	if ok {
		pubkey = idMan.ReadData(ir.ReadValue()).(*identity).pubkey
	}
	return
}

// uid -> Pubkey of member
func IdUidM (uid string) (pubkey Pubkey, ok bool) {
	ir := idUidT.NewReader()
	ok = ir.Search(&B.String{C: uid})
	if ok {
		id := idMan.ReadData(ir.ReadValue()).(*identity)
		ok = id.member
		if ok {
			pubkey = id.pubkey
		}
	}
	return
}

// uid -> identity
func IdUidComplete (uid string) (pubkey Pubkey, member bool, hash Hash, block_number int32, application, expires_on int64, ok bool) {
	ir := idUidT.NewReader()
	ok = ir.Search(&B.String{C: uid})
	if ok {
		id := idMan.ReadData(ir.ReadValue()).(*identity)
		pubkey = id.pubkey
		member = id.member
		hash = id.hash
		block_number = id.block_number
		application = id.application
		expires_on = id.expires_on
	}
	return
}

// Hash -> pubkey
func IdHash (hash Hash) (pub Pubkey, ok bool) {
	ir := idHashT.NewReader()
	ok = ir.Search(&hashKey{ref: hash})
	if ok {
		pub = idMan.ReadData(ir.ReadValue()).(*identity).pubkey
	}
	return
}

// Number of identities
func IdLen () int {
	return idUidT.NumberOfKeys()
}

// Number of members
func IdLenM () int {
	return idLenM
}

// Position next identity's pubkey for IdNextPubkey
func IdPosPubkey (pubkey Pubkey) *B.IndexReader {
	ir := idPubT.NewReader()
	_ = ir.Search(&pubKey{ref: pubkey})
	ir.Previous()
	return ir
}

// Browse all identity's pubkeys step by step
func IdNextPubkey (first bool, ir **B.IndexReader) (pubkey Pubkey, ok bool) {
	if first {
		*ir = idPubT.NewReader()
	}
	r := *ir
	r.Next()
	ok = r.PosSet()
	if ok {
		pubkey = r.CurrentKey().(*pubKey).ref
	}
	return
}

// Browse all members' pubkeys step by step
func IdNextPubkeyM (first bool, ir **B.IndexReader) (pubkey Pubkey, ok bool) {
	if first {
		*ir = idPubT.NewReader()
	}
	r := *ir
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
}

// Position next identity's uid for IdNextUid
func IdPosUid (uid string) *B.IndexReader {
	ir := idUidT.NewReader()
	ir.Search(&B.String{C: uid})
	ir.Previous()
	return ir
}

// Browse all identity's uid(s) lexicographically step by step
func IdNextUid (first bool, ir **B.IndexReader) (uid string, ok bool) {
	if first {
		*ir = idUidT.NewReader()
	}
	r := *ir
	r.Next()
	ok = r.PosSet()
	if ok {
		uid = r.CurrentKey().(*B.String).C
	}
	return
}

// Browse all members' uid(s) lexicographically step by step
func IdNextUidM (first bool, ir **B.IndexReader) (uid string, ok bool) {
	if first {
		*ir = idUidT.NewReader()
	}
	r := *ir
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
}

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
}

// Pubkey -> head of sub-index
func CertFrom (from Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	ir := certFromT.NewReader()
	ok = ir.Search(&pubKey{ref: from})
	if ok {
		pos.posT = database.OpenIndex(ir.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
}

// Pubkey -> head of sub-index
func CertTo (to Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	ir := certToT.NewReader()
	ok = ir.Search(&pubKey{ref: to})
	if ok {
		pos.posT = database.OpenIndex(ir.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
}

// Number of keys in sub-index
func (pos *CertPos) CertPosLen () int {
	M.Assert(pos != nil, 20)
	if pos.posT == nil {
		return 0
	}
	return pos.posT.Ind().NumberOfKeys()
}

// Browse all certification's pairs of Pubkey in a sub-index step by step
func (pos *CertPos) CertNextPos () (from, to Pubkey, ok bool) {
	ok = pos.posT != nil
	if ok {
		ir := pos.posT
		ir.Next()
		ok = ir.PosSet()
		if ok {
			c := certMan.ReadData(ir.ReadValue()).(*certification)
			from = c.from
			to = c.to
		}
	}
	return
}

// Browse all sub-indexes step by step in the lexicographic order of the from Pubkey
func CertNextFrom (first bool, pos *CertPos, ir **B.IndexReader) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(ir != nil, 21)
	if first {
		*ir = certFromT.NewReader()
	}
	r := *ir
	M.Assert(r != nil, 22)
	r.Next()
	ok = r.PosSet()
	if ok {
		pos.posT = database.OpenIndex(r.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
}

// Browse all sub-indexes step by step in the lexicographic order of the to Pubkey
func CertNextTo (first bool, pos *CertPos, ir **B.IndexReader) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(ir != nil, 21)
	if first {
		*ir = certToT.NewReader()
	}
	r := *ir
	M.Assert(r != nil, 22)
	r.Next()
	ok = r.PosSet()
	if ok {
		pos.posT = database.OpenIndex(r.ReadValue(), pubKeyMan, pubKeyFac).NewReader()
	} else {
		pos.posT = nil
	}
	return
}

func AllCertifiers (to string) StringArr {
	ir := idUidT.NewReader()
	if !ir.Search(&B.String{C: to}) {
		return nil
	}
	id := idMan.ReadData(ir.ReadValue()).(*identity)
	if id.certifiers == B.BNil {
		return nil
	}
	ind := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac)
	from := make(StringArr, ind.NumberOfKeys())
	ir = ind.NewReader()
	ir.Next()
	i := 0
	for ir.PosSet() {
		from[i] = ir.CurrentKey().(*B.String).C
		i++
		ir.Next()
	}
	M.Assert(i == len(from), 60)
	return from
}

func AllCertified (from string) StringArr {
	ir := idUidT.NewReader()
	if !ir.Search(&B.String{C: from}) {
		return nil
	}
	id := idMan.ReadData(ir.ReadValue()).(*identity)
	if id.certified == B.BNil {
		return nil
	}
	ind := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac)
	to := make(StringArr, ind.NumberOfKeys())
	ir = ind.NewReader()
	ir.Next()
	i := 0
	for ir.PosSet() {
		to[i] = ir.CurrentKey().(*B.String).C
		i++
		ir.Next()
	}
	M.Assert(i == len(to), 60)
	return to
}

func IsSentry (pubkey Pubkey) bool {
	e, ok := findMemberNum(pubkey)
	return ok && sentriesS.In(e)
}

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
}

// Return the number of sentries
func SentriesLen () int {
	return sentriesS.NbElems()
}

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
	}

	find := func (poSE *poSET) (set_1, set_2 U.Set, poS float64, ok bool) {
		el, ok, _ := poST.Search(poSE)
		if ok {
			p := el.Val().(*poSET)
			set_1 = p.set_1
			set_2 = p.set_2
			poS = p.poS
		}
		return
	}

	store := func (poSE *poSET, set_1, set_2 U.Set, poS float64) {
		poSE.set_1 = set_1
		poSE.set_2 = set_2
		poSE.poS = poS
		_, b, _ := poST.SearchIns(poSE); M.Assert(!b, 100)
	}
	
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
		poS = float64(set_1.NbElems()) / float64(sentriesS.NbElems())
		store(poSE, set_1, set_2, poS)
	}
	return
}

// Array of certifiers' pubkeys -> % of sentries reached in pars.stepMax - 1 steps
func PercentOfSentries (pubkeys PubkeysT) float64 {
	_, _, poS := PercentOfSentriesS(pubkeys)
	return poS
}

// Verify the distance rule for a set of certifiers' pubkeys
func DistanceRuleOk (pubkeys PubkeysT) bool {
	return PercentOfSentries(pubkeys) >= pars.Xpercent
}

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
}

// Updt
// Skip the string s from position i to the position of stop excluded; update i
func skipS (s []rune, stop rune, i *int) {
	for *i < len(s) && s[*i] != stop {
		*i++
	}
	*i++
}

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
}

// Cmds
// Extract Duniter parameters from JSON file
func params () {
	ok := J.ReadFile(dPars, &pars); M.Assert(ok, 100)
}

// Updt
// Add a block in timeT
func times (withList bool, bnb int, mTime, time int64) {
	t := &timeTy{bnb: int32(bnb), mTime: mTime, time: time}
	tRef := timeMan.WriteAllocateData(t)
	iw := timeT.Writer()
	b := iw.SearchIns(&intKey{ref: t.bnb}); M.Assert(!b, bnb, 100)
	iw.WriteValue(tRef)
	if withList {
		tL := &undoListT{next: undoList, typ: timeList, ref: tRef, aux: 0}
		undoList = undoListMan.WriteAllocateData(tL)
	}
	now = M.Max64(now, mTime)
	rNow = M.Max64(rNow, time)
}

// Updt
func removeCertifiersCertified (withList bool, idRef B.FilePos, id *identity) {
	idU1 := &B.String{C: id.uid}
	
	if id.certifiers != B.BNil {
		ir := idUidT.NewReader()
		c1 := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).NewReader()
		c1.Next()
		for c1.PosSet() {
			idU2 := c1.CurrentKey().(*B.String)
			b := ir.Search(idU2); M.Assert(b, idU2.C, 100)
			id2Ref := ir.ReadValue()
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
		ir := idUidT.NewReader()
		c1 := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).NewReader()
		c1.Next()
		for c1.PosSet() {
			idU2 := c1.CurrentKey().(*B.String)
			b := ir.Search(idU2); M.Assert(b, idU2.C, 103)
			id2Ref := ir.ReadValue()
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
}

// Updt
func revokeId (withList bool, p Pubkey) {
	ir := idPubT.NewReader()
	b := ir.Search(&pubKey{ref: p}); M.Assert(b, p, 100)
	idRef := ir.ReadValue()
	id := idMan.ReadData(idRef).(*identity)
	if withList {
		idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: id.application}
		undoList = undoListMan.WriteAllocateData(idL)
	}
	id.expires_on = BA.Revoked
	removeCertifiersCertified(withList, idRef, id)
	idMan.WriteData(idRef, id)
}

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
		id.application, _, b = TimeOf(int32(nb)); M.Assert(b, nb, 101)
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
			idMan.WriteData(idRef, id)
			b = iwU.SearchIns(idU); M.Assert(b, idU.C, 107)
			M.Assert(iwU.ReadValue() == idRef, 108)
			b = iwH.SearchIns(idH); M.Assert(b, idH.ref, 109)
			M.Assert(iwH.ReadValue() == idRef, 110)
			if withList {
				idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: oldId.expires_on, aux2: oldId.application}
				undoList = undoListMan.WriteAllocateData(idL)
			}
		} else {
			id.certifiers = B.BNil; id.certified = B.BNil
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
			idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: id.application}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		id.application, _, b = TimeOf(int32(nb)); M.Assert(b, nb, 116)
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
		skipS(ss, '"', &i)
		if ss[i] != ']' {
			i++
		}
		b = iwP.Search(idP); M.Assert(b, idP.ref, 118)
		idRef := iwP.ReadValue()
		id := idMan.ReadData(idRef).(*identity)
		if withList {
			idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: id.application}
			undoList = undoListMan.WriteAllocateData(idL)
		}
		id.application, _, b = TimeOf(int32(nb)); M.Assert(b, nb, 119)
		id.expires_on = - M.Abs64(id.expires_on)
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
				idL := &undoListT{next: undoList, typ: activeList, ref: idRef, aux: id.expires_on, aux2: id.application}
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
}

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
		idP := &pubKey{ref: c.from}
		var v B.FilePos
		if iwF.SearchIns(idP) {
			v = iwF.ReadValue()
		} else {
			v = database.CreateIndex(pubKeyS)
			iwF.WriteValue(v)
		}
		idP.ref = c.to
		iw := database.OpenIndex(v, pubKeyMan, pubKeyFac).Writer()
		var oldPC B.FilePos
		if iw.SearchIns(idP) {
			oldPC = iw.ReadValue()
		} else {
			oldPC = B.BNil
		}
		iw.WriteValue(pC)
		idP.ref = c.to
		if iwT.SearchIns(idP) {
			v = iwT.ReadValue()
		} else {
			v = database.CreateIndex(pubKeyS)
			iwT.WriteValue(v)
		}
		idP.ref = c.from
		iw = database.OpenIndex(v, pubKeyMan, pubKeyFac).Writer()
		iw.SearchIns(idP)
		iw.WriteValue(pC)
		if oldPC == B.BNil {
			idU := new(B.String)
			idP.ref = c.from
			b = iwP.Search(idP); M.Assert(b, idP.ref, 102)
			idRef := iwP.ReadValue()
			id := idMan.ReadData(idRef).(*identity)
			idU.C, b = IdPub(c.to); M.Assert(b, c.to, 103)
			if id.certified == B.BNil {
				id.certified = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).Writer()
			iw.SearchIns(idU)
			idP.ref = c.to
			b = iwP.Search(idP); M.Assert(b, idP.ref, 104)
			idRef = iwP.ReadValue()
			id = idMan.ReadData(idRef).(*identity)
			idU.C, b = IdPub(c.from); M.Assert(b, c.from, 105)
			if id.certifiers == B.BNil {
				id.certifiers = database.CreateIndex(0)
				idMan.WriteData(idRef, id)
			}
			iw = database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).Writer()
			iw.SearchIns(idU)
		} else {
			b = iwTi.Erase(&filePosKey{ref: oldPC}); M.Assert(b, 106)
		}
		b = iwTi.SearchIns(&filePosKey{ref: pC}); M.Assert(!b, 107)
		if withList {
			cL := &undoListT{next: undoList, typ: certAddList, ref: pC, aux: int64(oldPC)}
			undoList = undoListMan.WriteAllocateData(cL)
		} else if oldPC != B.BNil {
			certMan.EraseData(oldPC)
		}
	}
}

// Updt
// Remove c keys from certFromT and certToT
func removeCert (c *certification) {
	pKFrom := &pubKey{ref: c.from}
	pKTo := &pubKey{ref: c.to}
	
	iw1 := certFromT.Writer()
	b := iw1.Search(pKFrom); M.Assert(b, pKFrom.ref, 100)
	n := iw1.ReadValue()
	ind := database.OpenIndex(n, pubKeyMan, pubKeyFac)
	iw2 := ind.Writer()
	b = iw2.Erase(pKTo); M.Assert(b, pKTo.ref, 101)
	if ind.IsEmpty() {
		database.DeleteIndex(n)
		b = iw1.Erase(pKFrom); M.Assert(b, pKFrom.ref, 102)
	}
	
	iw1 = certToT.Writer()
	b = iw1.Search(pKTo); M.Assert(b, pKTo.ref, 103)
	n = iw1.ReadValue()
	ind = database.OpenIndex(n, pubKeyMan, pubKeyFac)
	iw2 = ind.Writer()
	b = iw2.Erase(pKFrom); M.Assert(b, pKFrom.ref, 104)
	if ind.IsEmpty() {
		database.DeleteIndex(n)
		b = iw1.Erase(pKTo); M.Assert(b, pKTo.ref, 105)
	}
}

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
		removeCert(c)
		withList := c.expires_on >= secureNow
		b := iw.Erase(&filePosKey{ref: pC}); M.Assert(b, 100)
		if withList {
			cL := &undoListT{next: undoList, typ: certRemoveList, ref: pC, aux: 0}
			undoList = undoListMan.WriteAllocateData(cL)
		} else {
			certMan.EraseData(pC)
		}
	}
}

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
}

// Updt
// Undo the last operations done from the secureGap last blocks
func removeSecureGap () {
	for undoList != B.BNil {
		l := undoListMan.ReadData(undoList).(*undoListT)
		switch l.typ {
			case timeList: {
				// Erase the timeTy data pointed by l.ref and the corresponding key in timeT
				t := timeMan.ReadData(l.ref).(*timeTy)
				b := timeT.Writer().Erase(&intKey{ref: t.bnb}); M.Assert(b, t.bnb, 100)
				timeMan.EraseData(l.ref)
			}
			case idAddList: {
				// Erase the identity data pointed by l.ref and the corresponding keys in idPubT, idHashT and idUidT
				id := idMan.ReadData(l.ref).(*identity)
				b := idPubT.Writer().Erase(&pubKey{ref: id.pubkey}); M.Assert(b, id.pubkey, 101)
				b = idUidT.Writer().Erase(&B.String{C: id.uid}); M.Assert(b, id.uid, 102)
				b = idHashT.Writer().Erase(&hashKey{ref: id.hash}); M.Assert(b, id.hash, 103)
				idMan.EraseData(l.ref)
			}
			case joinList: {
				// Let the identity no more be member; erase the last joinAndLeaveL data corresponding to l.ref; if this is also the first data, erase the corresponding joinAndLeave data and its key in joinAndLeaveT
				id := idMan.ReadData(l.ref).(*identity)
				id.member = false
				idMan.WriteData(l.ref, id)
				idLenM--
				p := &pubKey{ref: id.pubkey}
				iw := joinAndLeaveT.Writer()
				b := iw.Search(p); M.Assert(b, 104)
				jlRef := iw.ReadValue()
				jl := joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
				jlLRef := jl.list
				jlL := joinAndLeaveLMan.ReadData(jlLRef).(*joinAndLeaveL)
				M.Assert(jlL.leavingBlock == HasNotLeaved, 105)
				if jlL.next == B.BNil {
					joinAndLeaveMan.EraseData(jlRef)
					b = iw.Erase(p); M.Assert(b, 106)
				} else {
					jl.list = jlL.next
					joinAndLeaveMan.WriteData(jlRef, jl)
				}
				joinAndLeaveLMan.EraseData(jlLRef)
			}
			case activeList: {
				// Undo the identity.expires_on update
				id := idMan.ReadData(l.ref).(*identity)
				id.expires_on = l.aux
				id.application = l.aux2
				idMan.WriteData(l.ref, id)
			}
			case leaveList: {
				// Update the last joinAndLeaveL data corresponding to l.ref
				id := idMan.ReadData(l.ref).(*identity)
				id.member = true
				idMan.WriteData(l.ref, id)
				idLenM++
				ir := joinAndLeaveT.NewReader()
				b := ir.Search(&pubKey{ref: id.pubkey}); M.Assert(b, 107)
				jlRef := ir.ReadValue()
				jl := joinAndLeaveMan.ReadData(jlRef).(*joinAndLeave)
				jlL := joinAndLeaveLMan.ReadData(jl.list).(*joinAndLeaveL)
				M.Assert(jlL.leavingBlock != HasNotLeaved, 108)
				jlL.leavingBlock = HasNotLeaved
				joinAndLeaveLMan.WriteData(jl.list, jlL)
			}
			case idAddTimeList: {
				b := idTimeT.Writer().Erase(&filePosKey{ref: l.ref}); M.Assert(b, 109)
			}
			case idRemoveTimeList: {
				b := idTimeT.Writer().SearchIns(&filePosKey{ref: l.ref}); M.Assert(!b, 110)
			}
			case certAddList: {
				// Erase the keys corresponding to the certification pointed by l.ref in certFromT and certToT, or, if l.aux # B.BNil, update them; modify identity. certifiers and identity.certified as needed
				c := certMan.ReadData(l.ref).(*certification)
				iwT := certTimeT.Writer()
				b := iwT.Erase(&filePosKey{ref: l.ref}); M.Assert(b, 111)
				ir := idPubT.NewReader()
				p := new(pubKey)
				if B.FilePos(l.aux) == B.BNil {
					removeCert(c)
					u := new(B.String)
					p.ref = c.from
					b = ir.Search(p); M.Assert(b, 112)
					idRef := ir.ReadValue()
					id := idMan.ReadData(idRef).(*identity)
					u.C, b = IdPub(c.to); M.Assert(b, 113)
					ind := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac)
					iw := ind.Writer()
					b = iw.Erase(u); M.Assert(b, 114)
					if ind.IsEmpty() {
						database.DeleteIndex(id.certified); id.certified = B.BNil
						idMan.WriteData(idRef, id)
					}
					p.ref = c.to
					b = ir.Search(p); M.Assert(b, 115)
					idRef = ir.ReadValue()
					id = idMan.ReadData(idRef).(*identity)
					u.C, b = IdPub(c.from); M.Assert(b, 116)
					ind = database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac)
					iw = ind.Writer()
					b = iw.Erase(u); M.Assert(b, 117)
					if ind.IsEmpty() {
						database.DeleteIndex(id.certifiers); id.certifiers = B.BNil
						idMan.WriteData(idRef, id)
					}
				}else {
					p.ref = c.from
					irF := certFromT.NewReader()
					b = irF.Search(p); M.Assert(b, 118)
					n := irF.ReadValue()
					iw := database.OpenIndex(n, pubKeyMan, pubKeyFac).Writer()
					p.ref = c.to
					b = iw.Search(p); M.Assert(b, 119)
					iw.WriteValue(B.FilePos(l.aux))
					p.ref = c.to
					irT := certToT.NewReader()
					b = irT.Search(p); M.Assert(b, 120)
					n = irT.ReadValue()
					iw = database.OpenIndex(n, pubKeyMan, pubKeyFac).Writer()
					p.ref = c.from
					b = iw.Search(p); M.Assert(b, 121)
					iw.WriteValue(B.FilePos(l.aux))
					b := iwT.SearchIns(&filePosKey{ref: B.FilePos(l.aux)}); M.Assert(!b, 122)
				}
				certMan.EraseData(l.ref)
			}
			case certRemoveList: {
				// Insert the keys corresponding to the certification pointed by l.ref into certFromT, certToT and certTimeT
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
				b := iw.SearchIns(p); M.Assert(!b, 123)
				iw.WriteValue(l.ref)
				iwT := certToT.Writer()
				if iwT.SearchIns(p) {
					n = iwT.ReadValue()
				} else {
					n = database.CreateIndex(pubKeyS)
					iwT.WriteValue(n)
				}
				p.ref = c.from
				iw = database.OpenIndex(n, pubKeyMan, pubKeyFac).Writer()
				b = iw.SearchIns(p); M.Assert(!b, 124)
				iw.WriteValue(l.ref)
				b = certTimeT.Writer().SearchIns(&filePosKey{ref: l.ref}); M.Assert(!b, 125)
			}
			case remCertifiers: {
				id := idMan.ReadData(B.FilePos(l.aux)).(*identity)
				u := &B.String{C: id.uid}
				id = idMan.ReadData(l.ref).(*identity)
				if id.certifiers == B.BNil {
					id.certifiers = database.CreateIndex(0)
					idMan.WriteData(l.ref, id)
				}
				iw := database.OpenIndex(id.certifiers, uidKeyMan, uidKeyFac).Writer()
				b := iw.SearchIns(u); M.Assert(!b, 126)
			}
			case remCertified: {
				id := idMan.ReadData(B.FilePos(l.aux)).(*identity)
				u := &B.String{C: id.uid}
				id = idMan.ReadData(l.ref).(*identity)
				if id.certified == B.BNil {
					id.certified = database.CreateIndex(0)
					idMan.WriteData(l.ref, id)
				}
				iw := database.OpenIndex(id.certified, uidKeyMan, uidKeyFac).Writer()
				b := iw.SearchIns(u); M.Assert(!b, 127)
			}
		}
		undoListMan.EraseData(undoList)
		undoList = l.next
	}
}

// Cmds
// Initialize members and sentriesS
func calculateSentries (... interface{}) {
	members.len = IdLen()
	members.m = make(membersT, members.len + 1)
	var (ir *B.IndexReader; pos CertPos)
	i := 0
	p, ok := IdNextPubkey(true, &ir)
	for ok {
		M.Assert(i == 0 || p > members.m[i - 1].p, 100)
		members.m[i].p = p
		members.m[i].links = U.NewSet()
		i++
		p, ok = IdNextPubkey(false, &ir)
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
	n := IdLenM()
	if n == 0 {
		return
	}
	n = int(math.Ceil(math.Pow(float64(n), 1 / float64(pars.StepMax))))
	p, ok = IdNextPubkeyM(true, &ir)
	for ok {
		if CertFrom(p, &pos) && pos.CertPosLen() >= n && CertTo(p, &pos) && pos.CertPosLen() >= n {
			e, b := findMemberNum(p); M.Assert(b, 103)
			sentriesS.Incl(e)
		}
		p, ok = IdNextPubkeyM(false, &ir)
	}
	
	poST = A.New()
}

// Updt
// Insert datas from all the blocks from the secureGapth block before the last read
func scanBlocksUpdt (d *Q.DB) {
	lg.Println("Updating \"", dBaseName, "\"")
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
	s := C.Itoa(int(lastBlock - secureGap + 1))
	rs, err := d.Query("SELECT number, medianTime, time, joiners, actives, leavers, revoked, excluded, certifications FROM block WHERE NOT fork AND number >= " + s + " ORDER BY number ASC")
	M.Assert(err == nil, err, 100)
	defer rs.Close()
	var secureNow int64 = M.MaxInt64
	var medianTime int64 = -1
	var n = 0
	if maxN >= secureGap {
		n = maxN - secureGap + 1
	}
	getSecureNow := true
	for rs.Next() {
		var (
			number int
			m,
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
		medianTime = m.Unix()
		time := t.Unix()
		if getSecureNow && number >= n {
			getSecureNow = false
			secureNow = medianTime
		}
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
		identities(withList, joiners, actives, leavers, revoked, excluded, number, d)
		certifications(withList, certificationList, number)
	}
	M.Assert(rs.Err() == nil, rs.Err(), 60)
	if medianTime >= 0 {
		revokeExpiredIds(medianTime, secureNow)
		removeExpiredCerts(medianTime, secureNow) // Élimine toutes les certifications expirées avec réversibilité dans secureGap
	}
	database.WritePlace(undoListPlace, int64(undoList))
	lastBlock = int32(maxN)
	database.WritePlace(lastNPlace, int64(lastBlock))
	database.WritePlace(idLenPlace, int64(idLenM))
	lg.Println("\"", dBaseName, "\" updated")
	lg.Println("Number of members: ", idLenM)
}

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
}

// Updt
// Scan the Duniter database
func scan (... interface{}) {
	lg.Println("Opening Duniter database (bis)")
	d, err := Q.Open(driver, BA.DuniBase)
	M.Assert(err == nil, err, 100)
	defer d.Close()
	scanBlocksUpdt(d)
}

// Updt
// Scan the Duniter parameters in block 0
func scan1 () {
	lg.Println("Opening Duniter database")
	d, err := Q.Open(driver, BA.DuniBase)
	M.Assert(err == nil, err, 100)
	defer d.Close()
	paramsUpdt(d)
}

// Updt
func exportParameters () {
	lg.Println("Exporting money parameters")
	f, err := os.Create(dPars); M.Assert(err == nil, err, 100)
	defer f.Close()
	ok := J.FprintJsonOf(f, &pars); M.Assert(ok, 101)
	lg.Println("Money parameters exported")
}

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
}

// Updt
func readSyncTime () int64 {
	f, err := os.Open(duniSync); M.Assert(err == nil, err, 100)
	defer f.Close()
	var t int64
	_, err = fmt.Fscanf(f, "%d", &t); M.Assert(err == nil, err, 101)
	return t
}

// Updt
func writeSyncTime (t int64) {
	f, err := os.Create(duniSync); M.Assert(err == nil, err, 100)
	defer f.Close()
	_, err = fmt.Fprintf(f, "%d", t); M.Assert(err == nil, err, 101)
}

// Updt
func updateAllUpdt (stopProg <-chan bool, updateReady chan<- bool) {
	for {
		err := os.Remove(duniSync)
		M.Assert(err == nil || os.IsNotExist(err), err, 100)
		lg.Println("\"",  syncName, "\" erased")
		lg.Println("Looking for", duniSync, "\n")
		f, err := os.Open(duniSync)
		for os.IsNotExist(err) {
			select {
			case <-stopProg:
				closeB()
				lg.Println("Halting")
				return
			default:
			}
			time.Sleep(verifyPeriod)
			f, err = os.Open(duniSync)
		}
		M.Assert(err == nil, err, 101)
		f.Close()
		lg.Println("\"", syncName, "\" seen; reading it")
		t1 := time.Now().Add(syncDelay).Add(-verifyPeriod).Add(-secureDelay)
		t0 := readSyncTime()
		var done = make(chan bool)
		go doUpdates(done, updateReady)
		innerLoop:
		for {
			select {
			case <- done:
				break innerLoop
			default:
				if time.Now().After(t1) {
					t0 += addDelayInt
					writeSyncTime(t0)
				}
				time.Sleep(checkPeriod)
			}
		}
	}
}

// Updt
func saveBase () {
	os.Remove(dCopy1)
	os.Rename(dCopy, dCopy1)
	const bufferSize = 0X800
	f, err := os.Open(dBase)
	if err == nil {
		defer f.Close()
		lg.Println("Making a copy of \"", dBaseName, "\"")
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
}

// Cmds
func AddUpdateProc (name string, updateProc UpdateProc, params ... interface{}) {
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
}

// Cmds
func RemoveUpdateProc (name string) {
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
}

// Cmds
func FixSandBoxFUpdt (updateProc UpdateProc) {
	sbFirstUpdt = updateProc
}

// Cmds
func updateAll () {
	l := updateList
	for l != nil {
		l.update(l.params...)
		l = l.next
	}
}

// Cmds
func showUpdated () {
	m := J.NewMaker()
	m.StartObject()
	m.PushInteger(int64(LastBlock()));
	m.BuildField("last_block");
	m.BuildObject()
	f, err := os.Create(status); M.Assert(err == nil, err, 100)
	defer f.Close()
	J.Fprint(f, m.GetJson())
}

// Cmds
func addAction (newAction <-chan Actioner, getAction chan<- Actioner) {
	var actionQ = actionQueue{end: nil}
	for {
		select {
		case a := <-newAction:
			lg.Println("Adding action", a.Name(), "into queue")
			actionQ.put(a)
		default:
			if !actionQ.isEmpty() {
				getAction <- actionQ.get()
			} else {
				time.Sleep(checkPeriod)
			}
		}
	}
}

// Cmds
func updateCmds () {
	lg.Println("Starting update of commands")
	if firstUpdate {
		params()
		sbFirstUpdt()
		firstUpdate = false
	}
	updateAll()
	showUpdated()
	lg.Println("Update of commands done")
}

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
}

// Cmds
func doAction (a Actioner) {
	lg.Println("Starting action", a.Name())
	mutexCmds.RLock()
	mutex.RLock()
	a.Activate()
	mutex.RUnlock()
	mutexCmds.RUnlock()
	lg.Println("Action", a.Name(), "done")
}

// Cmds
func dispatchActions (updateReady <-chan bool, getAction <-chan Actioner) {
	if startUpdate {
		startUpdate = false
		mutexCmds.Lock()
		mutex.RLock()
		updateFirstCmds()
		mutex.RUnlock()
		mutexCmds.Unlock()
	}
	for {
		select {
		case <-updateReady:
			mutexCmds.Lock()
			mutex.RLock()
			updateCmds()
			mutex.RUnlock()
			mutexCmds.Unlock()
		case a := <-getAction:
			if a != nil && !firstUpdate {
				go doAction(a)
			}
		}
	}
}

// Updt
func lookForStop (stopProg chan<- bool) {
	for {
		f, err := os.Open(stop)
		if err == nil {
			f.Close()
			BA.SwitchOff(stop)
			stopProg <- true
			break
		}
		time.Sleep(stopPeriod)
	}
}

func Start (newAction <-chan Actioner) {
	lg.Println("Starting", "\n")
	m := J.NewMaker()
	m.StartObject()
	m.PushInteger(-1);
	m.BuildField("last_block");
	m.BuildObject()
	f, err := os.Create(status); M.Assert(err == nil, err, 100)
	defer f.Close()
	J.Fprint(f, m.GetJson())
	saveBase()
	openB()
	stopProg := make(chan bool)
	updateReady := make(chan bool)
	getAction :=  make(chan Actioner)
	go lookForStop(stopProg)
	go dispatchActions(updateReady, getAction)
	go addAction(newAction, getAction)
	updateAllUpdt(stopProg, updateReady)
}

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
}

func Initialize () {
	AddUpdateProcUpdt(scan)
	firstUpdate = virgin()
	startUpdate = !firstUpdate	
	AddUpdateProc(blockchainName, calculateSentries)
}

func init() {
	os.MkdirAll(system, 0777)
	os.MkdirAll(work, 0777)
	os.MkdirAll(jdir, 0777)
}
