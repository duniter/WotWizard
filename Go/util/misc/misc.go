/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package misc

import (
	
	F	"path/filepath"
	SC	"strconv"
		"log"
		"os"
		"strings"
		"testing"

)

const (
	
	MinByte = MinUint8
	MinRune = MinInt32
	
	MaxInt8 = 0x7F
	MaxInt16 = 0x7FFF
	MaxInt32 = 0x7FFFFFFF
	MaxInt64 = 0x7FFFFFFFFFFFFFFF
	
	MaxUint8 = uint8(0xFF)
	MaxUint16 = uint16(0xFFFF)
	MaxUint32 = uint32(0xFFFFFFFF)
	MaxUint64 = uint64(0xFFFFFFFFFFFFFFFF)
	
	MinInt8 = -MaxInt8 - 1
	MinInt16 = -MaxInt16 - 1
	MinInt32 = -MaxInt32 - 1
	MinInt64 = -MaxInt64 - 1
	
	MinUint8 = uint8(0)
	MinUint16 = uint16(0)
	MinUint32 = uint32(0)
	MinUint64 = uint64(0)
	
	MaxByte = MaxUint8
	MaxRune = MaxInt32

)

var lg *log.Logger = log.New(os.Stderr, "", log.LstdFlags)

func SetLog (log *log.Logger) {
	lg = log
}

func Odd (n int) bool {
	return n & 1 == 1
}

func Min (a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max (a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min32 (a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func Max32 (a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func Min64 (a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Max64 (a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinF32 (a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func MaxF32 (a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func MinF64 (a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func MaxF64 (a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Abs (a int) int {
	if a < 0 {
		a = -a
	}
	return a
}

func Abs32 (a int32) int32 {
	if a < 0 {
		a = -a
	}
	return a
}

func Abs64 (a int64) int64 {
	if a < 0 {
		a = -a
	}
	return a
}

func AbsF32 (a float32) float32 {
	if a < 0 {
		a = -a
	}
	return a
}

func AbsF64 (a float64) float64 {
	if a < 0 {
		a = -a
	}
	return a
}

func Assert (cond bool, flag ... interface{}) {
	if !cond {
		lg.Output(2, "*** ERROR ***");
		if len(flag) == 0 {
			lg.Output(2, "Assert error")
			panic("Assert error")
		}
		for _, f := range flag {
			switch e := f.(type) {
			case error:
				lg.Output(2, e.Error())
			case string:
				lg.Output(2, e)
			case int:
				lg.Output(2, SC.Itoa(e))
			default:
				lg.Print(f)
			}
		}
		panic(flag[len(flag) - 1])
	}
}

func Want (cond bool, t *testing.T) {
	if !cond {
		t.Fail()
	}
}

func hidePath (shown string) string {
	dir, name := F.Split(shown)
	return dir + "." + name
}

func showPath (hidden string) string {
	dir, name := F.Split(hidden)
	Assert(len(name) > 1 && name[0] == '.', 20)
	return dir + strings.Replace(name, ".", "", 1)
}

func InstantCreate (name string) (*os.File, error) {
	return os.Create(hidePath(name))
}

func InstantClose (f *os.File) error {
	hidden := f.Name()
	err := f.Close()
	if err != nil {
		return err
	}
	return os.Rename(hidden, showPath(hidden))
}
