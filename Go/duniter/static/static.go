/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

// The package graphQL is an implementation of the GraphQL query interface
// This package must be imported for its side effects when the GraphQL compiler and errors are stored in "util/graphQL/static" (file "compVar.go", variable "compiler" and file errorsVar.go, var strEn)
package static

import (
	
	F	"path/filepath"
	G	"util/graphQL"
		"io"
		"strings"

)

type (
	
	nopCloserS struct {
		io.ReadSeeker
	}
)

var (
	
	oldLinkDocs G.LinkGQL

)

func (nopCloserS) Close() error {
	return nil
}

func nopCloser(r io.ReadSeeker) G.ReadSeekCloser {
	return nopCloserS{r}
}

func linkStringDocs (path string) G.ReadSeekCloser {
	name := F.Base(path)
	switch name {
	case "TypeSystem.txt":
		return nopCloser(strings.NewReader(typeSystem))
	default:
		return oldLinkDocs(path)
	}
}

func init () {
	oldLinkDocs = G.SetLGQL(linkStringDocs)
}
