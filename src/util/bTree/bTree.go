/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package bTree

// Version with 64 bits file pointers.

// Manages a disk file like a heap, allocating and freeing it easily, with the possibility of indexing by BTrees.

// Management of concurrency in buffer pages allocations and deallocations

// Small Index Pages Version

import (

	A "util/avl"
	M "util/misc"

)
	
const (
	
	// Minimal number of allocated pages.
	minPageNb = 256
	
	// Nil pointer in a database.
	BNil = FilePos(0)
	
	valOct = 0x100
	
	// Sizes of data in bytes.
	BOS = 1 // bool size.
	BYS = 1 // byte size.
	SIS = 2 // int16 size.
	INS = 4 // int32 size.
	LIS = 8 // int64 size.
	SRS = 4 // float32 size.
	RES = 8 // float64 size.

)
	
type (
	
	// Position in a file, in bytes.
	FilePos int64
	
	// Slice of byte(s).
	Bytes []byte
	
	// Interface for a Factory.
	Factorer interface {
		
		// Create a file with name nF. Don't open this file.
		Create (nF string) bool
		
		// Open a file with name nF and return this file. Return nil if the file does not exist.
		Open (nF string) *File
	
	}
	
	// Factory for files and databases.
	Factory struct {
		Factorer
	}
	
	// Interface for file management.
	Filer interface {
		
		// Close the file.
		Close ()
		
		// Flush the file on disk.
		Flush ()
		
		// Set the reading position of the file to pos; the first position is 0.
		PosReader (pos FilePos)
		
		// Read an array of n bytes from the file at reading position.
		Read (n int) Bytes
		
		// Set the writing position of the file to pos; the first position is 0.
		PosWriter (pos FilePos)
		
		// Write the array a to the file at writing position.
		Write (b Bytes)
		
		// Return the length of the file.
		End () FilePos
		
		// Truncate the file at the length end.
		Truncate (end FilePos)
		
		// Read and return a float64 from the slice b at position pos; put pos at the position following the float64.
		BytesToFloat64 (b Bytes, pos *int) float64
		
		// Read and return a float32 from the slice b at position pos; put pos at the position following the float32.
		BytesToFloat32 (b Bytes, pos *int) float32
		
		// Return a pointer to a slice of bytes coding for f.
		Float64ToBytes (f float64) Bytes
		
		// Return a pointer to a slice of bytes coding for f.
		Float32ToBytes (f float32) Bytes
	
	}
	
	// A file containing a database.
	File struct {
		Filer
	}
	
	// A database
	Database struct {
		ref *File // File containing the database.
		root, // Root of the leftist tree of free clusters.
		end, // Length of file, or at least its used part.
		writtenLim FilePos // Position following the last cluster actually written on disk
		max int64 // Size of the largest free cluster.
		placeNb, // Number of fixed places in database.
		pageNb int // Number of allocated pages.
		pages *A.Tree // Buffer structured as a balanced tree of pages.
		pagesRing *pageT // Buffer structured as a ring of pages.
		
		stopPM,
		detPIn,
		updtPIn,
		verIn chan<- bool
		eraPIn,
		freHIn,
		freTIn,
		readPLIn,
		relIn,
		wriPIn chan<- FilePos
		readPIn chan<- *fpData
		newPIn,
		readSPIn chan<- *fpIntData
		
		detPOut,
		eraPOut,
		relOut,
		updtPOut,
		verOut,
		wriPOut <-chan bool
		readPLOut <-chan int
		newPOut,
		readPOut,
		readSPOut <-chan Data
		freHOut,
		freTOut <-chan *dataBool
	}
	
	// Interface for managers of index keys.
	KeyManagerer interface {
		
		// Comparison method. s1 and s2 are the keys to be compared. The result can be lt (s1 < s2), eq (s1 = s2) or gt (s1 > s2). KeyManagerer.CompP must induce a total order on the set of keys.
		CompP (s1, s2 Data) Comp
		
		// Create a prefix of a key. On input key1 and key2 are two keys, with key1 < key2. On output, key2 may be modified, but cannot be enlarged, so that it becomes a prefix of its previous value. A prefix Pref(key1, key2) of  the key key2 in relation to the key key1 is defined by:
		//	1) Pref(key1, key2) is a key, which can be compared to other keys by the mean of the method KeyManager.CompP;
		//	2) Pref(key1, key2) is the shortest key with key1 < Pref(key1, key2) <= key2;
		//	3) key1 may have a null length and, in this case, must be considered less than any other key.
		// Prefixes are useful if keys are long or if their lengths vary much. In this case, if KeyManager.PrefP is instantiated, database is shorter and searches are faster.
		PrefP (key1 Data, key2 *Data)
	
	}
	
	// Manager of index keys.
	KeyManager struct {
		KeyManagerer
		f DataFac // Factory of keys.
	}
	
	// String, satisfies Data.
	String struct {
		C string
	}
	
	// Factory of String; satisfies DataFac.
	StringFac struct {
	}
	
	// Key managerer for String; satisfies KeyManagerer.
	stringKeyManager struct {
	}
	
	// Reader of an index; each index may create several independant readers.
	IndexReader struct {
		ind *Index // Index of the IndexReader
		posI FilePos  // Current position in index
	}
	
	// Writer of an index; each index owns only one writer.
	IndexWriter struct {
		IndexReader // An IndexWriter can read too.
	}
	
	// Index, structured as a btree.
	Index struct {
		baseI *Database // Database owning the index
		manager *KeyManager // Key manager of index
		refI, // Reference of the index in its database
		rootI, // Root of btree
		stringI FilePos // Doubly linked ring of keys and associated datas (see the type stringI)
		writer *IndexWriter
		keySize, // Size of a key
		keysSize, // Size reserved for the concatenation of all key prefixes in btree pages
		height, // Height of btree
		size int // Number of keys
	}
	
	// Result of a comparaison.
	Comp = A.Comp

)

const (
	
	// Comp; result of a comparaison.
	Lt = Comp(-1) // Less than.
	Eq = Comp(0) // Equal.
	Gt = Comp(+1) // Greater than.
	
	// "Proofs" of base and index.
	guard1 = 1548793025
	guard2 = 2015876309

)
	
type (
	
	// Header of a cluster.
	clusterHead struct {
		size int64 // Size of the used part for reserved clusters, total size for free ones.
	}
	
	// Header of a reserved cluster; satisfies Data.
	rClusterHead struct {
		clusterHead
	}
	
	 // Factory of rClusterHead; satisfies DataFac
	rClusterHeadFac struct {
	}
	
	// Header of a free cluster; satisfies Data
	fClusterHead struct {
		// Structure of leftist tree. Cf. Knuth, The art of computer programming, vol. 3, ch. 5.2.3 and exercises 32 & 35.
		// For any p fClusterHead,
		//	p.size >= p.left.size, p.size >= p.right.size (priority queue: the root has the biggest size),
		//	p.rDist = 1 + p.right.rDist (rDist is the distance to the leaf (+ 1) when going right),
		//	p.left.rDist >= p.right.rDist (hence the name "leftist tree"),
		//	p.lDist = 1 + p.left.rDist (useful)
		//	p.father = bNil (root) or p = p.father.left or p = p.father.right
		clusterHead
		father,
		left,
		right FilePos
		lDist,
		rDist int8
	}
	
	// Factory of fClusterHead; satisfies DataFac
	fClusterHeadFac struct {
	}
	
	// Tail of a cluster. Pointers to clusterTail on disk point to its last field, i.e. the free / reserved boolean
	clusterTail struct {
	}
	
	// Tail of a reserved cluster; satisfies Data
	rClusterTail struct {
		clusterTail
	}
	
	// Factory of rClusterTail; satisfies DataFac
	rClusterTailFac struct {
	}
	
	// Tail of a free cluster; satisfies Data
	fClusterTail struct {
		clusterTail
		size int64 // Total size of the cluster.
	}
	
	// Factory of fClusterTail; satisfies DataFac
	fClusterTailFac struct {
	}

)

const (
	
	// Sizes of structures on disk
	rClusterHeadSize = BOS + LIS
	rClusterTailSize = BOS
	fClusterHeadSize = BOS + 4 * LIS + 2 * BYS
	fClusterTailSize = LIS + BOS
	
	// Size of the fixed part (without places) of a database, as written in Factory.CreateBase
	baseHead = LIS + 3 * INS + 2 * BOS

)

type (
	
	// Content of a page, used to store data or key.
	Data interface {
		Read (r *Reader)
		Write (w *Writer)
	}
	
	// Factory of Data.
	DataFac interface { 
		New (size int) Data
	}
	
	// Manager of Data
	DataMan struct {
		base *Database // Database
		fac DataFac // Factory of Data
	}
	
	// Page of buffered data
	pageT struct {
		nextP, prevP *pageT // Ring of pages
		pageP Data // Buffered data
		posP FilePos // Position of the page in the file
		sizeP, // Size of the page
		locked int // The page is still in use
		dirty bool // Dirty data; must be written to the file
	}

)

const (
	
	// 512: Standard size of sets of variable length keys
	sectSize = 0x200

)

type (
	
	// Element of a btree page
	element struct {
		ptr FilePos // Pointer to the son page
		endK int16 // Position of the last byte of the key (or  its prefix) in the field keys of the page
	}

)

const (
	
	// Size of element
	elemSize = LIS + SIS
	
	// Minimal number of elements in a btree page
	minEl = (M.MaxInt8 - 1) / 2 // 63

)

type (
	
	// Btree page of an index; satisfies Data
	pageI struct {
		keys FilePos // Pointer to the concatenation of all key prefixes in the page
		elNb int8 // Number of elements
		elems [2 * (minEl + 1)]element // Elements of the page
		// p.elems[0].endK = 0,
		// p.elems[0].ptr -> (pageI or stringI) corresponding to keys < the first key (or prefix) of p (i.e. p.keys[p.elems[0].endK .. p.elems[1].endK]),
		// p.elems[i].ptr (i > 0) -> (pageI or stringI) corresponding to keys >= the ith key (or prefix) of p (i.e. p.keys[p.elems[i - 1].endK .. p.elems[i].endK]) and < the (i + 1)th prefix, if present
	}
	
	// Factory of pageI; satisfies DataFac
	pageIFac struct {
	}
	
	// Key in indexes
	keyT Bytes
	
	// Key as Data; satisfies Data
	keyCont struct{
		c keyT
	}
	
	// Factory of keyCont; satisfies DataFac
	keyFac struct {
	}
	
	// Doubly linked ring of keys and associated datas in indexes; satisfies Data
	stringI struct {
		next, // Next item
		prev, // Previous item
		dataPos FilePos // Pointer to associated data
		key keyT // key
	}
	
	// Factory of stringI; satisfies DataFac
	stringIFac struct {
	}
	
	// Root data, on disk, of an index; satisfies Data
	info struct {
		iRoot, // Root of btree
		iString FilePos // Start item of the ring of keys
		iKSize, // Size of keys, or 0 for variable length keys
		iH, // Height of btree
		iSize int // Number of keys
	}
	
	// Factory of info; satisfies DataFac
	infoFac struct {
	}

)

const (
	
	// Size of pageI on disk
	pageIS = LIS + BYS + 2 * (minEl + 1) * elemSize
	
	// Size of the fixed part of stringI (without key) on disk.
	stringIS = 3 * LIS
	
	// Size of info on disk.
	infoS = 2 * LIS + 5 * INS

)

type (
	
	// Reader of a byte stream.
	Reader struct {
		len, // Length of the input stream of the reader.
		pos int // Position of the reader in the input stream.
		ref *File // Used for real <-> bytes translations.
		s Bytes // Contain the input stream.
	}
	
	// Writer on a byte stream.
	Writer struct {
		ref *File // Used for real <-> bytes translations.
		s *A.Tree // Output stream.
	}
	
	// List of open databases; used to close bases left open
	baseList struct {
		nextB *baseList // Next open database
		name string // Name of the base file
		ref *File // The base file
	}

)

var (
	
	fCH fClusterHeadFac
	rCH rClusterHeadFac
	fCT fClusterTailFac
	rCT rClusterTailFac
	
	pif pageIFac
	kf keyFac
	sif stringIFac
	inff infoFac
	
	// List of open databases
	bL *baseList = nil

)

