package basic

import (
	
	A	"util/avl"
	F	"path/filepath"
	M	"util/misc"
	R	"util/resources"
		"bufio"
		"bytes"
		"flag"
		"log"
		"os"
		"strings"
		"time"
		"unicode"

)

type (
	
	Comp = A.Comp

)

const (
	
	version = "4.0.0"
	
	duniBaseDef = "$HOME/.config/duniter/duniter_default/wotwizard-export.db"
	
	// Name of the file where the path to the Duniter database is written
	initName = "init.txt"
	logName = "log.txt"
	logOldName = "log1.txt"
	
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

)

func RsrcDir () string {
	return rsrcDir
}

func SwitchOff (path string) {
	err := os.Remove(path); M.Assert(err == nil, err, 100)
}

func Check (path string) bool {
	f, err := os.Open(path)
	if err == nil {
		f.Close()
	}
	return err == nil
}

func WaitFor (path string, present bool, delay time.Duration) {
	for Check(path) != present {
		time.Sleep(delay)
	}
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
}

// Extract the significant characters in s; only alphanumeric characters are significant, and their case of lowest rank is returned
func strip (s string) string {
	rs := bytes.Runes([]byte(s))
	buf := new(bytes.Buffer)
	for _, r := range rs {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		}
	}
	return string(buf.Bytes())
}

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
}

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
}

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
}

// Compare the characters of s1 and s2 and return whether the first one is a prefix of the second one or not
func Prefix (s1, s2 string) bool {
	return strings.HasPrefix(s2, s1)
}

func setLog () {
	err := os.Remove(logOldPath)
	M.Assert(err == nil || os.IsNotExist(err), err, 101)
	err = os.Rename(logPath, logOldPath)
	M.Assert(err == nil || os.IsNotExist(err), err, 102)
	f, err := os.Create(logPath)
	M.Assert(err == nil, 103)
	Lg = log.New(f, "", log.Ldate | log.Ltime | log.Lshortfile)
}

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
}

func init () {
	err := os.MkdirAll(rsrcDir, 0777); M.Assert(err == nil, err, 100)
	setLog()
	setDuniterPath()
}
