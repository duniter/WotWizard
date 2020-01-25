/* 
WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package parameters

import (
	
	B	"duniter/blockchain"
	G	"duniter/gqlReceiver"
	J	"util/json"
	M	"util/misc"

)

const (
	
	parsName = "Parameters"

)

type (
	
	action struct {
		output string
	}

)

func (a *action) Name () string {
	return parsName
}

func (a *action) Activate () {
	f, err := M.InstantCreate(a.output); M.Assert(err == nil, err, 100)
	ok := J.FprintJsonOf(f, B.Pars()); M.Assert(ok, 101)
	M.InstantClose(f)
}

func do (output string, newAction chan<- B.Actioner, fields ...string) {
	newAction <- &action{output: output}
}

func init () {
	G.AddAction(parsName, do, G.Arguments{})
}
