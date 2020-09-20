/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

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
	if !ok && (included(wd0, d) || included(wd0, "/tmp")) {
		wd = d
		ok = correct(wd)
	}
	if !ok {
		wd = wd0
		err := os.Mkdir(rsrcDir, 0777); M.Assert(err == nil, err, 106)
	}
	return F.Join(wd, "rsrc")
}