// Comparison method of two Strings. Use the lexical order.
func (man *stringKeyManager) CompP (key1, key2 Data) Comp {
	M.Assert(key1 != nil && key2 != nil, 20)
	k1 := key1.(*String); k2 := key2.(*String)
	if k1.C < k2.C {
		return Lt
	}
	if k1.C > k2.C {
		return Gt
	}
	return Eq
}

type (
	
	// Compare two strings.
	StringComparer = func (key1, key2 Data) Comp

)

// Create a prefix of the String key2, in relation with key1 and the comparison function compare; see PrefP.
func StringPrefP (key1, key2 *String, compare StringComparer) *String {
	M.Assert(key1 != nil && key2 != nil, 20)
	M.Assert(compare(key1, key2) == Lt, 21)
	l2 := len(key2.C)
	key := new(String); key.C = ""
	for l := 0; l < l2 && !(compare(key1, key) == Lt && compare(key, key2) <= Eq); l++ {
		key.C += string(key2.C[l])
	}
	return key
}

// Calculate the shortest prefix of a String, in the proper sense.
func (man *stringKeyManager) PrefP (key1 Data, key2 *Data) {
	*key2 = StringPrefP(key1.(*String), (*key2).(*String), man.CompP)
}

// Return the length (number of bytes) of base.
func (base *Database) End () FilePos {
	return base.end
}

// Return the number of fixed places in base.
func (base *Database) PlaceNb () int {
	return base.placeNb
}

// Initalize a reader from the input stream a at the position pos in a.
func (r *Reader) initReader (ref *File, a Bytes, pos int) {
	r.ref = ref
	r.s = a
	r.pos = pos
	r.len = len(a)
}

// Return the length of the input stream owned by r, in bytes.
func (r *Reader) Len () int {
	return r.len
}

// Return the position of r in its input stream, in bytes.
func (r *Reader) Pos () int {
	return r.pos
}

// Read a byte from the input stream.
func (r *Reader) InByte () byte {
	M.Assert(r.pos + BYS <= r.len, 20)
	c := r.s[r.pos]
	r.pos++
	return c
}

// Read a int8 from the input stream.
func (r *Reader) InInt8 () int8 {
	return int8(r.InByte())
}

// Read an array of bytes from the input stream.
func (r *Reader) InBytes (len int) Bytes {
	M.Assert(r.len - r.pos >= len * BYS, 20)
	bb := make(Bytes, len)
	for i := 0; i < len * BYS; i++ {
		bb[i] = r.s[r.pos]
		r.pos++
	}
	return bb
}

// Read a bool from the input stream.
func (r *Reader) InBool () bool {
	M.Assert(r.pos + BOS <= r.len, 20)
	b := r.s[r.pos] != 0
	r.pos++
	return b
}

// Read a string length from the input stream; preserves the position in the stream.
func (r *Reader) InStringLen () int {
	var (
		n = 0
		pos = r.pos
	)
	for pos < r.len && r.s[pos] != 0 {
		n++
		pos++
	}
	M.Assert(pos < r.len, 20)
	return n
}

// Read a string from the input stream.
func (r *Reader) InString () string {
	n := r.InStringLen()
	bb := make(Bytes, n)
	for i := 0; i < n; i++ {
		bb[i] = r.s[r.pos]
		r.pos++
	}
	r.pos++
	return string(bb)
}

// Read an uint16 from the input stream.
func (r *Reader) InUint16 () uint16 {
	M.Assert(r.pos + SIS <= r.len, 20)
	n := uint16(0)
	for i := r.pos + SIS - 1; i >= r.pos; i-- {
		n = n * valOct + uint16(r.s[i])
	}
	r.pos += SIS
	return n
}

// Read an int16 from the input stream.
func (r *Reader) InInt16 () int16 {
	return int16(r.InUint16())
}

// Read an uint32 from the input stream.
func (r *Reader) InUint32 () uint32 {
	M.Assert(r.pos + INS <= r.len, 20)
	n := uint32(0)
	for i := r.pos + INS - 1; i >= r.pos; i-- {
		n = n * valOct + uint32(r.s[i])
	}
	r.pos += INS
	return n
}

// Read an int32 from the input stream.
func (r *Reader) InInt32 () int32 {
	return int32(r.InUint32())
}

// Read an uint64 from the input stream.
func (r *Reader) InUint64 () uint64 {
	M.Assert(r.pos + LIS <= r.len, 20)
	n := uint64(0)
	for i := r.pos + LIS - 1; i >= r.pos; i-- {
		n = n * valOct + uint64(r.s[i])
	}
	r.pos += LIS
	return n
}

// Read a FilePos from the input stream.
func (r *Reader) InFilePos () FilePos {
	return FilePos(r.InUint64())
}

// Read an int64 from the input stream.
func (r *Reader) InInt64 () int64 {
	return int64(r.InUint64())
}

// Read a float32 from the input stream. Use File.BytesToFloat32.
func (r *Reader) InFloat32 () float32 {
	return r.ref.BytesToFloat32(r.s, &r.pos)
}

// Read a float64 from the input stream. Use File.BytesToFloat64.
func (r *Reader) InFloat64 () float64 {
	return r.ref.BytesToFloat64(r.s, &r.pos)
}

// Initialize a writer
func (w *Writer) initWriter (ref *File) {
	w.ref = ref
	w.s = A.New()
}

// Return the content of w.s (output stream of the writer) as an array of bytes
func (w *Writer) write () Bytes {
	n := w.s.NumberOfElems()
	a := make(Bytes, n)
	i := 0
	e := w.s.Next(nil)
	for e != nil {
		a[i] = e.Val().(byte)
		i++
		e = w.s.Next(e)
	}
	return a
}

// Write a byte to the output stream.
func (w *Writer) OutByte (c byte) {
	w.s.Append(c)
}

// Write an int8 to the output stream.
func (w *Writer) OutInt8 (c int8) {
	w.OutByte(byte(c))
}

// Write a Bytes to the output stream.
func (w *Writer) OutBytes (b Bytes) {
	for _, c := range b {
		w.OutByte(c)
	}
}

// Write a bool to the output stream.
func (w *Writer) OutBool (b bool) {
	if b {
		w.OutByte(1)
	} else {
		w.OutByte(0)
	}
}

// Write a string to the output stream.
func (w *Writer) OutString (c string) {
	w.OutBytes(Bytes(c))
	w.OutByte(0);
}

// Write a uint16 to the output stream.
func (w *Writer) OutUint16 (n uint16) {
	for i := 1; i <= SIS; i++ {
		w.OutByte(byte(n % valOct))
		n /= valOct
	}
}

// Write a int16 to the output stream.
func (w *Writer) OutInt16 (n int16) {
	w.OutUint16(uint16(n))
}

// Write a uint32 to the output stream.
func (w *Writer) OutUint32 (n uint32) {
	for i := 1; i <= INS; i++ {
		w.OutByte(byte(n % valOct))
		n /= valOct
	}
}

// Write a int32 to the output stream.
func (w *Writer) OutInt32 (n int32) {
	w.OutUint32(uint32(n))
}

// Write a uint64 to the output stream.
func (w *Writer) OutUint64 (n uint64) {
	for i := 1; i <= LIS; i++ {
		w.OutByte(byte(n % valOct))
		n /= valOct
	}
}

// Write a FilePos to the output stream.
func (w *Writer) OutFilePos (n FilePos) {
	w.OutUint64(uint64(n))
}

// Write a int64 to the output stream.
func (w *Writer) OutInt64 (n int64) {
	w.OutUint64(uint64(n))
}

// Write a float32 to the output stream. Use File.Float32ToBytes. *)
func (w *Writer) OutFloat32 (r float32) {
	w.OutBytes(w.ref.Float32ToBytes(r))
}

// Write a float64 to the output stream. Use File.Float64ToBytes. *)
func (w *Writer) OutFloat64 (r float64) {
	w.OutBytes(w.ref.Float64ToBytes(r))
}

// Read an integer on disk at the current reading position
func (ref *File) readInt () int {
	a := ref.Read(INS)
	var r Reader
	r.initReader(ref, a, 0)
	return int(r.InInt32())
}

// Read a uint64 on disk at the current reading position
func (ref *File) readInt64 () int64 {
	a := ref.Read(LIS)
	var r Reader
	r.initReader(ref, a, 0)
	return r.InInt64()
}

// Read a FilePos on disk at the current reading position
func (ref *File) readFilePos () FilePos {
	a := ref.Read(LIS)
	var r Reader
	r.initReader(ref, a, 0)
	return r.InFilePos()
}

// Read a bool on disk at the current reading position
func (ref *File) readBool () bool {
	a := ref.Read(BYS)
	var r Reader
	r.initReader(ref, a, 0)
	return a[0] != 0
}

// Write an integer to disk at the current writing position
func (ref *File) writeInt32 (n int32) {
	var w Writer
	w.initWriter(ref)
	w.OutInt32(int32(n))
	ref.Write(w.write())
}

// Write a uint64 to disk at the current writing position
func (ref *File) writeInt64 (n int64) {
	var w Writer
	w.initWriter(ref)
	w.OutInt64(n)
	ref.Write(w.write())
}

// Write a FilePos to disk at the current writing position
func (ref *File) writeFilePos (n FilePos) {
	var w Writer
	w.initWriter(ref)
	w.OutFilePos(n)
	ref.Write(w.write())
}

// Write a bool to disk at the current writing position
func (ref *File) writeBool (b bool) {
	a := make(Bytes, BYS)
	if b {
		a[0] = 1
	} else {
		a[0] = 0
	}
	ref.Write(a)
}

// Create a String.
func (f StringFac) New (size int) Data {
	M.Assert(size >= 0, 20)
	return new(String)
}

// Write s.C with the help of the writer w. The length of the production is 2 * (len(s.C)+ 1).
func (s *String) Write (w *Writer) {
	w.OutString(s.C)
}

// Read s.c with the help of the reader r.
func (s *String) Read (r *Reader) {
	s.C = r.InString()
}

// Create and return a new manager of the data created by fac in the database base.
func (base *Database) CreateDataMan (fac DataFac) *DataMan {
	M.Assert(fac != nil, 20)
	man := new(DataMan)
	man.base = base
	man.fac = fac
	return man
}

func (f fClusterHeadFac) New (sz int) Data {
	M.Assert(sz == fClusterHeadSize, 101)
	return new(fClusterHead)
}

func (h *fClusterHead) Write (w *Writer) {
	w.OutBool(true)
	w.OutInt64(h.size)
	w.OutFilePos(h.father)
	w.OutFilePos(h.left)
	w.OutFilePos(h.right)
	w.OutInt8(h.lDist)
	w.OutInt8(h.rDist)
}

func (h *fClusterHead) Read (r *Reader) {
	M.Assert(r.InBool(), 102)
	h.size = r.InInt64()
	h.father = r.InFilePos()
	h.left = r.InFilePos()
	h.right = r.InFilePos()
	h.lDist = r.InInt8()
	h.rDist = r.InInt8()
}

func (f fClusterTailFac) New (sz int) Data {
	M.Assert(sz == fClusterTailSize, 103)
	return new(fClusterTail)
}

func (t *fClusterTail) Write (w *Writer) {
	w.OutInt64(t.size)
	w.OutBool(true)
}

func (t *fClusterTail) Read  (r *Reader) {
	t.size = r.InInt64()
	M.Assert(r.InBool(), 104)
}

func (f rClusterHeadFac) New (sz int) Data {
	M.Assert(sz == rClusterHeadSize, 105)
	return new(rClusterHead)
}

func (h *rClusterHead) Write (w *Writer) {
	w.OutBool(false)
	w.OutInt64(h.size)
}

func (h *rClusterHead) Read (r *Reader) {
	M.Assert(!r.InBool(), 106)
	h.size = r.InInt64()
}

func (f rClusterTailFac) New (sz int) Data {
	M.Assert(sz == rClusterTailSize, 107)
	return new(rClusterTail)
}

func (t *rClusterTail) Write (w *Writer) {
	w.OutBool(false);
}

func (t *rClusterTail) Read (r *Reader) {
	M.Assert(!r.InBool(), 108)
}

