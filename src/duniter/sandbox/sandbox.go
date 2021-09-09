/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package sandbox

// Put the Duniter sandbox in AVL trees to access quickly sandbox data

import (
	
	A	"util/avl"
	B	"duniter/blockchain"
	BA	"duniter/basic"
	J	"util/json"
	M	"util/misc"
	Q	"database/sql"
		"os"
		"strings"
	_	"github.com/mattn/go-sqlite3"

)

type (
	
	Pubkey = B.Pubkey
	Hash = B.Hash
	
	identity struct {
		inBC bool
		hash Hash // Needed in sandbox: key of identities
		pubkey Pubkey
		uid string
		bnb int32
		expires_on int64
	}
	
	idUidE struct { // Sorted by uid
		*identity
	}
	
	idPubE struct { // Sorted by pubkey
		*identity
	}
	
	idHashE struct { // Sorted by hash
		*identity
	}
	
	certification struct {
		from,
		to Pubkey
		toHash Hash
		bnb int32
		expires_on int64
	}
	
	certFromE struct { // Sorted by from
		*certification
		list *A.Tree
	}
	
	certToE struct { // Sorted by to
		*certification
		list *A.Tree
	}
	
	CertPos struct { // Position in a certification subtree
		posT *A.Tree // The subtree
		posCur *A.Elem // The last seen element in the subtree
	}
		
	Identity struct {
		InBC bool
		Hash Hash // Needed in sandbox: key of identities
		Pubkey Pubkey
		Uid string
		Bnb int32
		Expires_on int64
	}
	
	Certification struct {
		From,
		To Pubkey
		ToHash Hash
		Bnb int32
		Expires_on int64
	}

	SandboxData struct {
		Block int
		Date int64
		Identities []Identity
		Certifications []Certification
	}
	
)

var (
	
	sBase = B.SBase()
	
	// AVL trees
	idUidT, // uid -> identity
	idPubT, // pubkey -> identity
	idHashT, // hash -> identity
	certFromT, // from -> certification
	certToT *A.Tree // toHash -> certification

)

// Comparison procedures for AVL trees

func (i1 *idUidE) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*idUidE)
	b := BA.CompP(i1.uid, ii2.uid)
	if b != A.Eq {
		return b
	}
	if i1.hash < ii2.hash {
		return A.Lt
	}
	if i1.hash > ii2.hash {
		return A.Gt
	}
	return A.Eq
} //Compare

func (i1 *idPubE) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*idPubE)
	if i1.pubkey < ii2.pubkey {
		return A.Lt
	}
	if i1.pubkey > ii2.pubkey {
		return A.Gt
	}
	if i1.hash < ii2.hash {
		return A.Lt
	}
	if i1.hash > ii2.hash {
		return A.Gt
	}
	return A.Eq
} //Compare

func (i1 *idHashE) Compare (i2 A.Comparer) A.Comp {
	ii2 := i2.(*idHashE)
	if i1.hash < ii2.hash {
		return A.Lt
	}
	if i1.hash > ii2.hash {
		return A.Gt
	}
	return A.Eq
} //Compare

func (c1 *certFromE) Compare (c2 A.Comparer) A.Comp {
	cc2 := c2.(*certFromE)
	if c1.from < cc2.from {
		return A.Lt
	}
	if c1.from > cc2.from {
		return A.Gt
	}
	return A.Eq
} //Compare

func (c1 *certToE) Compare (c2 A.Comparer) A.Comp {
	cc2 := c2.(*certToE)
	if c1.toHash < cc2.toHash {
		return A.Lt
	}
	if c1.toHash > cc2.toHash {
		return A.Gt
	}
	return A.Eq
} //Compare

// hash -> identity
func idHashId (hash Hash) *identity {
	e, ok, _ := idHashT.Search(&idHashE{&identity{hash: hash}})
	if ok {
		return e.Val().(*idHashE).identity
	}
	return nil
} //idHashId

// hash -> identity
func IdHash (hash Hash) (inBC bool, pubkey Pubkey, uid string, bnb int32, expires_on int64, ok bool) {
	id := idHashId(hash)
	ok = id != nil
	if ok {
		inBC = id.inBC
		pubkey = id.pubkey
		uid = id.uid
		bnb = id.bnb
		expires_on = id.expires_on
	}
	return
} //IdHash

