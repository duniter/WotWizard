/* 
duniterClient: WotWizard

Copyright (C) 2017-2020 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License  for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package run

import (
	
	_	"babel/static"
	_	"util/json/static"
	_	"util/graphQL/static"
	_	"duniterClient/static"
	
	W	"duniterClient/web"
	
	_	"duniterClient/certificationsPrint"
	_	"duniterClient/eventsPrint"
	_	"duniterClient/identitiesPrint"
	_	"duniterClient/identitySearch"
	_	"duniterClient/language"
	_	"duniterClient/membersPrint"
	_	"duniterClient/parametersPrint"
	_	"duniterClient/qualitiesPrint"
	_	"duniterClient/sentriesPrint"
	_	"duniterClient/wotWizardPrint"
	_	"duniterClient/wwViews"

	_	"duniterClient/tellLimitsPrint"
	
		"fmt"

)

const (
	
	version = "5.0.2"

)

func Start () {
	fmt.Println("WotWizard version", version, "\n")
	W.Start()
}