// Read n bytes from position ptr on disk
func (base *Database) readBase (ptr FilePos, n int) Bytes {
	base.ref.PosReader(ptr)
	return base.ref.Read(n)
}

// Write a to disk at position ptr
func (base *Database) writeBase (ptr FilePos, a Bytes) {
	base.ref.PosWriter(ptr)
	base.ref.Write(a)
}

// Ordering relation of pages in database buffer; order is given by the position of pages on disk
func (p1 *pageT) Compare (p2 A.Comparer) Comp {
	pp2 := p2.(*pageT)
	if p1.posP < pp2.posP {
		return A.Lt
	}
	if p1.posP > pp2.posP {
		return A.Gt
	}
	return A.Eq
}

// Look for the buffer page p at position pos; return true if found
func (base *Database) findPage (pos FilePos) (p *pageT, ok bool) {
	e, ok, _ := base.pages.Search(&pageT{posP: pos})
	if ok {
		p = e.Val().(*pageT)
	}
	return
}

type (
	
	fpData struct {
		pos FilePos
		fac DataFac
	}
	
	fpIntData struct {
		pos FilePos
		pageSize int
		fac DataFac
	}
	
	dataBool struct {
		cont Data
		free bool
	}

)

func (base *Database) pageManager (stopPM, detPIn, updtPIn, verIn <-chan bool, eraPIn, freHIn, freTIn, readPLIn, relIn, wriPIn <-chan FilePos, readPIn <-chan *fpData, newPIn, readSPIn <-chan *fpIntData, detPOut, eraPOut, relOut, updtPOut, verOut, wriPOut chan<- bool, readPLOut chan<- int, newPOut, readPOut, readSPOut chan<- Data, freHOut, freTOut chan<- *dataBool) {
	for {
		select {
		case <-stopPM:
			return
		case <-detPIn:
			base.detachPagesA()
			detPOut <- true
		case pos := <-eraPIn:
			base.erasePageA(pos)
			eraPOut <- true
		case pos := <-freHIn:
			cont, free := base.freeHeadA(pos)
			freHOut <- &dataBool{cont:cont, free: free}
		case pos := <-freTIn:
			cont, free := base.freeTailA(pos)
			freTOut <- &dataBool{cont:cont, free: free}
		case fid := <-newPIn:
			newPOut <- base.newPageA(fid.pageSize, fid.fac, fid.pos)
		case fd := <-readPIn:
			readPOut <- base.readPageA(fd.pos, fd.fac)
		case pos := <-readPLIn:
			readPLOut <- base.readPageLengthA(pos)
		case fid := <-readSPIn:
			readSPOut <- base.readSysPageA(fid.pos, fid.pageSize, fid.fac)
		case pos := <-relIn:
			base.releaseA(pos)
			relOut <- true
		case <-updtPIn:
			base.updatePagesA()
			updtPOut <- true
		case <-verIn:
			base.verifyNotLockedA()
			verOut <- true
		case pos := <-wriPIn:
			base.writePageA(pos)
			wriPOut <- true
		}
	}
}

// Promote p at the first place in the buffer ring
func (base *Database) promotePage (p *pageT){
	p.nextP.prevP = p.prevP
	p.prevP.nextP = p.nextP
	p.nextP = base.pagesRing.nextP
	p.prevP = base.pagesRing
	base.pagesRing.nextP.prevP = p
	base.pagesRing.nextP = p
	p.locked++ // Keep it in RAM
}

func adjustPos (d Data, size int, pos *FilePos) {
	switch d.(type) {
	case *fClusterTail, *rClusterTail:
		*pos -= FilePos(size) - BOS
	}
}

// Create an empty buffer page in ram for position pos and size pageSize on disk; if buffer is already full, remove the least promoted page
func (base *Database) createPage (pageSize int, pos FilePos) *pageT {
	var p *pageT
	if base.pages.NumberOfElems() < base.pageNb { // Allocate a new buffer page
		p = new(pageT)
		p.nextP = base.pagesRing.nextP
		p.prevP = base.pagesRing
		base.pagesRing.nextP.prevP = p
		base.pagesRing.nextP = p
		p.locked = 1 // Keep it in RAM
	} else { // No more new page available: recycle an old one, the oldest
		p = base.pagesRing.prevP
		for p != base.pagesRing && p.locked > 0 { //Don't erase a locked page
			p = p.prevP
		}
		M.Assert(p != base.pagesRing, 100) //Not enough pages, increase base.pageNb
		base.promotePage(p)
		if p.dirty { // Page not up to date on disk: write it
			if p.posP >= base.writtenLim { // Not yet written: write all pages with smaller addresses that are not still written (no hole on disk file)
				pp := new(pageT)
				pp.posP = base.writtenLim
				q, _, _ := base.pages.SearchNext(pp)
				for q != nil && q.Val().(*pageT).posP <= p.posP {
					switch qq := q.Val().(type) {
					case *pageT:
						po := qq.posP
						adjustPos(qq.pageP, qq.sizeP, &po)
						if qq.dirty { // Write it normally
							qq.dirty = false
							var w Writer
							w.initWriter(base.ref)
							qq.pageP.Write(&w)
							aP := w.write()
							base.writeBase(po, aP)
						// Useless on Linux (Ext4)
						//} else { // Write a series of zeroes
						//	var a = Bytes{0} 
						//	base.ref.PosWriter(po);
						//	for i := 0; i < qq.sizeP; i++ {
						//		base.ref.Write(a)
						//	}
						}
					}
					q = base.pages.Next(q)
				}
				// Update the limit of written clusters
				if q == nil {
					base.writtenLim = base.end
				} else {
					base.writtenLim = q.Val().(*pageT).posP
				}
			} else { // Write the oldest dirty page
				po := p.posP
				adjustPos(p.pageP, p.sizeP, &po)
				var w Writer
				w.initWriter(base.ref)
				p.pageP.Write(&w);
				aP := w.write()
				base.writeBase(po, aP)
			}
		}
		b := base.pages.Delete(p)
		M.Assert(b, 109)
	}
	p.dirty = false
	p.posP = pos
	p.pageP = nil
	p.sizeP = pageSize
	_, b, _ := base.pages.SearchIns(p)
	M.Assert(!b, 110)
	return p
}

// Create a new buffer page in ram for position pos and size pageSize (on disk); create a new empty Data in it with the help of fac and return it
func (base *Database) newPageA (pageSize int, fac DataFac, pos FilePos) Data {
	p := base.createPage(pageSize, pos)
	p.pageP = fac.New(pageSize)
	return p.pageP
}

// Create a new buffer page in ram for position pos and size pageSize (on disk); create a new empty Data in it with the help of fac and return it
func (base *Database) newPage (pageSize int, fac DataFac, pos FilePos) Data {
	base.newPIn <- &fpIntData{pageSize: pageSize, fac: fac, pos: pos}
	return <-base.newPOut
}

// Erase buffer page with position pos by removing it from base.pagesRing and base.pages
func (base *Database) erasePageA (pos FilePos) {
	if p, ok := base.findPage(pos); ok {
		p.nextP.prevP = p.prevP
		p.prevP.nextP = p.nextP
		b := base.pages.Delete(p)
		M.Assert(b, 111)
	}
}

// Erase buffer page with position pos by removing it from base.pagesRing and base.pages
func (base *Database) erasePage (pos FilePos) {
	base.eraPIn <- pos
	<- base.eraPOut
}

// Find the buffer page at position pos, or create it empty with size pageSize if not found; promote and return it
func (base *Database) selectSysPage (pos FilePos, pageSize int) *pageT {
	p, ok := base.findPage(pos)
	if ok {
		base.promotePage(p)
	} else {
		p = base.createPage(pageSize, pos)
	}
	return p
}

// Find, or create and fill if not found, the buffer page at position pos; creation, if any, is made with size pageSize and the help of fac
func (base *Database) readSysPageA (pos FilePos, pageSize int, fac DataFac) Data {
	p := base.selectSysPage(pos, pageSize)
	if p.pageP == nil {
		p.pageP = fac.New(p.sizeP)
		adjustPos(p.pageP, p.sizeP, &pos)
		a := base.readBase(pos, p.sizeP) // p.sizeP == pageSize
		var r Reader
		r.initReader(base.ref, a, 0)
		p.pageP.Read(&r)
	}
	return p.pageP
}

// Find, or create and fill if not found, the buffer page at position pos; creation, if any, is made with size pageSize and the help of fac
func (base *Database) readSysPage (pos FilePos, pageSize int, fac DataFac) Data {
	base.readSPIn <- &fpIntData{pos: pos, pageSize: pageSize, fac: fac}
	return <-base.readSPOut
}

// Mark the buffer page at position pos for writing on disk; the page must exist and not be empty
func (base *Database) writePageA (pos FilePos) {
	p, ok := base.findPage(pos)
	M.Assert(ok && (p.pageP != nil), 100)
	p.dirty = true
}

// Mark the buffer page at position pos for writing on disk; the page must exist and not be empty
func (base *Database) writePage (pos FilePos) {
	base.wriPIn <- pos
	<-base.wriPOut
}

// Release a locked page when it is no more in use
func (base *Database) releaseA (pos FilePos) {
	if pos != BNil {
		p, ok := base.findPage(pos)
		M.Assert(ok, 100)
		M.Assert(p.locked > 0, 101)
		p.locked--
	}
}

// Release a locked page when it is no more in use
func (base *Database) release (pos FilePos) {
	base.relIn <- pos
	<-base.relOut
}

// Write all buffer pages marked for writing on disk
func (base *Database) updatePagesA () {
	do :=
		func (v interface{}, _ ...interface{}) {
			p := v.(*pageT)
			n := p.posP
			adjustPos(p.pageP, p.sizeP, &n)
			if p.dirty {
				p.dirty = false
				var w Writer
				w.initWriter(base.ref)
				p.pageP.Write(&w)
				aP := w.write()
				base.writeBase(n, aP)
			// Useless on Linux (Ext4)
			//} else if p.posP >= base.writtenLim {
			//	a := Bytes{0}
			//	base.ref.PosWriter(n)
			//	for i := 0; i < p.sizeP; i++ {
			//		base.ref.Write(a)
			//	}
			}
		}
	base.pages.WalkThrough(do)
	base.writtenLim = base.end
}

// Write all buffer pages marked for writing on disk
func (base *Database) updatePages () {
	base.updtPIn <- true
	<-base.updtPOut
}

// Create a new file of name nF, with the help of fac, and a database inside this file, with placeNb fixed places. Don't open this database. Fixed places are locations where can be recorded integers, i.e. data pointers.
func (fac *Factory) CreateBase (nF string, placeNb int) bool {
	if !fac.Create(nF) {
		return false
	}
	ref := fac.Open(nF); M.Assert(ref != nil, 100)
	ref.PosWriter(0)
	ref.writeInt32(guard1) // First "proof"
	ref.writeFilePos(BNil) // root 
	ref.writeInt32(guard2) // Second "proof"
	ref.writeInt32(int32(placeNb))
	for i := 0; i < placeNb; i++ {
		ref.writeFilePos(BNil) // Places are empty
	}
	ref.writeBool(false) // Closed database
	ref.writeBool(false) // Tail of  a reserved cluster
	ref.Close()
	return true
}

