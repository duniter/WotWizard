/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package basic

import (
	
	A	"util/avl"
	F	"path/filepath"
	M	"util/misc"
	R	"util/resources"
		"bufio"
		"bytes"
		"errors"
		"flag"
		"fmt"
		"log"
		"os"
		"text/scanner"
		"strings"
		"unicode"

)

type (
	
	Comp = A.Comp

)

const (
	
	duniBaseDef = "$HOME/.config/duniter/duniter_default/wotwizard-export.db"
	
	// Name of the file where the path to the Duniter database is written
	initName = "init.txt"
	logName = "log.txt"
	logOldName = "log1.txt"
	minLogSize = 0x3200000 // 50MB
	
	serverDefaultAddress = "localhost:8080"
	serverAddressName = "serverAddress.txt"
	
	Never = M.MaxInt64 // In WotWizard window
	Revoked = M.MinInt64 // Limit date for revoked members
	Already = M.MinInt64 + 1 // Already available certification date
	
	Lt = A.Lt
	Eq = A.Eq
	Gt = A.Gt

)

var (
	
	Lg *log.Logger
	
	DuniDir,
	DuniBase string // Path to the Duniter database
	
	rsrcDir = F.Join(R.FindDir(), "duniter")
	initPath = F.Join(rsrcDir, initName)
	logPath = F.Join(rsrcDir, logName)
	logOldPath = F.Join(rsrcDir, logOldName)
	
	serverAddress = serverDefaultAddress

)

func RsrcDir () string {
	return rsrcDir
}

// Extract the significant characters in s; only alphanumeric characters are significant, and their case of lowest rank is returned
func ToDown (s string) string {
	rs := bytes.Runes([]byte(s))
	buf := new(bytes.Buffer)
	for _, r := range rs {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(M.Min32(unicode.ToUpper(r), unicode.ToLower(r)))
		}
	}
	return string(buf.Bytes())
} //ToDown

// Extract the significant characters in s; only alphanumeric characters are significant, and are returned
func strip (s string) string {
	rs := bytes.Runes([]byte(s))
	buf := new(bytes.Buffer)
	for _, r := range rs {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		}
	}
	return string(buf.Bytes())
} //strip

func compR (r1, r2 rune) Comp {
	ld1 := unicode.IsLetter(r1) || unicode.IsDigit(r1)
	ld2 := unicode.IsLetter(r2) || unicode.IsDigit(r2)
	if ld1 && !ld2 {
		return Lt
	}
	if !ld1 && ld2 {
		return Gt
	}
	if r1 < r2 {
		return Lt
	}
	if r1 > r2 {
		return Gt
	}
	return Eq
} //compR

func compS (s1, s2 string) Comp {
	rs1 := bytes.Runes([]byte(s1))
	rs2 := bytes.Runes([]byte(s2))
	l1 := len(rs1)
	l2 := len(rs2)
	l := M.Min(l1, l2)
	for i := 0; i < l; i++ {
		switch compR(rs1[i], rs2[i]) {
		case Lt:
			return Lt
		case Gt:
			return Gt
		case Eq:
		}
	}
	if l1 < l2 {
		return Lt
	}
	if l1 > l2 {
		return Gt
	}
	return Eq
} //compS

// Standard comparison procedure for identifiers; they are first compared with only significant characters and case ignored, and if still equal, with all characters and case taken into account
func CompP (s1, s2 string) Comp {
	ss1 := ToDown(s1); ss2 := ToDown(s2)
	if ss1 < ss2 {
		return Lt
	}
	if ss1 > ss2 {
		return Gt
	}
	ss1 = strip(s1); ss2 = strip(s2)
	if ss1 < ss2 {
		return Lt
	}
	if ss1 > ss2 {
		return Gt
	}
	return compS(s1, s2)
} //CompP

// Compare the characters of s1 and s2 and return whether the first one is a prefix of the second one or not
func Prefix (s1, s2 string) bool {
	return strings.HasPrefix(s2, s1)
} //Prefix

func setLog () {
	fi, err := os.Stat(logPath)
	M.Assert(err == nil || os.IsNotExist(err), err, 100)
	var f *os.File
	if err != nil || fi.Size() >= minLogSize {
		err := os.Remove(logOldPath)
		M.Assert(err == nil || os.IsNotExist(err), err, 101)
		err = os.Rename(logPath, logOldPath)
		M.Assert(err == nil || os.IsNotExist(err), err, 102)
		f, err = os.Create(logPath)
		M.Assert(err == nil, 103)
	} else {
		f, err = os.OpenFile(logPath, os.O_APPEND | os.O_WRONLY, 0644)
	}
	Lg = log.New(f, "", log.Ldate | log.Ltime | log.Lshortfile)
	M.SetLog(Lg)
} //setLog

// À vérifier : Est-ce que l'option -du fonctionne bien ?
func setDuniterPath () {
	
	storeDuniBase := func (du string) {
		DuniBase = du
		f, err := os.Create(initPath); M.Assert(err == nil, err, 100)
		_, err = f.WriteString(DuniBase); M.Assert(err == nil, err, 101)
		f.Close()
	}
	
	du := flag.String("du", "", "Path to the Duniter sql database")
	flag.Parse()
	ok := *du != "" && F.Ext(*du) == ".db"
	if ok {
		f, err := os.Open(*du)
		ok = err == nil
		if ok {
			f.Close()
		}
	}
	if ok {
		storeDuniBase(*du)
	} else {
		f, err := os.Open(initPath)
		if err == nil {
			sc := bufio.NewScanner(f)
			b := sc.Scan(); M.Assert(b, 100)
			DuniBase = sc.Text()
			b = sc.Scan(); M.Assert(!b && sc.Err() == nil, 101)
			f.Close()
		} else {
			storeDuniBase(os.ExpandEnv(duniBaseDef))
		}
	}
	DuniDir = F.Dir(DuniBase)
} //setDuniterPath

func ServerAddress () string {
	return serverAddress
}

func fixServerAddress () {
	name := F.Join(rsrcDir, serverAddressName)
	f, err := os.Open(name)
	if err == nil {
		defer f.Close()
		s := new(scanner.Scanner)
		s.Init(f)
		s.Error = func(s *scanner.Scanner, msg string) {panic(errors.New("File " + name + " incorrect"))}
		s.Mode = scanner.ScanStrings
		s.Scan()
		ss := s.TokenText()
		M.Assert(ss[0] == '"' && ss[len(ss) - 1] == '"', ss, 101)
		serverAddress = ss[1:len(ss) - 1]
	} else {
		f, err := os.Create(name)
		M.Assert(err == nil, err, 102)
		defer f.Close()
		fmt.Fprint(f, "\"" + serverAddress + "\"")
	}
}

func init () {
	err := os.MkdirAll(rsrcDir, 0777); M.Assert(err == nil, err, 100)
	setLog()
	setDuniterPath()
	fixServerAddress()
} //init