// Number of identities
func IdLen () int {
	return idHashT.NumberOfElems()
} //IdLen

// Position next identity's uid for IdNextUid
func IdPosUid (uid string) *A.Elem {
	pos, _, _ := idUidT.SearchNext(&idUidE{&identity{uid: uid, hash: ""}})
	pos = idUidT.Previous(pos)
	return pos
} //IdPosUid

// Browse all identity's uid(s) lexicographically step by step
func IdNextUid (first bool, pos **A.Elem) (uid string, hash Hash, ok bool) {
	if first {
		*pos = nil
	}
	*pos = idUidT.Next(*pos)
	ok = *pos != nil
	if ok {
		id := (*pos).Val().(*idUidE)
		uid = id.uid
		hash = id.hash
	}
	return
} //IdNextUid

// Position next identity's pubkey for IdNextPubkey
func IdPosPubkey (pubkey Pubkey) *A.Elem {
	pos, _, _ := idPubT.SearchNext(&idPubE{&identity{pubkey: pubkey, hash: ""}})
	pos = idPubT.Previous(pos)
	return pos
} //IdPosPubkey

// Browse all identity's pubkey(s)  step by step
func IdNextPubkey (first bool, pos **A.Elem) (pubkey Pubkey, hash Hash, ok bool) {
	if first {
		*pos = nil
	}
	*pos = idPubT.Next(*pos)
	ok = *pos != nil
	if ok {
		id := (*pos).Val().(*idPubE)
		pubkey = id.pubkey
		hash = id.hash
	}
	return
} //IdNextPubkey

// Position next identity's hash for IdNextHash
func IdPosHash (hash Hash) *A.Elem {
	pos, _, _ := idHashT.SearchNext(&idHashE{&identity{hash: hash}})
	pos = idHashT.Previous(pos)
	return pos
} //IdPosHash

// Browse all identity's hash(es)  step by step
func IdNextHash (first bool, pos **A.Elem) (hash Hash, ok bool) {
	if first {
		*pos = nil
	}
	*pos = idHashT.Next(*pos)
	ok = *pos != nil
	if ok {
		hash = (*pos).Val().(*idHashE).hash
	}
	return
} //IdNextHash

// (Pubkey, Hash) -> certification
func certC (from Pubkey, toHash Hash) *certification {
	c := &certification{from: from, toHash: toHash}
	if e, ok, _ := certFromT.Search(&certFromE{certification: c}); ok {
		cf := e.Val().(*certFromE)
		if e, ok, _ := cf.list.Search(&certToE{certification: c}); ok {
			ct := e.Val().(*certToE)
			return ct.certification
		}
	}
	return nil
} //certC

// (Pubkey, Hash) -> certification
func Cert (from Pubkey, toHash Hash) (to Pubkey, bnb int32, expires_on int64, ok bool) {
	c := certC(from, toHash)
	ok = c != nil
	if ok {
		to = c.to
		bnb = c.bnb
		expires_on = c.expires_on
	}
	return
} //Cert

// Pubkey -> head of subtree
func CertFrom (from Pubkey, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	var e *A.Elem
	if e, ok, _ = certFromT.Search(&certFromE{certification: &certification{from: from}}); ok {
		*pos = CertPos{posT: e.Val().(*certFromE).list, posCur: nil}
	} else {
		pos.posT = nil
	}
	return
} //CertFrom

// Hash -> head of subtree
func CertTo (toHash Hash, pos *CertPos) (ok bool) {
	M.Assert(pos != nil, 20)
	var e *A.Elem
	if e, ok, _ = certToT.Search(&certToE{certification: &certification{toHash: toHash}}); ok {
		*pos = CertPos{posT: e.Val().(*certToE).list, posCur: nil}
	} else {
		pos.posT = nil
	}
	return
} //CertTo

// Number of elements in subtree
func (pos *CertPos) CertPosLen () int {
	M.Assert(pos != nil, 20)
	if pos.posT == nil {
		return 0
	}
	return pos.posT.NumberOfElems()
} //CertPosLen