// Open and return, with the help of fac, the database created in the file of name nF. pageNb is the maximal number of allocated buffer pages. Return nil if the file does not exit or can't be opened.
func (fac *Factory) OpenBase (nF string, pageNb int) *Database {
	base := &Database{ref: fac.Open(nF)}
	if base.ref == nil {
		return nil
	}
	bL = &baseList{nextB: bL, name: nF, ref: base.ref}
	base.ref.PosReader(0)
	M.Assert(base.ref.readInt() == guard1, 20)
	base.root = base.ref.readFilePos()
	M.Assert(base.ref.readInt() == guard2, 21)
	base.placeNb = base.ref.readInt()
	base.ref.PosReader(FilePos((base.placeNb + 1) * LIS + 3 * INS))
	M.Assert(!base.ref.readBool(), 22)
	base.ref.PosWriter(FilePos((base.placeNb + 1) * LIS + 3 * INS))
	base.ref.writeBool(true)
	base.end = base.ref.End()
	base.writtenLim = base.end
	base.pageNb = M.Max(pageNb, minPageNb)
	base.pages = A.New()
	base.pagesRing = new(pageT)
	base.pagesRing.nextP = base.pagesRing
	base.pagesRing.prevP = base.pagesRing
	
	stopPM := make(chan bool); base.stopPM = stopPM
	detPIn := make(chan bool); base.detPIn = detPIn
	updtPIn := make(chan bool); base.updtPIn = updtPIn
	verIn := make(chan bool); base.verIn = verIn
	eraPIn := make(chan FilePos); base.eraPIn = eraPIn
	freHIn := make(chan FilePos); base.freHIn = freHIn
	freTIn := make(chan FilePos); base.freTIn = freTIn
	readPLIn := make(chan FilePos); base.readPLIn = readPLIn
	relIn := make(chan FilePos); base.relIn = relIn
	wriPIn := make(chan FilePos); base.wriPIn = wriPIn
	readPIn := make(chan *fpData); base.readPIn = readPIn
	newPIn := make(chan *fpIntData); base.newPIn = newPIn
	readSPIn := make(chan *fpIntData); base.readSPIn = readSPIn
	
	detPOut := make(chan bool); base.detPOut = detPOut
	eraPOut := make(chan bool); base.eraPOut = eraPOut
	relOut := make(chan bool); base.relOut = relOut
	updtPOut := make(chan bool); base.updtPOut = updtPOut
	verOut := make(chan bool); base.verOut = verOut
	wriPOut := make(chan bool); base.wriPOut = wriPOut
	readPLOut := make(chan int); base.readPLOut = readPLOut
	newPOut := make(chan Data); base.newPOut = newPOut
	readPOut := make(chan Data); base.readPOut = readPOut
	readSPOut := make(chan Data); base.readSPOut = readSPOut
	freHOut := make(chan *dataBool); base.freHOut = freHOut
	freTOut := make(chan *dataBool); base.freTOut = freTOut
	
	go base.pageManager(stopPM, detPIn, updtPIn, verIn, eraPIn, freHIn, freTIn, readPLIn, relIn, wriPIn, readPIn, newPIn, readSPIn, detPOut, eraPOut, relOut, updtPOut, verOut, wriPOut, readPLOut, newPOut, readPOut, readSPOut, freHOut, freTOut)
	
	if base.root == BNil {
		base.max = 0
	} else {
		base.max = base.readSysPage(base.root, fClusterHeadSize, fCH).(*fClusterHead).size
		base.release(base.root)
	}
	
	return base
}

// Verify, with the help of fac, if the file nF exists and if it contains a database; return true in this case; the base must be closed.
func (fac *Factory) TestBase (nF string) bool {
	ref := fac.Open(nF)
	if ref == nil {
		return false
	}
	ref.PosReader(0)
	ok := ref.readInt() == guard1
	if ok {
		ref.PosReader(INS + LIS)
		ok = ref.readInt() == guard2
	}
	ref.Close()
	return ok
}

// Close, with the help of fac, the database created in the file of name nF, if it exists and is open. It is a rescue func. Use it only in case of an accidentally kept open database. Normally, use Database.CloseBase.
func (fac *Factory) CloseBase (nF string) {
	var ll *baseList = nil; l := bL
	for l != nil && l.name != nF {
		ll = l; l = l.nextB
	}
	if l != nil {
		if ll == nil {
			bL = l.nextB
		} else {
			ll.nextB = l.nextB
		}
		l.ref.Close()
	}
	ref := fac.Open(nF)
	if ref != nil {
		ref.PosReader(0)
		M.Assert(ref.readInt() == guard1, 20)
		ref.PosReader(INS + LIS);
		M.Assert(ref.readInt() == guard2, 21)
		placeNb := ref.readInt()
		ref.PosWriter(FilePos((placeNb + 1) * LIS + 3 * INS))
		ref.writeBool(false)
		ref.Close()
	}
}

// Update the database base on disk.
func (base *Database) UpdateBase () {
	M.Assert(base.ref != nil, 20)
	base.updatePages()
	base.ref.PosWriter(INS)
	base.ref.writeFilePos(base.root)
	base.ref.Flush()
}

func (base *Database) verifyNotLockedA () {
	p := base.pagesRing.nextP
	for p != base.pagesRing {
		M.Assert(p.locked == 0, 101) // Otherwise, the page was never released
		p = p.nextP
	}
}

func (base *Database) verifyNotLocked () {
	base.verIn <- true
	<-base.verOut
}

func (base *Database) detachPagesA () {
	base.pages = nil
	base.pagesRing = nil
}

func (base *Database) detachPages () {
	base.detPIn <- true
	<-base.detPOut
}

// Close the database base.
func (base *Database) CloseBase () {
	M.Assert(base.ref != nil, 20)
	var ll *baseList = nil; l := bL
	for l != nil && l.ref != base.ref {
		ll = l; l = l.nextB
	}
	M.Assert(l != nil, 60)
	if ll == nil {
		bL = l.nextB
	} else {
		ll.nextB = l.nextB
	}
	base.updatePages()
	base.verifyNotLocked()
	base.ref.PosWriter(INS)
	base.ref.writeFilePos(base.root)
	base.ref.PosWriter(FilePos((base.placeNb + 1) * LIS + 3 * INS))
	base.ref.writeBool(false)
	base.ref.Close()
	base.ref = nil
	base.detachPages()
	base.stopPM <- true
}

// Read on disk, and return, the size of the page at position ptr
func (base *Database) readBaseLength (ptr FilePos) int {
	base.ref.PosReader(ptr - LIS) // Point to the last field (size) of the preceding rClusterHead
	return int(base.ref.readInt64())
}

// Join the two leftist trees of roots pointed by root1 and root2, and returns a pointer to the new root in root. base is the database. dad is a pointer to the father of root. cont1 and cont2 are the roots of the two trees. On return, dist contains the field rDist of the new root plus one. *)
func (base *Database) merge (dad, root1, root2 FilePos, cont1, cont2 *fClusterHead) (root FilePos, dist int8) {
	if root1 == BNil {
		root = root2
		if root2 == BNil {
			dist = 1
		} else {
			dist = cont2.rDist + 1
			if cont2.father != dad {
				cont2.father = dad
				base.writePage(root2)
			}
		}
	} else if root2 == BNil {
		root = root1
		dist = cont1.rDist + 1
		if cont1.father != dad {
			cont1.father = dad
			base.writePage(root1)
		}
	} else if cont1.size > cont2.size || cont1.size == cont2.size && root1 < root2 {
		root = root1
		oldCont := new(fClusterHead)
		*oldCont = *cont1
		cont1.father = dad
		var cont *fClusterHead
		ptr := cont1.right
		if ptr != BNil {
			cont = base.readSysPage(ptr, fClusterHeadSize, fCH).(*fClusterHead)
		}
		cont1.right, cont1.rDist = base.merge(root1, ptr, root2, cont, cont2)
		base.release(ptr)
		if cont1.lDist < cont1.rDist {
			cont1.left, cont1.right = cont1.right, cont1.left
			cont1.lDist, cont1.rDist = cont1.rDist, cont1.lDist
		}
		dist = cont1.rDist + 1
		if cont1.father != oldCont.father || cont1.left != oldCont.left || cont1.right != oldCont.right || cont1.lDist != oldCont.lDist || cont1.rDist != oldCont.rDist {
			base.writePage(root1)
		}
	} else {
		root = root2
		oldCont := new(fClusterHead)
		*oldCont = *cont2
		cont2.father = dad
		var cont *fClusterHead
		ptr := cont2.right
		if ptr != BNil {
			cont = base.readSysPage(ptr, fClusterHeadSize, fCH).(*fClusterHead)
		}
		cont2.right, cont2.rDist = base.merge(root2, root1, ptr, cont1, cont)
		base.release(ptr)
		if cont2.lDist < cont2.rDist {
			cont2.left, cont2.right = cont2.right, cont2.left
			cont2.lDist, cont2.rDist = cont2.rDist, cont2.lDist
		}
		dist = cont2.rDist + 1
		if cont2.father != oldCont.father || cont2.left != oldCont.left || cont2.right != oldCont.right || cont2.lDist != oldCont.lDist || cont2.rDist != oldCont.rDist {
			base.writePage(root2)
		}
	}
	return
}

// Insert the free header cluster cont, pointed by ptr, in the leftist tree of root pointed by root. If the root is modified, root points to the new one. base is the database. If the size of the greatest cluster is modified, max returns the new value. *)
func (base *Database) baseInsertFree (ptr FilePos, cont *fClusterHead, root *FilePos, max *int64) {
	cont.father = BNil
	cont.left = BNil
	cont.right = BNil
	cont.lDist = 1
	cont.rDist = 1
	oldRoot := *root
	var cont2 *fClusterHead = nil
	if *root == BNil {
		base.writePage(ptr)
	} else {
		cont2 = base.readSysPage(*root, fClusterHeadSize, fCH).(*fClusterHead)
	}
	*root, _ = base.merge(BNil, *root, ptr, cont2, cont)
	base.release(oldRoot)
	if *root == ptr {
		*max = cont.size
	} else if *root != oldRoot {
		*max = base.readSysPage(*root, fClusterHeadSize, fCH).(*fClusterHead).size
		base.release(*root)
	}
}

// Remove the free header cluster cont, pointed by ptr, from the leftist tree of root root. If the root is modified, root points to the new one. base is the database. If the size of the greatest cluster is modified, max returns the new value. *)
func (base *Database) baseRemoveFree (ptr FilePos, cont *fClusterHead, root *FilePos, max *int64) {
	var lCont, rCont *fClusterHead
	if cont.left != BNil {
		lCont = base.readSysPage(cont.left, fClusterHeadSize, fCH).(*fClusterHead)
	}
	if cont.right != BNil {
		rCont = base.readSysPage(cont.right, fClusterHeadSize, fCH).(*fClusterHead);
	}
	child, dist := base.merge(cont.father, cont.left, cont.right, lCont, rCont)
	base.release(cont.left); base.release(cont.right)
	nPtr := cont.father
	if nPtr == BNil {
		*root = child
		if *root == BNil {
			*max = 0
		} else {
			cont = base.readSysPage(*root, fClusterHeadSize, fCH).(*fClusterHead)
			*max = cont.size
			base.release(*root)
		}
	} else {
		cont = base.readSysPage(nPtr, fClusterHeadSize, fCH).(*fClusterHead)
		oldDist := cont.rDist
		if cont.left == ptr {
			cont.left = child
			cont.lDist = dist
		} else {
			M.Assert(cont.right == ptr, 60)
			cont.right = child
			cont.rDist = dist
		}
		for {
			if cont.rDist > cont.lDist {
				cont.left, cont.right = cont.right, cont.left
				cont.lDist, cont.rDist = cont.rDist, cont.lDist
			}
			newDist := cont.rDist
			dad := cont.father
			base.writePage(nPtr)
			base.release(nPtr)
			if newDist == oldDist {
				break
			}
			dist = newDist + 1
			nPtr = dad
			if nPtr == BNil {
				break
			}
			cont = base.readSysPage(nPtr, fClusterHeadSize, fCH).(*fClusterHead)
			oldDist = cont.rDist
		}
	}
}

