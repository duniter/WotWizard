/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

// Package strMapping implements a mapping between keys and longer strings.
// Mappings are stored in multiline strings (may be stored themselves in files), each line defining a mapping: <key> tabulation <mapped string>.
package strMapping

import (
	
	F	"path/filepath"
	M	"util/misc"
	R	"util/resources"
		"bufio"
		"errors"
		"fmt"
		"io"
		"os"
		"strings"
		"strconv"
		"sync"
		"text/scanner"

)

const (
	
	strMappingDir = "util/strMapping"
	languageName = "language.txt"
	
	stringsName = "strings.txt"
	header = "STRINGS"
	sep = '\t'

)

var (
	
	langPath string

)

type (
	
	// LinkStrings functions return an io.ReadCloser able to read a string containing mappings for the language 'lang'.
	LinkStrings func (lang string) (io.ReadCloser, bool)
	
	mapScan struct {
		s *bufio.Scanner
		rac *dico
	}
	
	value interface{}
	
	dico struct {
		next *dico
		key string
		val value
	}

)

var (
	
	lStr = make(map[string]LinkStrings)
	wd = R.FindDir()
	dicos *dico
	language = "en"
	mut = new(sync.RWMutex)

)

func linkFile (base string) LinkStrings {
	
	return func (lang string)  (io.ReadCloser, bool) {
		f, err := os.Open(F.Join(wd, base, lang, stringsName))
		return f, err == nil
	}

}

// Language returns the current language of the mapped strings
func Language () string {
	return language
}

func (rac *dico) search (key string) value {
	rac = rac.next
	for rac != nil && rac.key < key {
		rac = rac.next
	}
	if rac == nil || rac.key != key {
		return nil
	}
	return rac.val
}

func (rac *dico) insert (key string, val value) {
	for rac.next != nil && rac.next.key < key {
		rac = rac.next
	}
	if rac.next == nil || rac.next.key != key {
		rac.next = &dico{next: rac.next, key: key, val: val}
	}
}

func (m *mapScan) scan () bool {
	const sepLen = len(string(sep))
	b := m.s.Scan()
	if !b || m.s.Text() != header {
		fmt.Fprintln(os.Stderr, "Wrong header:", m.s.Text())
		return false
	}
	line := 1
	for m.s.Scan() {
		line++
		s := m.s.Text()
		if s == "" {
			continue
		}
		n := strings.IndexRune(s, sep)
		if n < 0 {
			fmt.Fprintln(os.Stderr, "Line", line, "without separator")
			return false
		}
		p := n + sepLen
		for strings.IndexRune(s[p:], sep) == 0 {
			p += sepLen
		}
		key := s[:n]
		val := s[p:]
		M.Assert(key != "" && val != "", s, 100)
		m.rac.insert(key, val)
	}
	err := m.s.Err()
	M.Assert(err == nil, err, 101)
	return true
}

func initDico (base string) bool {
	lang := language
	if lang == "en" {
		lang = ""
	}
	link, ok := lStr[base]
	if !ok {
		link = linkFile(base)
	}
	rc, ok := link(lang)
	if !ok {
		rc, ok = link("")
	}
	if !ok {
		return false
	}
	defer rc.Close()
	m := &mapScan{s: bufio.NewScanner(rc), rac: &dico{next: nil}}
	b := m.scan()
	M.Assert(b, errors.New("util/stringmapping: error with base = " + base + " and language = " + lang), 100)
	mut.Lock()
	dicos.insert(base, m.rac)
	mut.Unlock()
	return true
}

// Map translates a key string 'key' into a mapped string. Strings of the form "#base:message" are translated if there is a corresponding 'base' resource string (or file) for this base. Otherwise, the "#base:" prefix is stripped away.
// As an example, "#system:Cancel" may be translated to "Cancel" in the USA, and to "Abbrechen" in Germany; or to "Cancel" if the resource file or the appropriate entry is missing.
// Additional input parameters can be spliced into the resulting string. These parameters 'p' are inserted where "^0", "^1", "^2", etc... occur in the resulting string (^0 for the first parameter, ^1 for the second, etc...). The parameters are not mapped, but merely substituted.
// Map allows to remove country- and language-specific strings from a program source text, while at the same time providing a default string in the program source text such that the program always works, even if string resources are missing. 
func Map (key string, p ...string) string {
	M.Assert(key[0] == '#', 20)
	pos := strings.Index(key, ":")
	M.Assert(pos > 1, 21)
	pos++
	M.Assert(pos < len(key), 22)
	k := key[pos:]
	key = key[1:pos-1]
	mut.RLock()
	g := dicos.search(key)
	mut.RUnlock()
	if g == nil {
		if !initDico(key) {
			return k
		}
		g = dicos.search(key)
	}
	g = g.(*dico).search(k)
	if g == nil {
		return k
	}
	k = g.(string)
	for i, pp := range p {
		k = strings.Replace(k, "^" + strconv.Itoa(i), pp, -1)
	}
	return k
}

// Reinit flushes the contents of the .str files from main memory and fixes the current language to lang (two characters).
func Reinit (lang string) {
	M.Assert(len(lang) == 2, 20)
	language = lang
	dicos = &dico{next: nil}
	f, err := os.Create(langPath); M.Assert(err == nil, err, 100)
	defer f.Close()
	fmt.Fprintln(f, language)
}

// SetLStr fixes the LinkStrings function 'ls' for the base 'base'. Optional.
// The default LinkStrings function for the base 'base' associates the language 'lang' to the string contained in the file "<resource directory>/base/lang/strings.txt".
// For <resource directory>, see the package "util/resources".
func SetLStr (base string, lS LinkStrings) {
	lStr[base] = lS
}

func init () {
	dir := F.Join(wd, strMappingDir)
	langPath = F.Join(dir, languageName)
	os.MkdirAll(dir, 0777)
	f, err := os.Open(langPath)
	if err == nil {
		defer f.Close()
		s := new(scanner.Scanner)
		s.Init(f)
		s.Error = func(s *scanner.Scanner, msg string) {panic(errors.New("File" + langPath + "incorrect"))}
		s.Scan()
		lang := s.TokenText()
		Reinit(lang)
	} else {
		Reinit("en")
	}
}