// Browse all certification(s) in a subtree step by step
func (pos *CertPos) CertNextPos () (from Pubkey, toHash Hash, ok bool) {
	
	posCert := func (pos *CertPos) (c *certification) {
		p := pos.posCur.Val()
		M.Assert(p != nil, 20)
		switch cp := p.(type) {
			case *certFromE: {
				c = cp.certification
			}
			case *certToE: {
				c = cp.certification
			}
		}
		return
	} //posCert
	
	ok = pos.posT != nil
	if ok {
		pos.posCur = pos.posT.Next(pos.posCur)
		ok = pos.posCur != nil
		if ok {
			c := posCert(pos)
			from = c.from
			toHash = c.toHash
		}
	}
	return
} //CertNextPos

// Number of certifiers who certified in sandbox
func CertFromLen () int {
	return certFromT.NumberOfElems()
} //CertFromLen

// Browse all subtrees for all from Pubkey step by step
func CertNextFrom (first bool, pos *CertPos, p **A.Elem) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(p != nil, 21)
	if first {
		*p = nil
	}
	*p = certFromT.Next(*p)
	ok = *p != nil
	if ok {
		*pos = CertPos{posT: (*p).Val().(*certFromE).list, posCur: nil}
	} else {
		pos.posT = nil
	}
	return
} //CertNextFrom

// Number of certifiers who certified in sandbox
func CertToLen () int {
	return certToT.NumberOfElems()
} //CertToLen

// Browse all subtrees for all from Pubkey step by step
func CertNextTo (first bool, pos *CertPos, p **A.Elem) (ok bool) {
	M.Assert(pos != nil, 20)
	M.Assert(p != nil, 21)
	if first {
		*p = nil
	}
	*p = certToT.Next(*p)
	ok = *p != nil
	if ok {
		*pos = CertPos{posT: (*p).Val().(*certToE).list, posCur: nil}
	} else {
		pos.posT = nil
	}
	return
} //CertNextTo

// Extract hash out of buid
func extractBlockId (buid string) Hash {
	i := strings.Index(buid, "-")
	b := []byte(buid)
	return Hash(string(b[i + 1:]))
} //extractBlockId