// Reserve a page of size size on disk in base and return its position
func (base *Database) allocateBase (size int) FilePos {
	var ptr FilePos
	sz := int64(((size + rClusterHeadSize + rClusterTailSize - 1) / (fClusterHeadSize + fClusterTailSize) + 1) * (fClusterHeadSize + fClusterTailSize)) // Quantification of disk space with lump sizes multiple of fClusterHeadSize + fClusterTailSize
	szPos := FilePos(sz)
	if base.max < sz { // If not enough place, extend the base
		ptr = base.end
		base.end += szPos
	} else { // Take a part of the biggest lump
		szRest := base.max - sz
		szRPos := FilePos(szRest)
		ptr = base.root
		cont := base.readSysPage(ptr, fClusterHeadSize, fCH).(*fClusterHead)
		base.baseRemoveFree(ptr, cont, &base.root, &base.max)
		base.erasePage(ptr)
		if szRest == 0 {
			base.erasePage(ptr + szPos - 1)
		} else {
			contT := base.readSysPage(ptr + szPos + szRPos - 1, fClusterTailSize, fCT).(*fClusterTail)
			contT.size = szRest
			base.writePage(ptr + szPos + szRPos - 1)
			base.release(ptr + szPos + szRPos - 1)
			cont = base.newPage(fClusterHeadSize, fCH, ptr + szPos).(*fClusterHead)
			cont.size = szRest
			base.baseInsertFree(ptr + szPos, cont, &base.root, &base.max)
			base.release(ptr + szPos)
		}
	}
	contHR := base.newPage(rClusterHeadSize, rCH, ptr).(*rClusterHead)
	contHR.size = int64(size)
	base.writePage(ptr)
	base.release(ptr)
	_ = base.newPage(rClusterTailSize, rCT, ptr + szPos - 1).(*rClusterTail)
	base.writePage(ptr + szPos - 1)
	base.release(ptr + szPos - 1)
	ptr += rClusterHeadSize
	return ptr
}

// Return whether the header at position pos is free or not; return in cont this header
func (base *Database) freeHeadA (pos FilePos) (cont Data, free bool) {
	if p, ok := base.findPage(pos); ok {
		cont = p.pageP
		if _, free = cont.(*fClusterHead); free {
			p.locked++
		}
		return
	}
	base.ref.PosReader(pos)
	if free = base.ref.readBool(); free {
		cont = base.readSysPageA(pos, fClusterHeadSize, &fCH)
	}
	return
}

// Return whether the header at position pos is free or not; return in cont this header
func (base *Database) freeHead (pos FilePos) (cont Data, free bool) {
	base.freHIn <- pos
	db := <-base.freHOut
	return db.cont, db.free
}

// Return whether the tail at position pos is free or not; return in cont this tail
func (base *Database) freeTailA (pos FilePos) (cont Data, free bool) {
	if p, ok := base.findPage(pos); ok {
		cont = p.pageP
		if _, free = cont.(*fClusterTail); free {
			p.locked++
		}
		return
	}
	base.ref.PosReader(pos)
	if free = base.ref.readBool(); free {
		cont = base.readSysPageA(pos, fClusterTailSize, &fCT)
	}
	return
}

// Return whether the tail at position pos is free or not; return in cont this tail
func (base *Database) freeTail (pos FilePos) (cont Data, free bool) {
	base.freTIn <- pos
	db := <-base.freTOut
	return db.cont, db.free
}

// Return the reserved cluster at position ptr to the pool of free clusters on disk; aggregate adjacent free clusters into bigger ones; truncate the base file if possible
func (base *Database) deleteBase (ptr FilePos) {
	M.Assert(ptr >= FilePos(baseHead + rClusterHeadSize + base.placeNb * LIS), 112) // ptr points after the first head of reserved cluster
	ptr -= rClusterHeadSize
	contHR := base.readSysPage(ptr, rClusterHeadSize, rCH).(*rClusterHead)
	sz := ((contHR.size + rClusterHeadSize + rClusterTailSize - 1) / (fClusterHeadSize + fClusterTailSize) + 1) * (fClusterHeadSize + fClusterTailSize) // Quantification of disk space with lump sizes multiple of FClusterHeadSize + FClusterTailSize
	szPos := FilePos(sz)
	M.Assert(ptr + szPos <= base.end, 113)
	base.erasePage(ptr) // Erase head of reserved cluster...
	q := ptr + szPos
	base.erasePage(q - 1) // ...and its tail
	hasTail := false
	if q == base.end {
		base.end = ptr
		base.ref.Truncate(base.end) // If last cluster, truncate file...
	} else if contH, free := base.freeHead(q); free { // ... else, if next cluster is free, aggregate it
		contHF := contH.(*fClusterHead)
		sz += contHF.size; szPos = FilePos(sz)
		base.baseRemoveFree(q, contHF, &base.root, &base.max)
		base.erasePage(q)
		hasTail = true
	}
	q = ptr - 1
	if contT, free := base.freeTail(q); free { // If previous cluster is free...
		contTF := contT.(*fClusterTail)
		base.erasePage(q)
		q = ptr - FilePos(contTF.size)
		contHF := base.readSysPage(q, fClusterHeadSize, fCH).(*fClusterHead)
		base.baseRemoveFree(q, contHF, &base.root, &base.max)
		if ptr == base.end { // ... and if it's the last cluster, truncate file...
			base.erasePage(q)
			base.erasePage(ptr + szPos - 1)
			base.end = q
			base.ref.Truncate(base.end)
		} else { // ... else aggregate it
			contHF.size += sz
			qq := q + FilePos(contHF.size) - 1
			if hasTail {
				contTF = base.readSysPage(qq, fClusterTailSize, fCT).(*fClusterTail)
			} else {
				contTF = base.newPage(fClusterTailSize, fCT, qq).(*fClusterTail)
			}
			contTF.size = contHF.size
			base.writePage(qq)
			base.release(qq)
			base.baseInsertFree(q, contHF, &base.root, &base.max)
			base.release(q)
		}
	} else if ptr != base.end {
		contHF := base.newPage(fClusterHeadSize, fCH, ptr).(*fClusterHead)
		contHF.size = sz
		var contTF *fClusterTail
		if hasTail {
			contTF = base.readSysPage(ptr + szPos - 1, fClusterTailSize, fCT).(*fClusterTail)
		} else {
			contTF = base.newPage(fClusterTailSize, fCT, ptr + szPos - 1).(*fClusterTail)
		}
		contTF.size = sz
		base.writePage(ptr + szPos - 1)
		base.release(ptr + szPos - 1)
		base.baseInsertFree(ptr, contHF, &base.root, &base.max)
		base.release(ptr)
	}
	if base.end < base.writtenLim {
		base.writtenLim = base.end
	}
}

// Read and return the content of the fixed place place of the database base. The first place has number 0.
func (base *Database) ReadPlace (place int) int64 {
	M.Assert(base.ref != nil, 20)
	M.Assert(place >= 0 && place < base.placeNb, 21)
	base.ref.PosReader(FilePos((place + 1) * LIS + 3 * INS))
	return base.ref.readInt64()
}

// Write val in the fixed place place of the database base. The first place has number 0.
func (base *Database) WritePlace (place int, val int64) {
	M.Assert(base.ref != nil, 20)
	M.Assert(place >= 0 && place < base.placeNb, 21)
	base.ref.PosWriter(FilePos((place + 1) * LIS + 3 * INS))
	base.ref.writeInt64(val)
}

// Reserve a page of size pageSize on disk and create a new buffer page for it, with the help of fac; return the empty data, and the page position
func (base *Database) newDiskPage (pageSize int, fac DataFac, pos *FilePos) Data{
	*pos = base.allocateBase(pageSize)
	return base.newPage(pageSize, fac, *pos)
}

// Erase the page on disk at position pos and its buffer page
func (base *Database) eraseDiskPage (pos FilePos) {
	base.erasePage(pos)
	base.deleteBase(pos)
}

// For an already existing normal page (not cluster head nor tail page) on disk at position pos, promote and possibly create its buffer page and return it
func (base *Database) selectPage (pos FilePos) *pageT {
	var (p *pageT; ok bool)
	if p, ok = base.findPage(pos); ok {
		base.promotePage(p)
	} else {
		p = base.createPage(base.readBaseLength(pos), pos)
	}
	return p
}

// Find in buffer the size of the normal page (not cluster head nor tail page) at position pos and return it
func (base *Database) readPageLengthA (pos FilePos) int {
	size := base.selectPage(pos).sizeP
	base.releaseA(pos)
	return size
}

// Find in buffer the size of the normal page (not cluster head nor tail page) at position pos and return it
func (base *Database) readPageLength (pos FilePos) int {
	base.readPLIn <- pos
	return <-base.readPLOut
}

// Find, or create and fill if not found, the normal buffer page (not cluster head nor tail page) at position pos; creation of data, if any, is made with the help of fac
func (base *Database) readPageA (pos FilePos, fac DataFac) Data {
	p := base.selectPage(pos)
	if p.pageP == nil {
		p.pageP = fac.New(p.sizeP)
		a := base.readBase(pos, p.sizeP)
		var r Reader
		r.initReader(base.ref, a, 0)
		p.pageP.Read(&r)
	}
	return p.pageP
}

// Find, or create and fill if not found, the normal buffer page (not cluster head nor tail page) at position pos; creation of data, if any, is made with the help of fac
func (base *Database) readPage (pos FilePos, fac DataFac) Data {
	base.readPIn <- &fpData{pos: pos, fac: fac}
	return <-base.readPOut
}

// Read user data m in the database of m at position ptr. *)
func (m *DataMan) ReadData (ptr FilePos) Data {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(ptr != BNil, 21)
	var w Writer
	w.initWriter(m.base.ref) // Make a copy
	m.base.readPage(ptr, m.fac).Write(&w)
	m.base.release(ptr)
	a := w.write()
	var r Reader
	r.initReader(m.base.ref, a, 0)
	pa := m.fac.New(len(a))
	pa.Read(&r)
	return pa
}

// Consider using AllocateData or, better, WriteAllocateData instead. Allocate a cluster of size size, managed by m, in the database of m, and return its position. *)
func (m *DataMan) AllocateSize (size int) FilePos {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(size > 0, 21)
	var pos FilePos
	_ = m.base.newDiskPage(size, m.fac, &pos)
	m.base.release(pos)
	return pos
}

// Consider using WriteAllocateData instead. Allocate a cluster for data, managed by m, in the database of m, and return its position. Warning: AllocateData calls data.Write to find the size of data, be sure the value of data can be written and has its correct size.
func (m *DataMan) AllocateData (data Data) FilePos {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(data != nil, 21)
	var w Writer
	w.initWriter(m.base.ref)
	data.Write(&w)
	return m.AllocateSize(w.s.NumberOfElems())
}

// Allocate a cluster for data, managed by m, in the database of m, and write data into it; return the position of the allocated cluster.
func (m *DataMan) WriteAllocateData (data Data) FilePos {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(data != nil, 21)
	var w Writer
	w.initWriter(m.base.ref)
	data.Write(&w)
	a := w.write()
	var ptr FilePos
	pa := m.base.newDiskPage(len(a), m.fac, &ptr)
	var r Reader
	r.initReader(m.base.ref, a, 0)
	pa.Read(&r)
	m.base.writePage(ptr)
	m.base.release(ptr)
	return ptr
}

// Write data, managed by m, at the position ptr in the database of m.
func (m *DataMan) WriteData (ptr FilePos, data Data) {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(data != nil, 21)
	M.Assert(ptr != BNil, 22)
	var w Writer
	w.initWriter(m.base.ref)
	data.Write(&w)
	var r Reader
	r.initReader(m.base.ref, w.write(), 0)
	m.base.readPage(ptr, m.fac).Read(&r)
	m.base.writePage(ptr)
	m.base.release(ptr)
}

// Erase data managed by m at position ptr in the database of m.
func (m *DataMan) EraseData (ptr FilePos) {
	M.Assert(m.base.ref != nil, 20)
	M.Assert(ptr != BNil, 21)
	m.base.eraseDiskPage(ptr)
}

func (pa *pageI) Write (w *Writer) {
	w.OutFilePos(pa.keys)
	w.OutInt8(pa.elNb)
	for i := 0; i <= int(pa.elNb); i++ {
		w.OutFilePos(pa.elems[i].ptr)
		w.OutInt16(pa.elems[i].endK)
	}
}

func (pa *pageI) Read (r *Reader) {
	pa.keys = r.InFilePos()
	pa.elNb = r.InInt8()
	for i := 0; i <= int(pa.elNb); i++ {
		pa.elems[i].ptr = r.InFilePos()
		pa.elems[i].endK = r.InInt16()
	}
}

func (f pageIFac) New (sz int) Data {
	M.Assert(sz <= pageIS, 114)
	return new(pageI)
}

func (c *keyCont) Write (w *Writer) {
	w.OutBytes(Bytes(c.c))
}

func (c *keyCont) Read (r *Reader) {
	c.c = keyT(r.InBytes(len(c.c)))
}

