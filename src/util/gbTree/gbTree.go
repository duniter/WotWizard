/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package gbTree

// Implementation of bTree for Go. For use in Go, this package should be imported in place of util/bTree.

import (

	B	"util/bTree"
	F	"path/filepath"
		"os"
		"unsafe"

)

const (

	BNil = B.BNil
	Lt = B.Lt
	Eq = B.Eq
	Gt = B.Gt
	
	BOS = B.BOS
	BYS = B.BYS
	SIS = B.SIS
	INS = B.INS
	LIS = B.LIS
	SRS = B.SRS
	RES = B.RES

)

type (

	factorer struct {
	}

	filer struct {
		f *os.File
		rPos, wPos int64
	}
	
	File = B.File
	
	FilePos = B.FilePos
	
	Bytes = B.Bytes
	
	KeyManager = B.KeyManager
	
	KeyManagerer = B.KeyManagerer

	String = B.String

	StringFac = B.StringFac

	Database = B.Database

	IndexReader = B.IndexReader

	IndexWriter = B.IndexWriter

	Index = B.Index

	Data = B.Data

	DataFac = B.DataFac

	DataMan = B.DataMan

	Reader = B.Reader

	Writer = B.Writer
	
	Comp = B.Comp
	
	StringComparer = B.StringComparer

)

var (
	
	Fac *B.Factory
	fac factorer
	
)

func StringPrefP (key1, key2 *String, compare StringComparer) *String {
	return B.StringPrefP (key1, key2, compare)
}

func MakeKM (k KeyManagerer) *KeyManager {
	return B.MakeKM(k)
}

func StringKeyManager () *KeyManager {
	return B.StringKeyManager()
}

// Create a file with name nF. Don't open this file.
func (fac factorer) Create (nF string) bool {
	err := os.MkdirAll(F.Dir(nF), os.FileMode(0775))
	if err != nil {
		return false
	}
	f, err := os.Create (nF)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// Open a file with name nF and return this file. Return nil if the file does not exist.
func (fac factorer) Open (nF string) *File {
	f, err := os.OpenFile(nF, os.O_RDWR, 0777)
	if err != nil {
		return nil
	}
	fer := &filer{f: f}
	return &B.File{Filer: fer}
}

func init () {
	Fac = new(B.Factory)
	Fac.Factorer = fac
}

func (ref *filer) Close () {
	err := ref.f.Close()
	if err != nil {panic(20)}
}

func (ref *filer) Flush () {
	err := ref.f.Sync()
	if err != nil {panic(20)}
}

func (ref *filer) PosReader (pos FilePos) {
	ref.rPos = int64(pos)
}

func (ref *filer) Read (n int) Bytes {
	b := make(Bytes, n)
	_, err := ref.f.ReadAt(b, ref.rPos)
	ref.rPos += int64(n)
	if err != nil {panic(20)}
	return b
}

func (ref *filer) PosWriter (pos FilePos) {
	ref.wPos = int64(pos)
}

func (ref *filer) Write (b Bytes) {
	_, err := ref.f.WriteAt(b, ref.wPos)
	ref.wPos += int64(len(b))
	if err != nil {panic(20)}
}

func (ref *filer) End () FilePos {
	fi, err := ref.f.Stat()
	if err != nil {panic(20)}
	return FilePos(fi.Size())
}

func (ref *filer) Truncate (end FilePos) {
	err := ref.f.Truncate(int64(end))
	if err != nil {panic(20)}
}
	
const valOct = 0x100

func (ref *filer) Float64ToBytes (f float64) Bytes {
	var n uint64 = *(*uint64)(unsafe.Pointer(&f))
	s := int(unsafe.Sizeof(n))
	b := make(Bytes, s)
	for i := s - 1; i >= 0; i-- {
		b[i] = byte(n)
		n = n / valOct
	}
	return b
}

func (ref *filer) BytesToFloat64 (b Bytes, pos *int) float64 {
	var n uint64 = 0
	s := int(unsafe.Sizeof(n))
	for i := 0; i < s; i++ {
		n = n * valOct + uint64(b[*pos])
		(*pos)++
	}
	return *(*float64)(unsafe.Pointer(&n))
}

func (ref *filer) Float32ToBytes (f float32) Bytes {
	var n uint32 = *(*uint32)(unsafe.Pointer(&f))
	s := int(unsafe.Sizeof(n))
	b := make(Bytes, s)
	for i := s - 1; i >= 0; i-- {
		b[i] = byte(n)
		n = n / valOct
	}
	return b
}

func (ref *filer) BytesToFloat32 (b Bytes, pos *int) float32 {
	var n uint32 = 0
	s := int(unsafe.Sizeof(n))
	for i := 0; i < s; i++ {
		n = n * valOct + uint32(b[*pos])
		(*pos)++
	}
	return *(*float32)(unsafe.Pointer(&n))
}