// Scan the membership and the idty tables in the Duniter database and build idHashT, idPubT and idUidT; remove all items which reference a forked block
func membershipIds (d *Q.DB) {
	// Membership applications
	rows, err := d.Query("SELECT m.idtyHash, m.membership, m.issuer, m.number, m.userid, m.expires_on FROM membership m INNER JOIN block b ON m.blockHash = b.hash WHERE NOT b.fork ORDER BY m.blockNumber ASC")
	M.Assert(err == nil, err, 100)
	tr := A.New()
	for rows.Next() {
		var (
			h Q.NullString
			inOrOut,
			pubkey string
			bnb int32
			uid string
			expires_on int64
		)
		err = rows.Scan(&h, &inOrOut, &pubkey, &bnb, &uid, &expires_on)
		M.Assert(err == nil, err, 101)
		M.Assert(h.Valid, 102); hash := h.String
		id := &identity{hash: Hash(hash), expires_on: 0}
		idH := &idHashE{identity: id}
		if inOrOut == "IN" {
			e, _, _ := tr.SearchIns(idH)
			idH = e.Val().(*idHashE)
			idH.pubkey = Pubkey(pubkey)
			idH.uid = uid
			idH.bnb = bnb
			idH.expires_on = M.Max64(idH.expires_on, expires_on) // The last one is the good one 
		} else { M.Assert(inOrOut == "OUT", 103) // Leaving
			tr.Delete(idH)
		}
	}
	idHashT = A.New()
	e := tr.Next(nil)
	for e != nil { // For every membership applications
		idH := e.Val().(*idHashE)
		if p, ok := B.IdHash(idH.hash); ok { // If identity already in BC...
			if uid, b, _, _, _, exp, ok := B.IdPubComplete(p); ok && !b && exp != BA.Revoked { // ... and if no more member but not revoked
				M.Assert(uid == idH.uid, 112)
				id := &identity{inBC: true, hash: idH.hash, pubkey: p, uid: uid, bnb: idH.bnb, expires_on: M.Min64(M.Abs64(exp), idH.expires_on)}
				_, b, _ = idHashT.SearchIns(&idHashE{identity: id}); M.Assert(!b, 104)
			}
		} else {
			_, ok := B.IdPub(idH.pubkey)
			if !ok {
				_, ok = B.IdUid(idH.uid)
			}
			if !ok { // Not in BC
			// New identities
				row := d.QueryRow("SELECT pubkey, uid, buid, expires_on FROM idty WHERE revocation_sig IS NULL AND hash = '" + string(idH.hash) + "'")
				var (
					pubkey,
					uid,
					buid string
					e Q.NullInt64
				)
				err = row.Scan(&pubkey, &uid, &buid, &e)
				M.Assert(err == nil || err == Q.ErrNoRows, err, 105)
				if err == nil {
					M.Assert(e.Valid, 106); expires_on := e.Int64
					h := extractBlockId(buid)
					row2 := d.QueryRow("SELECT fork FROM block WHERE hash = '" + string(h) + "'")
					var r bool
					err = row2.Scan(&r)
					M.Assert(err == nil, err, 108)
					if !r {
						M.Assert(Pubkey(pubkey) == idH.pubkey && uid == idH.uid, 113)
						id := &identity{inBC: false, hash: idH.hash, pubkey: Pubkey(pubkey), uid: uid, bnb: idH.bnb, expires_on: M.Min64(idH.expires_on, expires_on)}
						_, b, _ := idHashT.SearchIns(&idHashE{identity: id}); M.Assert(!b, 109)
					}
				}
			}
		}
		e = tr.Next(e)
	}
	
	idUidT = A.New(); idPubT = A.New()
	e = idHashT.Next(nil)
	for e != nil {
		idH := e.Val().(*idHashE)
		_, b, _ := idUidT.SearchIns(&idUidE{identity: idH.identity}); M.Assert(!b, 110)
		_, b, _ = idPubT.SearchIns(&idPubE{identity: idH.identity}); M.Assert(!b, 111)
		e = idHashT.Next(e);
	}
} //membershipIds

// Builds certFromT and certToT from the Duniter database; remove all certifications where block_hash is in a fork
func certifications (d *Q.DB) {
	rows, err := d.Query("SELECT [from], [to], target, block_number, expires_on FROM cert INNER JOIN block ON cert.block_hash = block.hash WHERE NOT block.fork")
	M.Assert(err == nil, err, 100)
	now := B.Now()
	certFromT = A.New(); certToT = A.New()
	for rows.Next() {
		var (
			f,
			t,
			h string
			bnb int32
			e Q.NullInt64
		)
		err = rows.Scan(&f, &t, &h, &bnb, &e)
		from := Pubkey(f)
		to := Pubkey(t)
		toHash := Hash(h)
		M.Assert(e.Valid, 101); expires_on := e.Int64
		_, exp, cInBC := B.Cert(from, to)
		_, member, hash, _, _, _, inBC := B.IdPubComplete(to)
		if now <= expires_on && (idHashId(toHash) != nil || inBC && hash == toHash && member) && (!cInBC || expires_on - int64(B.Pars().SigWindow) > exp - int64(B.Pars().SigValidity) + int64(B.Pars().SigReplay)) {
			c := &certification{from: from, to: to, toHash: toHash, bnb: bnb, expires_on: expires_on}
			var (e *A.Elem; ok bool)
			
			if e, ok, _ = certFromT.SearchIns(&certFromE{certification: c}); !ok {
				e.Val().(*certFromE).list = A.New()
			}
			e.Val().(*certFromE).list.SearchIns(&certToE{certification: c})
			
			if e, ok, _ = certToT.SearchIns(&certToE{certification: c}); !ok {
				e.Val().(*certToE).list = A.New()
			}
			e.Val().(*certToE).list.SearchIns(&certFromE{certification: c})
		}
	}
} //certifications