func (f keyFac) New (sz int) Data {
	M.Assert(sz >= 0, 115)
	pa := new(keyCont)
	pa.c = make(keyT, sz)
	return pa
}

func (c *stringI) Write (w *Writer) {
	w.OutFilePos(c.next)
	w.OutFilePos(c.prev)
	w.OutFilePos(c.dataPos)
	w.OutBytes(Bytes(c.key))
}

func (c *stringI) Read (r *Reader) {
	c.next = r.InFilePos()
	c.prev = r.InFilePos()
	c.dataPos = r.InFilePos()
	c.key = keyT(r.InBytes(len(c.key)))
}

func (f stringIFac) New (sz int) Data {
	M.Assert(sz >= stringIS, 20)
	pa := new(stringI)
	sz -= stringIS
	pa.key = make(keyT, sz)
	return pa
}

func (i *info) Write (w *Writer) {
	w.OutInt32(guard1)
	w.OutFilePos(i.iRoot)
	w.OutFilePos(i.iString)
	w.OutInt32(int32(i.iKSize))
	w.OutInt32(int32(i.iH))
	w.OutInt32(int32(i.iSize))
	w.OutInt32(guard2)
}

func (i *info) Read (r *Reader) {
	M.Assert(r.InInt32() == guard1, 116)
	i.iRoot = r.InFilePos()
	i.iString = r.InFilePos()
	i.iKSize = int(r.InInt32())
	i.iH = int(r.InInt32())
	i.iSize = int(r.InInt32())
	M.Assert(r.InInt32() == guard2, 116)
}

func (f infoFac) New (sz int) Data {
	M.Assert(sz == infoS, 117, sz);
	return new(info)
}

// Make a KeyManager out of a KeyManagerer.
func MakeKM (k KeyManagerer) *KeyManager {
	km := new(KeyManager)
	km.KeyManagerer = k
	return km
}

// Create a KeyManeger for String(s).
func StringKeyManager () *KeyManager {
	return MakeKM (new(stringKeyManager))
}

// Create a new index in the database base and return its reference, but do not open it. keySize is the size of keys. If keys have a fixed size, put this size in keySize. If the size of keys does not vary much, put the greatest size in keySize. If the greatest size of keys is unknown, or if this size vary much, or if you want to use prefixes, put zero in keySize: you'll have to fix the size of each key later.
func (base *Database) CreateIndex (keySize int) FilePos {
	M.Assert(base.ref != nil, 20)
	M.Assert(keySize >= 0, 21)
	var pos FilePos
	p := base.newDiskPage(infoS, inff, &pos).(*info)
	p.iKSize = keySize
	p.iH = 0
	p.iSize = 0
	c := base.newDiskPage(stringIS, sif, &p.iString).(*stringI)
	p.iRoot = p.iString
	c.next = p.iString
	c.prev = p.iString
	c.dataPos = BNil
	base.writePage(p.iString)
	base.release(p.iString)
	base.writePage(pos)
	base.release(pos)
	return pos
}

// Open and return an index created with the reference ref by Database.CreateIndex in the database base. man is the key manager of the index and f is the factory of its keys. The current position of the index is reset.
func (base *Database) OpenIndex (ref FilePos, man *KeyManager, f DataFac) *Index {
	M.Assert(base.ref != nil, 20)
	M.Assert(man != nil, 21)
	M.Assert(ref != BNil, 22)
	p := base.readPage(ref, inff).(*info)
	ind := new(Index)
	ind.rootI = p.iRoot
	ind.stringI = p.iString
	ind.keySize = p.iKSize
	ind.height = p.iH
	ind.size = p.iSize
	base.release(ref)
	ind.baseI = base
	ind.refI = ref
	man.f = f
	ind.manager = man
	ind.writer = &IndexWriter{IndexReader{ind: ind, posI: ind.stringI}}
	if ind.keySize == 0 {
		ind.keysSize = sectSize
	} else {
		ind.keysSize = (2 * minEl + 1) * ind.keySize
	}
	return ind
}

func (base *Database) indDel (pos FilePos, h int) {
	if h > 0 {
		pa := base.readPage(pos, pif).(*pageI)
		base.eraseDiskPage(pa.keys)
		for i := 0; i <= int(pa.elNb); i++ {
			base.indDel(pa.elems[i].ptr, h - 1)
		}
	}
	base.eraseDiskPage(pos)
}

// Delete the index at position ref in the database base; this index must have been closed.
func (base *Database) DeleteIndex (ref FilePos) {
	M.Assert(base.ref != nil, 20)
	M.Assert(ref != BNil, 22)
	p := base.readPage(ref, inff).(*info)
	base.indDel(p.iRoot, p.iH)
	a := make(Bytes, infoS)
	for i := 0; i < infoS; i++ {
		a[i] = 0
	}
	base.writeBase(ref, a) // Erase the two "guards"
	base.eraseDiskPage(ref)
}

// Return the height of 'ind'
func (ind *Index) Height () int {
	return ind.height
}

// Create a new Reader for ind. A Reader keeps a position in an Index and can update this position or read values there.
func (ind *Index) NewReader () *IndexReader {
	return &IndexReader{ind: ind, posI: ind.stringI}
}

// Duplicate ir.
func (ir *IndexReader) Clone () *IndexReader {
	return &IndexReader{ind: ir.ind, posI: ir.posI}
}

// Return the Writer of ind. A Writer keeps a position in an Index and can update this position or read/write values there.
func (ind *Index) Writer () *IndexWriter {
	return ind.writer
}

// Calculate the prefix of c2 relative to c1, with the help of man.PrefP, and return it as a Key in p; l is the length of p (in bytes); ref is used for real <-> bytes translations
func (man *KeyManager) prefix (ref *File, c1, c2 Data) (p keyT, l int) {
	man.PrefP(c1, &c2)
	var w Writer
	w.initWriter(ref)
	c2.Write(&w)
	p = keyT(w.write())
	l = len(p)
	return
}

// Update the root of ind's btree on disk
func (ind *Index) updateRoot () {
	p := ind.baseI.readPage(ind.refI, inff).(*info)
	p.iRoot = ind.rootI
	p.iH = ind.height
	p.iSize = ind.size
	ind.baseI.writePage(ind.refI)
	ind.baseI.release(ind.refI)
}

// Change the allocation of p.keys when its content is increased by extra and becomes greater than its size 
func (ind *Index) adjustKeysSize (p *pageI, extra int) {
	l1 := ind.baseI.readPageLength(p.keys) // Size
	l := int(p.elems[p.elNb].endK) // Initial content size
	l2 := l + extra // New content size
	if l2 > l1 {
		l1 = (l2 + sectSize - 1) / sectSize * sectSize
		M.Assert(l1 <= M.MaxInt16 + 1, 100)
		var nKeys FilePos
		nK := ind.baseI.newDiskPage(l1, kf, &nKeys).(*keyCont)
		nC := nK.c
		k := ind.baseI.readPage(p.keys, kf,).(*keyCont)
		c := k.c
		for i := 0; i < l; i++ {
			nC[i] = c[i]
		}
		ind.baseI.writePage(nKeys)
		ind.baseI.release(nKeys)
		ind.baseI.eraseDiskPage(p.keys)
		p.keys = nKeys
	}
}

// Transfer nb successive elements from pa1 to pa2, starting at rank src in pa1 and at rank dst in pa2
func (ind *Index) transInter (pa1, pa2 *pageI, src, dst, nb int) {
	lC := int(pa1.elems[src + nb - 1].endK) - int(pa1.elems[src - 1].endK)
	ind.adjustKeysSize(pa2, lC)
	cl1 := ind.baseI.readPage(pa1.keys, kf, ).(*keyCont)
	cl2 := ind.baseI.readPage(pa2.keys, kf).(*keyCont)
	for i := int(pa2.elems[pa2.elNb].endK) - 1; i >= int(pa2.elems[dst - 1].endK); i-- {
		cl2.c[i + lC] = cl2.c[i]
	}
	u := int(pa2.elems[dst - 1].endK); v := int(pa1.elems[src - 1].endK)
	for i := 0; i < lC; i++ {
		cl2.c[i + u] = cl1.c[i + v]
	}
	ind.baseI.writePage(pa2.keys)
	ind.baseI.release(pa2.keys)
	diff := int(pa1.elems[src + nb - 1].endK) - int(pa1.elems[src - 1].endK)
	for i := int(pa1.elems[src + nb - 1].endK); i < int(pa1.elems[pa1.elNb].endK); i++ {
		cl1.c[i - diff] = cl1.c[i]
	}
	ind.baseI.writePage(pa1.keys)
	ind.baseI.release(pa1.keys)
	for i := int(pa2.elNb); i >= dst; i-- {
		pa2.elems[i].endK += int16(lC)
		pa2.elems[i + nb] = pa2.elems[i]
	}
	diff2 := pa2.elems[dst - 1].endK - pa1.elems[src - 1].endK
	for i := src; i < src + nb; i++ {
		pa1.elems[i].endK += diff2
	}
	for i := 0; i < nb; i++ {
		pa2.elems[dst + i] = pa1.elems[src + i]
	}
	for i := src + nb; i <= int(pa1.elNb); i++ {
		pa1.elems[i].endK -= int16(lC)
		pa1.elems[i - nb] = pa1.elems[i]
	}
	pa2.elNb += int8(nb)
	pa1.elNb -= int8(nb)
}

// Add, in the page pa of  the index ind, an element of rank dst corresponding to the key c of length lC and pointing to the page of position el (pageI inside index or stringI at its bottom)
func (ind *Index) transIn (pa *pageI, el FilePos, c keyT, lC, dst int) {
	ind.adjustKeysSize(pa, lC)
	k := ind.baseI.readPage(pa.keys, kf).(*keyCont)
	for i := int(pa.elems[pa.elNb].endK) - 1; i >= int(pa.elems[dst - 1].endK); i-- {
		k.c[i + lC] = k.c[i]
	}
	for i := 0; i < lC; i++ {
		k.c[int(pa.elems[dst - 1].endK) + i] = c[i]
	}
	ind.baseI.writePage(pa.keys)
	ind.baseI.release(pa.keys)
	for i := int(pa.elNb); i >= dst; i-- {
		pa.elems[i].endK += int16(lC)
		pa.elems[i + 1] = pa.elems[i]
	}
	pa.elems[dst].ptr = el
	pa.elems[dst].endK = pa.elems[dst - 1].endK + int16(lC)
	pa.elNb++
}

// Substract, in the page pa of  the index ind, an element of rank src corresponding to the key c of length lC and pointing to the page of position el (pageI inside index or stringI at its bottom)
func (ind *Index) transOut (pa *pageI, src int) (el FilePos, c keyT, lC int) {
	el = pa.elems[src].ptr
	lC = int(pa.elems[src].endK) - int(pa.elems[src - 1].endK)
	c = make(keyT, lC)
	k := ind.baseI.readPage(pa.keys, kf).(*keyCont)
	for i := 0; i < lC; i++ {
		c[i] = k.c[int(pa.elems[src - 1].endK) + i]
	}
	for i := int(pa.elems[src].endK); i < int(pa.elems[pa.elNb].endK); i++ {
		k.c[i - lC] = k.c[i]
	}
	ind.baseI.writePage(pa.keys)
	ind.baseI.release(pa.keys)
	for i := src + 1; i <= int(pa.elNb); i++ {
		pa.elems[i].endK -= int16(lC)
		pa.elems[i - 1] = pa.elems[i]
	}
	pa.elNb--
	return
}

// Return the Index of ir.
func (ir *IndexReader) Ind () *Index {
	return ir.ind
}

