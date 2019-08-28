package resources

// resources.FindDir looks for a directory containing resources for the executable that started the current process.
// It looks 1) for a "rsrc" directory in the directory where the executable started, 2) for a "rsrc" directory in the current directory, 3) for a "rsrc" directory in the {$GOPATH} directory.
	
import (
	
	M	"util/misc"
	F	"path/filepath"
		"os"

)

func FindDir () string {
	
	const rsrcDir = "rsrc"
	
	correct := func (dir string) bool {
		f, err := os.Open(dir); M.Assert(err == nil, err, 100)
		defer f.Close()
		fi, err := f.Stat(); M.Assert(err == nil, err, 101)
		M.Assert(fi.IsDir(), 102)
		list, err := f.Readdir(0)
		for _, fi := range list {
			if fi.IsDir() && fi.Name() == rsrcDir {
				return true
			}
		}
		return false
	}
	
	included := func (dirIn, dirOut string) bool {
		for dirIn != "/" && dirIn != dirOut {
			dirIn = F.Dir(dirIn)
		}
		return dirIn == dirOut
	}

	wd, err := os.Executable(); M.Assert(err == nil, err, 103)
	wd, err = F.EvalSymlinks(wd); M.Assert(err == nil, err, 104)
	wd = F.Dir(wd)
	wd0 := wd
	ok := correct(wd)
	if !ok {
		wd, err = os.Getwd(); M.Assert(err == nil, err, 105)
		ok = correct(wd)
	}
	d := os.Getenv("GOPATH")
	if !ok && (included(wd0, d) || wd0 == "/tmp") {
		wd = d
		ok = correct(wd)
	}
	if !ok {
		wd = wd0
		err := os.Mkdir(rsrcDir, 0777); M.Assert(err == nil, err, 106)
	}
	return F.Join(wd, "rsrc")
}