func export () {
	mk := J.NewMaker()
	mk.StartObject()
	mk.PushInteger(int64(B.LastBlock()))
	mk.BuildField("block")
	mk.PushInteger(B.Now())
	mk.BuildField("date")
	mk.StartArray()
	var el *A.Elem
	h, ok := IdNextHash (true, &el)
	for ok {
		inBC, p, uid, bnb, exp, b := IdHash(h); M.Assert(b, 100)
		mk.StartObject()
		mk.PushBoolean(inBC)
		mk.BuildField("inBC")
		mk.PushString(string(h))
		mk.BuildField("hash")
		mk.PushString(string(p))
		mk.BuildField("pubkey")
		mk.PushString(uid)
		mk.BuildField("uid")
		mk.PushInteger(int64(bnb))
		mk.BuildField("bnb")
		mk.PushInteger(exp)
		mk.BuildField("expires_on")
		mk.BuildObject()
		h, ok = IdNextHash(false, &el)
	}
	mk.BuildArray()
	mk.BuildField("identities")
	mk.StartArray()
	var pos CertPos
	ok = CertNextFrom(true, &pos, &el)
	for ok {
		from, toHash, ok2 := pos.CertNextPos()
		for ok2 {
			to, bnb, exp, b := Cert(from, toHash); M.Assert(b, 101)
			mk.StartObject()
			mk.PushString(string(from))
			mk.BuildField("from")
			mk.PushString(string(to))
			mk.BuildField("to")
			mk.PushString(string(toHash))
			mk.BuildField("toHash")
			mk.PushInteger(int64(bnb))
			mk.BuildField("bnb")
			mk.PushInteger(exp)
			mk.BuildField("expires_on")
			mk.BuildObject()
			from, toHash, ok2 = pos.CertNextPos()
		}
		ok = CertNextFrom(false, &pos, &el);
	}
	mk.BuildArray()
	mk.BuildField("certifications")
	mk.BuildObject()
	f, err := os.Create(sBase); M.Assert(err == nil, err, 102)
	mk.GetJson().Write(f)
} //export

func importSb (... interface{}) {
	sd := new(SandboxData)
	j := J.ReadFile(sBase); M.Assert(j != nil, 100)
	J.ApplyTo(j, sd)
	idUidT = A.New()
	idPubT = A.New()
	idHashT = A.New()
	certFromT = A.New()
	certToT = A.New()
	if sd.Identities != nil {
		for _, Id := range sd.Identities {
			id := identity{inBC: Id.InBC, hash: Id.Hash, pubkey: Id.Pubkey, uid: Id.Uid, bnb: Id.Bnb, expires_on: Id.Expires_on}
			_, b, _ := idHashT.SearchIns(&idHashE{identity: &id}); M.Assert(!b, 101)
			_, b, _ = idUidT.SearchIns(&idUidE{identity: &id}); M.Assert(!b, 102)
			_, b, _ = idPubT.SearchIns(&idPubE{identity: &id}); M.Assert(!b, 103)
		}
	}
	if sd.Certifications != nil {
		for _, C := range sd.Certifications {
			c := certification{from: C.From, to: C.To, toHash: C.ToHash, bnb: C.Bnb, expires_on: C.Expires_on}
			var (e *A.Elem; b bool)
			if e, b, _ = certFromT.SearchIns(&certFromE{certification: &c}); !b {
				e.Val().(*certFromE).list  = A.New()
			}
			_, b, _ = e.Val().(*certFromE).list.SearchIns(&certToE{certification: &c}); M.Assert(!b, 107)
			if e, b, _ = certToT.SearchIns(&certToE{certification: &c}); !b {
				e.Val().(*certToE).list  = A.New()
			}
			_, b, _ = e.Val().(*certToE).list.SearchIns(&certFromE{certification: &c}); M.Assert(!b, 108)
		}
	}
} //importSb

// Scan the sandbox in the Duniter database
func scan (... interface{}) {
	BA.Lg.Println("Updating sandbox")
	d, err := Q.Open(B.Driver(), BA.DuniBase)
	M.Assert(err == nil, err, 100)
	defer d.Close()
	membershipIds(d)
	certifications(d)
	export()
	BA.Lg.Println("Sandbox updated")
} //scan

func Initialize () {
	B.AddUpdateProcUpdt(scan)
	B.FixSandBoxFUpdt(importSb)
} //Initialize