// Search in the page of address p the key key (keyA in Bytes form); father is the address of the father of p, fatherNum is the rank of p in father, h is the height of p; at exit, if inc (tree has grown), el is the new link to insert (pageI inside index or stringI at its bottom), c is the new key (or prefix) to insert and lC is its length
func (iw *IndexWriter) searchIns (key Data, keyA keyT, p, father FilePos, fatherNum, h int) (found, inc bool, el FilePos, lC int, c keyT) {
	ind := iw.ind
	if h == 0 { // At the bottom of btree
		s := ind.baseI.readPage(p, sif).(*stringI)
		var (data Data; comp Comp)
		if p == ind.stringI { // Null length key
			data = ind.manager.f.New(0)
			comp = Gt
		} else {
			data = ind.manager.f.New(ind.baseI.readPageLength(p) - stringIS)
			var r Reader
			r.initReader(ind.baseI.ref, Bytes(s.key), 0)
			data.Read(&r)
			comp = ind.manager.CompP(key, data)
		}
		found = comp == Eq
		inc = !found
		if inc { // New key: insert it...
			nS := ind.baseI.newDiskPage(stringIS + len(keyA), sif, &el).(*stringI)
			iw.posI = el
			nS.dataPos = BNil
			M.Assert(len(nS.key) == len(keyA), 118)
			nS.key = keyA
			if comp == Lt { // ... before...
				nS.next = p
				nS.prev = s.prev
				ind.baseI.writePage(el)
				ind.baseI.release(el)
				pos := s.prev
				s.prev = el
				ind.baseI.writePage(p)
				nS = ind.baseI.readPage(pos, sif).(*stringI)
				nS.next = el
				ind.baseI.writePage(pos)
				ind.baseI.release(pos)
				if father != BNil {
					pa := ind.baseI.readPage(father, pif).(*pageI)
					pa.elems[fatherNum].ptr = el
					ind.baseI.writePage(father)
					ind.baseI.release(father)
				}
				el = p
				c, lC = ind.manager.prefix(ind.baseI.ref, key, data)
			} else { // ... or after
				nS.next = s.next
				nS.prev = p
				ind.baseI.writePage(el)
				ind.baseI.release(el)
				pos := s.next
				s.next = el
				ind.baseI.writePage(p)
				nS = ind.baseI.readPage(pos, sif).(*stringI)
				nS.prev = el
				ind.baseI.writePage(pos)
				ind.baseI.release(pos)
				c, lC = ind.manager.prefix(ind.baseI.ref, data, key)
			}
		} else { // Old key
			iw.posI = p
		}
		ind.baseI.release(p)
	} else { // Inside btree
		pa := ind.baseI.readPage(p, pif).(*pageI)
		kC := ind.baseI.readPage(pa.keys, kf).(*keyCont)
		lft := 1; rgt := int(pa.elNb) + 1
		// rgt > pa.elNb || pa.elems[rgt].'key' > key
		for lft < rgt {
			i := (lft + rgt) / 2
			data := ind.manager.f.New(int(pa.elems[i].endK) - int(pa.elems[i - 1].endK))
			var r Reader
			r.initReader(ind.baseI.ref, Bytes(kC.c), int(pa.elems[i - 1].endK))
			data.Read(&r)
			if ind.manager.CompP(key, data) == Lt {
				rgt = i
			} else {
				lft = i + 1
			}
		}
		ind.baseI.release(pa.keys)
		rgt--
		// rgt == 0 || pa.elems[rgt].'key' <= key
		var (elRet FilePos; lCRet int; cRet keyT)
		found, inc, elRet, lCRet, cRet = iw.searchIns(key, keyA, pa.elems[rgt].ptr, p, rgt, h - 1)
		if inc { // New link to be inserted
			ind.transIn(pa, elRet, cRet, lCRet, rgt + 1) // Insert it
			if pa.elNb <= 2 * minEl { // That's all!
				inc = false
			} else { // Too much links...
				n := 1 // Number of pages to consider
				m := 2 * minEl + 1 // Total number of elements in these pages
				seen1 := false; seen2 := false
				save1 := false; save2 := false
				var (lftP, rgtP FilePos; paP, pa1, pa2 *pageI)
				if father != BNil { // ... try first to tranfer links to adjacent pages
					paP = ind.baseI.readPage(father, pif).(*pageI)
					if fatherNum > 0 { // Previous page
						n++
						lftP = paP.elems[fatherNum - 1].ptr
						pa1 = ind.baseI.readPage(lftP, pif).(*pageI)
						m += int(pa1.elNb)
						seen1 = true;
					}
					if fatherNum < int(paP.elNb) { // Next page
						n++
						rgtP = paP.elems[fatherNum + 1].ptr
						pa2 = ind.baseI.readPage(rgtP, pif).(*pageI)
						m += int(pa2.elNb)
						seen2 = true
					}
				}
				if m <= 2 * n * minEl { // Transfer possible: do it...
					inc = false // ... and stop going up!
					m = (m + n - 2) / n // minEl <= m <= 2 * minEl and there is at least one element to transfer our of pa and the number of elements transferred is minimal
					// Transfers out of pa, which is full
					if fatherNum > 0 && m > int(pa1.elNb) { // Previous page
						el, c, lC = ind.transOut(paP, fatherNum)
						ind.transIn(pa1, pa.elems[0].ptr, c, lC, int(pa1.elNb) + 1)
						if m > int(pa1.elNb) {
							ind.transInter(pa, pa1, 1, int(pa1.elNb) + 1, m - int(pa1.elNb))
						}
						el, c, lC = ind.transOut(pa, 1)
						ind.transIn(paP, p, c, lC, fatherNum)
						pa.elems[0].ptr = el
						save1 = true
					}
					if fatherNum < int(paP.elNb) && m > int(pa2.elNb) { // Next page
						el, c, lC = ind.transOut(paP, fatherNum + 1)
						ind.transIn(pa2, pa2.elems[0].ptr, c, lC, 1)
						if m > int(pa2.elNb) {
							ind.transInter(pa, pa2, int(pa.elNb) + int(pa2.elNb) - m + 1, 1, m - int(pa2.elNb))
						}
						var nEl FilePos
						nEl, c, lC = ind.transOut(pa, int(pa.elNb))
						ind.transIn(paP, el, c, lC, fatherNum + 1)
						pa2.elems[0].ptr = nEl
						save2 = true
					}
					// Possible tranfers to pa, which is no more full
					if fatherNum > 0 && m < int(pa1.elNb) { // Previous page
						el, c, lC = ind.transOut(paP, fatherNum)
						ind.transIn(pa, pa.elems[0].ptr, c, lC, 1)
						if m + 1 < int(pa1.elNb) {
							ind.transInter(pa1, pa, m + 2, 1, int(pa1.elNb) - m - 1)
						}
						el, c, lC = ind.transOut(pa1, m + 1)
						ind.transIn(paP, p, c, lC, fatherNum)
						pa.elems[0].ptr = el
						save1 = true
					}
					if fatherNum < int(paP.elNb) && m < int(pa2.elNb) { // Next page
						el, c, lC = ind.transOut(paP, fatherNum + 1)
						ind.transIn(pa, pa2.elems[0].ptr, c, lC, int(pa.elNb) + 1)
						if m + 1 < int(pa2.elNb) {
							ind.transInter(pa2, pa, 1, int(pa.elNb) + 1, int(pa2.elNb) - m - 1)
						}
						var nEl FilePos
						nEl, c, lC = ind.transOut(pa2, 1)
						ind.transIn(paP, el, c, lC, fatherNum + 1)
						pa2.elems[0].ptr = nEl
						save2 = true
					}
					ind.baseI.writePage(father)
				} else { // Transfer impossible: split page, and up one level
					var nEl FilePos
					nEl, c, lC = ind.transOut(pa, minEl + 1)
					pa3 := ind.baseI.newDiskPage(pageIS, pif, &el).(*pageI)
					_ = ind.baseI.newDiskPage(ind.keysSize, kf, &pa3.keys)
					ind.baseI.release(pa3.keys)
					pa3.elNb = 0
					pa3.elems[0].ptr = nEl
					pa3.elems[0].endK = 0
					ind.transInter(pa, pa3, minEl + 1, 1, minEl)
					ind.baseI.writePage(el)
					ind.baseI.release(el)
				}
				if save1 {
					ind.baseI.writePage(lftP)
				}
				if save2 {
					ind.baseI.writePage(rgtP)
				}
				if seen1 {
					ind.baseI.release(lftP)
				}
				if seen2 {
					ind.baseI.release(rgtP)
				}
				ind.baseI.release(father)
			}
			ind.baseI.writePage(p)
		}
		ind.baseI.release(p)
	}
	return
}

// Seek in the index of iw the key key and insert it if it is not there already. The result indicates if the key was found. Fix the current position of iw on the key found or inserted.
func (iw *IndexWriter) SearchIns (key Data) bool {
	ind := iw.ind
	M.Assert(key != nil, 20)
	var w Writer
	w.initWriter(ind.baseI.ref)
	key.Write(&w)
	keyA := w.write()
	found, inc, el, lC, c := iw.searchIns(key, keyT(keyA), ind.rootI, BNil, 0, ind.height)
	if !found {
		ind.size++
	}
	if inc { // Top page split: increase height of btree
		ind.height++
		p := ind.rootI
		page := ind.baseI.newDiskPage(pageIS, pif, &ind.rootI).(*pageI)
		_ = ind.baseI.newDiskPage(ind.keysSize, kf, &page.keys)
		ind.baseI.release(page.keys)
		page.elNb = 0
		page.elems[0].ptr = p
		page.elems[0].endK = 0
		ind.transIn(page, el, c, lC, 1)
		ind.baseI.writePage(ind.rootI)
		ind.baseI.release(ind.rootI)
	}
	if !found {
		ind.updateRoot()
	}
	return found
}

// Seek in the index of ir the key key. The result indicates if the key was found. Fix the current position of ir on the found key or on the key which is immediately after the sought key in the event of unfruitful search.
func (ir *IndexReader) Search (key Data) bool {
	ind := ir.Ind()
	M.Assert(key != nil, 20)
	p := ind.rootI
	for h := 1; h <= ind.height; h++ {
		pa := ind.baseI.readPage(p, pif).(*pageI)
		kC := ind.baseI.readPage(pa.keys, kf).(*keyCont)
		lft := 1; rgt := int(pa.elNb) + 1
		// rgt > pa.elNb || pa.elems[rgt].'key' > key
		for lft < rgt {
			i := (lft + rgt) / 2
			data := ind.manager.f.New(int(pa.elems[i].endK) - int(pa.elems[i - 1].endK))
			var r Reader
			r.initReader(ind.baseI.ref, Bytes(kC.c), int(pa.elems[i - 1].endK))
			data.Read(&r)
			if ind.manager.CompP(key, data) == Lt {
				rgt = i
			} else {
				lft = i + 1
			}
		}
		rgt--
		// rgt == 0 || pa.elems[rgt].'key' <= key
		ind.baseI.release(pa.keys)
		ind.baseI.release(p)
		p = pa.elems[rgt].ptr
	}
	s := ind.baseI.readPage(p, sif).(*stringI)
	var comp Comp
	if p == ind.stringI {
		comp = Gt
	} else {
		data := ind.manager.f.New(ind.baseI.readPageLength(p) - stringIS)
		var r Reader
		r.initReader(ind.baseI.ref, Bytes(s.key), 0)
		data.Read(&r)
		comp = ind.manager.CompP(key, data)
	}
	if comp == Gt {
		ir.posI = s.next
	} else {
		ir.posI = p
	}
	ind.baseI.release(p)
	return comp == Eq
}

