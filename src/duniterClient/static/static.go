/* 
DuniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package static

// Package strMapping implements a mapping between keys and lengthier strings.

// This package must be imported for its side effects when the content of strMapping resource files is stored in "util/strMapping/static" (file "vars.go", variables "compiler" and "str")

import (
	
	SM	"util/strMapping"
		"io"
		"io/ioutil"
		"strings"

)

func linkString (lang string) (io.ReadCloser, bool) {
	switch lang {
	case "":
		return ioutil.NopCloser(strings.NewReader(strEn)), true
	case "fr":
		return ioutil.NopCloser(strings.NewReader(strFr)), true
	default:
		return nil, false
	}
	return nil, false
}

func init () {
	SM.SetLStr("duniterClient", linkString)
}