// The number of elements in the page pointed by elems[num].ptr (in the page pointed by p) has become less than minEl; Fix fixes this problem
func (ind *Index) fix (p FilePos, num int, dec *bool) {
	paP := ind.baseI.readPage(p, pif).(*pageI) // Father page
	join := false // Look if it's possible to join the modified page with an adjacent page at the same level
	var (q1, q2 FilePos = BNil, BNil; pa1, pa2 *pageI; right bool)
	if num > 0 { // Look at the previous page
		q1 = paP.elems[num - 1].ptr
		pa1 = ind.baseI.readPage(q1, pif).(*pageI)
		join = pa1.elNb == minEl
		right = false
	}
	if !join && num < int(paP.elNb) { // or at the next one
		q2 = paP.elems[num + 1].ptr
		pa2 = ind.baseI.readPage(q2, pif).(*pageI)
		join = pa2.elNb == minEl
		right = true
	}
	q := paP.elems[num].ptr
	pa := ind.baseI.readPage(q, pif).(*pageI)
	if join { // Join pa and pa1 or pa2
		if right {
			el, c, lC := ind.transOut(paP, num + 1)
			ind.transIn(pa2, pa2.elems[0].ptr, c, lC, 1)
			ind.transInter(pa, pa2, 1, 1, int(pa.elNb))
			pa2.elems[0].ptr = pa.elems[0].ptr
			paP.elems[num].ptr = el
			ind.baseI.writePage(q2)
		} else {
			_, c, lC := ind.transOut(paP, num)
			ind.transIn(pa1, pa.elems[0].ptr, c, lC, int(pa1.elNb) + 1)
			ind.transInter(pa, pa1, 1, int(pa1.elNb) + 1, int(pa.elNb))
			ind.baseI.writePage(q1)
		}
		M.Assert(pa.elNb == 0, 60) // Everything is in pa1
		ind.baseI.eraseDiskPage(pa.keys)
		ind.baseI.eraseDiskPage(q)
		*dec = paP.elNb < minEl // The father page may underflow too
	} else { // Transfer one element from pa1 to pa via paP
		*dec = false
		if right {
			el, c, lC := ind.transOut(paP, num + 1)
			ind.transIn(pa, pa2.elems[0].ptr, c, lC, int(pa.elNb) + 1)
			nEl, c, lC := ind.transOut(pa2, 1)
			ind.transIn(paP, el, c, lC, num + 1)
			pa2.elems[0].ptr = nEl
			ind.baseI.writePage(q2)
		} else {
			el, c, lC := ind.transOut(paP, num)
			ind.transIn(pa, pa.elems[0].ptr, c, lC, 1)
			el, c, lC = ind.transOut(pa1, int(pa1.elNb))
			ind.transIn(paP, q, c, lC, num)
			pa.elems[0].ptr = el
			ind.baseI.writePage(q1)
		}
		ind.baseI.writePage(q)
		ind.baseI.release(q)
	}
	ind.baseI.release(q1)
	ind.baseI.release(q2)
	ind.baseI.writePage(p)
	ind.baseI.release(p)
}

// Erase the element of the last key at level 1 in the subtree pointed by q, and put its key in the page pointed by p at position rgt; h is the current level
func (ind *Index) del (p, q FilePos, h, rgt int, dec *bool) {
	pa := ind.baseI.readPage(q, pif).(*pageI)
	if h == 1 {
		el, c, lC := ind.transOut(pa, int(pa.elNb))
		*dec = pa.elNb < minEl
		ind.baseI.writePage(q)
		pa = ind.baseI.readPage(p, pif).(*pageI)
		el, _, _ = ind.transOut(pa, rgt)
		ind.transIn(pa, el, c, lC, rgt)
		ind.baseI.writePage(p)
		ind.baseI.release(p)
	} else {
		ind.del(p, pa.elems[pa.elNb].ptr, h - 1, rgt, dec)
		if *dec {
			ind.fix(q, int(pa.elNb), dec)
		}
	}
	ind.baseI.release(q)
}

// Erase key in the subtree pointed by p; father is the father of p, fatherNum is the rank of p in father and h is the height of the subtree *)
func (iw *IndexWriter) era (key Data, p, father FilePos, fatherNum, h int, dec, minus1, found *bool, followP *FilePos, followNum *int) {
	
	ind := iw.ind
	
	// Compare c with the numth key (or prefix) of pa and return TRUE if they are equal *)
	equalKeys := func (c keyT, pa *pageI, num int) bool {
		l := len(c)
		m := int(pa.elems[num].endK) - int(pa.elems[num - 1].endK)
		if l != m {
			return false
		}
		if l > 0 {
			kC := ind.baseI.readPage(pa.keys, kf).(*keyCont)
			cc := kC.c
			m = int(pa.elems[num - 1].endK)
			for i := 0; i < l; i++ {
				if c[i] != cc[m + i] {
					ind.baseI.release(pa.keys)
					return false
				}
			}
			ind.baseI.release(pa.keys)
		}
		return true
	}
	
	if h == 0 { // At the bottom of btree
		*dec = false
		var pP, pN FilePos
		if p == ind.stringI { // Null key
			*minus1 = false
		} else { // Read the value into data1 ...
			s := ind.baseI.readPage(p, sif).(*stringI)
			data1 := ind.manager.f.New(ind.baseI.readPageLength(p) - stringIS)
			var r Reader
			r.initReader(ind.baseI.ref, Bytes(s.key), 0)
			data1.Read(&r)
			*minus1 = ind.manager.CompP(key, data1) == Eq // ... and compare with the key
			if *minus1 {
				pP = s.prev
				pN = s.next
			}
			ind.baseI.release(p)
		}
		*found = *minus1
		if *minus1 { // Key was minus1, delete it
			// Remove page from the ring ind.stringI
			sN := ind.baseI.readPage(pN, sif).(*stringI)
			sN.prev = pP
			ind.baseI.writePage(pN)
			sP := ind.baseI.readPage(pP, sif).(*stringI)
			sP.next = pN
			ind.baseI.writePage(pP)
			if fatherNum == 0 && father != BNil {
				pa := ind.baseI.readPage(father, pif).(*pageI)
				pa.elems[fatherNum].ptr = pP // Change this link to the previous stringI, since the corrresponding key (or prefix) will be deleted
				ind.baseI.writePage(father)
				ind.baseI.release(father)
			}
			if *followP != BNil { // If there are following keys (or prefixes), correct the following prefix, which may have changed, since its previous key has changed
				data1 := ind.manager.f.New(ind.baseI.readPageLength(pP) - stringIS)
				if pP != ind.stringI {
					var r Reader
					r.initReader(ind.baseI.ref, Bytes(sP.key), 0)
					data1.Read(&r)
				}
				data2 := ind.manager.f.New(ind.baseI.readPageLength(pN) - stringIS)
				if pN != ind.stringI {
					var r Reader
					r.initReader(ind.baseI.ref, Bytes(sN.key), 0)
					data2.Read(&r)
				}
				c, lC := ind.manager.prefix(ind.baseI.ref, data1, data2)
				pa := ind.baseI.readPage(*followP, pif).(*pageI)
				if !equalKeys(c, pa, *followNum) {
					el, _, _ := ind.transOut(pa, *followNum)
					ind.transIn(pa, el, c, lC, *followNum)
					ind.baseI.writePage(*followP)
				}
				ind.baseI.release(*followP)
			}
			ind.baseI.release(pN)
			ind.baseI.release(pP)
			if iw.posI == p { // Reset the position of the index if it was on the erased key
				iw.posI = ind.stringI
			}
			ind.baseI.eraseDiskPage(p)
		}
	} else { // Inside btree
		pa := ind.baseI.readPage(p, pif).(*pageI)
		kC := ind.baseI.readPage(pa.keys, kf).(*keyCont)
		// Find the position of key in pa (binary search)
		lft := 1; rgt := int(pa.elNb) + 1
		// rgt > pa.elNb || pa.elems[rgt].'key' > key
		for lft < rgt {
			i := (lft + rgt) / 2
			data1 := ind.manager.f.New(int(pa.elems[i].endK) - int(pa.elems[i - 1].endK))
			var r Reader
			r.initReader(ind.baseI.ref, Bytes(kC.c), int(pa.elems[i - 1].endK))
			data1.Read(&r)
			if ind.manager.CompP(key, data1) == Lt {
				rgt = i
			} else {
				lft = i + 1
			}
		}
		ind.baseI.release(pa.keys)
		rgt--
		// rgt == 0 || pa.elems[rgt].'key' <= key
		if rgt < int(pa.elNb) {
			*followP = p
			*followNum = rgt + 1
		}
		iw.era(key, pa.elems[rgt].ptr, p, rgt, h - 1, dec, minus1, found, followP, followNum)
		if *minus1 {
			// If rgt == 0, the corresponding key (or prefix) is upper in the tree and will be deleted later (unless it is ind.stringI)
			if rgt > 0 {
				if h == 1 { // At the lower level, the key can be erased directly
					_, _, _ = ind.transOut(pa, rgt)
					*dec = pa.elNb < minEl
					ind.baseI.writePage(p)
				} else { // At an upper level, erase the element of the previous key at level 1 and put this previous key in place of this one
					ind.del(p, pa.elems[rgt - 1].ptr, h - 1, rgt, dec)
					if *dec {
						ind.fix(p, rgt - 1, dec)
					}
				}
				*minus1 = false // key has already been found and erased
			}
		} else if *dec {
			ind.fix(p, rgt, dec)
		}
		ind.baseI.release(p)
	}
}

// Erase from the index of iw the key key. If key does not belong to the index, ind.Erase does nothing and returns false, otherwise it returns true. If the current position of iw was on the erased key, it is reset.
func (iw *IndexWriter) Erase (key Data) bool {
	ind := iw.ind
	M.Assert(key != nil, 20)
	followP := BNil
	var (dec, minus1, found bool; followNum int)
	iw.era(key, ind.rootI, BNil, 0, ind.height, &dec, &minus1, &found, &followP, &followNum)
	if dec { // Subtree has been erased...
		pa := ind.baseI.readPage(ind.rootI, pif).(*pageI)
		if pa.elNb == 0 { //... if it was the last, decrease the height of subtree
			ind.height--
			ind.baseI.eraseDiskPage(pa.keys)
			ind.baseI.eraseDiskPage(ind.rootI)
			ind.rootI = pa.elems[0].ptr
		} else {
			ind.baseI.release(ind.rootI)
		}
	}
	if found {
		ind.size--
		ind.updateRoot()
	}
	return found
}

// Test the emptiness of ind
func (ind *Index) IsEmpty () bool {
	M.Assert(ind.baseI != nil, 20)
	return ind.size == 0
}

// Return the number of different keys in ind
func (ind *Index) NumberOfKeys () int {
	M.Assert(ind.baseI != nil, 20)
	return ind.size
}

// Reset the current position of ir. After this action, ir.PosSet() returns false.
func (ir *IndexReader) ResetPos () {
	ind := ir.Ind()
	ir.posI = ind.stringI
}

// Test if ir is positioned on a key (return true) or if its current position is reset (false returned).
func (ir *IndexReader) PosSet () bool {
	ind := ir.ind
	return ir.posI != ind.stringI
}

// Position ir on the next key. If the current position is reset, ir is positioned on the first key. If ir is positioned on the last key, its current position becomes reset.
func (ir *IndexReader) Next () {
	ind := ir.ind
	s := ind.baseI.readPage(ir.posI, sif).(*stringI)
	ind.baseI.release(ir.posI)
	ir.posI = s.next
}

// Position ir on the previous key. If the current position is reset, ir is positioned on the last key. If ir is positioned on the first key, its current position becomes reset.
func (ir *IndexReader) Previous () {
	ind := ir.Ind()
	s := ind.baseI.readPage(ir.posI, sif).(*stringI)
	ind.baseI.release(ir.posI)
	ir.posI = s.prev
}

// Return the value of the key in the current position of ir, or nil if the current position of ir is reset.
func (ir *IndexReader) CurrentKey () Data {
	ind := ir.Ind()
	if ir.posI == ind.stringI {
		return nil
	}
	s := ind.baseI.readPage(ir.posI, sif).(*stringI)
	data := ind.manager.f.New(ind.baseI.readPageLength(ir.posI) - stringIS)
	var r Reader
	r.initReader(ind.baseI.ref, Bytes(s.key), 0) // Make a copy
	ind.baseI.release(ir.posI)
	data.Read(&r)
	return data
}

// Read, at the current position of ir, the associated reference of the Data of its Index and return it. When a new key is inserted, the reference which is initially attached to it has the value BNil. The reset position of an index has, itself, an associated reference, legible by IndexReader.ReadValue.
func (ir *IndexReader) ReadValue () FilePos {
	ind := ir.Ind()
	s := ind.baseI.readPage(ir.posI, sif).(*stringI)
	ind.baseI.release(ir.posI)
	return s.dataPos
}

// Write on the current position of iw the associated reference val of a Data into its Index. When a new key is inserted, the data which is initially attached to it has the value BNil. The reset position of an index has, itself, a associated data, modifiable by IndexWriter.WriteValue.
func (iw *IndexWriter) WriteValue (val FilePos) {
	ind := iw.ind
	ind.baseI.readPage(iw.posI, sif).(*stringI).dataPos = val
	ind.baseI.writePage(iw.posI)
	ind.baseI.release(iw.posI)
}
